package devtools

import (
	"testing"
)

func TestDevToolsNew(t *testing.T) {
	dt := New()

	if dt == nil {
		t.Fatal("New() returned nil")
	}

	if !dt.IsEnabled() {
		t.Error("DevTools should be enabled by default")
	}
}

func TestDevToolsLog(t *testing.T) {
	dt := New()

	// Should not panic
	dt.Log("test message")
	dt.Log("formatted %s %d", "message", 42)
}

func TestDevToolsEnableDisable(t *testing.T) {
	dt := New()

	dt.Disable()
	if dt.IsEnabled() {
		t.Error("DevTools should be disabled")
	}

	dt.Enable()
	if !dt.IsEnabled() {
		t.Error("DevTools should be enabled")
	}
}

func TestDevToolsRender(t *testing.T) {
	dt := New()

	// Should render when enabled
	output := dt.Render()
	if output == nil {
		t.Error("Render() should return non-nil Element when enabled")
	}

	// Should return nil when disabled
	dt.Disable()
	output = dt.Render()
	if output != nil {
		t.Error("Render() should return nil when disabled")
	}
}
