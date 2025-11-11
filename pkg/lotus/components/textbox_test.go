package components

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
)

func TestTextBox_New(t *testing.T) {
	tb := NewTextBox("test-box")

	if tb.ID != "test-box" {
		t.Errorf("Expected ID 'test-box', got '%s'", tb.ID)
	}

	if len(tb.Lines) != 0 {
		t.Error("Expected empty lines initially")
	}

	if tb.ScrollY != 0 {
		t.Error("Expected scroll at 0")
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
	tb.ScrollY = 5

	tb.Clear()

	if len(tb.Lines) != 0 {
		t.Error("Expected lines to be cleared")
	}

	if tb.ScrollY != 0 {
		t.Error("Expected scroll to be reset")
	}
}

func TestTextBox_ScrollUp(t *testing.T) {
	tb := NewTextBox().WithHeight(5)
	tb.Lines = make([]string, 20)
	tb.ScrollY = 10

	tb.ScrollUp()

	if tb.ScrollY != 9 {
		t.Errorf("Expected scroll at 9, got %d", tb.ScrollY)
	}

	// Can't scroll past 0
	tb.ScrollY = 0
	tb.ScrollUp()
	if tb.ScrollY != 0 {
		t.Error("Should not scroll past 0")
	}
}

func TestTextBox_ScrollDown(t *testing.T) {
	tb := NewTextBox().WithHeight(5)
	tb.Lines = make([]string, 20) // 20 lines, height 5, max scroll = 15
	tb.ScrollY = 10

	tb.ScrollDown()

	if tb.ScrollY != 11 {
		t.Errorf("Expected scroll at 11, got %d", tb.ScrollY)
	}

	// Can't scroll past max
	tb.ScrollY = 15
	tb.ScrollDown()
	if tb.ScrollY != 15 {
		t.Errorf("Should not scroll past max (15), got %d", tb.ScrollY)
	}
}

func TestTextBox_ScrollPageUp(t *testing.T) {
	tb := NewTextBox().WithHeight(10)
	tb.Lines = make([]string, 50)
	tb.ScrollY = 25

	tb.ScrollPageUp()

	if tb.ScrollY != 15 {
		t.Errorf("Expected scroll at 15 (25-10), got %d", tb.ScrollY)
	}

	// Clamps at 0
	tb.ScrollY = 5
	tb.ScrollPageUp()
	if tb.ScrollY != 0 {
		t.Errorf("Expected scroll at 0, got %d", tb.ScrollY)
	}
}

func TestTextBox_ScrollPageDown(t *testing.T) {
	tb := NewTextBox().WithHeight(10)
	tb.Lines = make([]string, 50) // Max scroll = 40
	tb.ScrollY = 10

	tb.ScrollPageDown()

	if tb.ScrollY != 20 {
		t.Errorf("Expected scroll at 20 (10+10), got %d", tb.ScrollY)
	}

	// Clamps at max
	tb.ScrollY = 35
	tb.ScrollPageDown()
	if tb.ScrollY != 40 {
		t.Errorf("Expected scroll at 40 (max), got %d", tb.ScrollY)
	}
}

func TestTextBox_ScrollToTop(t *testing.T) {
	tb := NewTextBox().WithHeight(10)
	tb.Lines = make([]string, 50)
	tb.ScrollY = 25

	tb.ScrollToTop()

	if tb.ScrollY != 0 {
		t.Errorf("Expected scroll at 0, got %d", tb.ScrollY)
	}
}

func TestTextBox_ScrollToBottom(t *testing.T) {
	tb := NewTextBox().WithHeight(10)
	tb.Lines = make([]string, 50) // Max scroll = 40

	tb.ScrollToBottom()

	if tb.ScrollY != 40 {
		t.Errorf("Expected scroll at 40, got %d", tb.ScrollY)
	}
}

func TestTextBox_GetVisibleLines(t *testing.T) {
	tb := NewTextBox().WithHeight(3)
	tb.Lines = []string{"Line 0", "Line 1", "Line 2", "Line 3", "Line 4"}

	// At scroll 0, should see first 3 lines
	visible := tb.GetVisibleLines()
	if len(visible) != 3 {
		t.Errorf("Expected 3 visible lines, got %d", len(visible))
	}
	if visible[0] != "Line 0" {
		t.Errorf("Expected 'Line 0', got '%s'", visible[0])
	}

	// Scroll to line 2
	tb.ScrollY = 2
	visible = tb.GetVisibleLines()
	if len(visible) != 3 {
		t.Errorf("Expected 3 visible lines, got %d", len(visible))
	}
	if visible[0] != "Line 2" {
		t.Errorf("Expected 'Line 2', got '%s'", visible[0])
	}
	if visible[2] != "Line 4" {
		t.Errorf("Expected 'Line 4', got '%s'", visible[2])
	}
}

func TestTextBox_AutoScroll(t *testing.T) {
	tb := NewTextBox().WithHeight(3).WithAutoScroll(true)

	// Add lines - should auto-scroll to bottom
	tb.AppendLine("Line 1")
	tb.AppendLine("Line 2")
	tb.AppendLine("Line 3")
	tb.AppendLine("Line 4")

	// Should be scrolled to show last lines
	maxScroll := tb.getMaxScrollY()
	if tb.ScrollY != maxScroll {
		t.Errorf("Expected auto-scroll to max (%d), got %d", maxScroll, tb.ScrollY)
	}
}

func TestTextBox_HandleKeyEvent(t *testing.T) {
	tb := NewTextBox().WithHeight(5).WithFocusable(true) // Enable focusable for keyboard scrolling
	tb.Lines = make([]string, 20)
	tb.ScrollY = 10

	tests := []struct {
		name     string
		event    tty.KeyEvent
		expected int
	}{
		{"up_arrow", tty.KeyEvent{Code: tty.SeqUp}, 9},
		{"down_arrow", tty.KeyEvent{Code: tty.SeqDown}, 11}, // From 10, scrolls down to 11
		{"vi_k", tty.KeyEvent{Char: "k", Key: 'k'}, 9},
		{"vi_j", tty.KeyEvent{Char: "j", Key: 'j'}, 11}, // From 10, scrolls down to 11
		{"home", tty.KeyEvent{Code: tty.SeqHome}, 0},
		{"end", tty.KeyEvent{Code: tty.SeqEnd}, 15},
		{"vi_g", tty.KeyEvent{Char: "g", Key: 'g'}, 0},
		{"vi_G", tty.KeyEvent{Char: "G", Key: 'G'}, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tb.ScrollY = 10 // Reset to middle
			handled := tb.HandleKeyEvent(tt.event)
			if !handled {
				t.Error("Event should be handled")
			}
			if tb.ScrollY != tt.expected {
				t.Errorf("Expected scroll at %d, got %d", tt.expected, tb.ScrollY)
			}
		})
	}
}

