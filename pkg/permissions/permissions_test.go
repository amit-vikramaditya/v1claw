package permissions

import (
	"testing"
)

func TestDenyByDefault(t *testing.T) {
	r := NewRegistry()
	for _, f := range AllFeatures() {
		if r.IsAllowed(f) {
			t.Errorf("%s should be denied by default", f)
		}
	}
}

func TestSetAndCheck(t *testing.T) {
	r := NewRegistry()

	r.Set(Camera, true)
	if !r.IsAllowed(Camera) {
		t.Error("camera should be allowed after Set(true)")
	}

	r.Set(Camera, false)
	if r.IsAllowed(Camera) {
		t.Error("camera should be denied after Set(false)")
	}
}

func TestCheckReturnsError(t *testing.T) {
	r := NewRegistry()

	err := r.Check(SMS, "test")
	if err == nil {
		t.Error("expected error for denied permission")
	}

	r.Set(SMS, true)
	err = r.Check(SMS, "test")
	if err != nil {
		t.Errorf("expected nil for allowed permission, got %v", err)
	}
}

func TestSetAll(t *testing.T) {
	r := NewRegistry()
	r.SetAll(map[Feature]bool{
		Camera:     true,
		Microphone: true,
		SMS:        false,
	})

	if !r.IsAllowed(Camera) {
		t.Error("camera should be allowed")
	}
	if !r.IsAllowed(Microphone) {
		t.Error("microphone should be allowed")
	}
	if r.IsAllowed(SMS) {
		t.Error("sms should be denied")
	}
	if r.IsAllowed(Location) {
		t.Error("location should be denied (not set)")
	}
}

func TestSnapshot(t *testing.T) {
	r := NewRegistry()
	r.Set(Location, true)

	snap := r.Snapshot()
	if !snap[Location] {
		t.Error("snapshot should show location=true")
	}

	// Mutating snapshot shouldn't affect registry.
	snap[Location] = false
	if !r.IsAllowed(Location) {
		t.Error("registry should not be affected by snapshot mutation")
	}
}

func TestEnabledFeatures(t *testing.T) {
	r := NewRegistry()
	r.Set(Camera, true)
	r.Set(Microphone, true)

	enabled := r.EnabledFeatures()
	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled, got %d", len(enabled))
	}
}

func TestGlobalSingleton(t *testing.T) {
	g1 := Global()
	g2 := Global()
	if g1 != g2 {
		t.Error("Global() should return same instance")
	}
}

func TestAllFeatures(t *testing.T) {
	features := AllFeatures()
	if len(features) != 8 {
		t.Errorf("expected 8 features, got %d", len(features))
	}
}
