package primitives

import (
	"github.com/speier/smith/pkg/lotus/tty"
)

// This file contains keyboard event handling for Input component

// HandleKey handles a key event and returns true if it was handled
// Returns false if the key should be handled by the application (e.g., Enter)
func (t *Input) HandleKey(event tty.KeyEvent) bool {
	return t.HandleKeyWithContext(Context{}, event)
}

// HandleKeyWithContext handles a key event with context support
func (t *Input) HandleKeyWithContext(ctx Context, event tty.KeyEvent) bool {
	oldValue := t.Value

	// Printable characters
	if event.IsPrintable() {
		t.InsertChar(event.Char)
		t.emitChange(ctx, oldValue)
		return true
	}

	// Backspace
	if event.IsBackspace() {
		t.DeleteChar()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Delete key
	if event.Code == tty.SeqDelete {
		t.DeleteForward()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Shift+Enter - insert newline (multi-line support)
	if event.Code == tty.SeqShiftEnter {
		t.InsertNewline()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Enter key - emit submit event (normal behavior)
	if event.IsEnter() {
		// Call OnSubmit callback
		if t.OnSubmit != nil {
			t.OnSubmit(ctx, t.Value)
			// Auto-clear after submit (like browser inputs)
			t.Value = ""
			t.CursorPos = 0
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

	if event.Code == tty.SeqUp {
		t.MoveUp()
		return true
	}

	if event.Code == tty.SeqDown {
		t.MoveDown()
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

	// Cmd+Backspace (Mac) or Ctrl+U - delete to beginning
	if event.Code == tty.SeqCmdBackspace || event.Key == '\x15' { // Ctrl+U is 0x15
		t.DeleteToBeginning()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Ctrl+K - delete to end
	if event.Key == '\x0b' { // Ctrl+K is 0x0b
		t.DeleteToEnd()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Ctrl+Backspace or Ctrl+W - delete word backward
	if event.Code == tty.SeqCtrlBackspace || event.Key == '\x17' { // Ctrl+W is 0x17
		t.DeleteWordBackward()
		t.emitChange(ctx, oldValue)
		return true
	}

	// Not handled - let application handle it (e.g., Tab, etc.)
	return false
}

// emitChange triggers OnChange callback if value changed
func (t *Input) emitChange(ctx Context, oldValue string) {
	if t.Value != oldValue && t.OnChange != nil {
		t.OnChange(ctx, t.Value)
	}
}

// adjustScroll adjusts horizontal scroll to keep cursor visible
func (t *Input) adjustScroll() {
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

// HandleKeyEvent implements Focusable interface (backward compatibility)
// Processes keyboard events when this component has focus
// Returns true if the event was handled, false to bubble up (e.g., Enter key)
func (t *Input) HandleKeyEvent(event tty.KeyEvent) bool {
	return t.HandleKeyWithContext(Context{}, event)
}
