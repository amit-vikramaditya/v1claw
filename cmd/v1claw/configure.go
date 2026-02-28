package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/amit-vikramaditya/v1claw/pkg/config"
)

type providerInfo struct {
	id      string
	name    string
	desc    string
	keyHint string
	keyURL  string
}

var traditional = []providerInfo{
	{id: "gemini", name: "Google Gemini", desc: "Free tier available — recommended", keyHint: "Gemini API key", keyURL: "https://aistudio.google.com/apikey"},
	{id: "openai", name: "OpenAI", desc: "GPT-5, GPT-4o, o3", keyHint: "OpenAI API key (starts with sk-)", keyURL: "https://platform.openai.com/api-keys"},
	{id: "anthropic", name: "Anthropic", desc: "Claude Opus 4.6, Sonnet", keyHint: "Anthropic API key (starts with sk-ant-)", keyURL: "https://console.anthropic.com/keys"},
	{id: "groq", name: "Groq", desc: "Llama 3.3, Fast inference, free tier", keyHint: "Groq API key", keyURL: "https://console.groq.com/keys"},
	{id: "deepseek", name: "DeepSeek", desc: "DeepSeek V3, Coder", keyHint: "DeepSeek API key", keyURL: "https://platform.deepseek.com/api_keys"},
	{id: "openrouter", name: "OpenRouter", desc: "100+ models, single API key", keyHint: "OpenRouter API key", keyURL: "https://openrouter.ai/keys"},
	{id: "nvidia", name: "NVIDIA NIM", desc: "NVIDIA hosted models", keyHint: "NVIDIA API key", keyURL: "https://build.nvidia.com"},
}

var providerModels = map[string][]string{
	"openai": {
		"gpt-5.2", "gpt-5.2-pro", "gpt-5", "gpt-5-mini", "gpt-4.1", "o3-deep-research", "o4-mini-deep-research", "gpt-4o", "gpt-4o-mini",
	},
	"gemini": {
		"gemini-3.1-pro-preview", "gemini-3.1-pro-preview-customtools", "gemini-3-flash-preview", "gemini-2.5-pro",
	},
	"anthropic": {
		"claude-opus-4-6", "claude-sonnet-4-6", "claude-haiku-4-5", "claude-3-7-sonnet-latest", "claude-3-5-sonnet-20241022", "claude-3-opus-20240229",
	},
	"deepseek": {
		"deepseek-reasoner", "deepseek-coder", "deepseek-v3",
	},
	"groq": {
		"llama-3.3-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768",
	},
	"openrouter": {
		"anthropic/claude-3.5-sonnet", "openai/gpt-4o", "meta-llama/llama-3.1-405b", "google/gemini-pro-1.5",
	},
}

func configureCmd() {
	configPath := getConfigPath()
	cfg := config.DefaultConfig()

	// Load existing config if available
	if _, err := os.Stat(configPath); err == nil {
		if loaded, err := loadConfig(); err == nil {
			cfg = loaded
		}
	}

	for {
		fmt.Println("\n" + logo + " V1Claw Configuration Menu")
		choice := ""
		prompt := &survey.Select{
			Message: "Select sections to configure:",
			Options: []string{
				"Workspace (Set paths & sandbox)",
				"Model (Providers, Keys, & Models)",
				"Identity (Soul & User)",
				"The Council Routing",
				"Save & Exit",
			},
		}
		survey.AskOne(prompt, &choice)

		switch {
		case strings.HasPrefix(choice, "Workspace"):
			configureWorkspace(cfg)
		case strings.HasPrefix(choice, "Model"):
			configureModels(cfg) // Returns a slice of unlocked models
		case strings.HasPrefix(choice, "Identity"):
			configureIdentity(cfg)
		case strings.HasPrefix(choice, "The Council"):
			configureCouncil(cfg)
		case strings.HasPrefix(choice, "Save"):
			if err := config.SaveConfig(configPath, cfg); err != nil {
				fmt.Printf("❌ Error saving config: %v\n", err)
			} else {
				createWorkspaceTemplates(cfg.WorkspacePath())
				fmt.Println("\n✓ Configuration saved securely to: " + configPath)
				fmt.Println("⚠️  IMPORTANT: If you are running `v1claw gateway` in the background (e.g. for Telegram),")
				fmt.Println("              you MUST restart the gateway for these changes to take effect.")
			}
			return
		}
	}
}

