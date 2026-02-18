// Package permissions provides a centralized, thread-safe permission
// registry for hardware and system features. All permissions default
// to BLOCKED — users must explicitly enable features they need.
package permissions

import (
	"fmt"
	"sync"

	"github.com/amit-vikramaditya/v1claw/pkg/logger"
)

// Feature identifies a hardware or system capability.
type Feature string

const (
	Camera      Feature = "camera"
	Microphone  Feature = "microphone"
	SMS         Feature = "sms"
	PhoneCalls  Feature = "phone_calls"
	Location    Feature = "location"
	Clipboard   Feature = "clipboard"
	Sensors     Feature = "sensors"
	ShellHardware Feature = "shell_hardware"
)

// AllFeatures returns all defined features.
func AllFeatures() []Feature {
	return []Feature{Camera, Microphone, SMS, PhoneCalls, Location, Clipboard, Sensors, ShellHardware}
}

// Registry holds the current permission state for each feature.
// It is safe for concurrent use.
type Registry struct {
	mu    sync.RWMutex
	perms map[Feature]bool
}

var (
	global     *Registry
	globalOnce sync.Once
)

// Global returns the singleton permission registry.
func Global() *Registry {
	globalOnce.Do(func() {
		global = &Registry{perms: make(map[Feature]bool)}
	})
	return global
}

// NewRegistry creates a fresh registry (for testing).
func NewRegistry() *Registry {
	return &Registry{perms: make(map[Feature]bool)}
}

// Set enables or disables a feature.
func (r *Registry) Set(f Feature, allowed bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.perms[f] = allowed
}

// IsAllowed returns true if the feature is explicitly enabled.
// Returns false by default (deny-by-default).
func (r *Registry) IsAllowed(f Feature) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.perms[f]
}

// Check verifies permission and logs an audit entry.
// Returns nil if allowed, error if denied.
func (r *Registry) Check(f Feature, caller string) error {
	allowed := r.IsAllowed(f)
	if allowed {
		logger.DebugCF("permissions", fmt.Sprintf("ALLOWED: %s accessed %s", caller, f), nil)
		return nil
	}
	logger.WarnCF("permissions", fmt.Sprintf("DENIED: %s attempted %s — permission not granted", caller, f), map[string]interface{}{
		"feature": string(f), "caller": caller,
	})
	return fmt.Errorf("permission denied: %s access is not enabled (set permissions.%s=true in config)", f, f)
}

// SetAll sets multiple features at once (used during config loading).
func (r *Registry) SetAll(perms map[Feature]bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for f, v := range perms {
		r.perms[f] = v
	}
}

// Snapshot returns a copy of all permission states.
func (r *Registry) Snapshot() map[Feature]bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[Feature]bool, len(r.perms))
	for k, v := range r.perms {
		out[k] = v
	}
	return out
}

// EnabledFeatures returns a slice of features that are currently allowed.
func (r *Registry) EnabledFeatures() []Feature {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var enabled []Feature
	for f, v := range r.perms {
		if v {
			enabled = append(enabled, f)
		}
	}
	return enabled
}
