package providers

import (
	"context"
	"fmt"

	copilot "github.com/github/copilot-sdk/go"
)

type GitHubCopilotProvider struct {
	uri         string
	connectMode string // `stdio` or `grpc``

	session *copilot.Session
}

func NewGitHubCopilotProvider(uri string, connectMode string, model string) (*GitHubCopilotProvider, error) {
	if connectMode == "" {
		connectMode = "grpc"
	}

	if model == "" {
		model = "gpt-4.1"
	}

	switch connectMode {
	case "stdio":
		return nil, fmt.Errorf("github copilot connect_mode=stdio is not implemented")
	case "grpc":
		client := copilot.NewClient(&copilot.ClientOptions{
			CLIUrl: uri,
		})
		if err := client.Start(context.Background()); err != nil {
			return nil, fmt.Errorf("can't connect to github copilot: %w", err)
		}

		session, err := client.CreateSession(context.Background(), &copilot.SessionConfig{
			Model: model,
			Hooks: &copilot.SessionHooks{},
		})
		if err != nil {
			client.Stop()
			return nil, fmt.Errorf("create github copilot session: %w", err)
		}

		return &GitHubCopilotProvider{
			uri:         uri,
			connectMode: connectMode,
			session:     session,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported github copilot connect mode: %s", connectMode)
	}
}

// Chat sends a chat request to GitHub Copilot
func (p *GitHubCopilotProvider) Chat(ctx context.Context, messages []Message, tools []ToolDefinition, model string, options map[string]interface{}) (*LLMResponse, error) {
	if p.session == nil {
		return nil, fmt.Errorf("github copilot session is not initialized")
	}

	// Build the prompt from message history.
	var prompt string
	for _, msg := range messages {
		if msg.Role == "user" {
			prompt = msg.Content
		}
	}
	if prompt == "" && len(messages) > 0 {
		prompt = messages[len(messages)-1].Content
	}

	content, err := p.session.Send(ctx, copilot.MessageOptions{
		Prompt: prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("github copilot send: %w", err)
	}

	return &LLMResponse{
		FinishReason: "stop",
		Content:      content,
	}, nil

}

func (p *GitHubCopilotProvider) GetDefaultModel() string {

	return "gpt-4.1"
}
