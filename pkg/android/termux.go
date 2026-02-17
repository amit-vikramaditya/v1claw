package android

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// TermuxAPI provides access to Android hardware via Termux:API package.
// Each method wraps a termux-* command from the termux-api package.
type TermuxAPI struct {
	available *bool
}

// NewTermuxAPI creates a new Termux API bridge.
func NewTermuxAPI() *TermuxAPI {
	return &TermuxAPI{}
}

// IsAvailable checks if termux-api commands are accessible.
func (t *TermuxAPI) IsAvailable() bool {
	if t.available != nil {
		return *t.available
	}
	_, err := exec.LookPath("termux-notification")
	avail := err == nil
	t.available = &avail
	if avail {
		logger.InfoC("android", "Termux:API detected — Android bridge enabled")
	}
	return avail
}

// --- Notifications ---

// Notify sends an Android notification.
func (t *TermuxAPI) Notify(ctx context.Context, title, content string, id string) error {
	args := []string{"--title", title, "--content", content}
	if id != "" {
		args = append(args, "--id", id)
	}
	return t.run(ctx, "termux-notification", args...)
}

// NotifyRemove removes a notification by ID.
func (t *TermuxAPI) NotifyRemove(ctx context.Context, id string) error {
	return t.run(ctx, "termux-notification-remove", id)
}

// --- Toast ---

// Toast shows a brief Android toast message.
func (t *TermuxAPI) Toast(ctx context.Context, message string, short bool) error {
	args := []string{message}
	if short {
		args = append(args, "-s")
	}
	return t.run(ctx, "termux-toast", args...)
}

// --- Vibrate ---

// Vibrate vibrates the device for the given duration.
func (t *TermuxAPI) Vibrate(ctx context.Context, durationMs int, force bool) error {
	args := []string{"-d", fmt.Sprintf("%d", durationMs)}
	if force {
		args = append(args, "-f")
	}
	return t.run(ctx, "termux-vibrate", args...)
}

// --- Battery ---

// BatteryStatus holds battery information.
type BatteryStatus struct {
	Health      string  `json:"health"`
	Percentage  int     `json:"percentage"`
	Plugged     string  `json:"plugged"`
	Status      string  `json:"status"`
	Temperature float64 `json:"temperature"`
}

// Battery returns current battery status.
func (t *TermuxAPI) Battery(ctx context.Context) (*BatteryStatus, error) {
	out, err := t.output(ctx, "termux-battery-status")
	if err != nil {
		return nil, err
	}
	var status BatteryStatus
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("parse battery status: %w", err)
	}
	return &status, nil
}

// --- Location ---

// Location holds GPS coordinates.
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
	Accuracy  float64 `json:"accuracy"`
	Provider  string  `json:"provider"`
}

// GetLocation returns current device location.
func (t *TermuxAPI) GetLocation(ctx context.Context, provider string) (*Location, error) {
	if provider == "" {
		provider = "gps"
	}
	out, err := t.output(ctx, "termux-location", "-p", provider)
	if err != nil {
		return nil, err
	}
	var loc Location
	if err := json.Unmarshal(out, &loc); err != nil {
		return nil, fmt.Errorf("parse location: %w", err)
	}
	return &loc, nil
}

// --- TTS (Text-to-Speech via Android) ---

// Speak uses Android's built-in TTS engine to speak text.
func (t *TermuxAPI) Speak(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, "termux-tts-speak", text)
	return cmd.Run()
}

// --- Torch ---

// Torch toggles the device flashlight.
func (t *TermuxAPI) Torch(ctx context.Context, on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	return t.run(ctx, "termux-torch", state)
}

// --- WiFi ---

// WifiInfo holds WiFi connection details.
type WifiInfo struct {
	BSSID       string `json:"bssid"`
	FrequencyMHz int   `json:"frequency_mhz"`
	IP          string `json:"ip"`
	LinkSpeedMbps int  `json:"link_speed_mbps"`
	RSSI        int    `json:"rssi"`
	SSID        string `json:"ssid"`
}

// GetWifiInfo returns current WiFi connection info.
func (t *TermuxAPI) GetWifiInfo(ctx context.Context) (*WifiInfo, error) {
	out, err := t.output(ctx, "termux-wifi-connectioninfo")
	if err != nil {
		return nil, err
	}
	var info WifiInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, fmt.Errorf("parse wifi info: %w", err)
	}
	return &info, nil
}

// --- Clipboard ---

// ClipboardGet returns the current clipboard contents.
func (t *TermuxAPI) ClipboardGet(ctx context.Context) (string, error) {
	out, err := t.output(ctx, "termux-clipboard-get")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ClipboardSet sets the clipboard contents.
func (t *TermuxAPI) ClipboardSet(ctx context.Context, text string) error {
	cmd := exec.CommandContext(ctx, "termux-clipboard-set", text)
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// --- Camera ---

// CameraPhoto takes a photo and saves to the given path.
func (t *TermuxAPI) CameraPhoto(ctx context.Context, outputPath string, cameraID int) error {
	return t.run(ctx, "termux-camera-photo", "-c", fmt.Sprintf("%d", cameraID), outputPath)
}

// --- SMS ---

// SMSMessage represents a text message.
type SMSMessage struct {
	Number string `json:"number"`
	Body   string `json:"body"`
	Date   string `json:"date"`
	Type   string `json:"type"` // "inbox", "sent"
}

// SendSMS sends a text message.
func (t *TermuxAPI) SendSMS(ctx context.Context, number, message string) error {
	return t.run(ctx, "termux-sms-send", "-n", number, message)
}

// GetSMS retrieves recent SMS messages.
func (t *TermuxAPI) GetSMS(ctx context.Context, limit int) ([]SMSMessage, error) {
	out, err := t.output(ctx, "termux-sms-list", "-l", fmt.Sprintf("%d", limit))
	if err != nil {
		return nil, err
	}
	var messages []SMSMessage
	if err := json.Unmarshal(out, &messages); err != nil {
		return nil, fmt.Errorf("parse sms: %w", err)
	}
	return messages, nil
}

// --- Telephony ---

// CallPhone initiates a phone call.
func (t *TermuxAPI) CallPhone(ctx context.Context, number string) error {
	return t.run(ctx, "termux-telephony-call", number)
}

// --- Sensor ---

// SensorInfo lists available sensors.
func (t *TermuxAPI) ListSensors(ctx context.Context) ([]string, error) {
	out, err := t.output(ctx, "termux-sensor", "-l")
	if err != nil {
		return nil, err
	}
	var result struct {
		Sensors []string `json:"sensors"`
	}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("parse sensors: %w", err)
	}
	return result.Sensors, nil
}

// --- Volume ---

// SetVolume sets the media volume (0-15).
func (t *TermuxAPI) SetVolume(ctx context.Context, stream string, volume int) error {
	return t.run(ctx, "termux-volume", stream, fmt.Sprintf("%d", volume))
}

// --- Helpers ---

func (t *TermuxAPI) run(ctx context.Context, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w (output: %s)", name, err, string(out))
	}
	return nil
}

func (t *TermuxAPI) output(ctx context.Context, name string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return out, nil
}
