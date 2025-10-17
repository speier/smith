package repl

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/speier/smith/internal/safety"
)

func TestNewBubbleModel(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewBubbleModel(tmpDir, "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	if model.projectPath != tmpDir {
		t.Errorf("Expected projectPath tmpDir, got '%s'", model.projectPath)
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
	tmpDir := t.TempDir()
	model, err := NewBubbleModel(tmpDir, "")
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
	tmpDir := t.TempDir()
	model, err := NewBubbleModel(tmpDir, "")
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
	tmpDir := t.TempDir()
	model, err := NewBubbleModel(tmpDir, "")
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
				if tt.msgContains != "" && !contains(lastMsg.Content, tt.msgContains) {
					t.Errorf("Command '%s': expected message to contain '%s', got '%s'", tt.command, tt.msgContains, lastMsg.Content)
				}
			}
		}
	}
}

func TestRenderHistory(t *testing.T) {
	tmpDir := t.TempDir()
	model, err := NewBubbleModel(tmpDir, "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	// Empty history
	history := model.renderHistory(10, 80)
	if history == "" {
		t.Error("Expected non-empty history (should show welcome or placeholder)")
	}

	// Add message
	model.messages = append(model.messages, Message{Content: "Test message", Type: "system", Timestamp: time.Now()})
	history = model.renderHistory(10, 80)
	if !contains(history, "Test") {
		t.Error("Expected history to contain test message")
	}
}

func TestView(t *testing.T) {
	// Create temp dir with minimal config so settings don't auto-open
	tmpDir, err := os.MkdirTemp("", "smith-view-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Create minimal local config
	smithDir := filepath.Join(tmpDir, ".smith")
	if err := os.MkdirAll(smithDir, 0755); err != nil {
		t.Fatalf("failed to create .smith dir: %v", err)
	}

	configPath := filepath.Join(smithDir, "config.yaml")
	configContent := `provider: copilot
model: gpt-4o
autoLevel: medium
version: 1
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	model, err := NewBubbleModel(tmpDir, "")
	if err != nil {
		t.Fatalf("NewBubbleModel failed: %v", err)
	}

	view := model.View()
	if view == "" {
		t.Error("Expected non-empty view output")
	}

	// Should show SMITH branding
	if !contains(view, "SMITH") {
		t.Error("Expected view to show SMITH branding")
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
