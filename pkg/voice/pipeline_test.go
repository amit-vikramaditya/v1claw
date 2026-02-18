package voice

import (
	"context"
	"testing"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/bus"
)

func TestPipelineCreation(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	p := NewPipeline(PipelineConfig{}, msgBus, nil, nil)
	if p == nil {
		t.Fatal("NewPipeline returned nil")
	}
	if p.cfg.Mode != ModeWakeWord {
		t.Errorf("default mode = %s, want wake-word", p.cfg.Mode)
	}
	if p.cfg.RecordDuration != 5 {
		t.Errorf("default record duration = %d, want 5", p.cfg.RecordDuration)
	}
	if p.cfg.SessionKey != "voice" {
		t.Errorf("default session key = %s, want voice", p.cfg.SessionKey)
	}
}

func TestPipelineStartWithoutTranscriber(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	p := NewPipeline(PipelineConfig{
		Recorder: RecorderConfig{Backend: "system"},
	}, msgBus, nil, nil)

	err := p.Start(context.Background())
	if err == nil {
		t.Error("expected error when starting without transcriber")
		p.Stop()
	}
}

func TestPipelineDoubleStart(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	transcriber := &GroqTranscriber{apiKey: "test-key"}
	p := NewPipeline(PipelineConfig{
		Recorder: RecorderConfig{Backend: "system"},
	}, msgBus, transcriber, nil)

	// Skip if recorder not available (CI).
	if !p.recorder.IsAvailable() {
		t.Skip("audio recorder not available")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := p.Start(ctx); err != nil {
		t.Skipf("cannot start pipeline: %v", err)
	}
	defer p.Stop()

	err := p.Start(ctx)
	if err == nil {
		t.Error("expected error on double start")
	}
}

func TestPipelineStopIdempotent(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	p := NewPipeline(PipelineConfig{}, msgBus, nil, nil)
	// Should not panic when stopping a pipeline that was never started.
	p.Stop()
	p.Stop()
}

func TestPipelineIsRunning(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	p := NewPipeline(PipelineConfig{}, msgBus, nil, nil)
	if p.IsRunning() {
		t.Error("pipeline should not be running before start")
	}
}

func TestPipelineModes(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	tests := []struct {
		mode     PipelineMode
		hasWake  bool
	}{
		{ModeWakeWord, true},
		{ModePushToTalk, false},
		{ModeAlwaysOn, false},
	}

	for _, tt := range tests {
		p := NewPipeline(PipelineConfig{Mode: tt.mode}, msgBus, nil, nil)
		if (p.wakeWord != nil) != tt.hasWake {
			t.Errorf("mode %s: wakeWord detector present = %v, want %v",
				tt.mode, p.wakeWord != nil, tt.hasWake)
		}
	}
}

func TestSpeakResponseFallback(t *testing.T) {
	msgBus := bus.NewMessageBus()
	defer msgBus.Close()

	p := NewPipeline(PipelineConfig{}, msgBus, nil, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Should not panic even with no TTS or player available.
	p.speakResponse(ctx, "Hello world")
	p.speakResponse(ctx, "")
}

func TestDrainAndClose(t *testing.T) {
	// Should not panic with nil.
	drainAndClose(nil)
}
