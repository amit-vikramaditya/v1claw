package vision

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Mocks ---

type mockVisionProvider struct {
	name      string
	available bool
	result    *AnalysisResult
	err       error
}

func (m *mockVisionProvider) Name() string      { return m.name }
func (m *mockVisionProvider) IsAvailable() bool { return m.available }
func (m *mockVisionProvider) Analyze(ctx context.Context, data []byte, mime, prompt string) (*AnalysisResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}
func (m *mockVisionProvider) AnalyzeURL(ctx context.Context, url, prompt string) (*AnalysisResult, error) {
	return m.Analyze(ctx, nil, "", prompt)
}

type mockCamera struct {
	name string
	data []byte
}

func (m *mockCamera) Name() string                                { return m.name }
func (m *mockCamera) StreamURL() string                           { return "" }
func (m *mockCamera) Capture(ctx context.Context) ([]byte, error) { return m.data, nil }

type mockOCR struct{}

func (m *mockOCR) Name() string { return "mock_ocr" }
func (m *mockOCR) ExtractText(ctx context.Context, data []byte, mime string) (string, error) {
	return "extracted text from image", nil
}

// --- Tests ---

func TestManager_RegisterProvider(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterProvider(&mockVisionProvider{name: "test", available: true})
	assert.Equal(t, 1, mgr.ProviderCount())
}

func TestManager_RegisterCamera(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterCamera(&mockCamera{name: "webcam"})
	assert.Equal(t, 1, mgr.CameraCount())
}

func TestManager_Analyze(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterProvider(&mockVisionProvider{
		name:      "gpt4v",
		available: true,
		result:    &AnalysisResult{Description: "A cat sitting on a desk"},
	})

	result, err := mgr.Analyze(context.Background(), []byte("fake-image"), "image/jpeg", "")
	require.NoError(t, err)
	assert.Equal(t, "A cat sitting on a desk", result.Description)
}

func TestManager_Analyze_Fallback(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterProvider(&mockVisionProvider{name: "broken", available: true, err: assert.AnError})
	mgr.RegisterProvider(&mockVisionProvider{
		name:      "backup",
		available: true,
		result:    &AnalysisResult{Description: "fallback result"},
	})

	result, err := mgr.Analyze(context.Background(), []byte("img"), "image/jpeg", "describe")
	require.NoError(t, err)
	assert.Equal(t, "fallback result", result.Description)
}

func TestManager_Analyze_NoProviders(t *testing.T) {
	mgr := NewManager()
	_, err := mgr.Analyze(context.Background(), []byte("img"), "image/jpeg", "")
	assert.Error(t, err)
}

func TestManager_CaptureAndAnalyze(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterProvider(&mockVisionProvider{
		name:      "vision",
		available: true,
		result:    &AnalysisResult{Description: "person detected"},
	})
	mgr.RegisterCamera(&mockCamera{name: "front_door", data: []byte("jpeg-data")})

	result, err := mgr.CaptureAndAnalyze(context.Background(), "front_door", "who is at the door?")
	require.NoError(t, err)
	assert.Equal(t, "person detected", result.Description)
	assert.Equal(t, "camera", result.Source.Type)
	assert.Equal(t, "front_door", result.Source.DeviceID)
}

func TestManager_CaptureAndAnalyze_NoCamera(t *testing.T) {
	mgr := NewManager()
	_, err := mgr.CaptureAndAnalyze(context.Background(), "missing", "")
	assert.Error(t, err)
}

func TestManager_OCR_WithProvider(t *testing.T) {
	mgr := NewManager()
	mgr.SetOCR(&mockOCR{})

	text, err := mgr.OCR(context.Background(), []byte("img"), "image/png")
	require.NoError(t, err)
	assert.Equal(t, "extracted text from image", text)
}

func TestManager_OCR_FallbackToVision(t *testing.T) {
	mgr := NewManager()
	mgr.RegisterProvider(&mockVisionProvider{
		name:      "vision",
		available: true,
		result:    &AnalysisResult{Description: "Hello World"},
	})

	text, err := mgr.OCR(context.Background(), []byte("img"), "image/png")
	require.NoError(t, err)
	assert.Equal(t, "Hello World", text)
}

func TestManager_Screenshot_NoProvider(t *testing.T) {
	mgr := NewManager()
	_, err := mgr.Screenshot(context.Background(), "", "")
	assert.Error(t, err)
}

func TestFormatAnalysis(t *testing.T) {
	result := &AnalysisResult{
		Description: "A living room",
		Objects: []DetectedObject{
			{Label: "couch", Confidence: 0.95},
			{Label: "lamp", Confidence: 0.87},
		},
		Text: "Welcome Home",
	}

	formatted := FormatAnalysis(result)
	assert.Contains(t, formatted, "A living room")
	assert.Contains(t, formatted, "couch")
	assert.Contains(t, formatted, "95%")
	assert.Contains(t, formatted, "OCR Text: Welcome Home")
}

func TestFormatAnalysis_Nil(t *testing.T) {
	assert.Equal(t, "No analysis available.", FormatAnalysis(nil))
}
