package reconciler

import (
	"sync"

	"github.com/speier/smith/pkg/lotus/components"
	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/parser"
	"github.com/speier/smith/pkg/lotus/renderer"
	"github.com/speier/smith/pkg/lotus/tty"
)

// Focusable is an alias to components.Focusable for backward compatibility
type Focusable = components.Focusable

// ComponentRenderer is an alias for core.ComponentRenderer
type ComponentRenderer = core.ComponentRenderer

// UI represents a complete terminal UI
type UI struct {
	Root         *core.Node
	Width        int
	Height       int
	focusedID    string               // ID of currently focused component
	components   map[string]Focusable // Registry of focusable components by ID
	previousTree *core.Element        // Previous element tree for diffing
}

// renderContext manages UI instances for automatic reconciliation
// This is like React's fiber tree - tracks component instances across renders
type renderContext struct {
	mu       sync.RWMutex
	contexts map[string]*UI // UI instances by context ID
}

var globalContext = &renderContext{
	contexts: make(map[string]*UI),
}

// Render renders markup/CSS with automatic reconciliation
// This is the React-like API - just call Render(), library handles instance management
//
// The contextID identifies the render "root" (like a React root).
// First call creates a new UI instance, subsequent calls reconcile (update) it.
//
// Example:
//
//	func (app *ChatApp) Render() string {
//	    markup := `<box>...</box>`
//	    css := `box { color: #fff; }`
//	    return lotus.Render("app", markup, css, app.width, app.height)
//	}
//
// No manual if app.ui == nil checks needed!
func Render(contextID, markup, css string, width, height int) string {
	globalContext.mu.Lock()
	defer globalContext.mu.Unlock()

	ui, exists := globalContext.contexts[contextID]
	if !exists {
		// First render - create new UI instance
		ui = NewUI(markup, css, width, height)
		globalContext.contexts[contextID] = ui
	} else {
		// Subsequent render - reconcile (update existing instance)
		// This preserves component registrations, focus state, etc.
		ui.Update(markup, css)

		// Handle dimension changes
		if ui.Width != width || ui.Height != height {
			ui.Reflow(width, height)
		}
	}

	return ui.RenderToTerminal(true)
}

// RenderWithElement renders using an Element tree (enables Virtual DOM diffing)
// This is the optimized path - diffs against previous tree when possible
func RenderWithElement(contextID string, element *core.Element, width, height int) string {
	globalContext.mu.Lock()
	defer globalContext.mu.Unlock()

	ui, exists := globalContext.contexts[contextID]
	if !exists {
		// First render - create from element
		markup := element.ToMarkup()
		css := element.ToCSS()
		ui = NewUI(markup, css, width, height)
		ui.previousTree = element
		globalContext.contexts[contextID] = ui
	} else {
		// Subsequent render - use diffing!
		ui.UpdateWithElement(element)

		// Handle dimension changes
		if ui.Width != width || ui.Height != height {
			ui.Reflow(width, height)
		}
	}

	return ui.RenderToTerminal(true)
}

// GetContext returns the UI instance for a context (for focus management, component registration)
// Returns nil if context doesn't exist yet (hasn't been rendered)
//
// Example:
//
//	ui := lotus.GetContext("app")
//	ui.RegisterComponent("input", app.input)
//	ui.SetFocus("input")
func GetContext(contextID string) *UI {
	globalContext.mu.RLock()
	defer globalContext.mu.RUnlock()
	return globalContext.contexts[contextID]
}

// DestroyContext removes a render context (cleanup)
// Like unmounting a React root
func DestroyContext(contextID string) {
	globalContext.mu.Lock()
	defer globalContext.mu.Unlock()
	delete(globalContext.contexts, contextID)
}

// RegisterComponent registers a focusable component in a context
// Convenience wrapper for GetContext(id).RegisterComponent()
func RegisterComponent(contextID, componentID string, component Focusable) {
	if ui := GetContext(contextID); ui != nil {
		ui.RegisterComponent(componentID, component)
	}
}

// SetFocus sets focus in a context
// Convenience wrapper for GetContext(id).SetFocus()
func SetFocus(contextID, componentID string) {
	if ui := GetContext(contextID); ui != nil {
		ui.SetFocus(componentID)
	}
}

// NewUI creates a new terminal UI from markup and CSS
func NewUI(markup, css string, width, height int) *UI {
	// Parse markup
	root := parser.Parse(markup)
	if root == nil {
		root = core.NewNode("box")
	}

	// Parse and apply styles (with automatic caching)
	styles := GetStyles(css)
	parser.ApplyStyles(root, styles)

	// Create UI
	ui := &UI{
		Root:       root,
		Width:      width,
		Height:     height,
		components: make(map[string]Focusable),
	}

	// Compute layout
	layout.Layout(root, width, height)

	return ui
}

// NewFullscreenUI creates a new terminal UI that auto-detects terminal size
func NewFullscreenUI(markup, css string) (*UI, error) {
	// Use terminal package for size detection
	term, err := tty.New()
	if err != nil {
		// Fallback to default size
		return NewUI(markup, css, 100, 40), nil
	}

	width, height := term.Size()
	return NewUI(markup, css, width, height), nil
}

