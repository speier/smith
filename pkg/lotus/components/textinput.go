// Package component provides reusable UI components for Lotus applications
package components

import (
	"github.com/speier/smith/pkg/lotus/tty"
)

// TextInputProps represents the props (configuration from parent) for TextInput
// React pattern: Props are immutable configuration passed by parent
type TextInputProps struct {
	Placeholder string
	Width       int
	Disabled    bool
	OnChange    func(value string)
	OnSubmit    func(value string)
}

// TextInput is a reusable text input field with cursor navigation and horizontal scrolling
type TextInput struct {
	// Component metadata
	ID string // Component ID for registration (React key)

	// Internal state (private to component)
	Value     string // Current input text
	CursorPos int    // Cursor position in the full text (0-indexed)
	Scroll    int    // Horizontal scroll offset

	// Props (configuration from parent)
	Width       int    // Visible width (set automatically from layout)
	Placeholder string // Placeholder text when empty
	Disabled    bool   // If true, component cannot receive focus

	// Event callbacks (React-like props)
	OnChange func(value string) // Called when text changes
	OnSubmit func(value string) // Called when Enter is pressed
}

// NewTextInput creates a new TextInput with the given ID
// React-like: TextInput(id) creates component instance
func NewTextInput(id string) *TextInput {
	return &TextInput{
		ID:        id,
		Value:     "",
		CursorPos: 0,
		Scroll:    0,
		Width:     50, // Default width
	}
}

// SetProps updates the component props (React pattern)
// Allows parent to configure component behavior
func (t *TextInput) SetProps(props TextInputProps) {
	t.Placeholder = props.Placeholder
	if props.Width > 0 {
		t.Width = props.Width
	}
	t.Disabled = props.Disabled
	t.OnChange = props.OnChange
	t.OnSubmit = props.OnSubmit
}

// Fluent API methods (React-like chaining)

// WithPlaceholder sets the placeholder text and returns the component for chaining
func (t *TextInput) WithPlaceholder(placeholder string) *TextInput {
	t.Placeholder = placeholder
	return t
}

// WithWidth sets the width and returns the component for chaining
func (t *TextInput) WithWidth(width int) *TextInput {
	t.Width = width
	return t
}

// WithOnChange sets the onChange callback and returns the component for chaining
func (t *TextInput) WithOnChange(onChange func(string)) *TextInput {
	t.OnChange = onChange
	return t
}

// WithOnSubmit sets the onSubmit callback and returns the component for chaining
func (t *TextInput) WithOnSubmit(onSubmit func(string)) *TextInput {
	t.OnSubmit = onSubmit
	return t
}

// WithValue sets the initial value and returns the component for chaining
func (t *TextInput) WithValue(value string) *TextInput {
	t.Value = value
	t.CursorPos = len(value)
	return t
}

// Render returns the markup for this component (React-like render)
func (t *TextInput) Render() string {
	display := t.GetDisplay()
	return `<input id="` + t.ID + `">` + display + `</input>`
}

// GetID returns the component ID (for auto-registration)
func (t *TextInput) GetID() string {
	return t.ID
}

// InsertChar inserts a character at the cursor position
func (t *TextInput) InsertChar(ch string) {
	t.Value = t.Value[:t.CursorPos] + ch + t.Value[t.CursorPos:]
	t.CursorPos++
	t.adjustScroll()
}

// DeleteChar deletes the character before the cursor (backspace)
func (t *TextInput) DeleteChar() {
	if t.CursorPos > 0 {
		t.Value = t.Value[:t.CursorPos-1] + t.Value[t.CursorPos:]
		t.CursorPos--
		t.adjustScroll()
	}
}

// DeleteForward deletes the character at the cursor (delete key)
func (t *TextInput) DeleteForward() {
	if t.CursorPos < len(t.Value) {
		t.Value = t.Value[:t.CursorPos] + t.Value[t.CursorPos+1:]
	}
}

// MoveLeft moves the cursor one position to the left
func (t *TextInput) MoveLeft() {
	if t.CursorPos > 0 {
		t.CursorPos--
		t.adjustScroll()
	}
}

// MoveRight moves the cursor one position to the right
func (t *TextInput) MoveRight() {
	if t.CursorPos < len(t.Value) {
		t.CursorPos++
		t.adjustScroll()
	}
}

// Home moves the cursor to the beginning
func (t *TextInput) Home() {
	t.CursorPos = 0
	t.Scroll = 0
}

// End moves the cursor to the end
func (t *TextInput) End() {
	t.CursorPos = len(t.Value)
	t.adjustScroll()
}

// MoveWordLeft moves the cursor to the start of the previous word
func (t *TextInput) MoveWordLeft() {
	if t.CursorPos == 0 {
		return
	}

	// Skip any spaces at current position
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] == ' ' {
		t.CursorPos--
	}

	// Move to start of current/previous word
	for t.CursorPos > 0 && t.Value[t.CursorPos-1] != ' ' {
		t.CursorPos--
	}

	t.adjustScroll()
}

