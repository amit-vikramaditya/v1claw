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
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	stepStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	boxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 3).
			MarginTop(1).
			MarginBottom(1)
	demoBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("10")).
			Padding(1, 3).
			MarginTop(1).
			MarginBottom(1)
)

// ─── Entry point ─────────────────────────────────────────────────────────────

func onboardCmd() {
	args := os.Args[2:]

	// Fast-exit flags that don't need the wizard.
	for _, arg := range args {
		if arg == "--refresh" {
			onboardRefresh()
			return
		}
	}

	// --auto mode: non-interactive setup via flags.
	//   v1claw onboard --auto --provider gemini --api-key sk-...
	//   v1claw onboard --auto --provider gemini --api-key sk-... --model gemini-3.1-pro-preview
	//   v1claw onboard --auto --provider gemini --api-key sk-... --workspace /my/path
	autoMode := false
	for _, arg := range args {
		if arg == "--auto" {
			autoMode = true
			break
		}
	}
	if autoMode {
		onboardAuto(args)
		return
	}

	printOnboardWelcome()

	cfg := config.DefaultConfig()

	// Load existing config if present, so re-running onboard doesn't wipe settings.
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); err == nil {
		if loaded, err := loadConfig(); err == nil {
			cfg = loaded
		}
	}

	// Step 1 – Workspace
	fmt.Println(stepStyle.Render("\n  Step 1 of 6 — Where should V1Claw store its files?"))
	if !onboardWorkspace(cfg) {
		return
	}

	// Step 2 – Provider
	fmt.Println(stepStyle.Render("\n  Step 2 of 6 — Choose your AI provider"))
	providerID, providerURL := onboardProvider(cfg)
	if providerID == "" {
		return
	}

	// Step 3 – API key + live validation
	fmt.Println(stepStyle.Render("\n  Step 3 of 6 — Enter your API key"))
	if !onboardAPIKey(cfg, providerID, providerURL) {
		return
	}

	// Step 4 – Identity (name the AI and yourself)
	fmt.Println(stepStyle.Render("\n  Step 4 of 6 — Give your assistant a name"))
	aiName, userName := onboardIdentity(cfg)

	// Step 5 – Tools
	fmt.Println(stepStyle.Render("\n  Step 5 of 6 — Enable web search (free, no key needed)"))
	onboardTools(cfg)

	// Step 6 – Save, seed workspace, and live test
	fmt.Println(stepStyle.Render("\n  Step 6 of 6 — Saving and running a live test…"))
	if !onboardSaveAndTest(cfg, configPath, aiName, userName) {
		return
	}

	printOnboardSuccess(cfg, configPath, aiName)
}

// ─── Welcome screen ───────────────────────────────────────────────────────────

func printOnboardWelcome() {
	welcome := `
 Welcome to V1Claw 🤖

 V1Claw is your personal AI assistant that runs on your own computer.
 It can:  search the web · read & write files · run commands
         remember things · and talk to you in plain English

 This wizard takes about 2 minutes.
 Press Ctrl+C at any time to quit.`

	fmt.Println(boxStyle.Render(titleStyle.Render(welcome)))
}

// ─── Step 1: Workspace ───────────────────────────────────────────────────────

func onboardWorkspace(cfg *config.Config) bool {
	home, _ := os.UserHomeDir()
	defaultPath := filepath.Join(home, ".v1claw", "workspace")

	recommended := fmt.Sprintf("~/.v1claw/workspace  %s", stepStyle.Render("← recommended"))

	var choice string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Where should the AI store its memory and workspace files?").
				Description("This folder holds conversation memory, skills, and workspace files.").
				Options(
					huh.NewOption(recommended, "default"),
					huh.NewOption("Current folder  "+stepStyle.Render("(./workspace)"), "current"),
					huh.NewOption("Custom path…", "custom"),
				).
				Value(&choice),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Println("Setup cancelled.")
		return false
	}

	switch choice {
	case "default":
		cfg.Workspace.Path = defaultPath
	case "current":
		cwd, _ := os.Getwd()
		cfg.Workspace.Path = filepath.Join(cwd, "workspace")
	case "custom":
		var customPath string
		customForm := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter the full path for the workspace:").
					Placeholder(defaultPath).
					Value(&customPath),
			),
		)
		if err := customForm.Run(); err != nil || strings.TrimSpace(customPath) == "" {
			cfg.Workspace.Path = defaultPath
		} else {
			cfg.Workspace.Path = expandHome(strings.TrimSpace(customPath))
		}
	}

	cfg.Workspace.Sandboxed = true
	fmt.Printf("\n  %s Workspace: %s\n", successStyle.Render("✓"), cfg.Workspace.Path)
	return true
}

