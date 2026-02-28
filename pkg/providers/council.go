package providers

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
)

// ExecuteWithCouncil handles the dynamic routing of a prompt to the primary AI.
// If the primary fails due to a rate limit (429) or service error (503), it
// automatically retries the prompt on the configured Fallback model.
func ExecuteWithCouncil(ctx context.Context, cfg *config.Config, prompt string) (string, error) {
	if !cfg.Council.Enabled {
		// Standard execution if council is not configured
		return RouteToProvider(ctx, cfg, cfg.Agents.Defaults.Provider, cfg.Agents.Defaults.Model, prompt)
	}

	// 1. Attempt Primary
	resp, err := RouteToProvider(ctx, cfg, cfg.Council.Primary, cfg.Council.PrimaryModel, prompt)
	if err == nil {
		return resp, nil
	}

	// Assess if the error is recoverable via Fallback
	errStr := strings.ToLower(err.Error())
	isRecoverable := strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "503") ||
		strings.Contains(errStr, "service unavailable") ||
		strings.Contains(errStr, "timeout")

	if !isRecoverable {
		// Some hard fatal error (e.g. invalid API key 401), we should not mask this by falling back.
		return "", fmt.Errorf("council: primary %s failed with unrecoverable error: %v", cfg.Council.Primary, err)
	}

	// 2. Trigger Fallback
	log.Printf("[Council Warning] Primary (%s) failed: %v. Rerouting to Fallback (%s)...", cfg.Council.Primary, err, cfg.Council.Fallback)

	fallbackResp, fallbackErr := RouteToProvider(ctx, cfg, cfg.Council.Fallback, cfg.Council.FallbackModel, prompt)
	if fallbackErr != nil {
		return "", fmt.Errorf("council: fallback %s also failed after primary failure: %v", cfg.Council.Fallback, fallbackErr)
	}

	return fallbackResp, nil
}

// RouteToProvider acts as the abstraction layer to map a provider string to its specific package executor.
func RouteToProvider(ctx context.Context, cfg *config.Config, providerID, model, prompt string) (string, error) {
	// This maps the string provider name to the actual pkg/providers sub-modules
	// e.g. "openai" -> openai.Generate(ctx, cfg, prompt, false)
	// For CLI tools ("copilot", "codex"), it uses os/exec.

	// Right now we will integrate with the existing generate interfaces
	// Note: This relies on the agent package or specific provider packages having a unified interface.

	// Temporary stub mapping until we inject it into agent/agent.go
	return "", fmt.Errorf("RouteToProvider not fully wired up to native providers yet")
}
