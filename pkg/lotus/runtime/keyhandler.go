package runtime

import (
	"github.com/speier/smith/pkg/lotus/commands"
	"github.com/speier/smith/pkg/lotus/tty"
)

// KeyHandler is a function that handles keyboard events
// Returns true if the event was handled (stops propagation)
type KeyHandler func(event tty.KeyEvent) bool

// KeyBinding represents a keyboard shortcut binding
type KeyBinding struct {
	Key         byte   // The key byte (e.g., 'o' for Ctrl+O, use byte('o'))
	Ctrl        bool   // Requires Ctrl modifier (key < 32)
	Code        string // Escape sequence code (e.g., tty.SeqUp)
	Description string // Help text
	Handler     any    // func(tty.KeyEvent) bool or func(Context, tty.KeyEvent) bool
}

// globalKeyHandlers stores registered global key handlers
var globalKeyHandlers []*KeyBinding

// globalCommandRegistry stores the global command registry
var globalCommandRegistry *commands.CommandRegistry

// RegisterGlobalCommands sets the global command registry
// Commands will be checked before input callbacks are invoked
func RegisterGlobalCommands(registry *commands.CommandRegistry) {
	globalCommandRegistry = registry
}

// RegisterGlobalKey registers a global keyboard shortcut
// Example: RegisterGlobalKey('o', true, "Open file", handler) for Ctrl+O
// Example: RegisterGlobalKey('k', true, "Command palette", handler) for Ctrl+K
// Handler can be func(tty.KeyEvent) bool or func(Context, tty.KeyEvent) bool
func RegisterGlobalKey(key byte, ctrl bool, description string, handler any) {
	globalKeyHandlers = append(globalKeyHandlers, &KeyBinding{
		Key:         key,
		Ctrl:        ctrl,
		Description: description,
		Handler:     handler,
	})
}

// RegisterGlobalKeyCode registers a global keyboard shortcut by escape sequence
// Example: RegisterGlobalKeyCode(tty.SeqF1, "Show help", handler)
// Handler can be func(tty.KeyEvent) bool or func(Context, tty.KeyEvent) bool
func RegisterGlobalKeyCode(code string, description string, handler any) {
	globalKeyHandlers = append(globalKeyHandlers, &KeyBinding{
		Code:        code,
		Description: description,
		Handler:     handler,
	})
}

// UnregisterAllGlobalKeys clears all global key handlers
func UnregisterAllGlobalKeys() {
	globalKeyHandlers = nil
}

// matchesBinding checks if an event matches a key binding
func matchesBinding(event tty.KeyEvent, binding *KeyBinding) bool {
	// Check escape sequence code first
	if binding.Code != "" {
		return event.Code == binding.Code
	}

	// For Ctrl+Key combinations
	if binding.Ctrl {
		// Ctrl+A = 1, Ctrl+B = 2, ... Ctrl+Z = 26
		// To match Ctrl+O, we check if event.Key matches the control code
		expectedCtrlCode := binding.Key - 'a' + 1
		if binding.Key >= 'A' && binding.Key <= 'Z' {
			expectedCtrlCode = binding.Key - 'A' + 1
		}
		return event.Key == expectedCtrlCode
	}

	// Regular key match
	return binding.Key == event.Key
}

// invokeKeyHandler invokes a key handler with signature detection
func invokeKeyHandler(handler any, ctx Context, event tty.KeyEvent) bool {
	if handler == nil {
		return false
	}

	switch fn := handler.(type) {
	case func(tty.KeyEvent) bool:
		return fn(event)
	case func(Context, tty.KeyEvent) bool:
		return fn(ctx, event)
	default:
		return false
	}
}

// handleGlobalKeys processes global key handlers
// Returns true if a handler consumed the event
func handleGlobalKeys(event tty.KeyEvent) bool {
	return handleGlobalKeysWithContext(event, Context{})
}

// handleGlobalKeysWithContext processes global key handlers with context support
// Returns true if a handler consumed the event
func handleGlobalKeysWithContext(event tty.KeyEvent, ctx Context) bool {
	for _, binding := range globalKeyHandlers {
		if matchesBinding(event, binding) {
			if invokeKeyHandler(binding.Handler, ctx, event) {
				return true
			}
		}
	}
	return false
}

// GetGlobalKeyBindings returns all registered global key bindings
// Useful for showing help/shortcuts to users
func GetGlobalKeyBindings() []*KeyBinding {
	return globalKeyHandlers
}