// ─── Step 2: Provider selection ──────────────────────────────────────────────

// onboardProvider returns (providerID, keyURL). Empty providerID means cancelled.
func onboardProvider(cfg *config.Config) (string, string) {
	var providerID string

	options := make([]huh.Option[string], 0, len(traditional))
	for _, p := range traditional {
		label := fmt.Sprintf("%-14s  %s", p.name, stepStyle.Render(p.desc))
		options = append(options, huh.NewOption(label, p.id))
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which AI provider would you like to use?").
				Description("Not sure? Pick Google Gemini — it has a free tier and works great.").
				Options(options...).
				Value(&providerID),
		),
	)

	if err := form.Run(); err != nil {
		fmt.Println("Setup cancelled.")
		return "", ""
	}

	// Find the key URL for the selected provider.
	var keyURL string
	for _, p := range traditional {
		if p.id == providerID {
			keyURL = p.keyURL
			break
		}
	}

	// Set default model.
	if models, ok := providerModels[providerID]; ok && len(models) > 0 {
		cfg.Agents.Defaults.Provider = providerID
		cfg.Agents.Defaults.Model = models[0]
	}

	fmt.Printf("\n  %s Provider: %s  (model: %s)\n",
		successStyle.Render("✓"), providerID, cfg.Agents.Defaults.Model)
	return providerID, keyURL
}

// ─── Step 3: API key + live validation ───────────────────────────────────────

func onboardAPIKey(cfg *config.Config, providerID, keyURL string) bool {
	// Keyless and special providers have their own flows.
	switch providerID {
	case "vertex":
		return onboardVertexAuth(cfg)
	case "bedrock":
		return onboardBedrockAuth()
	case "azure_openai":
		return onboardAzureAuth(cfg)
	}
	return onboardAPIKeyWithKey(cfg, providerID, keyURL)
}

// onboardVertexAuth configures Google Vertex AI (gcloud auth, no API key).
func onboardVertexAuth(cfg *config.Config) bool {
	fmt.Printf("\n  %s Google Vertex AI uses your gcloud credentials — no API key needed.\n\n",
		warnStyle.Render("→"))
	fmt.Printf("  If you haven't already, run:  %s\n\n",
		titleStyle.Render("gcloud auth application-default login"))

	var projectID string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Enter your Google Cloud Project ID:").
			Placeholder("my-gcp-project").
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("project ID cannot be empty")
				}
				return nil
			}).
			Value(&projectID),
	))
	if err := form.Run(); err != nil {
		fmt.Println("Setup cancelled.")
		return false
	}
	cfg.Providers.Vertex.ProjectID = strings.TrimSpace(projectID)
	if cfg.Providers.Vertex.Location == "" {
		cfg.Providers.Vertex.Location = "us-central1"
	}
	fmt.Printf("  %s Vertex AI configured (project: %s, region: %s)\n",
		successStyle.Render("✓"), cfg.Providers.Vertex.ProjectID, cfg.Providers.Vertex.Location)
	return true
}

// onboardBedrockAuth configures AWS Bedrock (uses ~/.aws/credentials, no API key).
func onboardBedrockAuth() bool {
	fmt.Printf("\n  %s AWS Bedrock uses your AWS credentials — no API key needed.\n\n",
		warnStyle.Render("→"))
	fmt.Printf("  Make sure  %s  is configured,\n  or set  %s  environment variables.\n\n",
		titleStyle.Render("~/.aws/credentials"),
		titleStyle.Render("AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY"))
	fmt.Printf("  %s AWS Bedrock configured (reads credentials from environment)\n",
		successStyle.Render("✓"))
	return true
}

