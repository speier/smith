package runtime

import "github.com/speier/smith/pkg/lotus/tty"

// Focusable is implemented by components that can receive keyboard focus
// When a component is focused, it receives keyboard events and shows cursor
type Focusable interface {
	// HandleKeyEvent processes a keyboard event
	// Returns true if the event was handled, false if it should bubble up
	HandleKeyEvent(event tty.KeyEvent) bool

	// GetCursorOffset returns the cursor position offset within the component
	// Used by Lotus to automatically position the terminal cursor
	GetCursorOffset() int

	// IsFocusable returns true if the component can currently receive focus
	// Allows components to disable focus dynamically (e.g., disabled input)
	IsFocusable() bool
}
