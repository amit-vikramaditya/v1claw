package knowledge

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

// OpenAIEmbedder implements EmbeddingProvider using the OpenAI-compatible embeddings API.
// Works with OpenAI, Groq, Ollama, and any compatible endpoint.
type OpenAIEmbedder struct {
	apiKey     string
	apiBase    string
	model      string
	dims       int
	httpClient *http.Client
}

// OpenAIEmbedderConfig configures the OpenAI embeddings provider.
type OpenAIEmbedderConfig struct {
	APIKey     string `json:"api_key"`
	APIBase    string `json:"api_base"`    // Default: "https://api.openai.com/v1"
	Model      string `json:"model"`       // Default: "text-embedding-3-small"
	Dimensions int    `json:"dimensions"`  // Default: 1536
}

// NewOpenAIEmbedder creates a new OpenAI-compatible embedding provider.
func NewOpenAIEmbedder(cfg OpenAIEmbedderConfig) *OpenAIEmbedder {
	if cfg.APIBase == "" {
		cfg.APIBase = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "text-embedding-3-small"
	}
	if cfg.Dimensions <= 0 {
		cfg.Dimensions = 1536
	}
	return &OpenAIEmbedder{
		apiKey:  cfg.APIKey,
		apiBase: cfg.APIBase,
		model:   cfg.Model,
		dims:    cfg.Dimensions,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *OpenAIEmbedder) Name() string       { return "openai" }
func (e *OpenAIEmbedder) Dimensions() int     { return e.dims }

func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float64, error) {
	results, err := e.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (e *OpenAIEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := map[string]interface{}{
		"model": e.model,
		"input": texts,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	url := e.apiBase + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("embeddings request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embeddings API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	embeddings := make([][]float64, len(texts))
	for _, d := range result.Data {
		if d.Index < len(embeddings) {
			embeddings[d.Index] = d.Embedding
		}
	}

	logger.DebugCF("knowledge", "Embeddings generated", map[string]interface{}{
		"count": len(texts), "dims": len(embeddings[0]),
	})

	return embeddings, nil
}