// onboardAzureAuth configures Azure OpenAI (endpoint + deployment + api_key).
func onboardAzureAuth(cfg *config.Config) bool {
	fmt.Printf("\n  %s  Get your endpoint and key from: %s\n\n",
		warnStyle.Render("→"), titleStyle.Render("https://portal.azure.com"))

	var endpoint, deployment, apiKey string
	form := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Azure OpenAI Endpoint:").
			Placeholder("https://mycompany.openai.azure.com").
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("endpoint cannot be empty")
				}
				return nil
			}).
			Value(&endpoint),
		huh.NewInput().
			Title("Deployment Name:").
			Placeholder("gpt-4o-prod").
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("deployment cannot be empty")
				}
				return nil
			}).
			Value(&deployment),
		huh.NewInput().
			Title("API Key:").
			Password(true).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return fmt.Errorf("API key cannot be empty")
				}
				return nil
			}).
			Value(&apiKey),
	))
	if err := form.Run(); err != nil {
		fmt.Println("Setup cancelled.")
		return false
	}
	cfg.Providers.AzureOpenAI.Endpoint = strings.TrimSpace(endpoint)
	cfg.Providers.AzureOpenAI.Deployment = strings.TrimSpace(deployment)
	cfg.Providers.AzureOpenAI.APIKey = strings.TrimSpace(apiKey)
	// Azure model is the deployment name.
	cfg.Agents.Defaults.Model = strings.TrimSpace(deployment)

	fmt.Printf("  %s Azure OpenAI configured (endpoint: %s, deployment: %s)\n",
		successStyle.Render("✓"), cfg.Providers.AzureOpenAI.Endpoint, cfg.Providers.AzureOpenAI.Deployment)
	return true
}

// onboardAPIKeyWithKey handles the standard API-key flow with live validation.
func onboardAPIKeyWithKey(cfg *config.Config, providerID, keyURL string) bool {
	fmt.Printf("\n  %s  Get your free API key at: %s\n\n",
		warnStyle.Render("→"), titleStyle.Render(keyURL))

	for attempt := 1; attempt <= 3; attempt++ {
		var apiKey string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Paste your API key:").
					Description("The key will be stored only on this machine.").
					Password(true).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return fmt.Errorf("API key cannot be empty")
						}
						return nil
					}).
					Value(&apiKey),
			),
		)

		if err := form.Run(); err != nil {
			fmt.Println("Setup cancelled.")
			return false
		}

		apiKey = strings.TrimSpace(apiKey)
		setProviderKey(cfg, providerID, apiKey)

		// Live validation.
		var validationErr error
		withSpinner("Testing connection to "+providerID+"…", func() {
			validationErr = validateProviderKey(cfg)
		})

		if validationErr == nil {
			fmt.Printf("  %s Connected! The API key works.\n", successStyle.Render("✓"))
			return true
		}

		if attempt < 3 {
			fmt.Printf("  %s Could not connect: %s\n", errorStyle.Render("✗"), simplifyProviderError(validationErr))
			fmt.Printf("  %s (attempt %d/3) — please check the key and try again\n\n", warnStyle.Render("→"), attempt)
		} else {
			fmt.Printf("  %s Could not validate the key after 3 attempts: %s\n",
				errorStyle.Render("✗"), simplifyProviderError(validationErr))

			var skip string
			skipForm := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("What would you like to do?").
						Options(
							huh.NewOption("Continue anyway  "+stepStyle.Render("(you can fix this later with: v1claw configure)"), "continue"),
							huh.NewOption("Quit and fix the key", "quit"),
						).
						Value(&skip),
				),
			)
			_ = skipForm.Run()
			if skip == "quit" {
				fmt.Println("\nRun  v1claw onboard  again when you have the correct key.")
				return false
			}
			// User chose to continue despite failure.
			return true
		}
	}
	return true
}

// validateProviderKey sends a minimal test message to the configured provider.
func validateProviderKey(cfg *config.Config) error {
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	msgs := []providers.Message{
		{Role: "user", Content: "Say \"OK\" and nothing else."},
	}

	opts := map[string]interface{}{
		"max_tokens":  10,
		"temperature": 0.0,
	}

	_, err = provider.Chat(ctx, msgs, nil, cfg.Agents.Defaults.Model, opts)
	return err
}

// simplifyProviderError strips verbose HTTP noise for end-users.
func simplifyProviderError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	switch {
	case strings.Contains(msg, "401") || strings.Contains(msg, "403") ||
		strings.Contains(strings.ToLower(msg), "invalid api key") ||
		strings.Contains(strings.ToLower(msg), "api key not valid") ||
		strings.Contains(strings.ToLower(msg), "not valid") ||
		strings.Contains(msg, "Unauthorized"):
		return "Invalid API key — double-check what you pasted."
	case strings.Contains(msg, "429") || strings.Contains(msg, "quota"):
		return "Rate limit or quota exceeded. Wait a moment or check your plan."
	case strings.Contains(msg, "timeout") || strings.Contains(msg, "context deadline"):
		return "Request timed out. Check your internet connection."
	case strings.Contains(msg, "no such host") || strings.Contains(msg, "dial"):
		return "Cannot reach the server. Check your internet connection."
	default:
		if len(msg) > 120 {
			return msg[:120] + "…"
		}
		return msg
	}
}