func configureWorkspace(cfg *config.Config) {
	fmt.Println("\n📁 Workspace & Security")

	cwd, _ := os.Getwd()
	defaultWorkspace := config.DefaultWorkspaceDir()

	workspaceOpts := []string{
		fmt.Sprintf("Default (%s)", defaultWorkspace),
		fmt.Sprintf("Current Directory (%s)", cwd),
		"Custom Path...",
	}

	workspaceChoice := ""
	promptWorkspace := &survey.Select{
		Message: "Where should V1Claw run its operations?",
		Options: workspaceOpts,
	}
	survey.AskOne(promptWorkspace, &workspaceChoice)

	if workspaceChoice == workspaceOpts[0] {
		cfg.Workspace.Path = defaultWorkspace
	} else if workspaceChoice == workspaceOpts[1] {
		cfg.Workspace.Path = cwd
	} else {
		customPath := ""
		promptCustom := &survey.Input{Message: "Enter absolute path to workspace:"}
		survey.AskOne(promptCustom, &customPath)
		if customPath != "" {
			customPath = filepath.Clean(customPath)
			if strings.HasPrefix(customPath, "~/") {
				home, _ := os.UserHomeDir()
				customPath = filepath.Join(home, customPath[2:])
			}
			cfg.Workspace.Path = customPath
		}
	}

	cfg.Agents.Defaults.Workspace = cfg.Workspace.Path

	securityChoice := ""
	promptSecurity := &survey.Select{
		Message: "Security Configuration: How much of your machine can I see?",
		Options: []string{
			fmt.Sprintf("Strict Sandbox (Recommended) - Restricted only to %s", cfg.Workspace.Path),
			"Global OS Access (Danger Zone) - I can read/edit ANY file on your disk.",
		},
	}
	survey.AskOne(promptSecurity, &securityChoice)

	if strings.HasPrefix(securityChoice, "Strict Sandbox") {
		cfg.Workspace.Sandboxed = true
		cfg.Agents.Defaults.RestrictToWorkspace = true
	} else {
		cfg.Workspace.Sandboxed = false
		cfg.Agents.Defaults.RestrictToWorkspace = false
	}
}

