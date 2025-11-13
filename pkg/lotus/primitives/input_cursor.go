package primitives

import (
	"time"

	"github.com/speier/smith/pkg/lotus/vdom"
)

// This file contains cursor-related functionality for Input component

// UpdateCursorBlink updates the cursor blink state based on elapsed time
// Returns true if cursor visibility changed (for re-rendering)
func (t *Input) UpdateCursorBlink() bool {
	if !t.CursorBlink || t.BlinkInterval == 0 {
		return false
	}

	now := time.Now()
	if now.Sub(t.lastBlinkTime) >= t.BlinkInterval {
		t.CursorVisible = !t.CursorVisible
		t.lastBlinkTime = now
		return true // Visibility changed
	}

	return false // No change
}

// SetCursorStyle sets the cursor visual style
func (t *Input) SetCursorStyle(style CursorStyle) {
	t.CursorStyle = style
}

// SetCursorBlink enables or disables cursor blinking
func (t *Input) SetCursorBlink(enabled bool) {
	t.CursorBlink = enabled
	if !enabled {
		t.CursorVisible = true // Always visible if not blinking
	}
}

// SetBlinkInterval sets how fast the cursor blinks
func (t *Input) SetBlinkInterval(interval time.Duration) {
	t.BlinkInterval = interval
}

// GetCursorChar returns the character to display for the cursor
func (t *Input) GetCursorChar() string {
	if !t.CursorVisible {
		return " " // Invisible during blink
	}

	switch t.CursorStyle {
	case CursorBlock:
		return "█"
	case CursorUnderline:
		return "_"
	case CursorBar:
		return "|"
	default:
		return "█"
	}
}

// GetCursorOffset implements Focusable interface
// Returns the cursor position offset within the visible text (including prompt)
func (t *Input) GetCursorOffset() int {
	_, offset := t.GetVisible()
	// Add prompt length ("> " = 2 characters)
	return 2 + offset
}

// GetDisplayWithCursor returns the text with cursor inserted at current position
// This is 100% testable without terminal rendering!
func (t *Input) GetDisplayWithCursor() string {
	visible, cursorOffset := t.GetVisible()

	// Cursor character
	cursorChar := t.GetCursorChar()

	// Insert cursor at the offset position
	if cursorOffset < 0 {
		cursorOffset = 0
	}
	if cursorOffset > len(visible) {
		cursorOffset = len(visible)
	}

	if cursorOffset == len(visible) {
		// Cursor at end
		return visible + cursorChar
	}

	// Cursor in middle - replace character at cursor position
	return visible[:cursorOffset] + cursorChar + visible[cursorOffset+1:]
}

// applyCursorStyle applies the appropriate cursor styling to a text element
func (t *Input) applyCursorStyle(textElem *vdom.Element) *vdom.Element {
	switch t.CursorStyle {
	case CursorBlock:
		// Block cursor: bright background with dark text (inverse video)
		return textElem.
			WithStyle("color", "#000000").
			WithStyle("background-color", "#ffffff")
	case CursorUnderline:
		// Underline cursor: underline decoration
		return textElem.WithStyle("text-decoration", "underline")
	case CursorBar:
		// Bar cursor: render a vertical bar before the character
		// Note: This is handled differently - we'll prefix with "|"
		return textElem
	default:
		return textElem
	}
}
