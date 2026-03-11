package providers

import (
	"testing"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
)

func TestCreateProvider_MoonshotExplicitProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Provider = "moonshot"
	cfg.Agents.Defaults.Model = "kimi-k2"
	cfg.Providers.Moonshot.APIKey = "moonshot-key"

	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider(moonshot) error = %v", err)
	}

	httpProvider, ok := provider.(*HTTPProvider)
	if !ok {
		t.Fatalf("CreateProvider(moonshot) returned %T, want *HTTPProvider", provider)
	}
	if httpProvider.apiKey != "moonshot-key" {
		t.Fatalf("apiKey = %q, want moonshot-key", httpProvider.apiKey)
	}
	if httpProvider.apiBase != "https://api.moonshot.cn/v1" {
		t.Fatalf("apiBase = %q, want https://api.moonshot.cn/v1", httpProvider.apiBase)
	}
}

func TestCreateProvider_OllamaExplicitProviderDoesNotRequireAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Provider = "ollama"
	cfg.Agents.Defaults.Model = "llama3.2"

	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider(ollama) error = %v", err)
	}

	httpProvider, ok := provider.(*HTTPProvider)
	if !ok {
		t.Fatalf("CreateProvider(ollama) returned %T, want *HTTPProvider", provider)
	}
	if httpProvider.apiKey != "" {
		t.Fatalf("apiKey = %q, want empty", httpProvider.apiKey)
	}
	if httpProvider.apiBase != "http://localhost:11434/v1" {
		t.Fatalf("apiBase = %q, want http://localhost:11434/v1", httpProvider.apiBase)
	}
}

func TestCreateProvider_VLLMExplicitProviderDoesNotRequireAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Provider = "vllm"
	cfg.Agents.Defaults.Model = "custom-local-model"

	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider(vllm) error = %v", err)
	}

	httpProvider, ok := provider.(*HTTPProvider)
	if !ok {
		t.Fatalf("CreateProvider(vllm) returned %T, want *HTTPProvider", provider)
	}
	if httpProvider.apiKey != "" {
		t.Fatalf("apiKey = %q, want empty", httpProvider.apiKey)
	}
	if httpProvider.apiBase != "http://localhost:8000/v1" {
		t.Fatalf("apiBase = %q, want http://localhost:8000/v1", httpProvider.apiBase)
	}
}

func TestCreateProvider_GitHubCopilotDefaultsToStdioCommand(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agents.Defaults.Provider = "github_copilot"
	cfg.Agents.Defaults.Model = "gpt-4.1"

	provider, err := CreateProvider(cfg)
	if err != nil {
		t.Fatalf("CreateProvider(github_copilot) error = %v", err)
	}

	copilotProvider, ok := provider.(*GitHubCopilotProvider)
	if !ok {
		t.Fatalf("CreateProvider(github_copilot) returned %T, want *GitHubCopilotProvider", provider)
	}
	if copilotProvider.connectMode != "stdio" {
		t.Fatalf("connectMode = %q, want stdio", copilotProvider.connectMode)
	}
	if copilotProvider.uri != "copilot" {
		t.Fatalf("uri = %q, want copilot", copilotProvider.uri)
	}
}
