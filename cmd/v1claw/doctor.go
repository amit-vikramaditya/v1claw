package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/auth"
	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/amit-vikramaditya/v1claw/pkg/providers"
)

// doctorCmd runs a series of quick health checks and prints a colour-coded report.
func doctorCmd() {
	if !runDoctor() {
		os.Exit(1)
	}
}

func runDoctor() bool {
	printLogo()
	fmt.Println(titleStyle.Render("  V1Claw Health Check\n"))

	pass := successStyle.Render("  ✓")
	fail := errorStyle.Render("  ✗")
	warn := warnStyle.Render("  ○")
	hint := func(msg string) { fmt.Println(stepStyle.Render("      → " + msg)) }

	allGood := true

	// ── 1. Config file ───────────────────────────────────────────────────────
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("%s  Config file        %s\n", pass, stepStyle.Render(configPath))
	} else {
		fmt.Printf("%s  Config file        not found\n", fail)
		hint("Run  v1claw onboard  to create it.")
		fmt.Println()
		printDoctorResult(false)
		return false
	}

	// ── 2. Load config ───────────────────────────────────────────────────────
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("%s  Load config        %s\n", fail, err.Error())
		hint("Config may be malformed. Try  v1claw configure  to repair it.")
		fmt.Println()
		printDoctorResult(allGood)
		return false
	}

	// ── 3. Workspace directory ───────────────────────────────────────────────
	ws := cfg.WorkspacePath()
	if ws == "" {
		fmt.Printf("%s  Workspace          not configured\n", fail)
		hint("Run  v1claw onboard  or  v1claw configure → Home.")
		allGood = false
	} else if info, statErr := os.Stat(ws); statErr != nil {
		fmt.Printf("%s  Workspace          %s  %s\n", warn, ws, stepStyle.Render("(will be created on first run)"))
	} else if !info.IsDir() {
		fmt.Printf("%s  Workspace          %s  (exists but is not a directory)\n", fail, ws)
		allGood = false
	} else {
		testFile := filepath.Join(ws, ".v1claw_write_test")
		if f, wErr := os.Create(testFile); wErr != nil {
			fmt.Printf("%s  Workspace          %s  (not writable)\n", fail, ws)
			allGood = false
		} else {
			f.Close()
			os.Remove(testFile)
			fmt.Printf("%s  Workspace          %s\n", pass, ws)
		}
	}

	// ── 4. Provider + model ──────────────────────────────────────────────────
	providerID := cfg.Agents.Defaults.Provider
	model := cfg.Agents.Defaults.Model
	if providerID == "" || model == "" {
		fmt.Printf("%s  AI provider        not configured\n", fail)
		hint("Run  v1claw onboard  or  v1claw configure → Brain.")
		allGood = false
	} else {
		fmt.Printf("%s  AI provider        %s  /  %s\n", pass, providerID, model)
	}

	// ── 5. Credentials ───────────────────────────────────────────────────────
	credentialLabel, credentialsReady, credentialHint := providerCredentialStatus(cfg, providerID)
	if providerID != "" && !credentialsReady {
		fmt.Printf("%s  Credentials        not ready for %s\n", fail, providerID)
		if credentialHint != "" {
			hint(credentialHint)
		}
		allGood = false
	} else if credentialLabel != "" {
		fmt.Printf("%s  Credentials        %s\n", pass, credentialLabel)
	}

	// ── 6. Live connectivity ping ─────────────────────────────────────────────
	if providerID != "" && model != "" && credentialsReady {
		type result struct{ err error }
		done := make(chan result, 1)
		go func() {
			p, pErr := providers.CreateProvider(cfg)
			if pErr != nil {
				done <- result{pErr}
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			msgs := []providers.Message{{Role: "user", Content: "ping"}}
			opts := map[string]interface{}{"max_tokens": 5, "temperature": 0.0}
			_, chatErr := p.Chat(ctx, msgs, nil, model, opts)
			done <- result{chatErr}
		}()

		start := time.Now()
		clearLen := len(providerID) + 40
		fmt.Printf("      …  LLM connectivity  testing %s…", providerID)
		r := <-done
		elapsed := time.Since(start).Round(time.Millisecond)
		fmt.Printf("\r%s\r", strings.Repeat(" ", clearLen))

		if r.err == nil {
			fmt.Printf("%s  LLM connectivity   OK  %s\n", pass, stepStyle.Render(fmt.Sprintf("(%s)", elapsed)))
		} else {
			fmt.Printf("%s  LLM connectivity   %s  %s\n",
				fail, simplifyProviderErrorFor(providerID, r.err), stepStyle.Render(fmt.Sprintf("(after %s)", elapsed)))
			hint(providerConnectionHint(cfg, providerID))
			allGood = false
		}
	}

	// ── 7. Tools ─────────────────────────────────────────────────────────────
	var toolList []string
	if cfg.Tools.Web.DuckDuckGo.Enabled {
		toolList = append(toolList, "DuckDuckGo")
	}
	if cfg.Tools.Web.Brave.Enabled {
		toolList = append(toolList, "Brave Search")
	}
	if cfg.Tools.Web.Perplexity.Enabled {
		toolList = append(toolList, "Perplexity")
	}
	if len(toolList) > 0 {
		fmt.Printf("%s  Web search         %s\n", pass, strings.Join(toolList, ", "))
	} else {
		fmt.Printf("%s  Web search         none enabled  (optional)\n", warn)
		hint("Enable with  v1claw configure → Tools.")
	}

	// ── 8. Messaging channels ────────────────────────────────────────────────
	channels := enabledChannelNames(cfg)
	if len(channels) > 0 {
		fmt.Printf("%s  Channels           %s\n", pass, strings.Join(channels, ", "))
	} else {
		fmt.Printf("%s  Channels           none configured  (optional)\n", warn)
		hint("Add channels with  v1claw configure → Channels.")
	}

	fmt.Println()
	printDoctorResult(allGood)
	return allGood
}

