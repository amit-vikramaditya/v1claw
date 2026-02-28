package config

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// KnownAIWorker defines the signature of a supported AI CLI.
type KnownAIWorker struct {
	Name             string
	Command          string
	PromptFlag       string
	AutoApproveFlags []string
	VersionFlag      string
}

// KnownWorkers is the master registry of AI CLIs Vikram knows how to "drive".
var KnownWorkers = []KnownAIWorker{
	{Name: "claude", Command: "claude", PromptFlag: "-p", AutoApproveFlags: []string{"--allowedTools", "Read,Edit,Bash"}, VersionFlag: "--version"},
	{Name: "gemini", Command: "gemini", PromptFlag: "--prompt", AutoApproveFlags: []string{"--approval-mode", "yolo"}, VersionFlag: "--version"},
	{Name: "codex", Command: "codex", PromptFlag: "exec", AutoApproveFlags: []string{"--dangerously-bypass-approvals-and-sandbox"}, VersionFlag: "--version"},
	{Name: "copilot", Command: "copilot", PromptFlag: "--prompt", AutoApproveFlags: []string{"--allow-all-tools", "--allow-all-paths"}, VersionFlag: "--version"},
	{Name: "vibe", Command: "vibe", PromptFlag: "-p", AutoApproveFlags: []string{}, VersionFlag: "--version"},
	{Name: "aider", Command: "aider", PromptFlag: "--message", AutoApproveFlags: []string{"--yes-always", "--no-auto-commits"}, VersionFlag: "--version"},
	{Name: "open-interpreter", Command: "interpreter", PromptFlag: "", AutoApproveFlags: []string{"-y"}, VersionFlag: "--version"},
}

// DiscoveryResult holds the details of a detected CLI.
type DiscoveryResult struct {
	KnownAIWorker
	Path    string
	Version string
	Status  string // "Available", "Error", "Needs Setup"
}

// CLIScanner handles searching for AI tools in the environment.
type CLIScanner struct{}

// Scan searches the PATH for known AI CLIs.
func (s *CLIScanner) Scan(ctx context.Context) []DiscoveryResult {
	results := []DiscoveryResult{}

	for _, kw := range KnownWorkers {
		path, err := exec.LookPath(kw.Command)
		if err != nil {
			// Not in PATH, skip
			continue
		}

		res := DiscoveryResult{
			KnownAIWorker: kw,
			Path:          path,
			Status:        "Available",
		}

		// Try to get version
		vCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		cmd := exec.CommandContext(vCtx, kw.Command, kw.VersionFlag)
		out, err := cmd.Output()
		cancel()
		if err == nil {
			res.Version = strings.TrimSpace(string(out))
		} else {
			res.Status = "Needs Setup (Found but error on --version)"
		}

		results = append(results, res)
	}

	return results
}

// RegisterSelectedWorkers saves the user-selected workers to workers.json.
func RegisterSelectedWorkers(workspace string, selected []DiscoveryResult) error {
	configs := []CLIWorkerConfig{}
	for _, s := range selected {
		configs = append(configs, CLIWorkerConfig{
			Name:             s.Name,
			Command:          s.Command,
			PromptFlag:       s.PromptFlag,
			AutoApproveFlags: s.AutoApproveFlags,
			Enabled:          true,
		})
	}

	// Re-use existing persistence logic
	return SaveCLIWorkers(workspace, configs)
}

// SaveCLIWorkers persists the workers configuration.
func SaveCLIWorkers(workspace string, workers []CLIWorkerConfig) error {
	data, err := json.MarshalIndent(workers, "", "  ")
	if err != nil {
		return err
	}

	configPath := filepath.Join(workspace, "workers.json")
	return os.WriteFile(configPath, data, 0644)
}
