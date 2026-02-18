package voice

import (
	"os"
	"testing"
)

func TestNewRecorderAutoDetect(t *testing.T) {
	r := NewRecorder(RecorderConfig{Backend: "auto"})
	if r == nil {
		t.Fatal("NewRecorder returned nil")
	}

	if os.Getenv("TERMUX_VERSION") != "" {
		if r.Name() != "termux" {
			t.Errorf("expected termux recorder in Termux, got %s", r.Name())
		}
	} else {
		if r.Name() != "system" {
			t.Errorf("expected system recorder outside Termux, got %s", r.Name())
		}
	}
}

func TestNewRecorderExplicit(t *testing.T) {
	r := NewRecorder(RecorderConfig{Backend: "termux"})
	if r.Name() != "termux" {
		t.Errorf("expected termux, got %s", r.Name())
	}

	r = NewRecorder(RecorderConfig{Backend: "system"})
	if r.Name() != "system" {
		t.Errorf("expected system, got %s", r.Name())
	}
}

func TestRecorderConfigDefaults(t *testing.T) {
	r := NewRecorder(RecorderConfig{})
	if r == nil {
		t.Fatal("NewRecorder returned nil")
	}
}

func TestIsTermux(t *testing.T) {
	// Just verify it doesn't panic; actual value depends on environment.
	_ = isTermux()
}

func TestNewPlayerAutoDetect(t *testing.T) {
	p := NewPlayer(PlayerConfig{Backend: "auto"})
	if p == nil {
		t.Fatal("NewPlayer returned nil")
	}

	if os.Getenv("TERMUX_VERSION") != "" {
		if p.Name() != "termux" {
			t.Errorf("expected termux player in Termux, got %s", p.Name())
		}
	} else {
		if p.Name() != "system" {
			t.Errorf("expected system player outside Termux, got %s", p.Name())
		}
	}
}

func TestNewPlayerExplicit(t *testing.T) {
	p := NewPlayer(PlayerConfig{Backend: "termux"})
	if p.Name() != "termux" {
		t.Errorf("expected termux, got %s", p.Name())
	}

	p = NewPlayer(PlayerConfig{Backend: "system"})
	if p.Name() != "system" {
		t.Errorf("expected system, got %s", p.Name())
	}
}
