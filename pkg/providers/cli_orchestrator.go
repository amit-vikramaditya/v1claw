package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// CLIInteractionProfile defines how to drive a specific AI CLI autonomously.
type CLIInteractionProfile struct {
	Name             string
	Command          string
	PromptFlag       string   // e.g., "-p" or "--prompt"
	AutoApproveFlags []string // e.g., ["--yolo"] or ["--allow-all-tools"]
	OutputFormatFlag string   // e.g., "--output-format json"
	RequiresTaskArg  bool     // Whether the task is a positional arg or flag-based
}

// LoadProfilesFromConfig converts user config workers into interaction profiles.
func LoadProfilesFromConfig(workers []config.CLIWorkerConfig) map[string]CLIInteractionProfile {
	profiles := make(map[string]CLIInteractionProfile)

	// Add built-ins first
	for k, v := range BuiltinProfiles {
		profiles[k] = v
	}

	// Overwrite or add from config
	for _, w := range workers {
		if !w.Enabled {
			continue
		}
		profiles[w.Name] = CLIInteractionProfile{
			Name:             w.Name,
			Command:          w.Command,
			PromptFlag:       w.PromptFlag,
			AutoApproveFlags: w.AutoApproveFlags,
			OutputFormatFlag: w.OutputFormatFlag,
			RequiresTaskArg:  true, // Default to true for most modern CLIs
		}
	}
	return profiles
}

// BuiltinProfiles contains proven automation patterns for popular AI CLIs.
var BuiltinProfiles = map[string]CLIInteractionProfile{
	"gemini": {
		Name:             "gemini",
		Command:          "gemini",
		PromptFlag:       "--prompt",
		AutoApproveFlags: []string{"--approval-mode", "yolo"},
		OutputFormatFlag: "--output-format json",
	},
	"codex": {
		Name:             "codex",
		Command:          "codex",
		PromptFlag:       "exec", // Codex uses a subcommand for non-interactive
		AutoApproveFlags: []string{"--dangerously-bypass-approvals-and-sandbox"},
		RequiresTaskArg:  true,
	},
	"copilot": {
		Name:             "copilot",
		Command:          "copilot",
		PromptFlag:       "--prompt",
		AutoApproveFlags: []string{"--allow-all-tools", "--allow-all-paths"},
		OutputFormatFlag: "--silent",
	},
	"vibe": {
		Name:             "vibe",
		Command:          "vibe",
		PromptFlag:       "-p",
		AutoApproveFlags: []string{}, // -p mode in vibe implies auto-approve
		OutputFormatFlag: "--output json",
	},
	"claude": {
		Name:             "claude",
		Command:          "claude",
		PromptFlag:       "-p",
		AutoApproveFlags: []string{"--allowedTools", "Read,Edit,Bash"},
	},
	"aider": {
		Name:             "aider",
		Command:          "aider",
		PromptFlag:       "--message",
		AutoApproveFlags: []string{"--yes-always", "--no-auto-commits"},
	},
	"open-interpreter": {
		Name:             "open-interpreter",
		Command:          "interpreter",
		PromptFlag:       "", // Positioned as last arg usually
		AutoApproveFlags: []string{"-y"},
	},
}

// AutonomousCLIWorker implements a "Virtual Human" wrapper for AI CLIs.
type AutonomousCLIWorker struct {
	Profile CLIInteractionProfile
}

// Chat implements the LLMProvider interface so AutonomousCLIWorker can be used as an agent brain.
// It extracts the last user message as the task and runs the CLI autonomously.
func (w *AutonomousCLIWorker) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	var lastTask string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastTask = messages[i].Content
			break
		}
	}
	if lastTask == "" {
		return nil, fmt.Errorf("no user message found for CLI worker %q", w.Profile.Name)
	}
	result, err := w.Execute(ctx, lastTask)
	if err != nil {
		return nil, err
	}
	return &LLMResponse{
		Content:      result,
		FinishReason: "stop",
	}, nil
}

// GetDefaultModel returns the CLI worker's name as its model identifier.
func (w *AutonomousCLIWorker) GetDefaultModel() string {
	return w.Profile.Name
}

func (w *AutonomousCLIWorker) Execute(ctx context.Context, task string) (string, error) {
	args := []string{}

	// 1. Add subcommand or prompt flag
	if w.Profile.PromptFlag != "" {
		args = append(args, w.Profile.PromptFlag)
	}

	// 2. Add the task itself.
	// Guard against argument injection: if the task string begins with a dash (-) it
	// could be mistaken by the target CLI as a flag.  We sanitise by prepending a
	// "--" end-of-options marker when the prompt flag is empty (positional mode),
	// and by limiting the task length to prevent excessively long argument strings.
	safeTask := task
	if len(safeTask) > 32768 {
		safeTask = safeTask[:32768]
	}
	if w.Profile.PromptFlag == "" && strings.HasPrefix(safeTask, "-") {
		// Insert end-of-options marker before positional task argument.
		args = append(args, "--")
	}
	args = append(args, safeTask)

	// 3. Add auto-approval flags to bypass "Human-in-Loop" prompts
	args = append(args, w.Profile.AutoApproveFlags...)

	// 4. Set output format for easier parsing if supported
	if w.Profile.OutputFormatFlag != "" {
		parts := strings.Fields(w.Profile.OutputFormatFlag)
		args = append(args, parts...)
	}

	logger.InfoCF("orchestrator", "Delegating task to autonomous worker", map[string]interface{}{
		"worker":  w.Profile.Name,
		"command": w.Profile.Command,
		"args":    args,
	})

	cmd := exec.CommandContext(ctx, w.Profile.Command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	out := strings.TrimSpace(stdout.String())

	if runErr != nil {
		// Include stderr in the error so callers know what went wrong
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = runErr.Error()
		}
		// Still return whatever stdout we got — it may contain a partial/explanation
		if out != "" {
			return out, fmt.Errorf("%s: %s", w.Profile.Name, errMsg)
		}
		return errMsg, fmt.Errorf("%s: %s", w.Profile.Name, errMsg)
	}

	// For CLIs that return JSON (e.g. gemini --output-format json), extract the
	// "response" field so the agent brain receives plain text, not a stats blob.
	if strings.HasPrefix(out, "{") {
		var envelope struct {
			Response string `json:"response"`
		}
		if err := json.Unmarshal([]byte(out), &envelope); err == nil && envelope.Response != "" {
			logger.InfoCF("orchestrator", "Extracted response from CLI JSON output", map[string]interface{}{
				"worker": w.Profile.Name,
			})
			return envelope.Response, nil
		}
	}

	return out, nil
}
