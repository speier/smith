package runtime

import (
	"fmt"
	"strings"

	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

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

// focusManager tracks which component has keyboard focus
type focusManager struct {
	focusables     []Focusable
	focusIndex     int
	focusedID      string                    // Track focused component by ID (user-provided, optional)
	focusedPath    string                    // Track focused component by tree path (auto-generated fallback)
	componentCache map[string]vdom.Component // Cache component instances by tree path for reuse
}

func newFocusManager() *focusManager {
	return &focusManager{
		focusables:     make([]Focusable, 0),
		focusIndex:     0,
		focusedID:      "",
		focusedPath:    "",
		componentCache: make(map[string]vdom.Component),
	}
}

func (fm *focusManager) collectFocusables(element *vdom.Element) {
	fm.collectFocusablesWithPath(element, "0")
}

func (fm *focusManager) collectFocusablesWithPath(element *vdom.Element, path string) {
	if element == nil {
		return
	}

	// Check if this element wraps a focusable component
	if element.Component != nil {
		if focusable, ok := element.Component.(Focusable); ok {
			if focusable.IsFocusable() {
				fm.focusables = append(fm.focusables, focusable)
			}
		}
	}

	// Recurse into children with path index
	for i, child := range element.Children {
		childPath := fmt.Sprintf("%s.%d", path, i)
		fm.collectFocusablesWithPath(child, childPath)
	}
}

func (fm *focusManager) rebuild(element *vdom.Element) {
	// Save currently focused component ID/path before rebuild
	if currentFocused := fm.getFocused(); currentFocused != nil {
		// Try ID first (user-provided, most stable)
		type IDGetter interface {
			GetID() string
		}
		if idGetter, ok := currentFocused.(IDGetter); ok {
			fm.focusedID = idGetter.GetID()
		}
		// Use focusIndex as fallback (position in list)
		fm.focusedPath = fmt.Sprintf("idx:%d", fm.focusIndex)
	}

	// CRITICAL: Reconcile components BEFORE collecting focusables
	// This replaces new component instances with cached ones, preserving state
	fm.reconcileComponents(element, "0")

	// Rebuild focusables list
	fm.focusables = make([]Focusable, 0)
	fm.collectFocusables(element)

	// Try to restore focus in order of preference:
	// 1. By user-provided ID (most stable)
	// 2. By position index (reasonable fallback)
	fm.focusIndex = 0 // Default to first

	if fm.focusedID != "" {
		// Try to find by ID
		type IDGetter interface {
			GetID() string
		}
		for i, f := range fm.focusables {
			if idGetter, ok := f.(IDGetter); ok {
				if idGetter.GetID() == fm.focusedID {
					fm.focusIndex = i
					goto restored
				}
			}
		}
	}

	// Fallback: restore by position if same number of focusables
	if fm.focusedPath != "" && strings.HasPrefix(fm.focusedPath, "idx:") {
		var idx int
		if _, err := fmt.Sscanf(fm.focusedPath, "idx:%d", &idx); err == nil {
			if idx < len(fm.focusables) {
				fm.focusIndex = idx
			}
		}
	}

restored:
	// Ensure focusIndex is valid
	if fm.focusIndex >= len(fm.focusables) {
		fm.focusIndex = 0
	}

	// Update focused state on all components and re-render them
	fm.updateFocusedStateAndRerender(element)
}

func (fm *focusManager) updateFocusedState() {
	focused := fm.getFocused()

	// Update all focusables - use a setter interface to avoid import cycles
	for _, f := range fm.focusables {
		isFocused := (f == focused)

		// Try to set focus via a common interface
		type FocusStateSetter interface {
			SetFocusState(bool)
		}
		if setter, ok := f.(FocusStateSetter); ok {
			setter.SetFocusState(isFocused)
		}
	}
}

// updateFocusedStateAndRerender updates focus state and re-renders focused components
// This ensures components render with correct focus state after reconciliation
func (fm *focusManager) updateFocusedStateAndRerender(element *vdom.Element) {
	// First update focus state
	fm.updateFocusedState()

	// Then re-render any components that changed focus state
	// This is needed because focus state affects rendering (cursor visibility, etc.)
	fm.rerenderFocusedComponents(element, "0")
}

// rerenderFocusedComponents walks the tree and re-renders components that are focusable
func (fm *focusManager) rerenderFocusedComponents(element *vdom.Element, path string) {
	if element == nil {
		return
	}

	// If this element has a focusable component, re-render it
	if element.Component != nil {
		if _, ok := element.Component.(Focusable); ok {
			// Re-render the component to update cursor visibility
			if rendered := element.Component.Render(); rendered != nil {
				element.Tag = rendered.Tag
				element.Props = rendered.Props
				element.Children = rendered.Children
			}
		}
	}

	// Recurse into children
	for i, child := range element.Children {
		childPath := fmt.Sprintf("%s.%d", path, i)
		fm.rerenderFocusedComponents(child, childPath)
	}
}

func (fm *focusManager) next() {
	if len(fm.focusables) == 0 {
		return
	}
	fm.focusIndex = (fm.focusIndex + 1) % len(fm.focusables)
	fm.updateFocusedState()
}

func (fm *focusManager) getFocused() Focusable {
	if len(fm.focusables) == 0 {
		return nil
	}
	if fm.focusIndex < 0 || fm.focusIndex >= len(fm.focusables) {
		return nil
	}
	return fm.focusables[fm.focusIndex]
}
