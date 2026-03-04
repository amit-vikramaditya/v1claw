package voice

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
	"github.com/amit-vikramaditya/v1claw/pkg/logger"
	"github.com/amit-vikramaditya/v1claw/pkg/permissions"
)

// PipelineMode controls how the voice pipeline listens.
type PipelineMode string

const (
	ModeWakeWord   PipelineMode = "wake-word"
	ModePushToTalk PipelineMode = "push-to-talk"
	ModeAlwaysOn   PipelineMode = "always-on"
)

// PipelineConfig configures the voice pipeline.
type PipelineConfig struct {
	Mode           PipelineMode   `json:"mode"`            // "wake-word", "push-to-talk", "always-on"
	RecordDuration int            `json:"record_duration"` // Seconds per recording chunk (default: 5)
	SilenceTimeout int            `json:"silence_timeout"` // Seconds of silence before stopping (default: 2)
	SessionKey     string         `json:"session_key"`     // Session key for agent (default: "voice")
	WakeWord       WakeWordConfig `json:"wake_word"`
	Recorder       RecorderConfig `json:"recorder"`
	Player         PlayerConfig   `json:"player"`
	TTS            TTSConfig      `json:"tts"`
}

// Pipeline orchestrates: Mic → STT → Agent → TTS → Speaker.
type Pipeline struct {
	cfg         PipelineConfig
	recorder    AudioRecorder
	player      AudioPlayer
	transcriber *GroqTranscriber
	ttsManager  *TTSManager
	wakeWord    *WakeWordDetector
	msgBus      *bus.MessageBus

	mu      sync.Mutex
	running bool
	stopCh  chan struct{}
}

// NewPipeline creates a voice pipeline.
func NewPipeline(cfg PipelineConfig, msgBus *bus.MessageBus, transcriber *GroqTranscriber, ttsManager *TTSManager) *Pipeline {
	if cfg.RecordDuration <= 0 {
		cfg.RecordDuration = 5
	}
	if cfg.SilenceTimeout <= 0 {
		cfg.SilenceTimeout = 2
	}
	if cfg.SessionKey == "" {
		cfg.SessionKey = "voice"
	}
	if cfg.Mode == "" {
		cfg.Mode = ModeWakeWord
	}

	p := &Pipeline{
		cfg:         cfg,
		recorder:    NewRecorder(cfg.Recorder),
		player:      NewPlayer(cfg.Player),
		transcriber: transcriber,
		ttsManager:  ttsManager,
		msgBus:      msgBus,
		stopCh:      make(chan struct{}),
	}

	if cfg.Mode == ModeWakeWord {
		p.wakeWord = NewWakeWordDetector(cfg.WakeWord)
	}

	return p
}

// Start begins the voice pipeline loop.
func (p *Pipeline) Start(ctx context.Context) error {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return fmt.Errorf("voice pipeline already running")
	}
	p.running = true
	p.mu.Unlock()

	if !p.recorder.IsAvailable() {
		return fmt.Errorf("audio recorder (%s) not available", p.recorder.Name())
	}
	if err := permissions.Global().Check(permissions.Microphone, "voice.Pipeline"); err != nil {
		return fmt.Errorf("voice pipeline requires microphone permission: %w", err)
	}
	if p.transcriber == nil || !p.transcriber.IsAvailable() {
		return fmt.Errorf("speech-to-text transcriber not available")
	}

	logger.InfoCF("voice", "Voice pipeline started", map[string]interface{}{
		"mode": string(p.cfg.Mode), "recorder": p.recorder.Name(), "player": p.player.Name(),
	})

	// Subscribe to outbound messages for TTS.
	go p.handleOutbound(ctx)

	// Main recording loop.
	go p.recordLoop(ctx)

	return nil
}

// Stop halts the voice pipeline.
func (p *Pipeline) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.running {
		return
	}
	p.running = false
	close(p.stopCh)
	logger.InfoC("voice", "Voice pipeline stopped")
}

// IsRunning returns whether the pipeline is active.
func (p *Pipeline) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}

// recordLoop continuously records and transcribes audio.
func (p *Pipeline) recordLoop(ctx context.Context) {
	duration := time.Duration(p.cfg.RecordDuration) * time.Second

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		default:
		}

		// Recheck microphone permission each iteration
		if err := permissions.Global().Check(permissions.Microphone, "voice.recordLoop"); err != nil {
			logger.WarnC("voice", "Microphone permission revoked, stopping recording")
			return
		}

		filePath, err := p.recorder.Record(ctx, duration)
		if err != nil {
			logger.WarnCF("voice", "Recording failed", map[string]interface{}{"error": err.Error()})
			time.Sleep(1 * time.Second)
			continue
		}

		// Transcribe in background to not block recording.
		go p.processRecording(ctx, filePath)
	}
}

