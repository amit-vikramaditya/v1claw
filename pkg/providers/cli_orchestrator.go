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
	StaticFlags      []string // Safe control flags that should remain enabled in sandboxed mode
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
			StaticFlags:      nil,
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
		StaticFlags:      []string{"--no-auto-commits"},
		AutoApproveFlags: []string{"--yes-always"},
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

// SandboxSafeProfile removes privileged automation flags that bypass approval
// or expand filesystem/tool access. Safe control flags remain intact.
func SandboxSafeProfile(profile CLIInteractionProfile) CLIInteractionProfile {
	safe := profile
	safe.AutoApproveFlags = nil
	return safe
}

// GetDefaultModel returns the CLI worker's name as its model identifier.
func (w *AutonomousCLIWorker) GetDefaultModel() string {
	return w.Profile.Name
}

func (w *AutonomousCLIWorker) Execute(ctx context.Context, task string) (string, error) {
	args := []string{}

	// Guard against argument injection: if the task string begins with a dash (-) it
	// could be mistaken by the target CLI as a flag. We sanitise by prepending a
	// "--" end-of-options marker when the task is positional, and by limiting the
	// task length to prevent excessively long argument strings.
	safeTask := task
	if len(safeTask) > 32768 {
		safeTask = safeTask[:32768]
	}

	promptFlag := w.Profile.PromptFlag
	hasPromptFlag := promptFlag != ""
	promptFlagIsOption := strings.HasPrefix(promptFlag, "-")

	// 1. Add a subcommand first if the profile uses one (e.g. "codex exec ...").
	if hasPromptFlag && !promptFlagIsOption {
		args = append(args, promptFlag)
	}

	// 2. Add control flags before the task for consistent CLI parsing.
	args = append(args, w.Profile.StaticFlags...)
	args = append(args, w.Profile.AutoApproveFlags...)
	if w.Profile.OutputFormatFlag != "" {
		parts := strings.Fields(w.Profile.OutputFormatFlag)
		args = append(args, parts...)
	}

	// 3. Add the prompt/task portion last.
	if hasPromptFlag && promptFlagIsOption {
		args = append(args, promptFlag, safeTask)
	} else {
		if strings.HasPrefix(safeTask, "-") {
			args = append(args, "--")
		}
		args = append(args, safeTask)
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
