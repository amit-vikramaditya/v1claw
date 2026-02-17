package voice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// WakeWordHandler is called when a wake phrase is detected.
type WakeWordHandler func(ctx context.Context, phrase string, confidence float64)

// WakeWordDetector listens for wake phrases in transcribed text.
// This is a software-based detector that works with any STT provider.
// For hardware wake word detection (Porcupine, OpenWakeWord), a separate
// adapter can be plugged in via the WakeWordEngine interface.
type WakeWordDetector struct {
	phrases   []string
	handler   WakeWordHandler
	cooldown  time.Duration
	lastWake  time.Time
}

// WakeWordConfig configures the wake word detector.
type WakeWordConfig struct {
	Enabled  bool     `json:"enabled"`
	Phrases  []string `json:"phrases"`  // e.g., ["hello v1", "hey v1", "hi v1"]
	Cooldown int      `json:"cooldown"` // Seconds between activations (default 3)
}

// NewWakeWordDetector creates a wake word detector with the given phrases.
func NewWakeWordDetector(cfg WakeWordConfig) *WakeWordDetector {
	phrases := cfg.Phrases
	if len(phrases) == 0 {
		phrases = []string{"hello v1", "hey v1", "hi v1"}
	}
	// Normalize all to lowercase.
	for i, p := range phrases {
		phrases[i] = strings.ToLower(strings.TrimSpace(p))
	}

	cooldown := time.Duration(cfg.Cooldown) * time.Second
	if cooldown <= 0 {
		cooldown = 3 * time.Second
	}

	logger.InfoCF("wakeword", "Wake word detector initialized", map[string]interface{}{
		"phrases": phrases, "cooldown_sec": cooldown.Seconds(),
	})

	return &WakeWordDetector{
		phrases:  phrases,
		cooldown: cooldown,
	}
}

// SetHandler sets the callback for wake word detection.
func (d *WakeWordDetector) SetHandler(handler WakeWordHandler) {
	d.handler = handler
}

// Check tests if the given text contains a wake phrase.
// Returns the matched phrase and the remaining command text after the phrase.
// If no wake phrase is found, returns ("", text, false).
func (d *WakeWordDetector) Check(text string) (phrase string, command string, detected bool) {
	lower := strings.ToLower(strings.TrimSpace(text))

	for _, p := range d.phrases {
		if strings.HasPrefix(lower, p) {
			// Extract command after wake phrase.
			rest := strings.TrimSpace(text[len(p):])
			// Strip punctuation between wake phrase and command.
			rest = strings.TrimLeft(rest, ",. !?")
			rest = strings.TrimSpace(rest)
			return p, rest, true
		}
		// Also check if wake phrase appears anywhere (for noisy STT).
		idx := strings.Index(lower, p)
		if idx >= 0 {
			rest := strings.TrimSpace(text[idx+len(p):])
			rest = strings.TrimLeft(rest, ",. !?")
			rest = strings.TrimSpace(rest)
			return p, rest, true
		}
	}

	return "", text, false
}

// Process checks text and fires the handler if a wake phrase is detected.
// Respects the cooldown to avoid rapid re-triggering.
func (d *WakeWordDetector) Process(ctx context.Context, text string) (string, bool) {
	phrase, command, detected := d.Check(text)
	if !detected {
		return text, false
	}

	// Respect cooldown.
	now := time.Now()
	if now.Sub(d.lastWake) < d.cooldown {
		return command, true
	}
	d.lastWake = now

	logger.InfoC("wakeword", fmt.Sprintf("Wake phrase detected: %q → command: %q", phrase, command))

	if d.handler != nil {
		d.handler(ctx, phrase, 1.0)
	}

	return command, true
}

// WakeWordEngine is an interface for hardware-based wake word detectors
// (Porcupine, OpenWakeWord, etc.) that can be plugged in later.
type WakeWordEngine interface {
	// Start begins listening for wake words on the audio stream.
	Start(ctx context.Context) error
	// Stop halts the engine.
	Stop()
	// SetHandler sets the callback.
	SetHandler(handler WakeWordHandler)
	// Name returns the engine name.
	Name() string
}
