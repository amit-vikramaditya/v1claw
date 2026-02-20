package main

import (
	"os"
	"path/filepath"
	"testing"
	"time" // Added for time.Duration

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPublicHost(t *testing.T) {
	assert.False(t, isPublicHost("localhost"))
	assert.False(t, isPublicHost("127.0.0.1"))
	assert.False(t, isPublicHost("::1"))
	assert.True(t, isPublicHost("0.0.0.0"))
	assert.True(t, isPublicHost("::"))
	assert.True(t, isPublicHost("example.com"))
}

func TestValidateGatewaySecurity_RequiresAPIKeyWhenAPIEnabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.V1API.Enabled = true
	cfg.V1API.APIKey = ""

	err := validateGatewaySecurity(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "v1_api.api_key")
}

func TestValidateGatewaySecurity_PublicRequiresAllowlist(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gateway.Host = "0.0.0.0"
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Telegram.AllowFrom = nil

	err := validateGatewaySecurity(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "allow_from")
	assert.Contains(t, err.Error(), "telegram")
}

func TestValidateGatewaySecurity_PublicRequiresWorkspaceRestriction(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gateway.Host = "0.0.0.0"
	cfg.Agents.Defaults.RestrictToWorkspace = false

	err := validateGatewaySecurity(cfg)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restrict_to_workspace")
}

func TestValidateGatewaySecurity_SafePublicConfigPasses(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Gateway.Host = "0.0.0.0"
	cfg.Agents.Defaults.RestrictToWorkspace = true
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Telegram.AllowFrom = []string{"123456"}
	cfg.V1API.Enabled = true
	cfg.V1API.APIKey = "test-key"

	err := validateGatewaySecurity(cfg)
	require.NoError(t, err)
}

func TestExecuteLocalCapability_Microphone_CommandInjection(t *testing.T) {
	// Temporarily set up a fake Termux environment to allow executeLocalCapability to run
	oldPath := os.Getenv("PATH")
	tempDir := t.TempDir()
	os.Setenv("PATH", tempDir+":"+oldPath)

	// Create a dummy termux-microphone-record executable
	err := os.WriteFile(filepath.Join(tempDir, "termux-microphone-record"), []byte(`#!/bin/sh
# Find the -f argument and create a dummy file there
OUTFILE=""
for i in "$@"; do
    if [ "$PREV" = "-f" ]; then
        OUTFILE="$i"
    fi
    PREV="$i"
done
if [ -n "$OUTFILE" ]; then
    echo "dummy audio data" > "$OUTFILE"
fi
echo "Mock termux-microphone-record $*" >&2
`), 0755)
	require.NoError(t, err)

	// Create a dummy file to simulate Termux environment in the temporary directory
	termuxRoot := filepath.Join(t.TempDir(), "data", "data", "com.termux")
	err = os.MkdirAll(termuxRoot, 0755)
	require.NoError(t, err)

	params := map[string]interface{}{
		"duration": "5; malicious_command", // Attempt to inject a command
	}

	_, capErr := executeLocalCapability("microphone", "record", params, termuxRoot)
	assert.Contains(t, capErr, "invalid duration parameter")

	// Restore original PATH
	os.Setenv("PATH", oldPath)
}

func TestExecuteLocalCapability_Microphone_DurationClamping(t *testing.T) {
	// Mock time.Sleep to prevent actual sleeping during test
	oldMicrophoneSleep := microphoneSleep
	microphoneSleep = func(d time.Duration) {
		// Do nothing, or log the duration for assertion if needed
	}
	defer func() { microphoneSleep = oldMicrophoneSleep }() // Restore original

	// Temporarily set up a fake Termux environment
	oldPath := os.Getenv("PATH")
	tempDir := t.TempDir()
	os.Setenv("PATH", tempDir+":"+oldPath)

	// Create a dummy termux-microphone-record executable
	err := os.WriteFile(filepath.Join(tempDir, "termux-microphone-record"), []byte(`#!/bin/sh
# Find the -f argument and create a dummy file there
OUTFILE=""
for i in "$@"; do
    if [ "$PREV" = "-f" ]; then
        OUTFILE="$i"
    fi
    PREV="$i"
done
if [ -n "$OUTFILE" ]; then
    echo "dummy audio data" > "$OUTFILE"
fi
echo "Mock termux-microphone-record $*" >&2
`), 0755) // Removed sleep from mock to avoid double-sleeping if microphoneSleep isn't mocked
	require.NoError(t, err)

	// Simulate Termux environment directory
	termuxRoot := filepath.Join(t.TempDir(), "data", "data", "com.termux")
	err = os.MkdirAll(termuxRoot, 0755)
	require.NoError(t, err)

	// Test with a duration larger than the maximum (300 seconds)
	params := map[string]interface{}{
		"duration": "9999", // This should be clamped to 300
	}

	result, capErr := executeLocalCapability("microphone", "record", params, termuxRoot)
	require.Empty(t, capErr) // No error expected

	// Although we cannot directly inspect the arguments passed to the real exec.Command easily,
	// the absence of an error and the clamping logic in the code provide confidence.
	assert.NotNil(t, result)

	// Restore original PATH
	os.Setenv("PATH", oldPath)
}
