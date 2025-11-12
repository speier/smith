package primitives

import (
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TextArea is a multi-line text input field (like HTML <textarea>)
// Uses Input internally but with multi-line behavior
type TextArea struct {
	input *Input
}

// NewTextArea creates a new TextArea
func NewTextArea() *TextArea {
	return &TextArea{
		input: NewInput(),
	}
}

// WithID sets the component ID
func (ta *TextArea) WithID(id string) *TextArea {
	ta.input.WithID(id)
	return ta
}

// WithPlaceholder sets the placeholder text
func (ta *TextArea) WithPlaceholder(placeholder string) *TextArea {
	ta.input.WithPlaceholder(placeholder)
	return ta
}

// WithValue sets the initial value
func (ta *TextArea) WithValue(value string) *TextArea {
	ta.input.Value = value
	return ta
}

// WithOnChange sets the change callback
func (ta *TextArea) WithOnChange(fn func(string)) *TextArea {
	ta.input.OnChange = fn
	return ta
}

// WithOnSubmit sets the submit callback (Ctrl+Enter or Shift+Enter)
func (ta *TextArea) WithOnSubmit(fn func(string)) *TextArea {
	ta.input.OnSubmit = fn
	return ta
}

// Clear clears the content
func (ta *TextArea) Clear() {
	ta.input.Clear()
}

// Render renders the text area
func (ta *TextArea) Render() *vdom.Element {
	return ta.input.Render()
}

// IsFocusable returns whether the component can receive focus
func (ta *TextArea) IsFocusable() bool {
	return ta.input.IsFocusable()
}

// HandleKeyEvent handles keyboard events
func (ta *TextArea) HandleKeyEvent(event tty.KeyEvent) bool {
	// Enter → insert newline (textarea behavior)
	if event.Key == tty.KeyEnter || event.Key == tty.KeyEnter2 {
		ta.input.InsertNewline()
		return true
	}

	// Shift+Enter → submit
	if event.Code == tty.SeqShiftEnter {
		if ta.input.OnSubmit != nil {
			ta.input.OnSubmit(ta.input.Value)
		}
		return true
	}

	// Delegate everything else to Input
	return ta.input.HandleKeyEvent(event)
}

// GetCursorOffset returns cursor position for rendering
func (ta *TextArea) GetCursorOffset() int {
	return ta.input.GetCursorOffset()
}

// SetFocusState sets the focus state
func (ta *TextArea) SetFocusState(focused bool) {
	ta.input.SetFocusState(focused)
}
