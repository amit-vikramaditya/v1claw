package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/amit-vikramaditya/v1claw/pkg/providers"
)

// doctorCmd runs a series of quick health checks and prints a colour-coded report.
func doctorCmd() {
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
		allGood = false
	}

	// ── 2. Load config ───────────────────────────────────────────────────────
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("%s  Load config        %s\n", fail, err.Error())
		hint("Config may be malformed. Try  v1claw configure  to repair it.")
		fmt.Println()
		printDoctorResult(allGood)
		return
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

	// ── 5. API key present ───────────────────────────────────────────────────
	apiKey := apiKeyFromConfig(cfg, providerID)
	if apiKey == "" && providerID != "" && !isLocalProvider(providerID) {
		fmt.Printf("%s  API key            not set for %s\n", fail, providerID)
		hint("Run  v1claw configure → Brain  to add the key.")
		allGood = false
	} else if apiKey != "" {
		fmt.Printf("%s  API key            %s\n", pass, maskKey(apiKey))
	} else if isLocalProvider(providerID) {
		fmt.Printf("%s  API key            local provider — no key needed\n", pass)
	}

	// ── 6. Live connectivity ping ─────────────────────────────────────────────
	if providerID != "" && model != "" && (apiKey != "" || isLocalProvider(providerID)) {
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
				fail, simplifyProviderError(r.err), stepStyle.Render(fmt.Sprintf("(after %s)", elapsed)))
			hint("Check your API key and internet connection.")
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
	var channels []string
	if cfg.Channels.Telegram.Enabled {
		channels = append(channels, "Telegram")
	}
	if cfg.Channels.Discord.Enabled {
		channels = append(channels, "Discord")
	}
	if len(channels) > 0 {
		fmt.Printf("%s  Channels           %s\n", pass, strings.Join(channels, ", "))
	} else {
		fmt.Printf("%s  Channels           none configured  (optional)\n", warn)
		hint("Add Telegram / Discord with  v1claw configure → Channels.")
	}

	fmt.Println()
	printDoctorResult(allGood)
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
	case "nvidia":
		return cfg.Providers.Nvidia.APIKey
	case "github_copilot":
		return cfg.Providers.GitHubCopilot.APIKey
	default:
		return ""
	}
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
