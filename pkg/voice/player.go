package voice

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// AudioPlayer plays audio through speakers.
type AudioPlayer interface {
	// Play plays audio data in the given format ("mp3", "wav", "opus").
	Play(ctx context.Context, audio io.Reader, format string) error
	// PlayFile plays an audio file.
	PlayFile(ctx context.Context, filePath string) error
	// IsAvailable returns true if the player can output audio.
	IsAvailable() bool
	// Name returns the player name.
	Name() string
}

// PlayerConfig configures audio playback.
type PlayerConfig struct {
	Backend string `json:"backend"` // "termux", "system", "auto"
}

// NewPlayer creates an AudioPlayer based on config.
func NewPlayer(cfg PlayerConfig) AudioPlayer {
	switch cfg.Backend {
	case "termux":
		return &termuxPlayer{}
	case "system":
		return &systemPlayer{}
	default: // "auto"
		if isTermux() {
			logger.InfoC("voice", "Auto-detected Termux — using termux-tts-speak for playback")
			return &termuxPlayer{}
		}
		logger.InfoC("voice", "Using system audio player")
		return &systemPlayer{}
	}
}

// --- Termux Player ---
// Uses termux-media-player for audio files, termux-tts-speak for direct text.

type termuxPlayer struct{}

func (p *termuxPlayer) Name() string { return "termux" }

func (p *termuxPlayer) IsAvailable() bool {
	_, err := exec.LookPath("termux-media-player")
	return err == nil
}

func (p *termuxPlayer) Play(ctx context.Context, audio io.Reader, format string) error {
	// Write audio to temp file, then play.
	tmpFile, err := os.CreateTemp("", "v1claw-play-*."+format)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, audio); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write audio: %w", err)
	}
	tmpFile.Close()

	return p.PlayFile(ctx, tmpFile.Name())
}

func (p *termuxPlayer) PlayFile(ctx context.Context, filePath string) error {
	logger.InfoCF("voice", "Playing audio via Termux", map[string]interface{}{"file": filePath})
	cmd := exec.CommandContext(ctx, "termux-media-player", "play", filePath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("termux-media-player: %w (output: %s)", err, string(out))
	}
	return nil
}

// --- System Player (Linux: aplay/ffplay, macOS: afplay) ---

type systemPlayer struct{}

func (p *systemPlayer) Name() string { return "system" }

func (p *systemPlayer) IsAvailable() bool {
	cmd := p.playerCommand()
	if cmd == "" {
		return false
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}

func (p *systemPlayer) playerCommand() string {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.LookPath("ffplay"); err == nil {
			return "ffplay"
		}
		return "aplay"
	case "darwin":
		return "afplay"
	default:
		return ""
	}
}

func (p *systemPlayer) Play(ctx context.Context, audio io.Reader, format string) error {
	tmpFile, err := os.CreateTemp("", "v1claw-play-*."+format)
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, audio); err != nil {
		tmpFile.Close()
		return fmt.Errorf("write audio: %w", err)
	}
	tmpFile.Close()

	return p.PlayFile(ctx, tmpFile.Name())
}

func (p *systemPlayer) PlayFile(ctx context.Context, filePath string) error {
	cmdName := p.playerCommand()
	if cmdName == "" {
		return fmt.Errorf("no audio player available for %s", runtime.GOOS)
	}

	logger.InfoCF("voice", "Playing audio via system", map[string]interface{}{
		"command": cmdName, "file": filePath,
	})

	var args []string
	switch cmdName {
	case "ffplay":
		args = []string{"-nodisp", "-autoexit", "-loglevel", "quiet", filePath}
	case "afplay":
		args = []string{filePath}
	case "aplay":
		args = []string{filePath}
	default:
		args = []string{filePath}
	}

	cmd := exec.CommandContext(ctx, cmdName, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w (output: %s)", cmdName, err, string(out))
	}
	return nil
}
