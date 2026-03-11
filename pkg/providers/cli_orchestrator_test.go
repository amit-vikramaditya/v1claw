package providers

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutonomousCLIWorker_Execute_SubcommandBeforeFlagsAndTask(t *testing.T) {
	script := writeArgPrinter(t)
	worker := &AutonomousCLIWorker{
		Profile: CLIInteractionProfile{
			Name:             "codex",
			Command:          script,
			PromptFlag:       "exec",
			AutoApproveFlags: []string{"--dangerously-bypass-approvals-and-sandbox"},
			OutputFormatFlag: "--json",
		},
	}

	out, err := worker.Execute(context.Background(), "fix the bug")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	want := []string{"exec", "--dangerously-bypass-approvals-and-sandbox", "--json", "fix the bug"}
	if len(lines) != len(want) {
		t.Fatalf("unexpected arg count: got %v want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestAutonomousCLIWorker_Execute_PromptFlagPlacedBeforeTask(t *testing.T) {
	script := writeArgPrinter(t)
	worker := &AutonomousCLIWorker{
		Profile: CLIInteractionProfile{
			Name:             "gemini",
			Command:          script,
			PromptFlag:       "--prompt",
			AutoApproveFlags: []string{"--approval-mode", "yolo"},
			OutputFormatFlag: "--output-format json",
		},
	}

	out, err := worker.Execute(context.Background(), "research this")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	want := []string{"--approval-mode", "yolo", "--output-format", "json", "--prompt", "research this"}
	if len(lines) != len(want) {
		t.Fatalf("unexpected arg count: got %v want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

func TestSandboxSafeProfile_RemovesAutoApproveFlagsKeepsStaticFlags(t *testing.T) {
	profile := CLIInteractionProfile{
		Name:             "aider",
		Command:          "aider",
		PromptFlag:       "--message",
		StaticFlags:      []string{"--no-auto-commits"},
		AutoApproveFlags: []string{"--yes-always"},
	}

	safe := SandboxSafeProfile(profile)

	if len(safe.AutoApproveFlags) != 0 {
		t.Fatalf("expected auto-approve flags to be stripped, got %v", safe.AutoApproveFlags)
	}
	if len(safe.StaticFlags) != 1 || safe.StaticFlags[0] != "--no-auto-commits" {
		t.Fatalf("expected safe static flags to remain, got %v", safe.StaticFlags)
	}
}

func TestAutonomousCLIWorker_Execute_StaticFlagsPreservedWithoutAutoApproveFlags(t *testing.T) {
	script := writeArgPrinter(t)
	worker := &AutonomousCLIWorker{
		Profile: CLIInteractionProfile{
			Name:             "aider",
			Command:          script,
			PromptFlag:       "--message",
			StaticFlags:      []string{"--no-auto-commits"},
			AutoApproveFlags: nil,
		},
	}

	out, err := worker.Execute(context.Background(), "fix this")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	want := []string{"--no-auto-commits", "--message", "fix this"}
	if len(lines) != len(want) {
		t.Fatalf("unexpected arg count: got %v want %v", lines, want)
	}
	for i := range want {
		if lines[i] != want[i] {
			t.Fatalf("arg %d = %q, want %q", i, lines[i], want[i])
		}
	}
}

func writeArgPrinter(t *testing.T) string {
	t.Helper()
	scriptPath := filepath.Join(t.TempDir(), "arg-printer.sh")
	script := "#!/bin/sh\nfor arg in \"$@\"; do\n  printf '%s\\n' \"$arg\"\ndone\n"
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		t.Fatalf("write script: %v", err)
	}
	return scriptPath
}
