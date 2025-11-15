package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
)

func TestMatchesBinding(t *testing.T) {
	tests := []struct {
		name     string
		event    tty.KeyEvent
		binding  *KeyBinding
		expected bool
	}{
		{
			name:  "Ctrl+O matches",
			event: tty.KeyEvent{Key: 15}, // Ctrl+O = 15
			binding: &KeyBinding{
				Key:  'o',
				Ctrl: true,
			},
			expected: true,
		},
		{
			name:  "Ctrl+K matches",
			event: tty.KeyEvent{Key: 11}, // Ctrl+K = 11
			binding: &KeyBinding{
				Key:  'k',
				Ctrl: true,
			},
			expected: true,
		},
		{
			name:  "Regular key matches",
			event: tty.KeyEvent{Key: 'a'},
			binding: &KeyBinding{
				Key:  'a',
				Ctrl: false,
			},
			expected: true,
		},
		{
			name:  "Regular key doesn't match",
			event: tty.KeyEvent{Key: 'b'},
			binding: &KeyBinding{
				Key:  'a',
				Ctrl: false,
			},
			expected: false,
		},
		{
			name:  "Escape sequence matches",
			event: tty.KeyEvent{Code: "[A"},
			binding: &KeyBinding{
				Code: "[A",
			},
			expected: true,
		},
		{
			name:  "Escape sequence doesn't match",
			event: tty.KeyEvent{Code: "[B"},
			binding: &KeyBinding{
				Code: "[A",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesBinding(tt.event, tt.binding)
			if result != tt.expected {
				t.Errorf("matchesBinding() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGlobalKeyRegistration(t *testing.T) {
	// Clear any existing handlers
	UnregisterAllGlobalKeys()

	// Register some handlers
	ctrlOCalled := false
	RegisterGlobalKey('o', true, "Open", func(event tty.KeyEvent) bool {
		ctrlOCalled = true
		return true
	})

	ctrlKCalled := false
	RegisterGlobalKey('k', true, "Command palette", func(event tty.KeyEvent) bool {
		ctrlKCalled = true
		return true
	})

	f1Called := false
	RegisterGlobalKeyCode("[OP", "Help", func(event tty.KeyEvent) bool {
		f1Called = true
		return true
	})

	// Check registration
	bindings := GetGlobalKeyBindings()
	if len(bindings) != 3 {
		t.Errorf("Expected 3 bindings, got %d", len(bindings))
	}

	// Test Ctrl+O
	if !handleGlobalKeys(tty.KeyEvent{Key: 15}) { // Ctrl+O = 15
		t.Error("Expected Ctrl+O to be handled")
	}
	if !ctrlOCalled {
		t.Error("Ctrl+O handler not called")
	}

	// Test Ctrl+K
	if !handleGlobalKeys(tty.KeyEvent{Key: 11}) { // Ctrl+K = 11
		t.Error("Expected Ctrl+K to be handled")
	}
	if !ctrlKCalled {
		t.Error("Ctrl+K handler not called")
	}

	// Test F1
	if !handleGlobalKeys(tty.KeyEvent{Code: "[OP"}) {
		t.Error("Expected F1 to be handled")
	}
	if !f1Called {
		t.Error("F1 handler not called")
	}

	// Test unhandled key
	if handleGlobalKeys(tty.KeyEvent{Key: 'a'}) {
		t.Error("Regular 'a' should not be handled by global handlers")
	}

	// Clean up
	UnregisterAllGlobalKeys()
	if len(GetGlobalKeyBindings()) != 0 {
		t.Error("Expected all bindings to be cleared")
	}
}

func TestKeyHandlerReturnValue(t *testing.T) {
	UnregisterAllGlobalKeys()

	// Handler that returns false (doesn't consume event)
	RegisterGlobalKey('x', true, "Test X", func(event tty.KeyEvent) bool {
		return false
	})

	// Should return false because handler returned false
	if handleGlobalKeys(tty.KeyEvent{Key: 24}) { // Ctrl+X = 24
		t.Error("Expected handleGlobalKeys to return false when handler returns false")
	}

	UnregisterAllGlobalKeys()

	// Handler that returns true (consumes event)
	RegisterGlobalKey('y', true, "Test Y", func(event tty.KeyEvent) bool {
		return true
	})

	// Should return true because handler returned true
	if !handleGlobalKeys(tty.KeyEvent{Key: 25}) { // Ctrl+Y = 25
		t.Error("Expected handleGlobalKeys to return true when handler returns true")
	}

	UnregisterAllGlobalKeys()
}
