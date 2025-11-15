package lotusui

import (
	"testing"
)

func TestTextBox_New(t *testing.T) {
	tb := NewTextBox("test-box")

	if tb.ID != "test-box" {
		t.Errorf("Expected ID 'test-box', got '%s'", tb.ID)
	}

	if len(tb.Lines) != 0 {
		t.Error("Expected empty lines initially")
	}
}

func TestTextBox_SetContent(t *testing.T) {
	tb := NewTextBox().WithHeight(5)

	lines := []string{"Line 1", "Line 2", "Line 3"}
	tb.SetContent(lines)

	if len(tb.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(tb.Lines))
	}

	if tb.Lines[0] != "Line 1" {
		t.Errorf("Expected 'Line 1', got '%s'", tb.Lines[0])
	}
}

func TestTextBox_AppendLine(t *testing.T) {
	tb := NewTextBox()

	tb.AppendLine("First")
	tb.AppendLine("Second")

	if len(tb.Lines) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(tb.Lines))
	}

	if tb.Lines[1] != "Second" {
		t.Errorf("Expected 'Second', got '%s'", tb.Lines[1])
	}
}

func TestTextBox_Clear(t *testing.T) {
	tb := NewTextBox()
	tb.Lines = []string{"Line 1", "Line 2"}

	tb.Clear()

	if len(tb.Lines) != 0 {
		t.Error("Expected lines to be cleared")
	}
}

func TestTextBox_Render(t *testing.T) {
	tb := NewTextBox()
	tb.AppendLine("Line 1")
	tb.AppendLine("Line 2")

	elem := tb.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}
}

func TestTextBox_RenderEmpty(t *testing.T) {
	tb := NewTextBox()

	elem := tb.Render()
	if elem == nil {
		t.Fatal("Render should not return nil for empty textbox")
	}
}

func TestTextBox_FluentAPI(t *testing.T) {
	tb := NewTextBox().
		WithHeight(20).
		WithWidth(80).
		WithFocusable(true)

	if tb.Height != 20 {
		t.Errorf("Expected height=20, got %d", tb.Height)
	}
	if tb.Width != 80 {
		t.Errorf("Expected width=80, got %d", tb.Width)
	}
	if !tb.Focusable {
		t.Error("Expected Focusable enabled")
	}
}

func TestTextBox_ScrollMethods(t *testing.T) {
	tb := NewTextBox()

	// Just verify methods don't panic (ScrollView handles the logic)
	tb.ScrollUp()
	tb.ScrollDown()
	tb.ScrollToTop()
	tb.ScrollToBottom()
	tb.ScrollPageUp()
	tb.ScrollPageDown()

	// Test passes if no panic
}
