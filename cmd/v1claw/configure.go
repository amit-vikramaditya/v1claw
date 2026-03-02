package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			MarginTop(1).
			MarginBottom(1)

	cyanStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	grayStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	redStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	greenStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)
)

func printLogo() {
	logoText := `
██░▄▄▄░██░▄▄░██░▄▄▄██░▀██░██░▄▄▀██░████░▄▄▀██░███░██
██░███░██░▀▀░██░▄▄▄██░█░█░██░█████░████░▀▀░██░█░█░██
██░▀▀▀░██░█████░▀▀▀██░██▄░██░▀▀▄██░▀▀░█░██░██▄▀▄▀▄██
`
	fmt.Println(cyanStyle.Render(logoText))
	fmt.Println(grayStyle.Render(" V1Claw - The Local-First Epistemic Engine"))
}

func printCurrentState(cfg *config.Config) {
	workspace := cfg.WorkspacePath()
	if workspace == "" {
		workspace = redStyle.Render("Not Set")
	}

	security := greenStyle.Render("Locked (Safe)")
	if !cfg.Workspace.Sandboxed {
		security = redStyle.Render("Unlocked (Danger)")
	}

	brain := cyanStyle.Render(cfg.Agents.Defaults.Model)
	if brain == "" {
		brain = redStyle.Render("Not Configured")
	}

	channels := "None"
	var activeChannels []string
	if cfg.Channels.Telegram.Enabled {
		activeChannels = append(activeChannels, "Telegram")
	}
	if cfg.Channels.Discord.Enabled {
		activeChannels = append(activeChannels, "Discord")
	}
	if len(activeChannels) > 0 {
		channels = strings.Join(activeChannels, ", ")
	}

	stateContent := fmt.Sprintf("┌  Current System State\n│\n│  Home: %s\n│  Security: %s\n│  Brain: %s\n│  Channels: %s\n└",
		workspace, security, brain, channels)

	fmt.Println(borderStyle.Render(stateContent))
}

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

	if _, err := os.Stat(configPath); err == nil {
		if loaded, err := loadConfig(); err == nil {
			cfg = loaded
		}
	}

	for {
		printLogo()
		printCurrentState(cfg)

		var choice string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to configure?").
					Options(
						huh.NewOption("🏠 The Home (Workspace & Security)", "workspace").Description("Where the AI lives and what files it can see"),
						huh.NewOption("🧠 The Brain (Providers & Council)", "models").Description("Connect Gemini, Claude, or local CLIs"),
						huh.NewOption("🧰 Tools & Skills (Search, Cron)", "tools").Description("Give your AI abilities to browse and run tasks"),
						huh.NewOption("📡 Channels (Telegram, Discord)", "channels").Description("Where you chat with your AI"),
						huh.NewOption("🧬 Identity (Soul & User)", "identity").Description("Define personalities and core instructions"),
						huh.NewOption("💾 Save & Exit", "save").Description("Commit changes and finish"),
					).
					Value(&choice),
			),
		)

		err := form.Run()
		if err != nil {
			fmt.Println("Configuration cancelled.")
			return
		}

		switch choice {
		case "workspace":
			configureWorkspace(cfg)
		case "models":
			configureModels(cfg)
		case "tools":
			configureTools(cfg)
		case "channels":
			configureChannels(cfg)
		case "identity":
			configureIdentity(cfg)
		case "save":
			if err := config.SaveConfig(configPath, cfg); err != nil {
				fmt.Printf("❌ Error saving config: %v\n", redStyle.Render(err.Error()))
			} else {
				createWorkspaceTemplates(cfg.WorkspacePath())
				fmt.Println(greenStyle.Bold(true).Render("\n✓ Configuration saved securely to: ") + configPath)
				fmt.Println(grayStyle.Render("⚠️  IMPORTANT: You MUST restart any running v1claw gateway for changes to take effect."))
			}
			return
		}
	}
}

