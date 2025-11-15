package primitives

import (
	"reflect"

	"github.com/speier/smith/pkg/lotus/commands"
	"github.com/speier/smith/pkg/lotus/tty"
)

// This file contains keyboard event handling for Input component

// invokeCallback invokes a callback that can be func(string) or func(Context, string)
// Uses reflection to handle concrete Context types (e.g., runtime.Context)
func invokeCallback(callback any, ctx Context, value string) {
	if callback == nil {
		return
	}

	// Fast path: check known signatures first
	switch fn := callback.(type) {
	case func(string):
		fn(value)
	case func(Context, string):
		fn(ctx, value)
	default:
		// Use reflection to handle func(concreteContextType, string)
		// where concreteContextType implements Context interface
		v := reflect.ValueOf(callback)
		t := v.Type()

		// Check if it's a function with 2 parameters
		if t.Kind() != reflect.Func || t.NumIn() != 2 {
			return // Ignore unsupported types
		}

		// Check if first param implements Context interface
		firstParam := t.In(0)
		contextInterface := reflect.TypeOf((*Context)(nil)).Elem()

		// Check if second param is string
		if t.In(1).Kind() != reflect.String {
			return
		}

		// If first param implements Context, call the function
		if firstParam.Implements(contextInterface) || (firstParam.Kind() == reflect.Struct && reflect.PointerTo(firstParam).Implements(contextInterface)) {
			// Convert Context interface to concrete type via reflection
			ctxValue := reflect.ValueOf(ctx)
			// If ctx is nil, create zero value
			if !ctxValue.IsValid() {
				ctxValue = reflect.Zero(firstParam)
			}

			v.Call([]reflect.Value{ctxValue, reflect.ValueOf(value)})
		}
	}
}

// HandleKey handles a key event and returns true if it was handled
// Returns false if the key should be handled by the application (e.g., Enter)
func (t *Input) HandleKey(event tty.KeyEvent) bool {
	return t.HandleKeyWithContext(event, nil)
}

// HandleKeyWithContext handles a key event with context support
func (t *Input) HandleKeyWithContext(event tty.KeyEvent, ctx Context) bool {
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

	// Shift+Enter - insert newline (multi-line support)
	if event.Code == tty.SeqShiftEnter {
		t.InsertNewline()
		t.emitChange(oldValue)
		return true
	}

	// Enter key - emit submit event (normal behavior)
	if event.IsEnter() {
		// Check global command registry for slash commands
		if commands.GetGlobalCommands().ExecuteWithContext(t.Value, ctx) {
			// Command was executed, clear input
			t.Value = ""
			t.CursorPos = 0
			return false // Let app handle it too (for rendering)
		}

		// Not a command, call OnSubmit callback
		if t.OnSubmit != nil {
			invokeCallback(t.OnSubmit, ctx, t.Value)
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
		t.emitChange(oldValue)
		return true
	}

	// Ctrl+K - delete to end
	if event.Key == '\x0b' { // Ctrl+K is 0x0b
		t.DeleteToEnd()
		t.emitChange(oldValue)
		return true
	}

	// Ctrl+Backspace or Ctrl+W - delete word backward
	if event.Code == tty.SeqCtrlBackspace || event.Key == '\x17' { // Ctrl+W is 0x17
		t.DeleteWordBackward()
		t.emitChange(oldValue)
		return true
	}

	// Not handled - let application handle it (e.g., Tab, etc.)
	return false
}

// emitChange triggers OnChange callback if value changed
func (t *Input) emitChange(oldValue string) {
	if t.Value != oldValue && t.OnChange != nil {
		invokeCallback(t.OnChange, nil, t.Value)
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
	return t.HandleKeyWithContext(event, nil)
}
