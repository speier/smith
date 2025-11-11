package components

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
)

func TestTextInput_DeleteToBeginning(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello World"
	input.CursorPos = 5 // After "Hello"

	input.DeleteToBeginning()

	if input.Value != " World" {
		t.Errorf("Expected ' World', got '%s'", input.Value)
	}
	if input.CursorPos != 0 {
		t.Errorf("Expected cursor at 0, got %d", input.CursorPos)
	}
	if input.Scroll != 0 {
		t.Errorf("Expected scroll at 0, got %d", input.Scroll)
	}
}

func TestTextInput_DeleteToEnd(t *testing.T) {
	input := NewTextInput().WithWidth(20)
	input.Value = "Hello World"
	input.CursorPos = 5 // After "Hello"

	input.DeleteToEnd()

	if input.Value != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", input.Value)
	}
	if input.CursorPos != 5 {
		t.Errorf("Expected cursor at 5, got %d", input.CursorPos)
	}
}

func TestTextInput_DeleteWordBackward(t *testing.T) {
	tests := []struct {
		name           string
		initialValue   string
		initialCursor  int
		expectedValue  string
		expectedCursor int
	}{
		{
			name:           "delete_one_word",
			initialValue:   "Hello World",
			initialCursor:  11,
			expectedValue:  "Hello ",
			expectedCursor: 6,
		},
		{
			name:           "delete_word_with_space",
			initialValue:   "Hello World  ",
			initialCursor:  13,
			expectedValue:  "Hello ",
			expectedCursor: 6,
		},
		{
			name:           "delete_from_middle_of_word",
			initialValue:   "Hello World",
			initialCursor:  8,
			expectedValue:  "Hello rld",
			expectedCursor: 6,
		},
		{
			name:           "delete_at_start",
			initialValue:   "Hello",
			initialCursor:  0,
			expectedValue:  "Hello",
			expectedCursor: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewTextInput().WithWidth(20)
			input.Value = tt.initialValue
			input.CursorPos = tt.initialCursor

			input.DeleteWordBackward()

			if input.Value != tt.expectedValue {
				t.Errorf("Expected '%s', got '%s'", tt.expectedValue, input.Value)
			}
			if input.CursorPos != tt.expectedCursor {
				t.Errorf("Expected cursor at %d, got %d", tt.expectedCursor, input.CursorPos)
			}
		})
	}
}

func TestTextInput_KeyboardShortcuts(t *testing.T) {
	tests := []struct {
		name           string
		initialValue   string
		initialCursor  int
		event          tty.KeyEvent
		expectedValue  string
		expectedCursor int
		description    string
	}{
		{
			name:           "ctrl_u_delete_to_beginning",
			initialValue:   "Hello World",
			initialCursor:  5,
			event:          tty.KeyEvent{Key: '\x15'}, // Ctrl+U
			expectedValue:  " World",
			expectedCursor: 0,
			description:    "Ctrl+U should delete to beginning",
		},
		{
			name:           "ctrl_k_delete_to_end",
			initialValue:   "Hello World",
			initialCursor:  5,
			event:          tty.KeyEvent{Key: '\x0b'}, // Ctrl+K
			expectedValue:  "Hello",
			expectedCursor: 5,
			description:    "Ctrl+K should delete to end",
		},
		{
			name:           "ctrl_w_delete_word_backward",
			initialValue:   "Hello World",
			initialCursor:  11,
			event:          tty.KeyEvent{Key: '\x17'}, // Ctrl+W
			expectedValue:  "Hello ",
			expectedCursor: 6,
			description:    "Ctrl+W should delete word backward",
		},
		{
			name:           "cmd_backspace_delete_to_beginning",
			initialValue:   "Hello World",
			initialCursor:  5,
			event:          tty.KeyEvent{Key: 27, Code: tty.SeqCmdBackspace},
			expectedValue:  " World",
			expectedCursor: 0,
			description:    "Cmd+Backspace should delete to beginning",
		},
		{
			name:           "ctrl_backspace_delete_word",
			initialValue:   "Hello World",
			initialCursor:  11,
			event:          tty.KeyEvent{Key: 27, Code: tty.SeqCtrlBackspace},
			expectedValue:  "Hello ",
			expectedCursor: 6,
			description:    "Ctrl+Backspace should delete word backward",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewTextInput().WithWidth(20)
			input.Value = tt.initialValue
			input.CursorPos = tt.initialCursor

			handled := input.HandleKey(tt.event)

			if !handled {
				t.Error("Expected event to be handled")
			}

			if input.Value != tt.expectedValue {
				t.Errorf("%s: Expected '%s', got '%s'", tt.description, tt.expectedValue, input.Value)
			}

			if input.CursorPos != tt.expectedCursor {
				t.Errorf("%s: Expected cursor at %d, got %d", tt.description, tt.expectedCursor, input.CursorPos)
			}
		})
	}
}

func TestTextInput_PromptAlwaysVisible(t *testing.T) {
	input := NewTextInput().WithPlaceholder("Type here...")
	input.Focused = true // Set focused to test cursor rendering

	// Empty - should show prompt + inverse T + ype here...
	elem := input.Render()
	hstack := elem.Children[0]
	if len(hstack.Children) != 3 {
		t.Fatalf("Expected 3 children for empty input (prompt + T + placeholder), got %d", len(hstack.Children))
	}
	if hstack.Children[0].Text != "> " {
		t.Error("Expected prompt '> ' for empty input")
	}
	if hstack.Children[1].Text != "T" {
		t.Error("Expected 'T' (inverse video) for empty input")
	}
	if hstack.Children[1].Props.Styles["background-color"] != "#ffffff" {
		t.Error("Expected white background for cursor char (bright, visible)")
	}
	if hstack.Children[2].Text != "ype here..." {
		t.Error("Expected placeholder 'ype here...' for empty input")
	}

	// With text - should still show prompt + inverse video cursor
	input.InsertChar("H")
	input.InsertChar("i")
	elem = input.Render()
	hstack = elem.Children[0]
	// Should have 3 children: prompt, "Hi", cursor (space with inverse, no afterCursor)
	if len(hstack.Children) != 3 {
		t.Fatalf("Expected 3 children for input with text (prompt + before + cursor), got %d", len(hstack.Children))
	}
	if hstack.Children[0].Text != "> " {
		t.Error("Expected prompt '> ' for input with text")
	}
	if hstack.Children[1].Text != "Hi" {
		t.Errorf("Expected 'Hi', got '%s'", hstack.Children[1].Text)
	}
	// Cursor should have inverse video background
	if hstack.Children[2].Props.Styles["background-color"] != "#ffffff" {
		t.Error("Expected white background for cursor in text (bright, visible)")
	}

	t.Logf("✓ Prompt always visible")
}
