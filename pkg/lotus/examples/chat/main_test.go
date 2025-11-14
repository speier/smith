package main

import (
	"testing"
)

func TestChatAppOnSubmit(t *testing.T) {
	app := NewChatApp()

	// Initial state: 2 welcome messages
	if len(app.messages) != 2 {
		t.Fatalf("Expected 2 initial messages, got %d", len(app.messages))
	}

	// Test normal message
	app.onSubmit("hello")
	if len(app.messages) != 4 { // 2 initial + "> hello" + "Echo: hello"
		t.Fatalf("Expected 4 messages after 'hello', got %d", len(app.messages))
	}

	// Test "long" command
	beforeLong := len(app.messages)
	app.onSubmit("long")
	afterLong := len(app.messages)

	// Should add: "> long" + 50 auto-scroll lines = 51 messages (no Echo for "long")
	expected := beforeLong + 51
	if afterLong != expected {
		t.Fatalf("Expected %d messages after 'long', got %d", expected, afterLong)
	}

	// Verify some of the auto-scroll messages exist
	found := false
	for _, msg := range app.messages {
		if msg == "[01] Auto-scroll test line" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find '[01] Auto-scroll test line' in messages")
	}
}

func TestChatAppEmptySubmit(t *testing.T) {
	app := NewChatApp()

	initialCount := len(app.messages)
	app.onSubmit("")

	if len(app.messages) != initialCount {
		t.Errorf("Empty submit should not add messages, got %d", len(app.messages))
	}
}
