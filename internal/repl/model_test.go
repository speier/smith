package repl

import (
	"testing"

	"github.com/speier/smith/internal/safety"
)

func TestNewBubbleModel(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	if model.projectPath != "/test/path" {
		t.Errorf("Expected projectPath '/test/path', got '%s'", model.projectPath)
	}

	if model.width != 80 {
		t.Errorf("Expected default width 80, got %d", model.width)
	}

	if model.height != 24 {
		t.Errorf("Expected default height 24, got %d", model.height)
	}

	if model.autoLevel != safety.AutoLevelMedium {
		t.Errorf("Expected default autoLevel Medium, got %s", model.autoLevel)
	}
}

func TestCycleAutoLevel(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	// Start at Medium (default)
	if model.autoLevel != safety.AutoLevelMedium {
		t.Fatalf("Expected starting level Medium, got %s", model.autoLevel)
	}

	// Cycle to High
	model.cycleAutoLevel()
	if model.autoLevel != safety.AutoLevelHigh {
		t.Errorf("Expected High after first cycle, got %s", model.autoLevel)
	}

	// Cycle to Low
	model.cycleAutoLevel()
	if model.autoLevel != safety.AutoLevelLow {
		t.Errorf("Expected Low after second cycle, got %s", model.autoLevel)
	}

	// Cycle back to Medium
	model.cycleAutoLevel()
	if model.autoLevel != safety.AutoLevelMedium {
		t.Errorf("Expected Medium after third cycle, got %s", model.autoLevel)
	}
}

func TestGetAutoLevelDisplay(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	tests := []struct {
		level    string
		expected string
	}{
		{safety.AutoLevelLow, "Low"},
		{safety.AutoLevelMedium, "Medium"},
		{safety.AutoLevelHigh, "High"},
	}

	for _, tt := range tests {
		model.autoLevel = tt.level
		result := model.getAutoLevelDisplay()
		if result != tt.expected {
			t.Errorf("For level %s, expected '%s', got '%s'", tt.level, tt.expected, result)
		}
	}
}

func TestHandleSlashCommand(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	tests := []struct {
		command      string
		shouldQuit   bool
		shouldAddMsg bool
		msgContains  string
	}{
		{"/quit", true, false, ""},
		{"/exit", true, false, ""},
		{"/help", false, true, "Commands:"},
		{"/status", false, true, "Task Status:"},
		{"/auth", false, true, "Authentication"},
		{"/unknown", false, true, "/help"},
	}

	for _, tt := range tests {
		model.quitting = false
		oldCount := len(model.messages)

		cmd := model.handleSlashCommand(tt.command)

		if model.quitting != tt.shouldQuit {
			t.Errorf("Command '%s': expected quitting=%v, got %v", tt.command, tt.shouldQuit, model.quitting)
		}

		// Check if quit command was returned
		if tt.shouldQuit && cmd == nil {
			t.Errorf("Command '%s': expected tea.Quit command to be returned", tt.command)
		}

		if tt.shouldAddMsg {
			if len(model.messages) <= oldCount {
				t.Errorf("Command '%s': expected message to be added", tt.command)
			} else {
				lastMsg := model.messages[len(model.messages)-1]
				if tt.msgContains != "" && !contains(lastMsg, tt.msgContains) {
					t.Errorf("Command '%s': expected message to contain '%s', got '%s'", tt.command, tt.msgContains, lastMsg)
				}
			}
		}
	}
}

func TestRenderHistory(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	// Empty history
	history := model.renderHistory(10, 80)
	if history == "" {
		t.Error("Expected non-empty history (should show welcome or placeholder)")
	}

	// Add message
	model.messages = append(model.messages, "Test message")
	history = model.renderHistory(10, 80)
	if !contains(history, "Test") {
		t.Error("Expected history to contain test message")
	}
}

func TestView(t *testing.T) {
	model, err := NewBubbleModel("/test/path", "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view output")
	}

	// Should contain project path
	if !contains(view, "/test/path") {
		t.Error("Expected view to show project path")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsInString(s, substr))
}

func containsInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
