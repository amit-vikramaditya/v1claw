package providers

import (
	"context"
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

func (w *AutonomousCLIWorker) Execute(ctx context.Context, task string) (string, error) {
	args := []string{}

	// 1. Add subcommand or prompt flag
	if w.Profile.PromptFlag != "" {
		args = append(args, w.Profile.PromptFlag)
	}

	// 2. Add the task itself
	args = append(args, task)

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
	output, err := cmd.CombinedOutput()

	// We return output even on error as it often contains the AI's explanation of the failure
	return string(output), err
}
