package devtools

import (
	"testing"
)

func TestDevToolsNew(t *testing.T) {
	dt := New()

	if dt == nil {
		t.Fatal("New() returned nil")
	}

	if !dt.enabled {
		t.Error("DevTools should be enabled by default")
	}

	if dt.GetID() != "devtools-panel" {
		t.Errorf("Expected ID 'devtools-panel', got '%s'", dt.GetID())
	}
}

func TestDevToolsLog(t *testing.T) {
	dt := New()

	// Should have initialization message
	if len(dt.messageList.Messages) == 0 {
		t.Error("Expected initialization message")
	}

	// Add a log message
	dt.Log("Test message %s %d", "foo", 42)

	if len(dt.messageList.Messages) < 2 {
		t.Error("Expected at least 2 messages after Log()")
	}

	// Check last message contains our text
	lastMsg := dt.messageList.Messages[len(dt.messageList.Messages)-1]
	if lastMsg.Content == "" {
		t.Error("Log message should not be empty")
	}
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
