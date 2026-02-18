package vision

import (
	"testing"
)

func TestTermuxCameraCreation(t *testing.T) {
	cam := NewTermuxCamera(TermuxCameraConfig{CameraID: 0})
	if cam.Name() != "termux-back" {
		t.Errorf("expected termux-back, got %s", cam.Name())
	}
	if cam.StreamURL() != "" {
		t.Error("termux camera should not have stream URL")
	}

	cam = NewTermuxCamera(TermuxCameraConfig{CameraID: 1})
	if cam.Name() != "termux-front" {
		t.Errorf("expected termux-front, got %s", cam.Name())
	}

	cam = NewTermuxCamera(TermuxCameraConfig{CameraID: 0, Name: "custom-cam"})
	if cam.Name() != "custom-cam" {
		t.Errorf("expected custom-cam, got %s", cam.Name())
	}
}

func TestTermuxCameraRegistration(t *testing.T) {
	mgr := NewManager()
	cam := NewTermuxCamera(TermuxCameraConfig{CameraID: 0})
	mgr.RegisterCamera(cam)

	if mgr.CameraCount() != 1 {
		t.Errorf("expected 1 camera, got %d", mgr.CameraCount())
	}
}
