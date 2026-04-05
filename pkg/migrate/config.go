package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/amit-vikramaditya/v1claw/pkg/config"
)

var supportedProviders = map[string]bool{
	"anthropic":      true,
	"openai":         true,
	"openrouter":     true,
	"groq":           true,
	"zhipu":          true,
	"vllm":           true,
	"gemini":         true,
	"deepseek":       true,
	"nvidia":         true,
	"moonshot":       true,
	"ollama":         true,
	"github_copilot": true,
}

var supportedChannels = map[string]bool{
	"telegram": true,
	"discord":  true,
	"whatsapp": true,
	"feishu":   true,
	"qq":       true,
	"dingtalk": true,
	"maixcam":  true,
	"slack":    true,
	"line":     true,
	"onebot":   true,
}

func findOpenClawConfig(openclawHome string) (string, error) {
	candidates := []string{
		filepath.Join(openclawHome, "openclaw.json"),
		filepath.Join(openclawHome, "config.json"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no config file found in %s (tried openclaw.json, config.json)", openclawHome)
}

func LoadOpenClawConfig(configPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("reading OpenClaw config: %w", err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parsing OpenClaw config: %w", err)
	}

	converted := convertKeysToSnake(raw)
	result, ok := converted.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected config format")
	}
	return result, nil
}

func ConvertConfig(data map[string]interface{}, openClawHome, v1ClawHome string) (*config.Config, []string, error) {
	cfg := config.DefaultConfig()
	var warnings []string

	if agents, ok := getMap(data, "agents"); ok {
		if defaults, ok := getMap(agents, "defaults"); ok {
			if v, ok := getString(defaults, "model"); ok {
				cfg.Agents.Defaults.Model = v
			}
			if v, ok := getFloat(defaults, "max_tokens"); ok {
				cfg.Agents.Defaults.MaxTokens = int(v)
			}
			if v, ok := getFloat(defaults, "temperature"); ok {
				cfg.Agents.Defaults.Temperature = v
			}
			if v, ok := getFloat(defaults, "max_tool_iterations"); ok {
				cfg.Agents.Defaults.MaxToolIterations = int(v)
			}
			if v, ok := getString(defaults, "workspace"); ok {
				cfg.Agents.Defaults.Workspace = rewriteWorkspacePath(v, openClawHome, v1ClawHome)
			}
		}
	}

	if providers, ok := getMap(data, "providers"); ok {
		for name, val := range providers {
			pMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}
			apiKey, _ := getString(pMap, "api_key")
			apiBase, _ := getString(pMap, "api_base")

			if !supportedProviders[name] {
				if apiKey != "" || apiBase != "" {
					warnings = append(warnings, fmt.Sprintf("Provider '%s' not supported in V1Claw, skipping", name))
				}
				continue
			}

			pc := config.ProviderConfig{APIKey: apiKey, APIBase: apiBase}
			switch name {
			case "anthropic":
				cfg.Providers.Anthropic = pc
			case "openai":
				cfg.Providers.OpenAI = pc
			case "openrouter":
				cfg.Providers.OpenRouter = pc
			case "groq":
				cfg.Providers.Groq = pc
			case "zhipu":
				cfg.Providers.Zhipu = pc
			case "vllm":
				cfg.Providers.VLLM = pc
			case "gemini":
				cfg.Providers.Gemini = pc
			case "deepseek":
				cfg.Providers.DeepSeek = pc
			case "nvidia":
				cfg.Providers.Nvidia = pc
			case "moonshot":
				cfg.Providers.Moonshot = pc
			case "ollama":
				cfg.Providers.Ollama = pc
			case "github_copilot":
				cfg.Providers.GitHubCopilot = pc
			}
		}
	}

	if channels, ok := getMap(data, "channels"); ok {
		for name, val := range channels {
			cMap, ok := val.(map[string]interface{})
			if !ok {
				continue
			}
			if !supportedChannels[name] {
				warnings = append(warnings, fmt.Sprintf("Channel '%s' not supported in V1Claw, skipping", name))
				continue
			}
			enabled, _ := getBool(cMap, "enabled")
			allowFrom := getStringSlice(cMap, "allow_from")

			switch name {
			case "telegram":
				cfg.Channels.Telegram.Enabled = enabled
				cfg.Channels.Telegram.AllowFrom = allowFrom
				if v, ok := getString(cMap, "token"); ok {
					cfg.Channels.Telegram.Token = v
				}
			case "whatsapp":
				cfg.Channels.WhatsApp.Enabled = enabled
				cfg.Channels.WhatsApp.AllowFrom = allowFrom
				if v, ok := getString(cMap, "bridge_url"); ok {
					cfg.Channels.WhatsApp.BridgeURL = v
				}
			case "discord", "feishu", "qq", "dingtalk", "slack", "line", "onebot", "maixcam":
				// Legacy channels — silently skip during migration
			}
		}
	}

	if gateway, ok := getMap(data, "gateway"); ok {
		if v, ok := getString(gateway, "host"); ok {
			cfg.Gateway.Host = v
		}
		if v, ok := getFloat(gateway, "port"); ok {
			cfg.Gateway.Port = int(v)
		}
	}

	if tools, ok := getMap(data, "tools"); ok {
		if web, ok := getMap(tools, "web"); ok {
			// Migrate old "search" config to "brave" if api_key is present
			if search, ok := getMap(web, "search"); ok {
				if v, ok := getString(search, "api_key"); ok {
					cfg.Tools.Web.Brave.APIKey = v
					if v != "" {
						cfg.Tools.Web.Brave.Enabled = true
					}
				}
				if v, ok := getFloat(search, "max_results"); ok {
					cfg.Tools.Web.Brave.MaxResults = int(v)
					cfg.Tools.Web.DuckDuckGo.MaxResults = int(v)
				}
			}
		}
	}

	return cfg, warnings, nil
}

