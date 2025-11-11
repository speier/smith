package components

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
)

func TestTextInput_InsertNewline(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5

	input.InsertNewline()

	if input.Value != "Hello\n" {
		t.Errorf("Expected 'Hello\\n', got '%s'", input.Value)
	}
	if input.CursorPos != 6 {
		t.Errorf("Expected cursor at 6, got %d", input.CursorPos)
	}
}

func TestTextInput_MultiLineNavigation(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Line 1\nLine 2\nLine 3"
	input.CursorPos = 0 // Start of first line

	// Move down to second line
	input.MoveDown()
	line, col := input.getCurrentLineAndCol()
	if line != 1 || col != 0 {
		t.Errorf("Expected line=1, col=0, got line=%d, col=%d", line, col)
	}

	// Move down to third line
	input.MoveDown()
	line, col = input.getCurrentLineAndCol()
	if line != 2 || col != 0 {
		t.Errorf("Expected line=2, col=0, got line=%d, col=%d", line, col)
	}

	// Move up to second line
	input.MoveUp()
	line, col = input.getCurrentLineAndCol()
	if line != 1 || col != 0 {
		t.Errorf("Expected line=1, col=0, got line=%d, col=%d", line, col)
	}

	// Move up to first line
	input.MoveUp()
	line, col = input.getCurrentLineAndCol()
	if line != 0 || col != 0 {
		t.Errorf("Expected line=0, col=0, got line=%d, col=%d", line, col)
	}

	// Try to move up past first line (should stay at first line)
	input.MoveUp()
	line, col = input.getCurrentLineAndCol()
	if line != 0 || col != 0 {
		t.Errorf("Expected line=0, col=0, got line=%d, col=%d", line, col)
	}
}

func TestTextInput_MultiLineNavigationWithColumn(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello World\nHi\nGoodbye"
	input.CursorPos = 5 // Middle of "Hello World"

	// Move down - should preserve column 5 on shorter line (goes to end)
	input.MoveDown()
	line, col := input.getCurrentLineAndCol()
	if line != 1 || col != 2 { // "Hi" is only 2 chars, so col = 2 (end)
		t.Errorf("Expected line=1, col=2, got line=%d, col=%d", line, col)
	}

	// Move down again - should preserve column 5 on longer line
	input.MoveDown()
	line, col = input.getCurrentLineAndCol()
	if line != 2 || col != 5 { // "Goodbye" is long enough for col 5
		t.Errorf("Expected line=2, col=5, got line=%d, col=%d", line, col)
	}
}

func TestTextInput_GetLines(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{
			name:     "empty",
			value:    "",
			expected: []string{""},
		},
		{
			name:     "single_line",
			value:    "Hello",
			expected: []string{"Hello"},
		},
		{
			name:     "two_lines",
			value:    "Line 1\nLine 2",
			expected: []string{"Line 1", "Line 2"},
		},
		{
			name:     "three_lines",
			value:    "A\nB\nC",
			expected: []string{"A", "B", "C"},
		},
		{
			name:     "trailing_newline",
			value:    "Hello\n",
			expected: []string{"Hello", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewTextInput().WithWidth(20)
			input.Value = tt.value

			lines := input.getLines()
			if len(lines) != len(tt.expected) {
				t.Errorf("Expected %d lines, got %d", len(tt.expected), len(lines))
				return
			}

			for i, expected := range tt.expected {
				if lines[i] != expected {
					t.Errorf("Line %d: expected '%s', got '%s'", i, expected, lines[i])
				}
			}
		})
	}
}