func configureWorkspace(cfg *config.Config) {
	fmt.Println(headerStyle.Render("┌  The Home (Workspace Configuration)"))

	cwd, _ := os.Getwd()
	defaultWorkspace := config.DefaultWorkspaceDir()

	var locationChoice string
	var customPath string
	var securityChoice string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Where should the AI store its memory and files?").
				Options(
					huh.NewOption(fmt.Sprintf("Default (%s)", defaultWorkspace), "default"),
					huh.NewOption(fmt.Sprintf("Current Folder (%s)", cwd), "current"),
					huh.NewOption("Custom Path...", "custom"),
				).
				Value(&locationChoice),
		),
	).WithShowHelp(false)

	if err := form.Run(); err != nil {
		return
	}

	if locationChoice == "custom" {
		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Enter absolute path to workspace").
					Placeholder("/Users/name/my-ai-workspace").
					Value(&customPath),
			),
		).Run()
	}

	// Update location
	switch locationChoice {
	case "default":
		cfg.Workspace.Path = defaultWorkspace
	case "current":
		cfg.Workspace.Path = cwd
	case "custom":
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

	huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("File Access Permissions").
				Description("Choose how much of your computer the AI can see.").
				Options(
					huh.NewOption("Locked (Safe) — AI is restricted to its workspace.", "locked"),
					huh.NewOption("Unlocked (Danger) — AI can read/edit ANY file on your machine.", "unlocked"),
				).
				Value(&securityChoice),
		),
	).Run()

	if securityChoice == "locked" {
		cfg.Workspace.Sandboxed = true
		cfg.Agents.Defaults.RestrictToWorkspace = true
	} else {
		cfg.Workspace.Sandboxed = false
		cfg.Agents.Defaults.RestrictToWorkspace = false
	}
}

func configureModels(cfg *config.Config) {
	fmt.Println(headerStyle.Render("┌  The Brain (AI Providers & Models)"))

	discoveredCLIs := config.DiscoverLocalCLIs()
	var providerOptions []huh.Option[string]
	optionMap := make(map[string]string)

	for _, tool := range discoveredCLIs {
		label := fmt.Sprintf("%s (Local)", tool.DisplayName)
		providerOptions = append(providerOptions, huh.NewOption(label, tool.ID).Description(tool.Description))
		optionMap[tool.ID] = label
	}

	for _, p := range traditional {
		providerOptions = append(providerOptions, huh.NewOption(p.name, p.id).Description(p.desc))
		optionMap[p.id] = p.name
	}

	var selectedIDs []string
	huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select the AI Providers you want to enable").
				Description("Choose one or more brains for your AI.").
				Options(providerOptions...).
				Value(&selectedIDs),
		),
	).Run()

	var unlockedModels []huh.Option[string]

	for _, providerID := range selectedIDs {
		isCLI := false
		for _, tool := range discoveredCLIs {
			if tool.ID == providerID {
				isCLI = true
				if cfg.Agents.Defaults.Provider == "" {
					cfg.Agents.Defaults.Provider = providerID
				}
				break
			}
		}
		if isCLI {
			continue
		}

		var pInfo providerInfo
		for _, p := range traditional {
			if p.id == providerID {
				pInfo = p
				break
			}
		}

		var apiKey string
		fmt.Printf("\n--- Key Required: %s ---\n", cyanStyle.Render(pInfo.name))
		fmt.Printf("Get a key here: %s\n", grayStyle.Render(pInfo.keyURL))

		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Enter your %s", pInfo.keyHint)).
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			),
		).Run()

		if apiKey != "" {
			setProviderKey(cfg, pInfo.id, apiKey)
			if models, ok := providerModels[pInfo.id]; ok {
				for _, m := range models {
					unlockedModels = append(unlockedModels, huh.NewOption(m, m))
				}
			}
		}
	}

	if len(unlockedModels) > 0 {
		var modelChoice string
		unlockedModels = append(unlockedModels, huh.NewOption("Custom override...", "custom"))

		huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Select your primary AI Model").
					Options(unlockedModels...).
					Value(&modelChoice),
			),
		).Run()

		if modelChoice == "custom" {
			var customModel string
			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Type the exact model ID").Value(&customModel),
				),
			).Run()
			if customModel != "" {
				cfg.Agents.Defaults.Model = customModel
			}
		} else if modelChoice != "" {
			cfg.Agents.Defaults.Model = modelChoice
			for pid, mList := range providerModels {
				for _, m := range mList {
					if m == modelChoice {
						cfg.Agents.Defaults.Provider = pid
					}
				}
			}
		}
	}

	// Dynamic Council Question
	var enableCouncil bool
	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("High Availability (The Council)").
				Description("Should we automatically switch to a backup AI if the primary fails?").
				Value(&enableCouncil),
		),
	).Run()

	if enableCouncil {
		configureCouncil(cfg)
	} else {
		cfg.Council.Enabled = false
	}
}

