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

// CreateTextArea creates a new multi-line text area with simplified API (like pi-tui)
// Usage: CreateTextArea(placeholder, onSubmit)
func CreateTextArea(placeholder string, onSubmit func(Context, string)) *TextArea {
	return NewTextArea().
		WithPlaceholder(placeholder).
		WithOnSubmit(onSubmit)
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

// WithOnChange sets the onChange callback (triggers on every change)
func (ta *TextArea) WithOnChange(fn func(Context, string)) *TextArea {
	ta.input.OnChange = fn
	return ta
}

// WithOnSubmit sets the onSubmit callback (triggers on Ctrl+Enter)
func (ta *TextArea) WithOnSubmit(fn func(Context, string)) *TextArea {
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

// HandleKey handles key events with context support
func (ta *TextArea) HandleKey(ctx Context, event tty.KeyEvent) bool {
	// Enter → insert newline (textarea behavior)
	if event.Key == tty.KeyEnter || event.Key == tty.KeyEnter2 {
		ta.input.InsertNewline()
		return true
	}

	// Shift+Enter → submit
	if event.Code == tty.SeqShiftEnter {
		if ta.input.OnSubmit != nil {
			ta.input.OnSubmit(ctx, ta.input.Value)
		}
		return true
	}

	// Delegate everything else to Input
	return ta.input.HandleKey(ctx, event)
}

// GetCursorOffset returns cursor position for rendering
func (ta *TextArea) GetCursorOffset() int {
	return ta.input.GetCursorOffset()
}

// SetFocusState sets the focus state
func (ta *TextArea) SetFocusState(focused bool) {
	ta.input.SetFocusState(focused)
}
