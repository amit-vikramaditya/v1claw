package integrations

import (
	"context"
	"fmt"
	"time"
)

// SmartHomeDevice represents an IoT device.
type SmartHomeDevice struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Type     string                 `json:"type"`  // "light", "switch", "sensor", "thermostat", "lock", "camera"
	State    string                 `json:"state"` // "on", "off", "locked", "unlocked", etc.
	Area     string                 `json:"area,omitempty"`
	Attrs    map[string]interface{} `json:"attributes,omitempty"`
	LastSeen time.Time              `json:"last_seen"`
}

// SmartHomeProvider is the interface for smart home backends.
type SmartHomeProvider interface {
	// ListDevices returns all known devices.
	ListDevices(ctx context.Context) ([]SmartHomeDevice, error)
	// GetDevice returns a specific device by ID.
	GetDevice(ctx context.Context, id string) (*SmartHomeDevice, error)
	// SetState changes a device's state (e.g., turn on/off).
	SetState(ctx context.Context, id, state string) error
	// CallService invokes a service on a device (e.g., set_temperature).
	CallService(ctx context.Context, domain, service string, data map[string]interface{}) error
	// Subscribe watches for device state changes.
	Subscribe(ctx context.Context, handler func(device SmartHomeDevice)) error
	// Name returns the provider name.
	Name() string
}

// SmartHomeConfig holds smart home integration configuration.
type SmartHomeConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"` // "homeassistant", "mqtt"
	// Home Assistant
	HAURL   string `json:"ha_url"`   // e.g., "http://homeassistant.local:8123"
	HAToken string `json:"ha_token"` // Long-lived access token
	// MQTT
	MQTTBroker   string `json:"mqtt_broker"` // e.g., "tcp://localhost:1883"
	MQTTUsername string `json:"mqtt_username"`
	MQTTPassword string `json:"mqtt_password"`
}

// SmartHomeManager provides a unified smart home interface.
type SmartHomeManager struct {
	provider SmartHomeProvider
	config   SmartHomeConfig
}

// NewSmartHomeManager creates a smart home manager.
func NewSmartHomeManager(cfg SmartHomeConfig, provider SmartHomeProvider) *SmartHomeManager {
	return &SmartHomeManager{config: cfg, provider: provider}
}

// TurnOn turns on a device.
func (m *SmartHomeManager) TurnOn(ctx context.Context, deviceID string) error {
	return m.provider.SetState(ctx, deviceID, "on")
}

// TurnOff turns off a device.
func (m *SmartHomeManager) TurnOff(ctx context.Context, deviceID string) error {
	return m.provider.SetState(ctx, deviceID, "off")
}

// ListByArea returns devices in a specific area.
func (m *SmartHomeManager) ListByArea(ctx context.Context, area string) ([]SmartHomeDevice, error) {
	all, err := m.provider.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []SmartHomeDevice
	for _, d := range all {
		if d.Area == area {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}

// ListByType returns devices of a specific type.
func (m *SmartHomeManager) ListByType(ctx context.Context, deviceType string) ([]SmartHomeDevice, error) {
	all, err := m.provider.ListDevices(ctx)
	if err != nil {
		return nil, err
	}
	var filtered []SmartHomeDevice
	for _, d := range all {
		if d.Type == deviceType {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}

// Status returns a human-readable status summary.
func (m *SmartHomeManager) Status(ctx context.Context) (string, error) {
	devices, err := m.provider.ListDevices(ctx)
	if err != nil {
		return "", err
	}

	on, off, other := 0, 0, 0
	for _, d := range devices {
		switch d.State {
		case "on":
			on++
		case "off":
			off++
		default:
			other++
		}
	}

	return fmt.Sprintf("🏠 Smart Home: %d devices (%d on, %d off, %d other)",
		len(devices), on, off, other), nil
}
