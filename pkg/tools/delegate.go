package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/amit-vikramaditya/v1claw/pkg/providers"
)

// DelegateTaskTool allows the brain to assign work to the most appropriate
// specialized CLI worker. The brain chooses the worker autonomously based on
// the task — it should never ask the user which worker to use.
type DelegateTaskTool struct {
	manager *SubagentManager
}

// workerSpecialties describes what each known CLI worker is best at.
// The brain uses this context to pick the right tool without user guidance.
var workerSpecialties = map[string]string{
	"standard":         "general tasks handled by the primary brain's own subagent",
	"gemini":           "research, analysis, multi-modal reasoning, brainstorming",
	"codex":            "code writing, implementation, refactoring, file creation",
	"copilot":          "code completion, review, debugging, GitHub-integrated tasks",
	"vibe":             "full-stack web app generation, UI scaffolding, agentically creating projects",
	"claude":           "long-context reasoning, writing, nuanced analysis",
	"aider":            "git-aware code editing, applying patches across a large codebase",
	"open-interpreter": "running code, executing shell commands, system automation",
}

// NewDelegateTaskTool creates a new delegation tool.
func NewDelegateTaskTool(manager *SubagentManager) *DelegateTaskTool {
	return &DelegateTaskTool{manager: manager}
}

func (t *DelegateTaskTool) Name() string {
	return "delegate_task"
}

func (t *DelegateTaskTool) Description() string {
	return "Delegate a task to the most appropriate specialized worker. " +
		"YOU choose which worker fits the task best — do not ask the user. " +
		"For multi-component projects, call this tool once per component using the best available worker for each. " +
		"The call blocks until the worker finishes and returns its result for you to review before continuing."
}

func (t *DelegateTaskTool) Parameters() map[string]interface{} {
	// Build worker list and per-worker specialization hints for the brain.
	workerList := []string{"standard"}
	var specialtyLines []string
	specialtyLines = append(specialtyLines, fmt.Sprintf("  standard — %s", workerSpecialties["standard"]))

	if t.manager != nil {
		for name := range t.manager.CLIProviders() {
			workerList = append(workerList, name)
			if spec, ok := workerSpecialties[name]; ok {
				specialtyLines = append(specialtyLines, fmt.Sprintf("  %s — %s", name, spec))
			} else {
				specialtyLines = append(specialtyLines, fmt.Sprintf("  %s — specialized CLI worker", name))
			}
		}
	}

	workerDesc := "Which worker to assign this task to. Choose the best fit:\n" +
		strings.Join(specialtyLines, "\n")

	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"worker_type": map[string]interface{}{
				"type":        "string",
				"enum":        workerList,
				"description": workerDesc,
			},
			"task": map[string]interface{}{
				"type":        "string",
				"description": "Complete, self-contained instruction for the worker. Include all context it needs — it has no memory of prior conversation.",
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

	messages := []providers.Message{
		{Role: "user", Content: task},
	}

	resp, err := provider.Chat(ctx, messages, nil, "default", map[string]interface{}{})
	outcome := buildWorkerOutcome(workerType, resp, err)
	return NewToolResult(outcome)
}

// buildWorkerOutcome formats the CLI worker result into a compact string for the LLM.
func buildWorkerOutcome(workerType string, resp *providers.LLMResponse, err error) string {
	if err != nil {
		return fmt.Sprintf("Worker '%s': FAILED\nError: %s", workerType, err.Error())
	}
	summary := resp.Content
	if len(summary) > 500 {
		summary = summary[:497] + "..."
	}
	return fmt.Sprintf("Worker '%s': SUCCESS\n%s", workerType, summary)
}
