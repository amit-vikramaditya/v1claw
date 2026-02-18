package android

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTermuxAPI_NewTermuxAPI(t *testing.T) {
	api := NewTermuxAPI()
	assert.NotNil(t, api)
}

func TestTermuxAPI_IsAvailable(t *testing.T) {
	api := NewTermuxAPI()
	// On non-Android systems, termux commands won't exist.
	// Just verify the check doesn't panic.
	_ = api.IsAvailable()
}

func TestTermuxAPI_IsAvailable_Cached(t *testing.T) {
	api := NewTermuxAPI()
	// Call twice to test caching.
	result1 := api.IsAvailable()
	result2 := api.IsAvailable()
	assert.Equal(t, result1, result2)
}

func TestBatteryStatus_Struct(t *testing.T) {
	status := BatteryStatus{
		Health:      "GOOD",
		Percentage:  85,
		Plugged:     "UNPLUGGED",
		Status:      "DISCHARGING",
		Temperature: 25.0,
	}
	assert.Equal(t, 85, status.Percentage)
	assert.Equal(t, "GOOD", status.Health)
}

func TestLocation_Struct(t *testing.T) {
	loc := Location{
		Latitude:  37.7749,
		Longitude: -122.4194,
		Altitude:  10.0,
		Accuracy:  5.0,
		Provider:  "gps",
	}
	assert.InDelta(t, 37.7749, loc.Latitude, 0.001)
}

func TestWifiInfo_Struct(t *testing.T) {
	info := WifiInfo{
		SSID:          "HomeNetwork",
		IP:            "192.168.1.100",
		RSSI:          -50,
		LinkSpeedMbps: 100,
	}
	assert.Equal(t, "HomeNetwork", info.SSID)
}

func TestSMSMessage_Struct(t *testing.T) {
	msg := SMSMessage{
		Number: "+1234567890",
		Body:   "Hello V1",
		Type:   "inbox",
	}
	assert.Equal(t, "Hello V1", msg.Body)
}
