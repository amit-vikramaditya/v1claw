package tools

import (
	"context"
	"fmt"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
	"github.com/amit-vikramaditya/v1claw/pkg/providers"
)

// DelegateTaskTool allows the manager to assign a task to a specific specialized worker.
type DelegateTaskTool struct {
	manager *SubagentManager
	// cliProviders is now passed via ToolContext.Bus.GetHandler(workerType).Chat
}

// NewDelegateTaskTool creates a new delegation tool.
func NewDelegateTaskTool(manager *SubagentManager) *DelegateTaskTool {
	return &DelegateTaskTool{
		manager: manager,
	}
}

func (t *DelegateTaskTool) Name() string {
	return "delegate_task"
}

func (t *DelegateTaskTool) Description() string {
	return "Delegate a complex coding, refactoring, or search task to a specialized subordinate AI worker (e.g., 'codex', 'claude_code'). The worker will run autonomously and report back."
}

func (t *DelegateTaskTool) Parameters() map[string]interface{} {
	// Dynamically build workers list from available CLI providers (from SubagentManager)
	workerList := []string{"standard"}
	if t.manager != nil {
		for k := range t.manager.CLIProviders() {
			workerList = append(workerList, k)
		}
	}

	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"worker_type": map[string]interface{}{
				"type":        "string",
				"enum":        workerList,
				"description": "The type of subordinate to use. 'standard' for general tasks, or a configured CLI worker name.",
			},
			"task": map[string]interface{}{
				"type":        "string",
				"description": "The detailed instruction for the subordinate.",
			},
		},
		"required": []string{"worker_type", "task"},
	}
}

func (t *DelegateTaskTool) Execute(ctx context.Context, tc ToolContext, args map[string]interface{}) *ToolResult {
	workerType, ok := args["worker_type"].(string)
	if !ok {
		return ErrorResult("worker_type is required")
	}
	task, ok := args["task"].(string)
	if !ok {
		return ErrorResult("task is required")
	}

	if t.manager == nil {
		return ErrorResult("subagent manager is not configured")
	}

	// If it's a standard task, just use the existing spawn logic
	if workerType == "standard" {
		// Pass the original ToolContext to Spawn
		msg, err := t.manager.Spawn(ctx, task, "Standard_Worker_Task", tc)
		if err != nil {
			return ErrorResult(err.Error()).WithError(err)
		}
		return AsyncResult(msg)
	}

	// For specialized CLI workers, use the CLIProviders map from SubagentManager
	provider, exists := t.manager.CLIProviders()[workerType]
	if !exists {
		return ErrorResult(fmt.Sprintf("Worker type '%s' is not available.", workerType))
	}

	// Spin up a specialized worker in the background
	go func() {
		messages := []providers.Message{
			{Role: "user", Content: task},
		}

		// The CLI provider will handle execution. We wait for Chat to return.
		resp, err := provider.Chat(context.Background(), messages, nil, "default", map[string]interface{}{}) // context.Background should be ctx, but CLIProviders does not accept it.

		var result string
		if err != nil {
			result = fmt.Sprintf("Worker '%s' failed: %v", workerType, err)
		} else {
			result = resp.Content
		}

		// Announce completion to the manager's bus
		if tc.Bus != nil {
			announceContent := fmt.Sprintf("Delegated task to '%s' completed.\n\nResult:\n%s", workerType, result)
			tc.Bus.PublishInbound(bus.InboundMessage{
				Channel:    "system",
				SenderID:   fmt.Sprintf("worker:%s", workerType),
				ChatID:     fmt.Sprintf("%s:%s", tc.Channel, tc.ChatID),
				Content:    announceContent,
				SessionKey: tc.SessionKey,
			})
		}
	}()

	return AsyncResult(fmt.Sprintf("Delegated task to worker '%s'. Will report back asynchronously.", workerType))
}