func TestTextBox_WithContent(t *testing.T) {
	tb := NewTextBox().WithContent("Line 1\nLine 2\nLine 3")

	if len(tb.Lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(tb.Lines))
	}

	if tb.Lines[1] != "Line 2" {
		t.Errorf("Expected 'Line 2', got '%s'", tb.Lines[1])
	}
}

func TestTextBox_FluentAPI(t *testing.T) {
	tb := NewTextBox().
		WithHeight(10).
		WithWidth(50).
		WithAutoScroll(true).
		WithFocusable(true).
		WithLines([]string{"Test"})

	if tb.Height != 10 {
		t.Errorf("Expected height 10, got %d", tb.Height)
	}

	if tb.Width != 50 {
		t.Errorf("Expected width 50, got %d", tb.Width)
	}

	if !tb.AutoScroll {
		t.Error("Expected auto-scroll enabled")
	}

	if !tb.Focusable {
		t.Error("Expected focusable enabled")
	}

	if len(tb.Lines) != 1 {
		t.Error("Expected lines to be set")
	}
}

func TestTextBox_Focusable(t *testing.T) {
	// Default: not focusable
	tb := NewTextBox()
	if tb.IsFocusable() {
		t.Error("Expected TextBox to be non-focusable by default")
	}

	// Enable focusable
	tb.WithFocusable(true)
	if !tb.IsFocusable() {
		t.Error("Expected TextBox to be focusable when enabled")
	}
}

func TestTextBox_HorizontalScroll(t *testing.T) {
	tb := NewTextBox().WithHeight(3).WithWidth(10)
	tb.Lines = []string{
		"This is a very long line that exceeds the width",
		"Short",
		"Another long line here",
	}
	tb.ScrollX = 5

	visible := tb.GetVisibleLines()

	// Each line should be cut from position 5
	if visible[0] != "is a very long line that exceeds the width" {
		t.Errorf("Unexpected horizontal scroll result: '%s'", visible[0])
	}

	if visible[1] != "" {
		t.Errorf("Expected empty after scroll, got '%s'", visible[1])
	}
}

func TestTextBox_Performance_LargeContent(t *testing.T) {
	tb := NewTextBox().WithHeight(20)

	// Add 10,000 lines
	for i := 0; i < 10000; i++ {
		tb.AppendLine("Line content here")
	}

	// Getting visible lines should be fast (only processes visible range)
	visible := tb.GetVisibleLines()

	// Should only return 20 lines (the viewport)
	if len(visible) != 20 {
		t.Errorf("Expected 20 visible lines, got %d", len(visible))
	}

	// Scroll to middle - still should be fast
	tb.ScrollY = 5000
	visible = tb.GetVisibleLines()

	if len(visible) != 20 {
		t.Errorf("Expected 20 visible lines at middle, got %d", len(visible))
	}
}