func MergeConfig(existing, incoming *config.Config) *config.Config {
	if existing.Providers.Anthropic.APIKey == "" {
		existing.Providers.Anthropic = incoming.Providers.Anthropic
	}
	if existing.Providers.OpenAI.APIKey == "" {
		existing.Providers.OpenAI = incoming.Providers.OpenAI
	}
	if existing.Providers.OpenRouter.APIKey == "" {
		existing.Providers.OpenRouter = incoming.Providers.OpenRouter
	}
	if existing.Providers.Groq.APIKey == "" {
		existing.Providers.Groq = incoming.Providers.Groq
	}
	if existing.Providers.Zhipu.APIKey == "" {
		existing.Providers.Zhipu = incoming.Providers.Zhipu
	}
	if existing.Providers.VLLM.APIKey == "" && existing.Providers.VLLM.APIBase == "" {
		existing.Providers.VLLM = incoming.Providers.VLLM
	}
	if existing.Providers.Gemini.APIKey == "" {
		existing.Providers.Gemini = incoming.Providers.Gemini
	}

	if !existing.Channels.Telegram.Enabled && incoming.Channels.Telegram.Enabled {
		existing.Channels.Telegram = incoming.Channels.Telegram
	}
	if !existing.Channels.WhatsApp.Enabled && incoming.Channels.WhatsApp.Enabled {
		existing.Channels.WhatsApp = incoming.Channels.WhatsApp
	}

	if existing.Tools.Web.Brave.APIKey == "" {
		existing.Tools.Web.Brave = incoming.Tools.Web.Brave
	}

	return existing
}

func camelToSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				prev := rune(s[i-1])
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					result.WriteRune('_')
				} else if unicode.IsUpper(prev) && i+1 < len(s) && unicode.IsLower(rune(s[i+1])) {
					result.WriteRune('_')
				}
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func convertKeysToSnake(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, val := range v {
			result[camelToSnake(key)] = convertKeysToSnake(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = convertKeysToSnake(val)
		}
		return result
	default:
		return data
	}
}

func rewriteWorkspacePath(workspacePath, openClawHome, v1ClawHome string) string {
	if workspacePath == "" {
		return ""
	}

	if !isDefaultOpenClawWorkspacePath(workspacePath, openClawHome) {
		return workspacePath
	}

	targetHome := expandHome(v1ClawHome)
	if targetHome == "" {
		targetHome = config.HomeDir()
	}
	return filepath.Join(targetHome, "workspace")
}

func isDefaultOpenClawWorkspacePath(workspacePath, openClawHome string) bool {
	candidates := []string{
		"~/.openclaw/workspace",
		"~\\.openclaw\\workspace",
	}
	if openClawHome != "" {
		candidates = append(candidates, filepath.Join(openClawHome, "workspace"))
	}

	normalizedWorkspace := normalizeComparablePath(workspacePath)
	expandedWorkspace := normalizeComparablePath(expandHome(workspacePath))
	for _, candidate := range candidates {
		if normalizedWorkspace == normalizeComparablePath(candidate) {
			return true
		}
		if expandedWorkspace == normalizeComparablePath(expandHome(candidate)) {
			return true
		}
	}
	return false
}

func normalizeComparablePath(pathValue string) string {
	pathValue = strings.TrimSpace(pathValue)
	if pathValue == "" {
		return ""
	}
	return path.Clean(strings.ReplaceAll(pathValue, "\\", "/"))
}

func getMap(data map[string]interface{}, key string) (map[string]interface{}, bool) {
	v, ok := data[key]
	if !ok {
		return nil, false
	}
	m, ok := v.(map[string]interface{})
	return m, ok
}

func getString(data map[string]interface{}, key string) (string, bool) {
	v, ok := data[key]
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func getFloat(data map[string]interface{}, key string) (float64, bool) {
	v, ok := data[key]
	if !ok {
		return 0, false
	}
	f, ok := v.(float64)
	return f, ok
}

func getBool(data map[string]interface{}, key string) (bool, bool) {
	v, ok := data[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

func getStringSlice(data map[string]interface{}, key string) []string {
	v, ok := data[key]
	if !ok {
		return []string{}
	}
	arr, ok := v.([]interface{})
	if !ok {
		return []string{}
	}
	result := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}
