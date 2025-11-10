package components

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/tty"
)

func TestTextInput_InsertChar(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)

	input.InsertChar("H")
	input.InsertChar("i")

	if input.Value != "Hi" {
		t.Errorf("Expected 'Hi', got '%s'", input.Value)
	}

	if input.CursorPos != 2 {
		t.Errorf("Expected cursor at 2, got %d", input.CursorPos)
	}
}

func TestTextInput_DeleteChar(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5

	input.DeleteChar()

	if input.Value != "Hell" {
		t.Errorf("Expected 'Hell', got '%s'", input.Value)
	}

	if input.CursorPos != 4 {
		t.Errorf("Expected cursor at 4, got %d", input.CursorPos)
	}
}

func TestTextInput_DeleteForward(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 1

	input.DeleteForward()

	if input.Value != "Hllo" {
		t.Errorf("Expected 'Hllo', got '%s'", input.Value)
	}

	if input.CursorPos != 1 {
		t.Errorf("Expected cursor at 1, got %d", input.CursorPos)
	}
}

func TestTextInput_MoveLeft(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5

	input.MoveLeft()

	if input.CursorPos != 4 {
		t.Errorf("Expected cursor at 4, got %d", input.CursorPos)
	}

	// Can't move past beginning
	input.CursorPos = 0
	input.MoveLeft()
	if input.CursorPos != 0 {
		t.Errorf("Expected cursor to stay at 0, got %d", input.CursorPos)
	}
}

func TestTextInput_MoveRight(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 0

	input.MoveRight()

	if input.CursorPos != 1 {
		t.Errorf("Expected cursor at 1, got %d", input.CursorPos)
	}

	// Can't move past end
	input.CursorPos = 5
	input.MoveRight()
	if input.CursorPos != 5 {
		t.Errorf("Expected cursor to stay at 5, got %d", input.CursorPos)
	}
}

func TestTextInput_HomeEnd(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello World"
	input.CursorPos = 6

	input.Home()
	if input.CursorPos != 0 {
		t.Errorf("Expected cursor at 0, got %d", input.CursorPos)
	}

	input.End()
	if input.CursorPos != 11 {
		t.Errorf("Expected cursor at 11, got %d", input.CursorPos)
	}
}

func TestTextInput_MoveWordLeft(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello World Test"
	input.CursorPos = 16 // At end

	input.MoveWordLeft()
	if input.CursorPos != 12 {
		t.Errorf("Expected cursor at 12 (start of 'Test'), got %d", input.CursorPos)
	}

	input.MoveWordLeft()
	if input.CursorPos != 6 {
		t.Errorf("Expected cursor at 6 (start of 'World'), got %d", input.CursorPos)
	}
}

func TestTextInput_MoveWordRight(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello World Test"
	input.CursorPos = 0

	input.MoveWordRight()
	if input.CursorPos != 5 {
		t.Errorf("Expected cursor at 5 (end of 'Hello'), got %d", input.CursorPos)
	}

	input.MoveWordRight()
	if input.CursorPos != 11 {
		t.Errorf("Expected cursor at 11 (end of 'World'), got %d", input.CursorPos)
	}
}

func TestTextInput_GetVisible(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(5) // Only 5 chars visible
	input.Value = "Hello World"

	// At start
	input.CursorPos = 0
	visible, offset := input.GetVisible()
	if visible != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", visible)
	}
	if offset != 0 {
		t.Errorf("Expected offset 0, got %d", offset)
	}

	// Scroll to show "World"
	input.CursorPos = 10
	input.adjustScroll()
	visible, _ = input.GetVisible()
	if visible != "World" {
		t.Errorf("Expected 'World', got '%s'", visible)
	}
}

func TestTextInput_Clear(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Value = "Hello"
	input.CursorPos = 5
	input.Scroll = 2

	input.Clear()

	if input.Value != "" {
		t.Errorf("Expected empty value, got '%s'", input.Value)
	}
	if input.CursorPos != 0 {
		t.Errorf("Expected cursor at 0, got %d", input.CursorPos)
	}
	if input.Scroll != 0 {
		t.Errorf("Expected scroll at 0, got %d", input.Scroll)
	}
}

func TestTextInput_HandleKey(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)

	// Test printable character
	event := tty.KeyEvent{Key: 'a', Char: "a"}
	handled := input.HandleKey(event)
	if !handled {
		t.Error("Expected printable char to be handled")
	}
	if input.Value != "a" {
		t.Errorf("Expected 'a', got '%s'", input.Value)
	}

	// Test Enter (should NOT be handled)
	event = tty.KeyEvent{Key: tty.KeyEnter}
	handled = input.HandleKey(event)
	if handled {
		t.Error("Expected Enter to NOT be handled")
	}
}

func TestTextInput_GetDisplay(t *testing.T) {
	input := NewTextInput("test-input").WithWidth(20)
	input.Placeholder = "Type here..."

	// Empty input shows placeholder
	display := input.GetDisplay()
	if display != "Type here..." {
		t.Errorf("Expected placeholder, got '%s'", display)
	}

	// With text, shows text
	input.Value = "Hello"
	display = input.GetDisplay()
	if display != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", display)
	}
}
