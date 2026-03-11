package main

import (
	"testing"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestRunDoctor_ReturnsFalseWithoutConfig(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	assert.False(t, runDoctor())
}

func TestEnabledChannelNames(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Slack.Enabled = true
	cfg.Channels.OneBot.Enabled = true

	assert.Equal(t, []string{"Telegram", "Slack", "OneBot"}, enabledChannelNames(cfg))
}

func TestDisableAllChannels(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Discord.Enabled = true
	cfg.Channels.Slack.Enabled = true
	cfg.Channels.WhatsApp.Enabled = true
	cfg.Channels.LINE.Enabled = true
	cfg.Channels.DingTalk.Enabled = true
	cfg.Channels.Feishu.Enabled = true
	cfg.Channels.QQ.Enabled = true
	cfg.Channels.OneBot.Enabled = true
	cfg.Channels.MaixCam.Enabled = true

	disableAllChannels(cfg)

	assert.Empty(t, enabledChannelNames(cfg))
}

func TestEnabledPermissionIDs(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Permissions.Camera = true
	cfg.Permissions.Notifications = true
	cfg.Permissions.Screen = true

	assert.Equal(t, []string{"camera", "screen", "notifications"}, enabledPermissionIDs(cfg))
}

func TestProviderCredentialStatus_BedrockDoesNotRequireAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()

	label, ready, hint := providerCredentialStatus(cfg, "bedrock")
	assert.Equal(t, "AWS credentials / profile", label)
	assert.True(t, ready)
	assert.Empty(t, hint)
}

func TestProviderCredentialStatus_GeminiRequiresCredentials(t *testing.T) {
	cfg := config.DefaultConfig()

	label, ready, hint := providerCredentialStatus(cfg, "gemini")
	assert.Empty(t, label)
	assert.False(t, ready)
	assert.Contains(t, hint, "configure")
}
