package runtime

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus/vdom"
)

// PropsUpdater is an optional interface for components that need to update
// callbacks and props when reconciliation reuses a cached instance
type PropsUpdater interface {
	UpdateProps(newComponent vdom.Component)
}

// reconcileComponents walks the tree and replaces component instances with cached ones
// This preserves component state (like Input value) across renders
func (fm *focusManager) reconcileComponents(element *vdom.Element, path string) {
	if element == nil {
		return
	}

	// If this element has a component, check cache
	if element.Component != nil {
		// Generate a stable key for this component position
		// Use type + path for uniqueness
		componentType := fmt.Sprintf("%T", element.Component)
		cacheKey := fmt.Sprintf("%s@%s", componentType, path)

		// Check if we have a cached instance at this position
		if cached, exists := fm.componentCache[cacheKey]; exists {
			// Update props/callbacks on the cached component if it implements PropsUpdater
			if updater, ok := cached.(PropsUpdater); ok {
				updater.UpdateProps(element.Component)
			}

			// Reuse the cached instance - this preserves state!
			element.Component = cached

			// CRITICAL: Re-render the component to update the element tree
			// The tree was rendered from the NEW component, but we want the cached one's output
			if rendered := cached.Render(); rendered != nil {
				// Replace the element's content with freshly rendered tree from cached component
				element.Tag = rendered.Tag
				element.Props = rendered.Props
				element.Children = rendered.Children
			}
		} else {
			// First time seeing this component at this position - cache it
			fm.componentCache[cacheKey] = element.Component
		}
	}

	// Recurse into children
	for i, child := range element.Children {
		childPath := fmt.Sprintf("%s.%d", path, i)
		fm.reconcileComponents(child, childPath)
	}
}
