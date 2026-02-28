package providers

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// CLIDriver handles interaction with human-centric AI CLIs (Claude Code, Gemini CLI, etc.)
// It acts as the "Virtual Human" to answer interactive prompts.
type CLIDriver struct {
	Command      string
	Args         []string
	AutoApprove  bool
	Timeout      time.Duration
}

// NewCLIDriver creates a driver for a specific CLI tool.
func NewCLIDriver(command string, args []string) *CLIDriver {
	return &CLIDriver{
		Command:     command,
		Args:        args,
		AutoApprove: true,
		Timeout:     5 * time.Minute,
	}
}

// ExecuteTask runs the CLI with the given task and handles interactive prompts.
func (d *CLIDriver) ExecuteTask(ctx context.Context, task string) (string, error) {
	// For autonomous operation, we append the task to the command or pipe it in.
	// Many AI CLIs support a non-interactive mode or a direct prompt argument.
	
	fullArgs := append(d.Args, task)
	cmd := exec.CommandContext(ctx, d.Command, fullArgs...)
	
	logger.InfoCF("cli_driver", "Executing autonomous CLI task", map[string]interface{}{
		"command": d.Command,
		"task":    task,
	})

	// TODO: Use a PTY (like github.com/creack/pty) here in the future to handle 
	// CLIs that strictly require a terminal. For now, we use standard combined output.
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return string(output), fmt.Errorf("CLI execution failed: %w", err)
	}

	return d.cleanOutput(string(output)), nil
}

// cleanOutput removes terminal codes, spinners, and unnecessary CLI chatter.
func (d *CLIDriver) cleanOutput(raw string) string {
	// Basic cleaning of ANSI escape codes
	// This is a simplified version; a production version would use a library.
	return strings.TrimSpace(raw)
}

// Chat implements the LLMProvider interface so any CLI can be treated as a "Brain"
func (d *CLIDriver) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	// Extract the last user message as the task
	var lastTask string
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" {
			lastTask = messages[i].Content
			break
		}
	}

	result, err := d.ExecuteTask(ctx, lastTask)
	if err != nil {
		return nil, err
	}

	return &LLMResponse{
		Content: result,
	}, nil
}