func configureTools(cfg *config.Config) {
	fmt.Println(headerStyle.Render("┌  Tools & Autonomous Skills"))

	var selectedTools []string
	huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select tools to equip").
				Description("Give your AI abilities to interact with the world.").
				Options(
					huh.NewOption("DuckDuckGo Search", "ddg").Description("Free web search, no key required"),
					huh.NewOption("Tavily Search", "tavily").Description("Premium research tool, requires Tavily API key"),
					huh.NewOption("Academic Literature", "academic").Description("Search peer-reviewed papers (Consensus feature)"),
					huh.NewOption("Terminal Access", "shell").Description("Allow AI to run bash commands in its workspace"),
					huh.NewOption("File System", "fs").Description("Allow AI to read and write files"),
				).
				Value(&selectedTools),
		),
	).Run()

	// Map tools to config
	for _, t := range selectedTools {
		switch t {
		case "ddg":
			cfg.Tools.Web.DuckDuckGo.Enabled = true
		case "tavily":
			cfg.Tools.Web.Brave.Enabled = true // Re-using brave slot for now as placeholder or we update config struct
		case "shell":
			cfg.Permissions.ShellHardware = true
		}
	}

	var enableCron bool
	huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Autonomous Background Thinking?").
				Description("The AI will wake up on a schedule to research topics on its own.").
				Value(&enableCron),
		),
	).Run()

	if enableCron {
		var schedule string
		huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Background Schedule").
					Options(
						huh.NewOption("Every 12 hours (Morning & Evening)", "720"),
						huh.NewOption("Every 1 hour", "60"),
					).
					Value(&schedule),
			),
		).Run()
		// TODO: Map schedule to cfg.Heartbeat.Interval
	}
}

func configureChannels(cfg *config.Config) {
	fmt.Println(headerStyle.Render("┌  Channels (Where to talk)"))

	var selectedChannels []string
	huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select communication channels").
				Options(
					huh.NewOption("Telegram", "telegram").Description("Talk to your AI via Telegram bot"),
					huh.NewOption("Discord", "discord").Description("Connect to your Discord server"),
					huh.NewOption("Slack", "slack").Description("Integrate with your Slack workspace"),
				).
				Value(&selectedChannels),
		),
	).Run()

	for _, ch := range selectedChannels {
		if ch == "telegram" {
			var token string
			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Enter Telegram Bot Token").
						Description("Get this from @BotFather").
						Value(&token),
				),
			).Run()
			if token != "" {
				cfg.Channels.Telegram.Enabled = true
				cfg.Channels.Telegram.Token = token
			}
		}
		// Repeat for other channels...
	}
}

func configureIdentity(cfg *config.Config) {
	fmt.Println(headerStyle.Render("┌  Identity (Who are we?)"))

	var aiName string = "V1Claw"
	var aiRole string = "helpful personal assistant"
	var userName string
	var userPrefs string

	huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("What is my name?").Value(&aiName),
			huh.NewInput().Title("What is my core purpose?").Value(&aiRole),
			huh.NewInput().Title("What is your name?").Value(&userName),
			huh.NewInput().Title("Any specific preferences?").Value(&userPrefs),
		),
	).Run()

	initMemory(cfg.WorkspacePath(), aiName, aiRole, userName, userPrefs)
}

func configureCouncil(cfg *config.Config) {
	var persona string
	huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Routing Persona").
				Description("How should the AI behave during complex tasks?").
				Options(
					huh.NewOption("Software Engineer", "coder").Description("High code accuracy, multi-agent focus"),
					huh.NewOption("Writer / Researcher", "writer").Description("High context, better prose"),
					huh.NewOption("General Assistant", "speed").Description("Cost-optimized, fast response"),
				).
				Value(&persona),
		),
	).Run()

	cfg.Council.Enabled = true
	cfg.Council.Persona = persona
}
