package voice

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- TTS Manager Tests ---

type mockTTSProvider struct {
	name      string
	available bool
	failErr   error
}

func (m *mockTTSProvider) Name() string      { return m.name }
func (m *mockTTSProvider) IsAvailable() bool { return m.available }
func (m *mockTTSProvider) Synthesize(ctx context.Context, text string, opts TTSOptions) (io.ReadCloser, string, error) {
	if m.failErr != nil {
		return nil, "", m.failErr
	}
	return io.NopCloser(bytes.NewReader([]byte("audio:" + text))), "mp3", nil
}

func TestTTSManager_Register(t *testing.T) {
	mgr := NewTTSManager()
	mgr.Register(&mockTTSProvider{name: "test1", available: true})
	mgr.Register(&mockTTSProvider{name: "test2", available: true})
	assert.Equal(t, 2, mgr.ProviderCount())
}

func TestTTSManager_Synthesize(t *testing.T) {
	mgr := NewTTSManager()
	mgr.Register(&mockTTSProvider{name: "primary", available: true})

	reader, format, err := mgr.Synthesize(context.Background(), "Hello V1", DefaultTTSOptions())
	require.NoError(t, err)
	assert.Equal(t, "mp3", format)

	data, _ := io.ReadAll(reader)
	reader.Close()
	assert.Equal(t, "audio:Hello V1", string(data))
}

func TestTTSManager_Fallback(t *testing.T) {
	mgr := NewTTSManager()
	mgr.Register(&mockTTSProvider{name: "broken", available: true, failErr: assert.AnError})
	mgr.Register(&mockTTSProvider{name: "backup", available: true})

	reader, _, err := mgr.Synthesize(context.Background(), "test", DefaultTTSOptions())
	require.NoError(t, err)
	reader.Close()
}

func TestTTSManager_NoProviders(t *testing.T) {
	mgr := NewTTSManager()
	_, _, err := mgr.Synthesize(context.Background(), "test", DefaultTTSOptions())
	assert.Error(t, err)
}

func TestDefaultTTSOptions(t *testing.T) {
	opts := DefaultTTSOptions()
	assert.Equal(t, 1.0, opts.Speed)
	assert.Equal(t, 0.0, opts.Pitch)
	assert.Equal(t, "mp3", opts.Format)
}

func TestOpenAITTS_IsAvailable(t *testing.T) {
	tts := NewOpenAITTS(OpenAITTSConfig{APIKey: ""})
	assert.False(t, tts.IsAvailable())

	tts2 := NewOpenAITTS(OpenAITTSConfig{APIKey: "sk-test"})
	assert.True(t, tts2.IsAvailable())
}

func TestOpenAITTS_Name(t *testing.T) {
	tts := NewOpenAITTS(OpenAITTSConfig{})
	assert.Equal(t, "openai", tts.Name())
}

func TestEdgeTTS_Name(t *testing.T) {
	tts := NewEdgeTTS()
	assert.Equal(t, "edge", tts.Name())
}

// --- Wake Word Tests ---

func TestWakeWordDetector_BasicDetection(t *testing.T) {
	det := NewWakeWordDetector(WakeWordConfig{
		Phrases: []string{"hello v1", "hey v1"},
	})

	tests := []struct {
		input   string
		phrase  string
		command string
		found   bool
	}{
		{"Hello V1, what time is it?", "hello v1", "what time is it?", true},
		{"hey v1 turn on the lights", "hey v1", "turn on the lights", true},
		{"Hi V1", "", "Hi V1", false}, // Not in configured phrases
		{"good morning", "", "good morning", false},
		{"Hello V1", "hello v1", "", true},
		{"HELLO V1, how are you?", "hello v1", "how are you?", true},
	}

	for _, tt := range tests {
		phrase, command, found := det.Check(tt.input)
		assert.Equal(t, tt.found, found, "input: %s", tt.input)
		if found {
			assert.Equal(t, tt.phrase, phrase, "input: %s", tt.input)
			assert.Equal(t, tt.command, command, "input: %s", tt.input)
		}
	}
}

func TestWakeWordDetector_DefaultPhrases(t *testing.T) {
	det := NewWakeWordDetector(WakeWordConfig{})

	_, _, found1 := det.Check("hello v1 test")
	_, _, found2 := det.Check("hey v1 test")
	_, _, found3 := det.Check("hi v1 test")

	assert.True(t, found1)
	assert.True(t, found2)
	assert.True(t, found3)
}

func TestWakeWordDetector_Process(t *testing.T) {
	det := NewWakeWordDetector(WakeWordConfig{Cooldown: 0})

	var handlerCalled bool
	det.SetHandler(func(ctx context.Context, phrase string, confidence float64) {
		handlerCalled = true
	})

	command, detected := det.Process(context.Background(), "Hello V1, play music")
	assert.True(t, detected)
	assert.Equal(t, "play music", command)
	assert.True(t, handlerCalled)
}

func TestWakeWordDetector_Cooldown(t *testing.T) {
	det := NewWakeWordDetector(WakeWordConfig{Cooldown: 1})

	callCount := 0
	det.SetHandler(func(ctx context.Context, phrase string, confidence float64) {
		callCount++
	})

	det.Process(context.Background(), "Hello V1, first")
	det.Process(context.Background(), "Hello V1, second")

	assert.Equal(t, 1, callCount) // Second call within cooldown

	// Wait for cooldown.
	time.Sleep(1100 * time.Millisecond)
	det.Process(context.Background(), "Hello V1, third")
	assert.Equal(t, 2, callCount)
}

func TestWakeWordDetector_NoMatch(t *testing.T) {
	det := NewWakeWordDetector(WakeWordConfig{})

	command, detected := det.Process(context.Background(), "good morning")
	assert.False(t, detected)
	assert.Equal(t, "good morning", command)
}
