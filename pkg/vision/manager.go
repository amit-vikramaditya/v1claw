package vision

import (
	"context"
	"fmt"
	"sync"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// Manager coordinates vision providers with fallback.
type Manager struct {
	mu        sync.RWMutex
	providers map[string]VisionProvider
	cameras   map[string]CameraProvider
	screen    ScreenProvider
	ocr       OCRProvider
	primary   string
}

// NewManager creates a vision manager.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[string]VisionProvider),
		cameras:   make(map[string]CameraProvider),
	}
}

// RegisterProvider adds a vision analysis provider.
func (m *Manager) RegisterProvider(p VisionProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[p.Name()] = p
	if m.primary == "" {
		m.primary = p.Name()
	}
}

// RegisterCamera adds a camera source.
func (m *Manager) RegisterCamera(c CameraProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cameras[c.Name()] = c
}

// SetScreen sets the screen capture provider.
func (m *Manager) SetScreen(s ScreenProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.screen = s
}

// SetOCR sets the OCR provider.
func (m *Manager) SetOCR(o OCRProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ocr = o
}

// Analyze sends image bytes to the primary vision provider.
func (m *Manager) Analyze(ctx context.Context, imageData []byte, mimeType, prompt string) (*AnalysisResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if prompt == "" {
		prompt = "Describe what you see in this image."
	}

	// Try primary.
	if p, ok := m.providers[m.primary]; ok && p.IsAvailable() {
		result, err := p.Analyze(ctx, imageData, mimeType, prompt)
		if err == nil {
			return result, nil
		}
		logger.DebugC("vision", fmt.Sprintf("Primary provider %s failed: %v", m.primary, err))
	}

	// Fallback.
	for name, p := range m.providers {
		if name == m.primary || !p.IsAvailable() {
			continue
		}
		result, err := p.Analyze(ctx, imageData, mimeType, prompt)
		if err == nil {
			return result, nil
		}
	}

	return nil, fmt.Errorf("no vision provider available")
}

// CaptureAndAnalyze captures from a camera and analyzes the image.
func (m *Manager) CaptureAndAnalyze(ctx context.Context, cameraName, prompt string) (*AnalysisResult, error) {
	m.mu.RLock()
	cam, ok := m.cameras[cameraName]
	m.mu.RUnlock()

	if !ok {
		if cameraName != "" {
			return nil, fmt.Errorf("camera %q not found", cameraName)
		}
		// No name specified; try first available camera.
		m.mu.RLock()
		for _, c := range m.cameras {
			cam = c
			break
		}
		m.mu.RUnlock()
	}

	if cam == nil {
		return nil, fmt.Errorf("no camera available")
	}

	imageData, err := cam.Capture(ctx)
	if err != nil {
		return nil, fmt.Errorf("capture: %w", err)
	}

	result, err := m.Analyze(ctx, imageData, "image/jpeg", prompt)
	if err != nil {
		return nil, err
	}

	result.Source = ImageSource{Type: "camera", DeviceID: cam.Name()}
	return result, nil
}

// Screenshot captures and analyzes the screen.
func (m *Manager) Screenshot(ctx context.Context, windowTitle, prompt string) (*AnalysisResult, error) {
	m.mu.RLock()
	screen := m.screen
	m.mu.RUnlock()

	if screen == nil {
		return nil, fmt.Errorf("no screen capture provider available")
	}

	imageData, err := screen.CaptureScreen(ctx, windowTitle)
	if err != nil {
		return nil, fmt.Errorf("screenshot: %w", err)
	}

	result, err := m.Analyze(ctx, imageData, "image/png", prompt)
	if err != nil {
		return nil, err
	}

	result.Source = ImageSource{Type: "screenshot"}
	return result, nil
}

// OCR extracts text from image bytes.
func (m *Manager) OCR(ctx context.Context, imageData []byte, mimeType string) (string, error) {
	m.mu.RLock()
	ocr := m.ocr
	m.mu.RUnlock()

	if ocr == nil {
		// Fall back to vision provider with OCR prompt.
		result, err := m.Analyze(ctx, imageData, mimeType, "Extract all text from this image. Return only the text, no descriptions.")
		if err != nil {
			return "", err
		}
		return result.Description, nil
	}

	return ocr.ExtractText(ctx, imageData, mimeType)
}

// ProviderCount returns the number of registered vision providers.
func (m *Manager) ProviderCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.providers)
}

// CameraCount returns the number of registered cameras.
func (m *Manager) CameraCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cameras)
}
