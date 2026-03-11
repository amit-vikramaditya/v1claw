package providers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGitHubCopilotProvider_StdioModeUsesCLIWorker(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "copilot")
	script := "#!/bin/sh\necho 'copilot cli response'\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("write mock copilot script: %v", err)
	}

	provider, err := NewGitHubCopilotProvider(scriptPath, "stdio", "gpt-4.1")
	if err != nil {
		t.Fatalf("NewGitHubCopilotProvider returned error: %v", err)
	}

	resp, err := provider.Chat(context.Background(), []Message{
		{Role: "user", Content: "say hello"},
	}, nil, "", nil)
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if resp.Content != "copilot cli response" {
		t.Fatalf("unexpected content: %q", resp.Content)
	}
	if provider.GetDefaultModel() != "gpt-4.1" {
		t.Fatalf("unexpected default model: %q", provider.GetDefaultModel())
	}
}

func TestGitHubCopilotProvider_StdioSandboxStripsDangerousFlags(t *testing.T) {
	provider, err := NewGitHubCopilotProviderWithSandbox("copilot", "stdio", "gpt-4.1", true)
	if err != nil {
		t.Fatalf("NewGitHubCopilotProviderWithSandbox returned error: %v", err)
	}

	if provider.cliWorker == nil {
		t.Fatal("expected stdio mode to use a CLI worker")
	}
	if len(provider.cliWorker.Profile.AutoApproveFlags) != 0 {
		t.Fatalf("expected sandboxed copilot profile to strip auto-approve flags, got %v", provider.cliWorker.Profile.AutoApproveFlags)
	}
}