// ─── Step 4: Identity ────────────────────────────────────────────────────────

func onboardIdentity(cfg *config.Config) (string, string) {
	aiName := "V1"
	userName := "User"

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("What should we call your AI assistant?").
				Description("This is the name it will use to introduce itself.").
				Placeholder("V1").
				Value(&aiName),
			huh.NewInput().
				Title("What should the AI call you?").
				Placeholder("User").
				Value(&userName),
		),
	)

	if err := form.Run(); err != nil {
		return aiName, userName // Keep defaults on cancel.
	}

	if strings.TrimSpace(aiName) == "" {
		aiName = "V1"
	}
	if strings.TrimSpace(userName) == "" {
		userName = "User"
	}

	aiName = sanitizeOnboardingField(aiName, false)
	userName = sanitizeOnboardingField(userName, false)

	fmt.Printf("  %s AI name: %s   Your name: %s\n", successStyle.Render("✓"), aiName, userName)
	// Return both names; memory init happens after workspace templates are seeded.
	return aiName, userName
}

// ─── Step 5: Tools ───────────────────────────────────────────────────────────

func onboardTools(cfg *config.Config) {
	var enableSearch bool

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable web search?").
				Description("Uses DuckDuckGo — completely free, no API key required.").
				Affirmative("Yes, enable web search").
				Negative("Skip for now").
				Value(&enableSearch),
		),
	)

	if err := form.Run(); err != nil {
		return
	}

	if enableSearch {
		cfg.Tools.Web.DuckDuckGo.Enabled = true
		fmt.Printf("  %s Web search enabled (DuckDuckGo)\n", successStyle.Render("✓"))
	} else {
		fmt.Printf("  %s Web search skipped  %s\n", warnStyle.Render("○"), stepStyle.Render("(enable later with: v1claw configure)"))
	}
}

// ─── Step 6: Save + live test ─────────────────────────────────────────────────

func onboardSaveAndTest(cfg *config.Config, configPath string, aiName string, userName string) bool {
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("  %s Could not save config: %v\n", errorStyle.Render("✗"), err)
		return false
	}

	// Seed workspace template files first, then write identity memory on top.
	createWorkspaceTemplates(cfg.Workspace.Path)
	if cfg.Workspace.Path != "" {
		initMemory(cfg.Workspace.Path, aiName, "Your helpful personal AI assistant", userName, "")
	}

	fmt.Printf("  %s Config saved to: %s\n", successStyle.Render("✓"), configPath)

	// Run live test.
	fmt.Printf("\n  %s\n\n", titleStyle.Render("Sending a live test message to your AI…"))

	var response string
	var testErr error

	withSpinner("Waiting for response…", func() {
		response, testErr = runOnboardTestMessage(cfg, aiName)
	})

	if testErr != nil {
		fmt.Printf("  %s Live test failed: %s\n", warnStyle.Render("⚠"), simplifyProviderError(testErr))
		fmt.Printf("  %s This is usually fine — your API key was already validated.\n", stepStyle.Render("→"))
		fmt.Printf("  %s Run  v1claw doctor  to check your setup anytime.\n\n", stepStyle.Render("→"))
		return true
	}

	if aiName == "" {
		aiName = "AI"
	}

	demoContent := fmt.Sprintf("  🤖 %s says:\n\n  %s", aiName, wrapText(response, 64, "  "))
	fmt.Println(demoBoxStyle.Render(demoContent))

	return true
}

// runOnboardTestMessage sends a friendly intro message and returns the reply.
// aiName is used in the system prompt so the AI introduces itself by name.
func runOnboardTestMessage(cfg *config.Config, aiName string) (string, error) {
	provider, err := providers.CreateProvider(cfg)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if aiName == "" {
		aiName = "V1"
	}

	msgs := []providers.Message{
		{Role: "system", Content: fmt.Sprintf("You are %s, a helpful personal AI assistant. Keep your response to 1-2 friendly sentences.", aiName)},
		{Role: "user", Content: "Say hello and introduce yourself! Tell me one thing you can help me with."},
	}

	opts := map[string]interface{}{
		"max_tokens":  120,
		"temperature": 0.7,
	}

	resp, err := provider.Chat(ctx, msgs, nil, cfg.Agents.Defaults.Model, opts)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Content), nil
}

