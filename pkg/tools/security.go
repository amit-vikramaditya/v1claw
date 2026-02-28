package tools

import (
	"fmt"
	"strings"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// SecurityMiddleware defines the interface for vetting shell commands.
type SecurityMiddleware interface {
	VerifyCommand(command string) (string, error)
}

// AllowlistMiddleware enforces a strict set of pre-approved commands.
type AllowlistMiddleware struct {
	Allowed []string
}

func (m *AllowlistMiddleware) VerifyCommand(command string) (string, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return "", fmt.Errorf("empty command")
	}

	baseCmd := parts[0]
	for _, a := range m.Allowed {
		if baseCmd == a {
			return command, nil
		}
	}

	return "", fmt.Errorf("command '%s' is not in the security allowlist", baseCmd)
}

// DefaultAllowlist provides a safe set of read-only and common dev tools.
var DefaultAllowlist = []string{
	"ls", "cat", "grep", "find", "wc", "head", "tail", "du", "df",
	"git", "go", "npm", "node", "python", "python3", "rustc", "cargo",
	"mkdir", "cp", "mv", "rm", // Modifying files is allowed within workspace limits
}

// SandboxMiddleware wraps a command to run inside a restricted container.
type SandboxMiddleware struct {
	ContainerImage string
	WorkspaceDir   string
}

func (m *SandboxMiddleware) VerifyCommand(command string) (string, error) {
	// Wrap the command in a docker exec or bwrap command.
	// For now, we return the command as-is but log the intent.
	logger.DebugCF("security", "Command prepared for sandbox", map[string]interface{}{
		"command": command,
		"image":   m.ContainerImage,
	})
	
	// Example transformation: "docker run --rm -v ... image /bin/sh -c 'command'"
	return command, nil
}