func printDoctorResult(allGood bool) {
	if allGood {
		fmt.Println(successStyle.Render("  All checks passed — V1Claw is ready! 🚀"))
		fmt.Println()
		fmt.Printf("%s\n\n", stepStyle.Render("  Run  v1claw agent  to start chatting."))
	} else {
		fmt.Println(warnStyle.Render("  Some checks failed. Fix the items above and re-run  v1claw doctor."))
		fmt.Println()
	}
}

// apiKeyFromConfig returns the configured API key for the named provider.
func apiKeyFromConfig(cfg *config.Config, providerID string) string {
	switch strings.ToLower(providerID) {
	case "gemini":
		return cfg.Providers.Gemini.APIKey
	case "openai", "gpt":
		return cfg.Providers.OpenAI.APIKey
	case "anthropic", "claude":
		return cfg.Providers.Anthropic.APIKey
	case "groq":
		return cfg.Providers.Groq.APIKey
	case "deepseek":
		return cfg.Providers.DeepSeek.APIKey
	case "openrouter":
		return cfg.Providers.OpenRouter.APIKey
	case "zhipu", "glm":
		return cfg.Providers.Zhipu.APIKey
	case "moonshot":
		return cfg.Providers.Moonshot.APIKey
	case "nvidia":
		return cfg.Providers.Nvidia.APIKey
	case "vllm":
		return cfg.Providers.VLLM.APIKey
	case "github_copilot":
		return cfg.Providers.GitHubCopilot.APIKey
	default:
		return ""
	}
}

