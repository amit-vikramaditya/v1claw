package config

import (
	"os/exec"
)

// DiscoveredTool represents an AI CLI tool found on the local system.
type DiscoveredTool struct {
	ID          string // Internal identifier (e.g., "copilot")
	DisplayName string // User-facing name (e.g., "GitHub Copilot CLI")
	Path        string // Absolute path to the executable
	Description string // Short description of what the tool is
}

// DiscoverLocalCLIs scans the local system's PATH for known AI CLI tools.
func DiscoverLocalCLIs() []DiscoveredTool {
	tools := []DiscoveredTool{
		{ID: "copilot", DisplayName: "GitHub Copilot CLI", Description: "Powered by OpenAI via GitHub"},
		{ID: "codex", DisplayName: "Codex CLI", Description: "Command-line AI Assistant"},
		{ID: "vibe", DisplayName: "Vibe CLI", Description: "Local Developer AI"},
		{ID: "gemini", DisplayName: "Google Gemini CLI", Description: "Official Google Gemini Tool"},
		{ID: "ollama", DisplayName: "Ollama (Local)", Description: "Runs local open-source models"},
	}

	var discovered []DiscoveredTool

	for _, tool := range tools {
		if path, err := exec.LookPath(tool.ID); err == nil {
			tool.Path = path
			discovered = append(discovered, tool)
		}
	}

	return discovered
}
