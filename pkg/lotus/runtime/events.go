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
	if element.Component != nil {
		// Component might handle keys via HandleKey
		// (e.g., wrapper components like Tabs that delegate focus to children)
		if handler, ok := element.Component.(interface {
			HandleKey(Context, tty.KeyEvent) bool
		}); ok {
			ctx := Context{} // Empty context for global handlers
			if handler.HandleKey(ctx, event) {
				return true
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
