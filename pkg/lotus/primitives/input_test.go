package primitives

import (
	"testing"
	"time"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestInput_InsertChar(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)

	input.InsertChar("H")
	input.InsertChar("i")

	if input.Value != "Hi" {
		t.Errorf("Expected 'Hi', got '%s'", input.Value)
	}

	if input.CursorPos != 2 {
		t.Errorf("Expected cursor at 2, got %d", input.CursorPos)
	}
}

func TestInput_DeleteChar(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_DeleteForward(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_MoveLeft(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_MoveRight(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_HomeEnd(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_MoveWordLeft(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_MoveWordRight(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_GetVisible(t *testing.T) {
	input := NewInput("test-input").WithWidth(5) // Only 5 chars visible
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

func TestInput_Clear(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

func TestInput_HandleKey(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)

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

func TestInput_GetDisplay(t *testing.T) {
	input := NewInput("test-input").WithWidth(20)
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

// --- CURSOR TESTS (100% testable!) ---

func TestInput_CursorStyle(t *testing.T) {
	input := NewInput()

	// Default is block
	if input.CursorStyle != CursorBlock {
		t.Errorf("Expected default CursorBlock, got %v", input.CursorStyle)
	}

	// Test changing styles
	input.SetCursorStyle(CursorUnderline)
	if input.CursorStyle != CursorUnderline {
		t.Error("Expected CursorUnderline")
	}

	input.SetCursorStyle(CursorBar)
	if input.CursorStyle != CursorBar {
		t.Error("Expected CursorBar")
	}
}

func TestInput_CursorChar(t *testing.T) {
	input := NewInput()
	input.CursorVisible = true

	tests := []struct {
		style    CursorStyle
		expected string
	}{
		{CursorBlock, "█"},
		{CursorUnderline, "_"},
		{CursorBar, "|"},
	}

	for _, tt := range tests {
		input.SetCursorStyle(tt.style)
		char := input.GetCursorChar()
		if char != tt.expected {
			t.Errorf("For style %v, expected '%s', got '%s'", tt.style, tt.expected, char)
		}
	}

	// When cursor is invisible, should return space
	input.CursorVisible = false
	char := input.GetCursorChar()
	if char != " " {
		t.Errorf("Expected space for invisible cursor, got '%s'", char)
	}
}

func TestInput_CursorBlink(t *testing.T) {
	input := NewInput()

	// Default is blinking
	if !input.CursorBlink {
		t.Error("Expected cursor to blink by default")
	}

	// Disable blinking
	input.SetCursorBlink(false)
	if input.CursorBlink {
		t.Error("Expected cursor blink to be disabled")
	}
	if !input.CursorVisible {
		t.Error("Cursor should be visible when blink is disabled")
	}

	// Enable blinking
	input.SetCursorBlink(true)
	if !input.CursorBlink {
		t.Error("Expected cursor blink to be enabled")
	}
}

func TestInput_CursorBlinkUpdate(t *testing.T) {
	input := NewInput()
	input.SetBlinkInterval(100 * time.Millisecond)
	input.CursorVisible = true

	// Initially no change (not enough time passed)
	changed := input.UpdateCursorBlink()
	if changed {
		t.Error("Cursor should not have changed immediately")
	}

	// Wait for blink interval
	time.Sleep(110 * time.Millisecond)

	// Should toggle visibility
	changed = input.UpdateCursorBlink()
	if !changed {
		t.Error("Cursor should have toggled")
	}
	if input.CursorVisible {
		t.Error("Cursor should be invisible after first blink")
	}

	// Wait and toggle again
	time.Sleep(110 * time.Millisecond)
	changed = input.UpdateCursorBlink()
	if !changed {
		t.Error("Cursor should have toggled again")
	}
	if !input.CursorVisible {
		t.Error("Cursor should be visible after second blink")
	}
}

func TestInput_CursorBlinkDisabled(t *testing.T) {
	input := NewInput()
	input.SetCursorBlink(false)

	// Should never change
	time.Sleep(600 * time.Millisecond)
	changed := input.UpdateCursorBlink()
	if changed {
		t.Error("Cursor should not blink when blinking is disabled")
	}
}

func TestInput_CursorPosition(t *testing.T) {
	input := NewInput().WithWidth(20)

	// Test cursor at different positions
	tests := []struct {
		value    string
		cursorAt int
	}{
		{"", 0},
		{"H", 1},
		{"Hello", 5},
		{"Hello World", 11},
	}

	for _, tt := range tests {
		input.Value = tt.value
		input.CursorPos = tt.cursorAt

		if input.CursorPos != tt.cursorAt {
			t.Errorf("For value '%s', expected cursor at %d, got %d", tt.value, tt.cursorAt, input.CursorPos)
		}
	}
}

func TestInput_GetDisplayWithCursor(t *testing.T) {
	input := NewInput().WithWidth(20)
	input.SetCursorBlink(false) // Keep cursor always visible
	input.SetCursorStyle(CursorBlock)

	tests := []struct {
		name     string
		value    string
		cursor   int
		expected string
	}{
		{
			name:     "empty_with_cursor",
			value:    "",
			cursor:   0,
			expected: "█",
		},
		{
			name:     "cursor_at_start",
			value:    "Hello",
			cursor:   0,
			expected: "█ello",
		},
		{
			name:     "cursor_in_middle",
			value:    "Hello",
			cursor:   2,
			expected: "He█lo",
		},
		{
			name:     "cursor_at_end",
			value:    "Hello",
			cursor:   5,
			expected: "Hello█",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input.Value = tt.value
			input.CursorPos = tt.cursor
			input.Scroll = 0

			display := input.GetDisplayWithCursor()
			if display != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, display)
			}
		})
	}
}

func TestInput_GetDisplayWithCursor_DifferentStyles(t *testing.T) {
	input := NewInput().WithWidth(20)
	input.Value = "Test"
	input.CursorPos = 2
	input.Scroll = 0
	input.SetCursorBlink(false)

	tests := []struct {
		style    CursorStyle
		expected string
	}{
		{CursorBlock, "Te█t"},
		{CursorUnderline, "Te_t"},
		{CursorBar, "Te|t"},
	}

	for _, tt := range tests {
		input.SetCursorStyle(tt.style)
		display := input.GetDisplayWithCursor()
		if display != tt.expected {
			t.Errorf("For style %v, expected '%s', got '%s'", tt.style, tt.expected, display)
		}
	}
}

func TestInput_Placeholder(t *testing.T) {
	input := NewInput().WithWidth(20)
	input.Placeholder = "Enter text..."

	// Placeholder shown when empty
	display := input.GetDisplay()
	if display != "Enter text..." {
		t.Errorf("Expected placeholder, got '%s'", display)
	}

	// Placeholder hidden when typing
	input.InsertChar("H")
	display = input.GetDisplay()
	if display != "H" {
		t.Errorf("Expected 'H', got '%s'", display)
	}

	// Placeholder shown again after clearing
	input.Clear()
	display = input.GetDisplay()
	if display != "Enter text..." {
		t.Errorf("Expected placeholder after clear, got '%s'", display)
	}
}

func TestInput_TypingFlow(t *testing.T) {
	input := NewInput().WithWidth(20)
	input.Placeholder = "Type here..."
	input.SetCursorBlink(false)

	// Start with placeholder
	if input.GetDisplay() != "Type here..." {
		t.Error("Expected placeholder at start")
	}

	// Type "H"
	input.InsertChar("H")
	if input.Value != "H" {
		t.Errorf("Expected 'H', got '%s'", input.Value)
	}
	if input.CursorPos != 1 {
		t.Errorf("Expected cursor at 1, got %d", input.CursorPos)
	}

	// Type "e"
	input.InsertChar("e")
	if input.Value != "He" {
		t.Errorf("Expected 'He', got '%s'", input.Value)
	}

	// Type "llo"
	input.InsertChar("l")
	input.InsertChar("l")
	input.InsertChar("o")
	if input.Value != "Hello" {
		t.Errorf("Expected 'Hello', got '%s'", input.Value)
	}
	if input.CursorPos != 5 {
		t.Errorf("Expected cursor at 5, got %d", input.CursorPos)
	}

	// Backspace
	input.DeleteChar()
	if input.Value != "Hell" {
		t.Errorf("Expected 'Hell', got '%s'", input.Value)
	}
	if input.CursorPos != 4 {
		t.Errorf("Expected cursor at 4, got %d", input.CursorPos)
	}

	// Move cursor to middle and insert
	input.CursorPos = 2
	input.InsertChar("X")
	if input.Value != "HeXll" {
		t.Errorf("Expected 'HeXll', got '%s'", input.Value)
	}
}

func TestInput_ScrollingWithCursor(t *testing.T) {
	input := NewInput().WithWidth(5) // Only 5 chars visible
	input.SetCursorBlink(false)

	// Type more than visible width
	input.Value = "Hello World"
	input.CursorPos = 11 // At end
	input.adjustScroll()

	// Should be scrolled to show last part
	visible, offset := input.GetVisible()

	// With width 5 and cursor at 11, scroll should be 11-5+1 = 7
	// So visible should be Value[7:12] = "orld" (but limited to length)
	// Actually Value[7:11] = "orld"
	expectedVisible := "orld"
	if visible != expectedVisible {
		t.Errorf("Expected '%s', got '%s' (scroll=%d)", expectedVisible, visible, input.Scroll)
	}

	// Cursor offset should be relative to scroll position
	// CursorPos(11) - Scroll(7) = 4
	expectedOffset := 4
	if offset != expectedOffset {
		t.Errorf("Expected cursor offset %d, got %d", expectedOffset, offset)
	}
}

func TestInput_WithPlaceholder(t *testing.T) {
	input := NewInput().WithPlaceholder("Enter name...")

	if input.Placeholder != "Enter name..." {
		t.Errorf("Expected placeholder 'Enter name...', got '%s'", input.Placeholder)
	}

	// Fluent API should return same instance
	same := input.WithPlaceholder("New placeholder")
	if same != input {
		t.Error("WithPlaceholder should return same instance")
	}
	if input.Placeholder != "New placeholder" {
		t.Error("Placeholder should be updated")
	}
}

func TestInput_RenderWithPlaceholder(t *testing.T) {
	input := NewInput().WithPlaceholder("Say something...")
	input.Focused = true // Set focused to test cursor rendering

	// Test 1: Empty input should render cursor + placeholder in HStack
	elem := input.Render()
	if elem == nil {
		t.Fatal("Render should return an element")
	}

	// Root should be a box (from Render)
	if elem.Tag != "box" {
		t.Errorf("Expected box element, got %s", elem.Tag)
	}

	// Should have 1 child (the HStack)
	if len(elem.Children) != 1 {
		t.Fatalf("Expected 1 child (HStack), got %d", len(elem.Children))
	}

	hstack := elem.Children[0]
	if hstack.Tag != "box" {
		t.Errorf("Expected box (HStack), got %s", hstack.Tag)
	}

	// HStack should have flex-direction: row
	if hstack.Props.Styles["flex-direction"] != "row" {
		t.Errorf("Expected flex-direction: row for HStack")
	}

	// HStack should have 3 children: prompt, styled first char (cursor), and rest of placeholder
	if len(hstack.Children) != 3 {
		t.Errorf("Expected 3 children (prompt + styled char + rest of placeholder), got %d", len(hstack.Children))
	}

	// First child should be prompt text "> "
	promptText := hstack.Children[0]
	if promptText.Type != vdom.TextElement {
		t.Errorf("Expected TextElement for prompt, got %v", promptText.Type)
	}
	if promptText.Text != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", promptText.Text)
	}

	// Second child should be first char with inverse video (cursor effect)
	cursorText := hstack.Children[1]
	if cursorText.Type != vdom.TextElement {
		t.Errorf("Expected TextElement for cursor char, got %v", cursorText.Type)
	}
	if cursorText.Text != "S" {
		t.Errorf("Expected 'S' with cursor styling, got '%s'", cursorText.Text)
	}
	if cursorText.Props.Styles["color"] != "#000000" {
		t.Errorf("Expected #000000 (black) color for cursor char, got '%s'", cursorText.Props.Styles["color"])
	}
	if cursorText.Props.Styles["background-color"] != "#ffffff" {
		t.Errorf("Expected #ffffff background for cursor char (inverse), got '%s'", cursorText.Props.Styles["background-color"])
	}

	// Third child should be rest of placeholder with gray color (#808080)
	placeholderText := hstack.Children[2]
	if placeholderText.Type != vdom.TextElement {
		t.Errorf("Expected TextElement for placeholder, got %v", placeholderText.Type)
	}
	// Note: cursor shows [S], so rest is "ay something..."
	if placeholderText.Text != "ay something..." {
		t.Errorf("Expected placeholder text 'ay something...', got '%s'", placeholderText.Text)
	}
	if placeholderText.Props.Styles["color"] != "#808080" {
		t.Errorf("Expected #808080 color for placeholder, got '%s'", placeholderText.Props.Styles["color"])
	}

	// Test 2: With text, should render with inverse video cursor
	input.InsertChar("H")
	input.InsertChar("i")
	elem = input.Render()

	// Should be Box with HStack
	if len(elem.Children) != 1 {
		t.Errorf("Expected 1 child (HStack), got %d", len(elem.Children))
	}

	hstackWithText := elem.Children[0]
	if hstackWithText.Tag != "box" {
		t.Errorf("Expected box (HStack), got %s", hstackWithText.Tag)
	}

	// HStack should have 3 children: prompt, "Hi", cursor (space)
	// (cursor is at end, so it shows as space with inverse, no afterCursor)
	if len(hstackWithText.Children) != 3 {
		t.Errorf("Expected 3 children (prompt + before + cursor), got %d", len(hstackWithText.Children))
	}

	// First child should be prompt "> "
	promptWithText := hstackWithText.Children[0]
	if promptWithText.Type != vdom.TextElement || promptWithText.Text != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", promptWithText.Text)
	}

	// Second child should be "Hi"
	beforeText := hstackWithText.Children[1]
	if beforeText.Text != "Hi" {
		t.Errorf("Expected 'Hi', got '%s'", beforeText.Text)
	}

	// Third child should be cursor (space with inverse video since cursor at end)
	cursorElem := hstackWithText.Children[2]
	if cursorElem.Props.Styles["background-color"] != "#ffffff" {
		t.Errorf("Expected #ffffff background for cursor, got '%s'", cursorElem.Props.Styles["background-color"])
	}
}

func TestInput_PlaceholderWithPrompt(t *testing.T) {
	input := NewInput().WithPlaceholder("Type here...")
	input.Focused = true // Set focused to test cursor rendering

	elem := input.Render()
	hstack := elem.Children[0]

	// Should have 3 children: prompt, T (underlined), ype here...
	if len(hstack.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(hstack.Children))
	}

	// First: prompt "> "
	promptText := hstack.Children[0]
	if promptText.Text != "> " {
		t.Errorf("Expected prompt '> ', got '%s'", promptText.Text)
	}

	// Second: first char 'T' with inverse video (cursor effect)
	cursorText := hstack.Children[1]
	if cursorText.Text != "T" {
		t.Errorf("Expected 'T' with cursor styling, got '%s'", cursorText.Text)
	}
	if cursorText.Props.Styles["background-color"] != "#ffffff" {
		t.Errorf("Expected #ffffff background for cursor char, got '%s'", cursorText.Props.Styles["background-color"])
	}

	// Third: rest of placeholder
	placeholderText := hstack.Children[2]
	if placeholderText.Text != "ype here..." {
		t.Errorf("Expected placeholder 'ype here...', got '%s'", placeholderText.Text)
	}
}
