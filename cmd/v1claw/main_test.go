package main

import (
	"testing"

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
