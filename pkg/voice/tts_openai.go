package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// OpenAITTS implements TTSProvider using the OpenAI-compatible TTS API.
// Works with OpenAI, Groq, and any OpenAI-compatible endpoint.
type OpenAITTS struct {
	apiKey     string
	apiBase    string
	model      string
	httpClient *http.Client
}

// OpenAITTSConfig holds configuration for the OpenAI TTS provider.
type OpenAITTSConfig struct {
	APIKey  string `json:"api_key"`
	APIBase string `json:"api_base"` // Default: "https://api.openai.com/v1"
	Model   string `json:"model"`    // Default: "tts-1"
}

// NewOpenAITTS creates a new OpenAI-compatible TTS provider.
func NewOpenAITTS(cfg OpenAITTSConfig) *OpenAITTS {
	if cfg.APIBase == "" {
		cfg.APIBase = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "tts-1"
	}
	return &OpenAITTS{
		apiKey:  cfg.APIKey,
		apiBase: cfg.APIBase,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (t *OpenAITTS) Name() string { return "openai" }

func (t *OpenAITTS) IsAvailable() bool { return t.apiKey != "" }

func (t *OpenAITTS) Synthesize(ctx context.Context, text string, opts TTSOptions) (io.ReadCloser, string, error) {
	voice := opts.Voice
	if voice == "" {
		voice = "alloy"
	}
	format := opts.Format
	if format == "" {
		format = "mp3"
	}
	speed := opts.Speed
	if speed <= 0 {
		speed = 1.0
	}

	reqBody := map[string]interface{}{
		"model":           t.model,
		"input":           text,
		"voice":           voice,
		"response_format": format,
		"speed":           speed,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("marshal request: %w", err)
	}

	url := t.apiBase + "/audio/speech"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	logger.DebugCF("tts", "OpenAI TTS request", map[string]interface{}{
		"voice": voice, "format": format, "text_len": len(text),
	})

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("TTS request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("TTS API error (status %d): %s", resp.StatusCode, string(errBody))
	}

	logger.DebugC("tts", fmt.Sprintf("OpenAI TTS: synthesized %d chars → %s", len(text), format))
	return resp.Body, format, nil
}
