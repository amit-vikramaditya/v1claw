package vision

import (
	"context"
	"fmt"
	"time"
)

// ImageSource identifies where an image came from.
type ImageSource struct {
	Type     string `json:"type"`      // "camera", "screenshot", "file", "url"
	DeviceID string `json:"device_id,omitempty"`
	Path     string `json:"path,omitempty"`
	URL      string `json:"url,omitempty"`
}

// AnalysisResult holds the result of image analysis.
type AnalysisResult struct {
	Description string            `json:"description"`
	Objects     []DetectedObject  `json:"objects,omitempty"`
	Text        string            `json:"text,omitempty"` // OCR text
	Labels      []string          `json:"labels,omitempty"`
	Confidence  float64           `json:"confidence"`
	Timestamp   time.Time         `json:"timestamp"`
	Source      ImageSource       `json:"source"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// DetectedObject represents a detected object in an image.
type DetectedObject struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	BoundingBox *BBox  `json:"bounding_box,omitempty"`
}

// BBox is a bounding box for object detection.
type BBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// VisionProvider analyzes images using multimodal LLMs or local models.
type VisionProvider interface {
	// Analyze sends an image (as bytes) to the vision model with a prompt.
	Analyze(ctx context.Context, imageData []byte, mimeType, prompt string) (*AnalysisResult, error)
	// AnalyzeURL analyzes an image from a URL.
	AnalyzeURL(ctx context.Context, imageURL, prompt string) (*AnalysisResult, error)
	// Name returns the provider name.
	Name() string
	// IsAvailable returns whether the provider is configured.
	IsAvailable() bool
}

// CameraProvider captures images from cameras.
type CameraProvider interface {
	// Capture takes a snapshot and returns JPEG bytes.
	Capture(ctx context.Context) ([]byte, error)
	// StreamURL returns an RTSP/MJPEG stream URL if available.
	StreamURL() string
	// Name returns the camera name.
	Name() string
}

// OCRProvider extracts text from images.
type OCRProvider interface {
	// ExtractText performs OCR on image bytes.
	ExtractText(ctx context.Context, imageData []byte, mimeType string) (string, error)
	// Name returns the provider name.
	Name() string
}

// ScreenProvider captures screenshots.
type ScreenProvider interface {
	// CaptureScreen takes a screenshot of the entire screen or a window.
	CaptureScreen(ctx context.Context, windowTitle string) ([]byte, error)
	// Name returns the provider name.
	Name() string
}

// FormatAnalysis returns a human-readable summary.
func FormatAnalysis(r *AnalysisResult) string {
	if r == nil {
		return "No analysis available."
	}
	result := fmt.Sprintf("👁️ %s", r.Description)
	if len(r.Objects) > 0 {
		result += fmt.Sprintf("\nDetected %d objects:", len(r.Objects))
		for _, obj := range r.Objects {
			result += fmt.Sprintf("\n  • %s (%.0f%%)", obj.Label, obj.Confidence*100)
		}
	}
	if r.Text != "" {
		result += fmt.Sprintf("\nOCR Text: %s", r.Text)
	}
	return result
}
