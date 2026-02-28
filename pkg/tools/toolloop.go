package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
	"github.com/amit-vikramaditya/v1claw/pkg/logger"
	"github.com/amit-vikramaditya/v1claw/pkg/providers"
	"github.com/amit-vikramaditya/v1claw/pkg/utils"
)

// ToolLoopConfig configures the tool execution loop.
type ToolLoopConfig struct {
	Provider      providers.LLMProvider
	Model         string
	Tools         *ToolRegistry
	MaxIterations int
	LLMOptions    map[string]any
}

// ToolLoopResult contains the result of running the tool loop.
type ToolLoopResult struct {
	Content    string
	Iterations int
}

// RunToolLoop executes the LLM + tool call iteration loop.
// This is the core agent logic that can be reused by both main agent and subagents.
func RunToolLoop(ctx context.Context, config ToolLoopConfig, messages []providers.Message, tc ToolContext) (*ToolLoopResult, error) { // Updated to accept ToolContext
	iteration := 0
	var finalContent string

	for iteration < config.MaxIterations {
		iteration++

		logger.DebugCF("toolloop", "LLM iteration",
			map[string]any{
				"iteration": iteration,
				"max":       config.MaxIterations,
			})

		// 1. Build tool definitions
		var providerToolDefs []providers.ToolDefinition
		if config.Tools != nil {
			providerToolDefs = config.Tools.ToProviderDefs()
		}

		// 2. Set default LLM options
		llmOpts := config.LLMOptions
		if llmOpts == nil {
			llmOpts = map[string]any{
				"max_tokens":  4096,
				"temperature": 0.7,
			}
		}

		// 3. Call LLM
		response, err := config.Provider.Chat(ctx, messages, providerToolDefs, config.Model, llmOpts)
		if err != nil {
			logger.ErrorCF("toolloop", "LLM call failed",
				map[string]any{
					"iteration": iteration,
					"error":     err.Error(),
				})
			return nil, fmt.Errorf("LLM call failed: %w", err)
		}

		// 4. If no tool calls, we're done
		if len(response.ToolCalls) == 0 {
			finalContent = response.Content
			logger.InfoCF("toolloop", "LLM response without tool calls (direct answer)",
				map[string]any{
					"iteration":     iteration,
					"content_chars": len(finalContent),
				})
			break
		}

		// 5. Log tool calls
		toolNames := make([]string, 0, len(response.ToolCalls))
		for _, tc := range response.ToolCalls {
			toolNames = append(toolNames, tc.Name)
		}
		logger.InfoCF("toolloop", "LLM requested tool calls",
			map[string]any{
				"tools":     toolNames,
				"count":     len(response.ToolCalls),
				"iteration": iteration,
			})

		// 6. Build assistant message with tool calls
		assistantMsg := providers.Message{
			Role:    "assistant",
			Content: response.Content,
		}
		for _, tc := range response.ToolCalls {
			argumentsJSON, _ := json.Marshal(tc.Arguments)
			assistantMsg.ToolCalls = append(assistantMsg.ToolCalls, providers.ToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: &providers.FunctionCall{
					Name:      tc.Name,
					Arguments: string(argumentsJSON),
				},
			})
		}
		messages = append(messages, assistantMsg)

		// 7. Execute tool calls
		for _, toolCall := range response.ToolCalls { // Renamed 'tc' to 'toolCall'
			argsJSON, _ := json.Marshal(toolCall.Arguments)
			argsPreview := utils.Truncate(string(argsJSON), 200)
			logger.InfoCF("toolloop", fmt.Sprintf("Tool call: %s(%s)", toolCall.Name, argsPreview),
				map[string]any{
					"tool":      toolCall.Name,
					"iteration": iteration,
				})

			// Execute tool with ToolContext.
			// Create a new ToolContext for this specific tool execution from the parent ToolContext.
			toolExecutionTC := ToolContext{
				Channel:    tc.Channel,
				ChatID:     tc.ChatID,
				SessionKey: tc.SessionKey,
				SenderID:   tc.SenderID,
				Async:      tc.Async, // Pass the async callback if available
				Bus:        tc.Bus,
			}

			var toolResult *ToolResult
			if config.Tools != nil {
				toolResult = config.Tools.ExecuteWithContext(ctx, toolCall.Name, toolCall.Arguments, toolExecutionTC)
			} else {
				toolResult = ErrorResult("No tools available")
			}

			// Determine content for LLM
			contentForLLM := toolResult.ForLLM
			if contentForLLM == "" && toolResult.Err != nil {
				contentForLLM = toolResult.Err.Error()
			}

			// Add tool result message
			toolResultMsg := providers.Message{
				Role:       "tool",
				Content:    contentForLLM,
				ToolCallID: toolCall.ID,
			}
			messages = append(messages, toolResultMsg)
		}
	}

	return &ToolLoopResult{
		Content:    finalContent,
		Iterations: iteration,
	}, nil
}

// SubagentManager manages the lifecycle and execution of subagents.
// It also holds a reference to the main LLM provider for subagent instantiation.
type SubagentManager struct {
	provider     providers.LLMProvider
	model        string
	workspace    string
	bus          *bus.MessageBus
	tools        *ToolRegistry
	mu           sync.RWMutex
	cliProviders map[string]providers.LLMProvider // Registered CLI workers
}

// NewSubagentManager creates a new SubagentManager.
func NewSubagentManager(provider providers.LLMProvider, model, workspace string, msgBus *bus.MessageBus, cliProviders map[string]providers.LLMProvider) *SubagentManager {
	return &SubagentManager{
		provider:     provider,
		model:        model,
		workspace:    workspace,
		bus:          msgBus,
		cliProviders: cliProviders,
	}
}

