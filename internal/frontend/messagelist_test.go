package frontend

import (
	"strings"
	"testing"
)

func TestMessageListScrolling(t *testing.T) {
	ml := NewMessageList()

	// Add many messages to trigger scrolling
	for i := 0; i < 50; i++ {
		if i%2 == 0 {
			ml.AddMessage("user", "Test message "+string(rune('0'+i)))
		} else {
			// Add multi-line assistant response
			ml.AddMessage("assistant", "This is a response\nwith multiple lines\nto test scrolling\nbehavior properly")
		}
	}

	// Render should produce a large output
	elem := ml.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Check that we have many children (messages)
	if len(elem.Children) < 50 {
		t.Errorf("Expected at least 50 message elements, got %d", len(elem.Children))
	}
}

func TestMessageListFormatting(t *testing.T) {
	ml := NewMessageList()

	// Test user message formatting
	ml.AddMessage("user", "Hello")
	ml.AddMessage("assistant", "Hi there!")
	ml.AddMessage("system", "System message")

	elem := ml.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Should have 3 messages
	// Note: actual count may vary due to spacing elements
	if len(ml.Messages) != 3 {
		t.Errorf("Expected 3 messages, got %d", len(ml.Messages))
	}
}

func TestMessageListStreaming(t *testing.T) {
	ml := NewMessageList()

	// Test streaming indicator
	ml.SetStreaming(true, "Partial response")
	elem := ml.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Should show streaming indicator
	if !ml.Streaming {
		t.Error("Streaming should be enabled")
	}
	if ml.StreamBuf != "Partial response" {
		t.Errorf("StreamBuf should be 'Partial response', got %q", ml.StreamBuf)
	}

	// Test thinking indicator
	ml.SetStreaming(true, "")
	elem = ml.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}
}

func TestMessageFormatMarkdown(t *testing.T) {
	ml := NewMessageList()

	// Test that markdown gets rendered for assistant with newlines preserved
	ml.AddMessage("assistant", "Line 1\nLine 2\nLine 3")
	formatted := ml.formatMessage("assistant", "Line 1\nLine 2\nLine 3")

	// Should preserve newlines
	if strings.Count(formatted, "\n") < 2 {
		t.Error("Expected newlines to be preserved in markdown rendering")
	}

	// Test markdown formatting
	ml.AddMessage("assistant", "**bold** and `code`")
	formatted = ml.formatMessage("assistant", "**bold** and `code`")

	// Should have some ANSI codes from glamour
	if !strings.Contains(formatted, "\x1b") {
		t.Error("Expected ANSI codes in markdown-rendered output")
	}

	// User messages should not be markdown-rendered
	userFormatted := ml.formatMessage("user", "**not bold**")
	if !strings.Contains(userFormatted, "**") {
		t.Error("User message should preserve markdown syntax")
	}
}
