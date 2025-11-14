package runtime

import (
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// handleEventInTreeGlobal handles global keyboard events (for components that should receive events regardless of focus)
// For example, Tabs component should handle Left/Right arrows even when a child input has focus
func handleEventInTreeGlobal(element *vdom.Element, event tty.KeyEvent, focusMgr *focusManager) bool {
	if element == nil {
		return false
	}

	// Check if this element wraps a component that wants global events
	// For now, only non-focusable components get a chance (like Tabs wrapper)
	if element.Component != nil {
		if focusable, ok := element.Component.(Focusable); ok {
			// Skip if this is a focusable component (those are handled via focus manager)
			if !focusable.IsFocusable() {
				// Non-focusable component - give it a chance to handle global events
				if focusable.HandleKeyEvent(event) {
					return true
				}
			}
			// Focused components already handled above, skip them here
		} else {
			// Component doesn't implement Focusable but might handle keys
			// (e.g., wrapper components like Tabs that delegate focus to children)
			if handler, ok := element.Component.(interface{ HandleKeyEvent(tty.KeyEvent) bool }); ok {
				if handler.HandleKeyEvent(event) {
					return true
				}
			}
		}
	}

	// Recurse into children
	for _, child := range element.Children {
		if handleEventInTreeGlobal(child, event, focusMgr) {
			return true
		}
	}

	return false
}