func TestTextInput_GetCurrentLineAndCol(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		cursorPos    int
		expectedLine int
		expectedCol  int
	}{
		{
			name:         "empty",
			value:        "",
			cursorPos:    0,
			expectedLine: 0,
			expectedCol:  0,
		},
		{
			name:         "first_line_start",
			value:        "Hello\nWorld",
			cursorPos:    0,
			expectedLine: 0,
			expectedCol:  0,
		},
		{
			name:         "first_line_end",
			value:        "Hello\nWorld",
			cursorPos:    5,
			expectedLine: 0,
			expectedCol:  5,
		},
		{
			name:         "second_line_start",
			value:        "Hello\nWorld",
			cursorPos:    6,
			expectedLine: 1,
			expectedCol:  0,
		},
		{
			name:         "second_line_middle",
			value:        "Hello\nWorld",
			cursorPos:    8,
			expectedLine: 1,
			expectedCol:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewTextInput().WithWidth(20)
			input.Value = tt.value
			input.CursorPos = tt.cursorPos

			line, col := input.getCurrentLineAndCol()
			if line != tt.expectedLine {
				t.Errorf("Expected line %d, got %d", tt.expectedLine, line)
			}
			if col != tt.expectedCol {
				t.Errorf("Expected col %d, got %d", tt.expectedCol, col)
			}
		})
	}
}

func TestTextInput_ShiftEnterInsertsNewline(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5

	// Shift+Enter inserts newline
	event := tty.KeyEvent{Key: 27, Code: tty.SeqShiftEnter}
	handled := input.HandleKey(event)

	if !handled {
		t.Error("Expected Shift+Enter to be handled")
	}

	if input.Value != "Hello\n" {
		t.Errorf("Expected 'Hello\\n', got '%s'", input.Value)
	}

	if input.CursorPos != 6 {
		t.Errorf("Expected cursor at 6, got %d", input.CursorPos)
	}
}

func TestTextInput_EnterStillSubmits(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5

	submitted := false
	input.OnSubmit = func(value string) {
		submitted = true
	}

	// Regular Enter submits
	event := tty.KeyEvent{Key: 13} // Enter
	input.HandleKey(event)

	if !submitted {
		t.Error("Expected Enter to trigger submit")
	}
}

func TestTextInput_MultiLineRendering(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Focused = true // Set focused to test cursor rendering
	input.Value = "Line 1\nLine 2"
	input.CursorPos = 0

	elem := input.Render()

	// Root should be Box
	if elem.Tag != "box" {
		t.Errorf("Expected box element, got %s", elem.Tag)
	}

	// Should have VStack for multi-line
	if len(elem.Children) != 1 {
		t.Fatalf("Expected 1 child (VStack), got %d", len(elem.Children))
	}

	vstack := elem.Children[0]
	if vstack.Tag != "box" {
		t.Errorf("Expected box (VStack), got %s", vstack.Tag)
	}

	// VStack should have flex-direction: column
	if vstack.Props.Styles["flex-direction"] != "column" {
		t.Error("Expected flex-direction: column for VStack")
	}

	// Should have 2 lines rendered
	if len(vstack.Children) != 2 {
		t.Errorf("Expected 2 lines, got %d", len(vstack.Children))
	}

	t.Logf("✓ Multi-line rendering works!")
}

func TestTextInput_UpDownArrowKeys(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Line 1\nLine 2\nLine 3"
	input.CursorPos = 0

	// Down arrow
	downEvent := tty.KeyEvent{Key: 27, Code: tty.SeqDown}
	handled := input.HandleKey(downEvent)
	if !handled {
		t.Error("Expected Down arrow to be handled")
	}

	line, _ := input.getCurrentLineAndCol()
	if line != 1 {
		t.Errorf("Expected line 1 after Down, got %d", line)
	}

	// Up arrow
	upEvent := tty.KeyEvent{Key: 27, Code: tty.SeqUp}
	handled = input.HandleKey(upEvent)
	if !handled {
		t.Error("Expected Up arrow to be handled")
	}

	line, _ = input.getCurrentLineAndCol()
	if line != 0 {
		t.Errorf("Expected line 0 after Up, got %d", line)
	}

	t.Logf("✓ Up/Down arrow navigation works!")
}