func configureModels(cfg *config.Config) {
	fmt.Println("\n🔧 AI Providers & Models")

	discoveredCLIs := config.DiscoverLocalCLIs()

	var options []string
	optionMap := make(map[string]string) // Label -> ID mapping

	// Explicitly separate CLI tools for UX
	if len(discoveredCLIs) > 0 {
		fmt.Printf("\n✨ Good news! I found %d local CLIs on your machine that I can tap into.\n", len(discoveredCLIs))
	}

	for _, tool := range discoveredCLIs {
		label := fmt.Sprintf("%s (Auto-Detected Local Tool)", tool.DisplayName)
		options = append(options, label)
		optionMap[label] = tool.ID
	}

	for _, p := range traditional {
		label := fmt.Sprintf("%s (%s)", p.name, p.desc)
		options = append(options, label)
		optionMap[label] = p.id
	}

	var selectedLabels []string
	promptSurvey := &survey.MultiSelect{
		Message: "Select the AI Providers you want to enable:",
		Options: options,
	}
	survey.AskOne(promptSurvey, &selectedLabels)

	var selectedIDs []string
	for _, label := range selectedLabels {
		selectedIDs = append(selectedIDs, optionMap[label])
	}

	var unlockedModels []string

	// Now ask for keys sequentially
	for _, providerID := range selectedIDs {
		// Is it a CLI?
		isCLI := false
		for _, tool := range discoveredCLIs {
			if tool.ID == providerID {
				isCLI = true
				fmt.Printf("\n✓ Registered %s (CLI Tool found, zero-key bridging enabled).\n", providerID)

				if cfg.Agents.Defaults.Provider == "" {
					cfg.Agents.Defaults.Provider = providerID
				}
				break
			}
		}

		if isCLI {
			continue
		}

		// It must be a Cloud Provider
		var pInfo providerInfo
		for _, p := range traditional {
			if p.id == providerID {
				pInfo = p
				break
			}
		}

		fmt.Printf("\n--- Key Required: %s ---\n", pInfo.name)
		fmt.Printf("Get a key here: %s\n", pInfo.keyURL)

		apiKey := ""
		prompt := &survey.Password{Message: fmt.Sprintf("Enter your %s:", pInfo.keyHint)}
		survey.AskOne(prompt, &apiKey)

		if apiKey != "" {
			setProviderKey(cfg, pInfo.id, apiKey)
			fmt.Println("✓ API key securely saved.")

			// Append the models tied to this newly unlocked cloud provider
			if models, ok := providerModels[pInfo.id]; ok {
				unlockedModels = append(unlockedModels, models...)
			}
		}
	}

	// Final Step: Model Selection
	if len(unlockedModels) > 0 {
		fmt.Println("\n🧠 Model Selection")
		unlockedModels = append(unlockedModels, "Custom override (Type your own)")

		modelChoice := ""
		promptModel := &survey.Select{
			Message: "Select your primary AI Model:",
			Options: unlockedModels,
		}
		survey.AskOne(promptModel, &modelChoice)

		if modelChoice == "Custom override (Type your own)" {
			customModel := ""
			survey.AskOne(&survey.Input{Message: "Type the exact model ID:"}, &customModel)
			if customModel != "" {
				cfg.Agents.Defaults.Model = customModel
			}
		} else if modelChoice != "" {
			cfg.Agents.Defaults.Model = modelChoice
			// Attempt to link it back to the provider logic for the default
			for pid, mList := range providerModels {
				for _, m := range mList {
					if m == modelChoice {
						cfg.Agents.Defaults.Provider = pid
					}
				}
			}
		}
	}
}

func configureIdentity(cfg *config.Config) {
	fmt.Println("\n🤖 Identity Configuration")
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("What is my name? [current: %s]: ", "V1Claw") // Wait, V1claw has no persistent Identity struct outside of memory docs. We will just init it later.
	var aiName, aiRole, userName, userPrefs string

	if scanner.Scan() {
		aiName = strings.TrimSpace(scanner.Text())
	}
	if aiName == "" {
		aiName = "V1Claw"
	}

	fmt.Print("What is my core purpose/role? [current: helpful personal assistant]: ")
	if scanner.Scan() {
		aiRole = strings.TrimSpace(scanner.Text())
	}
	if aiRole == "" {
		aiRole = "helpful personal assistant"
	}

	fmt.Println("\n👤 User Configuration")
	fmt.Print("What is your name? ")
	if scanner.Scan() {
		userName = strings.TrimSpace(scanner.Text())
	}

	fmt.Print("Any specific preferences I should learn right now? ")
	if scanner.Scan() {
		userPrefs = strings.TrimSpace(scanner.Text())
	}

	initMemory(cfg.WorkspacePath(), aiName, aiRole, userName, userPrefs)
	fmt.Println("✓ Identity synchronized to Memory.")
}

func configureCouncil(cfg *config.Config) {
	fmt.Println("\n🛡️ The Council Routing")

	persona := ""
	promptPersona := &survey.Select{
		Message: "How do you plan to use V1Claw the most?",
		Options: []string{
			"Software Engineer (High Code Accuracy, Multi-Agent)",
			"Writer / Researcher (High Context, Better Prose)",
			"General Assistant (Cost-Optimized, Fast)",
		},
	}
	survey.AskOne(promptPersona, &persona)

	cfg.Council.Enabled = true
	if strings.Contains(persona, "Software Engineer") {
		cfg.Council.Persona = "coder"
	} else if strings.Contains(persona, "Writer") {
		cfg.Council.Persona = "writer"
	} else {
		cfg.Council.Persona = "speed"
	}

	fmt.Println("✓ Routing Persona updated. (Note: Leader/Fallback model definitions are auto-handled by 'Model' settings)")
}
