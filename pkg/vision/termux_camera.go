package vision

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
	"github.com/amit-vikramaditya/v1claw/pkg/permissions"
)

// TermuxCamera implements CameraProvider using termux-camera-photo.
type TermuxCamera struct {
	cameraID int    // 0 = back, 1 = front
	name     string
	tempDir  string
}

// TermuxCameraConfig configures the Termux camera.
type TermuxCameraConfig struct {
	CameraID int    `json:"camera_id"` // 0 = back camera, 1 = front camera
	Name     string `json:"name"`      // e.g., "termux-back", "termux-front"
	TempDir  string `json:"temp_dir"`
}

// NewTermuxCamera creates a Termux camera provider.
func NewTermuxCamera(cfg TermuxCameraConfig) *TermuxCamera {
	name := cfg.Name
	if name == "" {
		if cfg.CameraID == 1 {
			name = "termux-front"
		} else {
			name = "termux-back"
		}
	}
	tempDir := cfg.TempDir
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	return &TermuxCamera{
		cameraID: cfg.CameraID,
		name:     name,
		tempDir:  tempDir,
	}
}

func (c *TermuxCamera) Name() string { return c.name }

func (c *TermuxCamera) StreamURL() string { return "" }

// IsAvailable checks if termux-camera-photo is accessible.
func (c *TermuxCamera) IsAvailable() bool {
	_, err := exec.LookPath("termux-camera-photo")
	return err == nil
}

// Capture takes a photo using termux-camera-photo and returns JPEG bytes.
func (c *TermuxCamera) Capture(ctx context.Context) ([]byte, error) {
	if err := permissions.Global().Check(permissions.Camera, "vision.TermuxCamera"); err != nil {
		return nil, err
	}

	outPath := filepath.Join(c.tempDir, fmt.Sprintf("v1claw-cam-%d-%d.jpg", c.cameraID, time.Now().UnixNano()))

	logger.InfoCF("vision", "Capturing photo via Termux", map[string]interface{}{
		"camera_id": c.cameraID, "output": outPath,
	})

	captureCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(captureCtx, "termux-camera-photo",
		"-c", fmt.Sprintf("%d", c.cameraID),
		outPath,
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("termux-camera-photo: %w (output: %s)", err, string(out))
	}

	defer os.Remove(outPath)

	data, err := os.ReadFile(outPath)
	if err != nil {
		return nil, fmt.Errorf("read captured photo: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("captured photo is empty")
	}

	logger.InfoCF("vision", "Photo captured", map[string]interface{}{
		"size_bytes": len(data), "camera": c.name,
	})
	return data, nil
}
