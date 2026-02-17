package voice

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/sipeed/picoclaw/pkg/logger"
)

// EdgeTTS implements TTSProvider using Microsoft Edge TTS (free, no API key).
// Requires the `edge-tts` CLI tool (pip install edge-tts).
type EdgeTTS struct {
	available *bool
}

// NewEdgeTTS creates an Edge TTS provider.
func NewEdgeTTS() *EdgeTTS {
	return &EdgeTTS{}
}

func (e *EdgeTTS) Name() string { return "edge" }

func (e *EdgeTTS) IsAvailable() bool {
	if e.available != nil {
		return *e.available
	}
	_, err := exec.LookPath("edge-tts")
	avail := err == nil
	e.available = &avail
	if avail {
		logger.InfoC("tts", "Edge TTS available (edge-tts CLI found)")
	}
	return avail
}

func (e *EdgeTTS) Synthesize(ctx context.Context, text string, opts TTSOptions) (io.ReadCloser, string, error) {
	voice := opts.Voice
	if voice == "" {
		voice = "en-US-GuyNeural"
	}

	rate := ""
	if opts.Speed != 0 && opts.Speed != 1.0 {
		pct := int((opts.Speed - 1.0) * 100)
		if pct >= 0 {
			rate = fmt.Sprintf("+%d%%", pct)
		} else {
			rate = fmt.Sprintf("%d%%", pct)
		}
	}

	args := []string{"--voice", voice, "--text", text, "--write-media", "/dev/stdout"}
	if rate != "" {
		args = append(args, "--rate", rate)
	}

	cmd := exec.CommandContext(ctx, "edge-tts", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", fmt.Errorf("edge-tts stdout pipe: %w", err)
	}

	logger.DebugCF("tts", "Edge TTS synthesizing", map[string]interface{}{
		"voice": voice, "text_len": len(text),
	})

	if err := cmd.Start(); err != nil {
		return nil, "", fmt.Errorf("edge-tts start: %w", err)
	}

	// Return a wrapper that waits for the process when closed.
	return &cmdReadCloser{ReadCloser: stdout, cmd: cmd}, "mp3", nil
}

// ListEdgeVoices returns available Edge TTS voices.
func ListEdgeVoices(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "edge-tts", "--list-voices")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("edge-tts --list-voices: %w", err)
	}
	var voices []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Name: ") {
			voices = append(voices, strings.TrimPrefix(line, "Name: "))
		}
	}
	return voices, nil
}

// cmdReadCloser wraps an io.ReadCloser that also waits for a command to finish.
type cmdReadCloser struct {
	io.ReadCloser
	cmd *exec.Cmd
}

func (c *cmdReadCloser) Close() error {
	err := c.ReadCloser.Close()
	c.cmd.Wait()
	return err
}