// ─── Success screen ───────────────────────────────────────────────────────────

func printOnboardSuccess(cfg *config.Config, configPath string, aiName string) {
	if aiName == "" {
		aiName = "V1"
	}

	successContent := fmt.Sprintf(`
 🎉  %s is ready!

 Start chatting now:

   v1claw agent                         ← interactive chat
   v1claw agent -m "your question"      ← single question

 Try these demo prompts:

   v1claw agent -m "Who are you and what can you do?"
   v1claw agent -m "Search the web for today's top AI news"
   v1claw agent -m "Create a file called hello.txt saying Hello World"

 Useful commands:

   v1claw doctor      ← check that everything is working
   v1claw configure   ← change settings, add channels (Telegram, Discord)
   v1claw skills list ← see available skills
`,
		aiName,
	)

	fmt.Println(boxStyle.
		BorderForeground(lipgloss.Color("10")).
		Render(successStyle.Render(successContent)))

	fmt.Printf("  Config: %s\n", stepStyle.Render(configPath))
	fmt.Printf("  Workspace: %s\n\n", stepStyle.Render(cfg.Workspace.Path))
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// withSpinner runs fn while showing an animated spinner in the terminal.
func withSpinner(label string, fn func()) {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})

	go func() {
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				fmt.Printf("\r  %s  %s  ", stepStyle.Render(frames[i%len(frames)]), label)
				i++
				time.Sleep(80 * time.Millisecond)
			}
		}
	}()

	fn()
	close(done)
	// Clear the spinner line.
	fmt.Printf("\r%s\r", strings.Repeat(" ", len(label)+12))
}

// wrapText soft-wraps s to maxWidth characters, indenting continuation lines with indent.
func wrapText(s string, maxWidth int, indent string) string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	var current strings.Builder
	lineLen := 0

	for _, word := range words {
		if lineLen > 0 && lineLen+1+len(word) > maxWidth {
			lines = append(lines, current.String())
			current.Reset()
			lineLen = 0
		}
		if lineLen > 0 {
			current.WriteByte(' ')
			lineLen++
		}
		current.WriteString(word)
		lineLen += len(word)
	}

	if current.Len() > 0 {
		lines = append(lines, current.String())
	}

	return strings.Join(lines, "\n"+indent)
}

// ─── Auto (non-interactive) mode ─────────────────────────────────────────────

// onboardAuto runs a fully non-interactive setup using flags:
//
//	--provider <id>    Required. e.g. gemini, openai, anthropic, groq, deepseek, openrouter
//	--api-key  <key>   Required for API-key providers (not needed for vertex/bedrock).
//	--model    <model> Optional. Defaults to the first model for the provider.
//	--workspace <path> Optional. Defaults to ~/.v1claw/workspace.
func onboardAuto(args []string) {
	providerID := flagValue(args, "--provider")
	apiKey := flagValue(args, "--api-key")
	model := flagValue(args, "--model")
	workspace := flagValue(args, "--workspace")

	// Validate required flags.
	if providerID == "" {
		fmt.Printf("  %s --provider is required for --auto mode.\n\n", errorStyle.Render("✗"))
		fmt.Printf("  Example:\n")
		fmt.Printf("    v1claw onboard --auto --provider gemini --api-key YOUR_KEY\n\n")
		fmt.Printf("  Providers: gemini, openai, anthropic, groq, deepseek, openrouter, openrouter, nvidia\n")
		fmt.Printf("  Enterprise (no key needed): vertex, bedrock\n\n")
		os.Exit(1)
	}

	needsKey := providerID != "vertex" && providerID != "bedrock"
	if needsKey && apiKey == "" {
		fmt.Printf("  %s --api-key is required for provider %q.\n\n", errorStyle.Render("✗"), providerID)
		fmt.Printf("  Example:\n")
		fmt.Printf("    v1claw onboard --auto --provider %s --api-key YOUR_KEY\n\n", providerID)
		os.Exit(1)
	}

	fmt.Printf("\n%s V1Claw non-interactive setup\n\n", titleStyle.Render("⚡"))

	configPath := getConfigPath()
	cfg := config.DefaultConfig()
	// Load existing config if present — don't wipe existing settings.
	if _, err := os.Stat(configPath); err == nil {
		if loaded, err := loadConfig(); err == nil {
			cfg = loaded
		}
	}

	// Set workspace.
	if workspace != "" {
		cfg.Workspace.Path = expandHome(workspace)
	} else if cfg.Workspace.Path == "" {
		home, _ := os.UserHomeDir()
		cfg.Workspace.Path = filepath.Join(home, ".v1claw", "workspace")
	}

	// Set provider + model.
	cfg.Agents.Defaults.Provider = providerID
	if model != "" {
		cfg.Agents.Defaults.Model = model
	} else if models, ok := providerModels[providerID]; ok && len(models) > 0 {
		cfg.Agents.Defaults.Model = models[0]
	}

	// Set API key.
	if apiKey != "" {
		setProviderKey(cfg, providerID, apiKey)
	}

	fmt.Printf("  %s Provider:  %s\n", successStyle.Render("✓"), providerID)
	fmt.Printf("  %s Model:     %s\n", successStyle.Render("✓"), cfg.Agents.Defaults.Model)
	fmt.Printf("  %s Workspace: %s\n", successStyle.Render("✓"), cfg.Workspace.Path)

	// Validate the key (live test).
	if needsKey {
		var testErr error
		withSpinner("Testing connection to "+providerID+"…", func() {
			testErr = validateProviderKey(cfg)
		})
		if testErr != nil {
			fmt.Printf("  %s Connection test failed: %s\n", errorStyle.Render("✗"), simplifyProviderError(testErr))
			fmt.Printf("  %s Config was NOT saved. Check your API key and try again.\n\n", warnStyle.Render("→"))
			os.Exit(1)
		}
		fmt.Printf("  %s API key validated — connection works.\n", successStyle.Render("✓"))
	}

	// Save config + seed workspace.
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("  %s Could not save config: %v\n", errorStyle.Render("✗"), err)
		os.Exit(1)
	}
	createWorkspaceTemplates(cfg.Workspace.Path)

	fmt.Printf("  %s Config saved to: %s\n\n", successStyle.Render("✓"), configPath)
	fmt.Printf("  %s\n\n", titleStyle.Render("Setup complete. Try it:"))
	fmt.Printf("    v1claw agent -m \"Hello! What can you do?\"\n\n")
}