func providerCredentialStatus(cfg *config.Config, providerID string) (string, bool, string) {
	switch strings.ToLower(providerID) {
	case "":
		return "", false, ""
	case "vertex", "vertex_ai", "vertexai":
		if strings.TrimSpace(cfg.Providers.Vertex.ProjectID) == "" {
			return "", false, "Set Vertex project_id with  v1claw configure → Brain."
		}
		return "gcloud / ADC credentials", true, ""
	case "bedrock", "aws_bedrock", "aws":
		return "AWS credentials / profile", true, ""
	case "ollama":
		apiBase := strings.TrimSpace(cfg.Providers.Ollama.APIBase)
		if apiBase == "" {
			apiBase = defaultProviderAPIBase("ollama")
		}
		return "local Ollama endpoint at " + apiBase, true, ""
	case "vllm":
		apiBase := strings.TrimSpace(cfg.Providers.VLLM.APIBase)
		if apiBase == "" {
			apiBase = defaultProviderAPIBase("vllm")
		}
		if apiKey := strings.TrimSpace(cfg.Providers.VLLM.APIKey); apiKey != "" {
			return maskKey(apiKey) + " @ " + apiBase, true, ""
		}
		return "OpenAI-compatible endpoint at " + apiBase, true, ""
	case "github_copilot", "copilot":
		connectMode := strings.TrimSpace(cfg.Providers.GitHubCopilot.ConnectMode)
		if connectMode == "" {
			connectMode = "stdio"
		}
		target := strings.TrimSpace(cfg.Providers.GitHubCopilot.APIBase)
		if target == "" {
			target = defaultGitHubCopilotTarget(connectMode)
		}
		return fmt.Sprintf("Copilot %s via %s", connectMode, target), true, ""
	case "claude-cli", "claudecode", "claude-code":
		return "Claude CLI auth", true, ""
	case "codex-cli", "codex-code":
		return "Codex CLI auth", true, ""
	case "openai", "gpt":
		return providerConfigCredentialStatus(cfg.Providers.OpenAI, "openai")
	case "anthropic", "claude":
		return providerConfigCredentialStatus(cfg.Providers.Anthropic, "anthropic")
	default:
		apiKey := apiKeyFromConfig(cfg, providerID)
		if apiKey != "" {
			return maskKey(apiKey), true, ""
		}
		if isLocalProvider(providerID) {
			return "local provider — no key needed", true, ""
		}
		return "", false, "Run  v1claw configure → Brain  to add the required credentials."
	}
}

func providerConfigCredentialStatus(providerCfg config.ProviderConfig, authProvider string) (string, bool, string) {
	switch strings.ToLower(strings.TrimSpace(providerCfg.AuthMethod)) {
	case "oauth", "token":
		cred, err := auth.GetCredential(authProvider)
		if err != nil {
			return "", false, fmt.Sprintf("Run  v1claw auth login --provider %s  after configuring auth storage.", authProvider)
		}
		if cred == nil {
			return "", false, fmt.Sprintf("Run  v1claw auth login --provider %s.", authProvider)
		}
		return fmt.Sprintf("stored auth (%s)", providerCfg.AuthMethod), true, ""
	case "codex-cli":
		if _, _, _, err := providers.ReadCodexCliCredentials(); err != nil {
			return "", false, "Run  codex login  to refresh Codex CLI credentials."
		}
		return "Codex CLI credentials", true, ""
	}

	if apiKey := strings.TrimSpace(providerCfg.APIKey); apiKey != "" {
		return maskKey(apiKey), true, ""
	}

	return "", false, "Run  v1claw configure → Brain  to add the required credentials."
}

// maskKey returns a display-safe string like "AIza…a1b2".
func maskKey(key string) string {
	if len(key) <= 8 {
		return strings.Repeat("●", len(key))
	}
	return key[:4] + "…" + key[len(key)-4:]
}

// isLocalProvider returns true for providers that don't need an API key.
func isLocalProvider(providerID string) bool {
	switch strings.ToLower(providerID) {
	case "ollama", "lmstudio", "localai", "llamacpp":
		return true
	}
	return false
}