// SetTools sets the tool registry for subagents.
func (sm *SubagentManager) SetTools(tools *ToolRegistry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.tools = tools
}

// Spawn starts a subagent in a separate goroutine and returns a message indicating its creation.
// The subagent will run for a limited number of iterations and report its final answer via the message bus.
func (sm *SubagentManager) Spawn(ctx context.Context, task, label string, tc ToolContext) (string, error) { // Updated to accept ToolContext
	sm.mu.RLock()
	subagentTools := sm.tools // Get a reference to the tools
	sm.mu.RUnlock()

	if subagentTools == nil {
		return "", fmt.Errorf("subagent tools not set")
	}

	logSource := fmt.Sprintf("subagent:%s", label)

	// Create a new context builder for the subagent.
	// We do NOT use the main agent's context builder here to prevent context pollution.
	subagentContextBuilder := &ContextBuilder{
		workspace: sm.workspace,
		// No need to set tools registry here as the tools are passed directly via ToolLoopConfig
	}

	// Build initial messages for the subagent.
	// The subagent gets its own fresh prompt, independent of the main agent's history.
	subagentMessages := subagentContextBuilder.BuildMessages(
		nil,  // No history for subagent init
		"",   // No summary for subagent init
		task, // The task is the primary user message for the subagent
		subagentTools,
		tc.Channel, // Use ToolContext for channel
		tc.ChatID,  // Use ToolContext for chatID
	)

	go func() {
		// Recover from panics in subagent goroutine.
		defer func() {
			if r := recover(); r != nil {
				logger.ErrorCF(logSource, "Subagent panicked", map[string]interface{}{"error": fmt.Sprintf("%v", r)})
				if tc.Async != nil { // Use tc.Async
					tc.Async(context.Background(), ErrorResult(fmt.Sprintf("Subagent panicked: %v", r)))
				}
			}
		}()

		logger.InfoCF(logSource, "Subagent spawned", map[string]interface{}{"task": task, "label": label})

		// Create a new ToolLoopConfig for the subagent.
		// The subagent gets its own tool registry but shares the main LLM provider.
		toolLoopConfig := ToolLoopConfig{
			Provider:      sm.provider,
			Model:         sm.model,
			Tools:         subagentTools,
			MaxIterations: 10,  // Subagents have a limited number of iterations
			LLMOptions:    nil, // Use default LLM options
		}

		// Pass the ToolContext directly to RunToolLoop
		loopResult, err := RunToolLoop(ctx, toolLoopConfig, subagentMessages, tc)
		if err != nil {
			logger.ErrorCF(logSource, "Subagent failed", map[string]interface{}{"error": err.Error()})
			if tc.Async != nil {
				tc.Async(context.Background(), ErrorResult(fmt.Sprintf("Subagent failed: %v", err)))
			}
			return
		}

		logger.InfoCF(logSource, "Subagent completed", map[string]interface{}{"iterations": loopResult.Iterations})

		if tc.Async != nil {
			tc.Async(context.Background(), NewToolResult(loopResult.Content))
		} else {
			logger.InfoCF(logSource, "Subagent completed without explicit callback", map[string]interface{}{"result": loopResult.Content})
		}
	}()

	return fmt.Sprintf("Subagent '%s' spawned for task: %s", label, task), nil
}

// RunToolLoop executes the LLM + tool call iteration loop for synchronous subagents.
func (sm *SubagentManager) RunToolLoop(ctx context.Context, task, label, channel, chatID, sessionKey string) (*ToolLoopResult, error) {
	sm.mu.RLock()
	subagentTools := sm.tools // Get a reference to the tools
	sm.mu.RUnlock()

	if subagentTools == nil {
		return nil, fmt.Errorf("subagent tools not set")
	}

	// Create a new context builder for the subagent.
	subagentContextBuilder := &ContextBuilder{
		workspace: sm.workspace,
	}

	// Build initial messages for the subagent.
	subagentMessages := subagentContextBuilder.BuildMessages(
		nil,  // No history for subagent init
		"",   // No summary for subagent init
		task, // The task is the primary user message for the subagent
		subagentTools,
		channel,
		chatID,
	)

	toolLoopConfig := ToolLoopConfig{
		Provider:      sm.provider,
		Model:         sm.model,
		Tools:         subagentTools,
		MaxIterations: 10,  // Subagents have a limited number of iterations
		LLMOptions:    nil, // Use default LLM options
	}

	// Create a ToolContext for the synchronous subagent loop
	tc := ToolContext{
		Channel:    channel,
		ChatID:     chatID,
		SessionKey: sessionKey,
		// No Async callback needed as this is synchronous
		Bus: sm.bus,
	}

	loopResult, err := RunToolLoop(ctx, toolLoopConfig, subagentMessages, tc) // Pass ToolContext
	if err != nil {
		return nil, fmt.Errorf("subagent failed: %w", err)
	}
	return loopResult, nil
}

func (sm *SubagentManager) CLIProviders() map[string]providers.LLMProvider {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.cliProviders
}

// ContextBuilder (placeholder for the actual implementation in agent/context.go)
// It's defined here because SubagentManager (in tools) needs it.
type ContextBuilder struct {
	workspace string
	// Removed logger *logger.Logger to avoid undefined error with placeholder
	// Add other necessary fields as per agent/context.go
}

// BuildMessages (placeholder)
func (cb *ContextBuilder) BuildMessages(history []providers.Message, summary, userMessage string, tools *ToolRegistry, channel, chatID string) []providers.Message {
	// Simplified for compilation within tools package
	messages := []providers.Message{}
	if summary != "" {
		messages = append(messages, providers.Message{Role: "system", Content: "Summary: " + summary})
	}
	if userMessage != "" {
		messages = append(messages, providers.Message{Role: "user", Content: userMessage})
	}
	return messages
}
