package voice

import (
	"context"
	"fmt"
	"io"
)

// TTSProvider synthesizes speech from text.
type TTSProvider interface {
	// Synthesize converts text to audio and writes to w.
	// Returns the audio format (e.g., "mp3", "wav", "opus").
	Synthesize(ctx context.Context, text string, opts TTSOptions) (io.ReadCloser, string, error)
	// Name returns the provider name.
	Name() string
	// IsAvailable returns true if the provider is configured and ready.
	IsAvailable() bool
}

// TTSOptions controls speech synthesis.
type TTSOptions struct {
	Voice  string  // Voice identifier (provider-specific).
	Speed  float64 // Speech rate multiplier (1.0 = normal).
	Pitch  float64 // Pitch adjustment (-1.0 to 1.0).
	Format string  // Output format: "mp3", "wav", "opus". Empty = provider default.
}

// DefaultTTSOptions returns sensible defaults.
func DefaultTTSOptions() TTSOptions {
	return TTSOptions{
		Speed:  1.0,
		Pitch:  0.0,
		Format: "mp3",
	}
}

// TTSConfig holds TTS configuration.
type TTSConfig struct {
	Enabled    bool   `json:"enabled"`
	Provider   string `json:"provider"`   // "edge", "openai", "piper"
	Voice      string `json:"voice"`      // Default voice
	Speed      float64 `json:"speed"`
	CacheDir   string `json:"cache_dir"`  // Cache synthesized audio
}

// TTSManager manages TTS providers with fallback.
type TTSManager struct {
	providers map[string]TTSProvider
	primary   string
}

// NewTTSManager creates a TTS manager.
func NewTTSManager() *TTSManager {
	return &TTSManager{
		providers: make(map[string]TTSProvider),
	}
}

// Register adds a TTS provider.
func (m *TTSManager) Register(provider TTSProvider) {
	m.providers[provider.Name()] = provider
	if m.primary == "" {
		m.primary = provider.Name()
	}
}

// SetPrimary sets the primary provider by name.
func (m *TTSManager) SetPrimary(name string) {
	if _, ok := m.providers[name]; ok {
		m.primary = name
	}
}

// Synthesize uses the primary provider, falling back to others on failure.
func (m *TTSManager) Synthesize(ctx context.Context, text string, opts TTSOptions) (io.ReadCloser, string, error) {
	// Try primary first.
	if p, ok := m.providers[m.primary]; ok && p.IsAvailable() {
		reader, format, err := p.Synthesize(ctx, text, opts)
		if err == nil {
			return reader, format, nil
		}
	}

	// Fallback to any available provider.
	for name, p := range m.providers {
		if name == m.primary || !p.IsAvailable() {
			continue
		}
		reader, format, err := p.Synthesize(ctx, text, opts)
		if err == nil {
			return reader, format, nil
		}
	}

	return nil, "", fmt.Errorf("no TTS provider available")
}

// ProviderCount returns number of registered providers.
func (m *TTSManager) ProviderCount() int {
	return len(m.providers)
}
