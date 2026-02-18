package voice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// AudioRecorder captures audio from a microphone.
type AudioRecorder interface {
	// Record captures audio for the given duration and returns the file path.
	// The caller is responsible for removing the file when done.
	Record(ctx context.Context, duration time.Duration) (filePath string, err error)
	// IsAvailable returns true if the recorder can capture audio.
	IsAvailable() bool
	// Name returns the recorder name.
	Name() string
}

// RecorderConfig configures audio recording.
type RecorderConfig struct {
	Backend    string `json:"backend"`     // "termux", "system", "auto"
	SampleRate int    `json:"sample_rate"` // Default: 16000
	Format     string `json:"format"`      // "wav", "ogg"; default "wav"
	TempDir    string `json:"temp_dir"`    // Directory for temp audio files
}

// NewRecorder creates an AudioRecorder based on config.
// "auto" detects Termux, then falls back to system recorder.
func NewRecorder(cfg RecorderConfig) AudioRecorder {
	if cfg.SampleRate <= 0 {
		cfg.SampleRate = 16000
	}
	if cfg.Format == "" {
		cfg.Format = "wav"
	}
	if cfg.TempDir == "" {
		cfg.TempDir = os.TempDir()
	}

	switch cfg.Backend {
	case "termux":
		return &termuxRecorder{cfg: cfg}
	case "system":
		return &systemRecorder{cfg: cfg}
	default: // "auto"
		if isTermux() {
			logger.InfoC("voice", "Auto-detected Termux — using termux-microphone-record")
			return &termuxRecorder{cfg: cfg}
		}
		logger.InfoC("voice", "Using system audio recorder")
		return &systemRecorder{cfg: cfg}
	}
}

// isTermux checks if running inside Termux.
func isTermux() bool {
	return os.Getenv("TERMUX_VERSION") != ""
}

// --- Termux Recorder ---

type termuxRecorder struct {
	cfg RecorderConfig
}

func (r *termuxRecorder) Name() string { return "termux" }

func (r *termuxRecorder) IsAvailable() bool {
	_, err := exec.LookPath("termux-microphone-record")
	return err == nil
}

func (r *termuxRecorder) Record(ctx context.Context, duration time.Duration) (string, error) {
	outPath := filepath.Join(r.cfg.TempDir, fmt.Sprintf("v1claw-rec-%d.wav", time.Now().UnixNano()))

	logger.InfoCF("voice", "Recording audio via Termux", map[string]interface{}{
		"duration_sec": duration.Seconds(), "output": outPath,
	})

	// termux-microphone-record writes until stopped or limit reached.
	args := []string{
		"-f", outPath,
		"-l", fmt.Sprintf("%d", int(duration.Seconds())),
		"-r", fmt.Sprintf("%d", r.cfg.SampleRate),
		"-e", "pcm_16bit",
	}
	recordCtx, cancel := context.WithTimeout(ctx, duration+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(recordCtx, "termux-microphone-record", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("termux-microphone-record: %w (output: %s)", err, string(out))
	}

	// Wait for the recording duration, then stop.
	select {
	case <-time.After(duration):
	case <-ctx.Done():
	}
	// Stop recording.
	stopCmd := exec.CommandContext(ctx, "termux-microphone-record", "-q")
	_ = stopCmd.Run()

	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("recording file not created: %w", err)
	}

	logger.InfoCF("voice", "Recording complete", map[string]interface{}{"file": outPath})
	return outPath, nil
}

// --- System Recorder (Linux: arecord, macOS: sox/rec) ---

type systemRecorder struct {
	cfg RecorderConfig
}

func (r *systemRecorder) Name() string { return "system" }

func (r *systemRecorder) IsAvailable() bool {
	cmd, _ := r.recordCommand()
	if cmd == "" {
		return false
	}
	_, err := exec.LookPath(cmd)
	return err == nil
}

// recordCommand returns the system command and base args for recording.
func (r *systemRecorder) recordCommand() (string, []string) {
	switch runtime.GOOS {
	case "linux":
		// arecord (ALSA)
		return "arecord", []string{
			"-f", "S16_LE",
			"-r", fmt.Sprintf("%d", r.cfg.SampleRate),
			"-c", "1",
			"-t", "wav",
		}
	case "darwin":
		// sox/rec (install: brew install sox)
		return "rec", []string{
			"-r", fmt.Sprintf("%d", r.cfg.SampleRate),
			"-c", "1",
			"-b", "16",
		}
	default:
		return "", nil
	}
}

func (r *systemRecorder) Record(ctx context.Context, duration time.Duration) (string, error) {
	cmdName, baseArgs := r.recordCommand()
	if cmdName == "" {
		return "", fmt.Errorf("no audio recorder available for %s", runtime.GOOS)
	}

	outPath := filepath.Join(r.cfg.TempDir, fmt.Sprintf("v1claw-rec-%d.wav", time.Now().UnixNano()))

	logger.InfoCF("voice", "Recording audio via system", map[string]interface{}{
		"command": cmdName, "duration_sec": duration.Seconds(), "output": outPath,
	})

	args := append(baseArgs, outPath)
	// Add duration limit.
	switch cmdName {
	case "arecord":
		args = append(args, "-d", fmt.Sprintf("%d", int(duration.Seconds())))
	case "rec":
		args = append(args, "trim", "0", fmt.Sprintf("%d", int(duration.Seconds())))
	}

	recordCtx, cancel := context.WithTimeout(ctx, duration+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(recordCtx, cmdName, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		os.Remove(outPath)
		return "", fmt.Errorf("%s: %w (output: %s)", cmdName, err, string(out))
	}

	if _, err := os.Stat(outPath); err != nil {
		return "", fmt.Errorf("recording file not created: %w", err)
	}

	logger.InfoCF("voice", "Recording complete", map[string]interface{}{"file": outPath})
	return outPath, nil
}