// MoveWordRight moves the cursor to the start of the next word
func (t *TextInput) MoveWordRight() {
	if t.CursorPos >= len(t.Value) {
		return
	}

	// Skip any spaces at current position
	for t.CursorPos < len(t.Value) && t.Value[t.CursorPos] == ' ' {
		t.CursorPos++
	}

	// Move to end of current/next word
	for t.CursorPos < len(t.Value) && t.Value[t.CursorPos] != ' ' {
		t.CursorPos++
	}

	t.adjustScroll()
}

// Clear clears the input and resets cursor
func (t *TextInput) Clear() {
	t.Value = ""
	t.CursorPos = 0
	t.Scroll = 0
}

// GetVisible returns the visible portion of the text and the cursor offset within it
func (t *TextInput) GetVisible() (visible string, cursorOffset int) {
	visibleWidth := t.Width
	if visibleWidth < 1 {
		visibleWidth = 10 // Minimum width
	}

	endPos := t.Scroll + visibleWidth
	if endPos > len(t.Value) {
		endPos = len(t.Value)
	}

	visible = t.Value[t.Scroll:endPos]
	cursorOffset = t.CursorPos - t.Scroll
	return visible, cursorOffset
}

// GetDisplay returns the text to display (visible portion or placeholder)
func (t *TextInput) GetDisplay() string {
	visible, _ := t.GetVisible()
	if len(visible) == 0 && t.Placeholder != "" {
		return t.Placeholder
	}
	if len(visible) == 0 {
		return " " // Empty space to maintain layout
	}
	return visible
}

// HandleKey handles a key event and returns true if it was handled
// Returns false if the key should be handled by the application (e.g., Enter)
func (t *TextInput) HandleKey(event tty.KeyEvent) bool {
	oldValue := t.Value

	// Printable characters
	if event.IsPrintable() {
		t.InsertChar(event.Char)
		t.emitChange(oldValue)
		return true
	}

	// Backspace
	if event.IsBackspace() {
		t.DeleteChar()
		t.emitChange(oldValue)
		return true
	}

	// Delete key
	if event.Code == tty.SeqDelete {
		t.DeleteForward()
		t.emitChange(oldValue)
		return true
	}

	// Enter key - emit submit event
	if event.IsEnter() {
		if t.OnSubmit != nil {
			t.OnSubmit(t.Value)
		}
		return false // Let app handle it too
	}

	// Arrow keys
	if event.Code == tty.SeqLeft {
		t.MoveLeft()
		return true
	}

	if event.Code == tty.SeqRight {
		t.MoveRight()
		return true
	}

	// Word navigation (Ctrl+Left/Right)
	if event.Code == tty.SeqCtrlLeft {
		t.MoveWordLeft()
		return true
	}

	if event.Code == tty.SeqCtrlRight {
		t.MoveWordRight()
		return true
	}

	// Cmd+Left/Right on Mac (beginning/end of line)
	if event.Code == tty.SeqCmdLeft {
		t.Home()
		return true
	}

	if event.Code == tty.SeqCmdRight {
		t.End()
		return true
	}

	// Home key
	if event.Code == tty.SeqHome {
		t.Home()
		return true
	}

	// End key
	if event.Code == tty.SeqEnd {
		t.End()
		return true
	}

	// Not handled - let application handle it (e.g., Tab, etc.)
	return false
}

// emitChange triggers OnChange callback if value changed
func (t *TextInput) emitChange(oldValue string) {
	if t.Value != oldValue && t.OnChange != nil {
		t.OnChange(t.Value)
	}
}

// adjustScroll adjusts horizontal scroll to keep cursor visible
func (t *TextInput) adjustScroll() {
	visibleWidth := t.Width
	if visibleWidth < 1 {
		visibleWidth = 10
	}

	// If cursor is past the visible area, scroll right
	if t.CursorPos-t.Scroll >= visibleWidth {
		t.Scroll = t.CursorPos - visibleWidth + 1
	}

	// If cursor is before the visible area, scroll left
	if t.CursorPos < t.Scroll {
		t.Scroll = t.CursorPos
	}
}

// --- Focusable interface implementation ---

// HandleKeyEvent implements Focusable interface
// Processes keyboard events when this component has focus
// Returns true if the event was handled, false to bubble up (e.g., Enter key)
func (t *TextInput) HandleKeyEvent(event tty.KeyEvent) bool {
	return t.HandleKey(event)
}

// GetCursorOffset implements Focusable interface
// Returns the cursor position offset within the visible text (including prompt)
func (t *TextInput) GetCursorOffset() int {
	_, offset := t.GetVisible()
	// Add prompt length ("> " = 2 characters)
	return 2 + offset
}

// IsFocusable implements Focusable interface
// Returns true if this component can receive focus
func (t *TextInput) IsFocusable() bool {
	return !t.Disabled
}