// processRecording transcribes audio, checks wake word, and sends to agent.
func (p *Pipeline) processRecording(ctx context.Context, filePath string) {
	defer os.Remove(filePath)

	result, err := p.transcriber.Transcribe(ctx, filePath)
	if err != nil {
		logger.WarnCF("voice", "Transcription failed", map[string]interface{}{"error": err.Error()})
		return
	}

	text := result.Text
	if text == "" {
		return
	}

	logger.DebugCF("voice", "Transcribed", map[string]interface{}{"text": text})

	// Wake word mode: only process if wake phrase detected.
	if p.cfg.Mode == ModeWakeWord && p.wakeWord != nil {
		command, detected := p.wakeWord.Process(ctx, text)
		if !detected {
			return
		}
		text = command
		if text == "" {
			// Wake word detected but no command — acknowledge and wait.
			logger.InfoC("voice", "Wake word detected, waiting for command...")
			p.speakResponse(ctx, "Yes?")
			return
		}
	}

	logger.InfoCF("voice", "Sending voice input to agent", map[string]interface{}{"text": text})

	// Publish to message bus as inbound message.
	p.msgBus.PublishInbound(bus.InboundMessage{
		Channel:    "voice",
		SenderID:   "voice-user",
		ChatID:     "voice",
		Content:    text,
		SessionKey: p.cfg.SessionKey,
		Metadata: map[string]string{
			"source": "microphone",
		},
	})
}

// handleOutbound listens for agent responses and speaks them.
func (p *Pipeline) handleOutbound(ctx context.Context) {
	sub := p.msgBus.SubscribeOutbound()
	defer sub.Unsubscribe()
	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case msg, ok := <-sub.C:
			if !ok {
				return
			}

			// Only speak responses targeted at the voice channel.
			if msg.Channel != "voice" {
				continue
			}

			// Speak asynchronously to prevent pipeline deafness to inbound events
			go p.speakResponse(ctx, msg.Content)
		}
	}
}

// speakResponse converts text to speech and plays it.
func (p *Pipeline) speakResponse(ctx context.Context, text string) {
	if text == "" {
		return
	}

	// Try TTS manager (OpenAI, Edge, etc.) first for high-quality voice.
	if p.ttsManager != nil && p.ttsManager.ProviderCount() > 0 {
		audioReader, format, err := p.ttsManager.Synthesize(ctx, text, DefaultTTSOptions())
		if err == nil {
			defer audioReader.Close()
			if p.player.IsAvailable() {
				if err := p.player.Play(ctx, audioReader, format); err != nil {
					logger.WarnCF("voice", "Audio playback failed", map[string]interface{}{"error": err.Error()})
				} else {
					return
				}
			}
		} else {
			logger.DebugCF("voice", "TTS synthesis failed, trying Termux speak", map[string]interface{}{"error": err.Error()})
		}
	}

	// Fallback: if on Termux, use native Android TTS directly.
	if isTermux() {
		if err := permissions.Global().Check(permissions.ShellHardware, "voice.speakResponse"); err == nil {
			if _, err := exec.LookPath("termux-tts-speak"); err == nil {
				cmd := exec.CommandContext(ctx, "termux-tts-speak", text)
				if err := cmd.Run(); err != nil {
					logger.WarnCF("voice", "termux-tts-speak failed", map[string]interface{}{"error": err.Error()})
				}
				return
			}
		}
	}

	// Last resort: just log the response.
	logger.InfoCF("voice", "No audio output available, text response", map[string]interface{}{"text": text})
}

// SpeakText is a convenience method to speak text directly (for notifications, etc.).
func (p *Pipeline) SpeakText(ctx context.Context, text string) {
	p.speakResponse(ctx, text)
}

// SetTTSManager allows updating the TTS manager after creation.
func (p *Pipeline) SetTTSManager(m *TTSManager) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ttsManager = m
}

// exec is imported for the Termux fallback in speakResponse.
// The import is at the top of the file via os/exec.
// This helper drains an io.ReadCloser (used when we need to discard TTS output).
func drainAndClose(rc io.ReadCloser) {
	if rc != nil {
		io.Copy(io.Discard, rc)
		rc.Close()
	}
}