// flagValue returns the value for a --flag <value> pair in args,
// or "" if the flag is not present.
func flagValue(args []string, flag string) string {
	for i, arg := range args {
		if arg == flag && i+1 < len(args) {
			return args[i+1]
		}
		// Support --flag=value syntax too.
		prefix := flag + "="
		if strings.HasPrefix(arg, prefix) {
			return arg[len(prefix):]
		}
	}
	return ""
}

// ─── Refresh mode ────────────────────────────────────────────────────────────

// onboardRefresh upgrades an existing install non-interactively:
//  1. Load existing config (preserving all user values).
//  2. Re-save it — this round-trip adds any new fields with their schema defaults
//     without overwriting anything the user previously set.
//  3. Re-sync workspace templates (skipping files that already exist).
//
// Equivalent to nanobot's "N" (keep) path in onboard. Run it after upgrading
// v1claw to get new config fields without re-running the full wizard.
func onboardRefresh() {
	configPath := getConfigPath()
	fmt.Printf("\n%s Refreshing V1Claw config…\n\n", titleStyle.Render("↻"))

	cfg, err := loadConfig()
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("  %s No config found at %s — run  v1claw onboard  first.\n",
				errorStyle.Render("✗"), configPath)
			return
		}
		fmt.Printf("  %s Failed to load config: %v\n", errorStyle.Render("✗"), err)
		return
	}

	// Re-save: JSON marshal → write. New fields in DefaultConfig will be added
	// with their zero/default values; existing values are preserved.
	if err := config.SaveConfig(configPath, cfg); err != nil {
		fmt.Printf("  %s Failed to save config: %v\n", errorStyle.Render("✗"), err)
		return
	}
	fmt.Printf("  %s Config refreshed — new schema fields added with defaults.\n",
		successStyle.Render("✓"))
	fmt.Printf("    %s\n\n", stepStyle.Render(configPath))

	// Re-sync workspace templates (only creates missing files, never overwrites).
	workspace := cfg.WorkspacePath()
	if workspace != "" {
		createWorkspaceTemplates(workspace)
		fmt.Printf("  %s Workspace templates refreshed (existing files untouched).\n",
			successStyle.Render("✓"))
		fmt.Printf("    %s\n\n", stepStyle.Render(workspace))
	}

	fmt.Printf("  %s Run  v1claw doctor  to verify everything is configured.\n\n",
		stepStyle.Render("→"))
}