// Update updates the UI with new markup and CSS without losing component registrations
// This is React-like reconciliation - reuse the UI instance instead of creating new ones
func (ui *UI) Update(markup, css string) {
	// Parse new markup
	root := parser.Parse(markup)
	if root == nil {
		root = core.NewNode("box")
	}

	// Parse and apply styles (with automatic caching)
	styles := GetStyles(css)
	parser.ApplyStyles(root, styles)

	// Replace root but keep component registrations and focus
	ui.Root = root

	// Recompute layout with current dimensions
	layout.Layout(root, ui.Width, ui.Height)
}

// UpdateWithElement updates the UI using an Element tree (enables diffing)
// This is the React-like Virtual DOM update path
func (ui *UI) UpdateWithElement(element *core.Element) {
	// If we have a previous tree, try to diff
	if ui.previousTree != nil && element != nil {
		patches := Diff(ui.previousTree, element)

		// If patches are simple (text/style only), apply them
		if ui.canUseFastPath(patches) {
			_ = ui.ApplyPatches(patches)
			// Store new tree for next diff
			ui.previousTree = element
			return
		}
	}

	// Fall back to full render if:
	// - No previous tree
	// - Patches are complex (structure changes)
	// - Diffing not applicable

	markup := element.ToMarkup()
	css := element.ToCSS()
	ui.Update(markup, css)

	// Store tree for next diff
	ui.previousTree = element
}

// canUseFastPath checks if patches can be applied without full re-render
func (ui *UI) canUseFastPath(patches []Patch) bool {
	for _, patch := range patches {
		switch patch.(type) {
		case UpdateTextPatch, UpdateStylePatch:
			// These are fast
			continue
		default:
			// Structure changes need full render
			return false
		}
	}
	return true
}

// ApplyPatches applies a list of patches to the UI
func (ui *UI) ApplyPatches(patches []Patch) error {
	for _, patch := range patches {
		if err := patch.Apply(ui); err != nil {
			// If any patch fails, fall back to full render
			return err
		}
	}

	// After applying patches, re-layout (patches may have changed sizes)
	if ui.Root != nil {
		layout.Layout(ui.Root, ui.Width, ui.Height)
	}

	return nil
}

// RenderToTerminal renders the UI to a terminal string
// If clear is true, clears the screen first (full rebuild)
// If clear is false, just redraws over existing content (incremental update)
func (ui *UI) RenderToTerminal(clear bool) string {
	return renderer.RenderNode(ui.Root)
}

// FindByID finds a node by its ID
func (ui *UI) FindByID(id string) *core.Node {
	return ui.Root.FindByID(id)
}

// Reflow recomputes the layout with new dimensions
func (ui *UI) Reflow(width, height int) {
	ui.Width = width
	ui.Height = height
	layout.Layout(ui.Root, width, height)
}

// GetCursorPosition calculates the terminal cursor position for an input element
// Returns (row, col) in 1-indexed ANSI terminal coordinates, or (0, 0) if element not found
// The offset parameter specifies the character position within the input text
func (ui *UI) GetCursorPosition(inputID string, offset int) (row, col int) {
	node := ui.Root.FindByID(inputID)
	if node == nil {
		return 0, 0
	}

	// Convert 0-indexed layout coordinates to 1-indexed ANSI coordinates
	row = node.Y + 1
	col = node.X + offset + 1
	return row, col
}

// SetFocus sets focus to a component by ID
// The focused component will receive keyboard events and show cursor
func (ui *UI) SetFocus(id string) {
	ui.focusedID = id
}

// GetFocus returns the ID of the currently focused component
func (ui *UI) GetFocus() string {
	return ui.focusedID
}

// ClearFocus removes focus from any component
func (ui *UI) ClearFocus() {
	ui.focusedID = ""
}

// RegisterComponent registers a focusable component with the UI
// The component can then receive focus and keyboard events
func (ui *UI) RegisterComponent(id string, component Focusable) {
	if ui.components == nil {
		ui.components = make(map[string]Focusable)
	}
	ui.components[id] = component
}

// UnregisterComponent removes a component from the registry
func (ui *UI) UnregisterComponent(id string) {
	delete(ui.components, id)
}

// GetComponent returns a registered component by ID
func (ui *UI) GetComponent(id string) Focusable {
	return ui.components[id]
}

// HandleKey routes a keyboard event to the focused component
// Returns true if the event was handled by the component
// Returns false if no component is focused or the component didn't handle it
func (ui *UI) HandleKey(event tty.KeyEvent) bool {
	if ui.focusedID == "" {
		return false
	}

	component := ui.components[ui.focusedID]
	if component == nil || !component.IsFocusable() {
		return false
	}

	return component.HandleKeyEvent(event)
}

// UpdateCursor positions the terminal cursor at the focused component
// Should be called after rendering to show cursor at correct position
// Returns true if cursor was positioned, false if no focused component
func (ui *UI) UpdateCursor(term interface {
	ShowCursor()
	HideCursor()
	MoveCursor(row, col int)
}) bool {
	if ui.focusedID == "" {
		term.HideCursor()
		return false
	}

	// Get the focused component
	component := ui.components[ui.focusedID]
	if component == nil || !component.IsFocusable() {
		term.HideCursor()
		return false
	}

	// Find the focused node
	node := ui.Root.FindByID(ui.focusedID)
	if node == nil {
		term.HideCursor()
		return false
	}

	// Get cursor offset from the component
	cursorOffset := component.GetCursorOffset()

	// Calculate position
	row, col := ui.GetCursorPosition(ui.focusedID, cursorOffset)
	if row > 0 && col > 0 {
		term.ShowCursor()
		term.MoveCursor(row, col)
		return true
	}

	term.HideCursor()
	return false
}
