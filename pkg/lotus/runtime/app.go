package runtime

import (
	"fmt"
	"os"

	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/reconciler"
	"github.com/speier/smith/pkg/lotus/tty"
)

// App interface represents a Lotus application (like React.Component)
// Components implement Render() to describe their UI declaratively.
//
// Example:
//
//	type MyApp struct {}
//
//	func (app *MyApp) Render() *lotus.Element {
//	    return lotus.VStack(
//	        lotus.Text("Hello World"),
//	    ).Render()
//	}
//
// Note: The return type *lotus.Element is a type alias for *core.Element.
// Your IDE may show a warning that these types don't match - this is a false positive.
// The Go compiler correctly recognizes them as identical types.
// This is a known limitation of gopls (the Go language server) with cross-package type aliases.
type App interface {
	// Render returns the element tree (React's render() pattern).
	//
	// Return *lotus.Element from your implementation - it's the same type as *core.Element
	// due to the type alias defined in the lotus package.
	//
	// The framework automatically:
	//  - Extracts components for focus management
	//  - Generates CSS from inline styles
	//  - Performs Virtual DOM diffing
	//  - Updates the terminal display
	Render() *core.Element
}

// RenderFunc is a functional component (React hooks pattern)
// Use this for simple components without state management needs.
//
// Example:
//
//	func ChatApp() lotus.RenderFunc {
//	    messages := lotus.NewMessageList("messages")
//	    input := lotus.NewTextInput("input")
//
//	    return func() *lotus.Element {
//	        return lotus.VStack(
//	            messages.Render(),
//	            input.Render(),
//	        ).Render()
//	    }
//	}
//
//	lotus.Run("app", ChatApp())
type RenderFunc func() *core.Element

// Render implements App interface for RenderFunc (makes functions compatible with App)
func (fn RenderFunc) Render() *core.Element {
	return fn()
}

// ComponentProvider is an optional interface apps can implement
// to provide components for automatic registration and focus management
type ComponentProvider interface {
	// GetComponents returns a map of component IDs to focusable components
	// Framework will auto-register these and set initial focus to the first one
	GetComponents() map[string]interface{}
}

// TerminalApp wraps terminal setup boilerplate for Lotus apps
// This is like ReactDOM - handles mounting, event routing, rendering
type TerminalApp struct {
	term       *tty.Terminal
	app        App
	contextID  string
	width      int
	height     int
	devTools   DevToolsProvider // DevTools instance
	hmr        interface{}      // HMR instance (avoid import cycle)
	devEnabled bool
	hmrEnabled bool
}

// DevToolsProvider is an interface for DevTools to avoid import cycles
type DevToolsProvider interface {
	Log(format string, args ...interface{})
	Render() *core.Element
	IsEnabled() bool
}

// HMRProvider is an interface for HMR to avoid import cycles
type HMRProvider interface {
	Start() error
	Stop()
}

// devToolsFactory creates a new DevTools instance (set by devtools package init)
var devToolsFactory func() DevToolsProvider

// SetDevToolsFactory allows devtools package to register itself
func SetDevToolsFactory(factory func() DevToolsProvider) {
	devToolsFactory = factory
}

// Option configures a Lotus app
type Option func(*TerminalApp)

// WithDevTools enables or disables the DevTools panel
func WithDevTools(enabled bool) Option {
	return func(ta *TerminalApp) {
		ta.devEnabled = enabled
	}
}

// WithHMR enables or disables Hot Module Reload
func WithHMR(enabled bool) Option {
	return func(ta *TerminalApp) {
		ta.hmrEnabled = enabled
	}
}

// Run creates and runs a Lotus terminal app (like ReactDOM.render())
// This is the React-like entry point - handles all terminal setup automatically
//
// Example:
//
//	type MyApp struct { width, height int }
//	func (a *MyApp) Render() (string, string) {
//	    return `<box>Hello</box>`, `box { color: #0f0; }`
//	}
//
//	app := &MyApp{}
//	lotus.Run("app", app)  // That's it!
//
// With options:
//
//	lotus.Run("app", app,
//	    lotus.WithDevTools(true),
//	    lotus.WithHMR(true),
//	)
func Run(contextID string, app App, opts ...Option) error {
	term, err := tty.New()
	if err != nil {
		return fmt.Errorf("creating terminal: %w", err)
	}

	width, height := term.Size()

	ta := &TerminalApp{
		term:      term,
		app:       app,
		contextID: contextID,
		width:     width,
		height:    height,
	}

	// Auto-detect dev mode from environment (can be overridden by opts)
	if os.Getenv("LOTUS_DEV") == "true" {
		ta.devEnabled = true
		ta.hmrEnabled = true
	}

	// Apply options (overrides env var if explicitly set)
	for _, opt := range opts {
		opt(ta)
	}

	// Restore state if this is a HMR restart (automatic!)
	if statePath := os.Getenv("LOTUS_STATE_PATH"); statePath != "" {
		_ = LoadAppState(app, statePath)
		_ = os.Remove(statePath) // Cleanup
	}

	// Initialize DevTools if enabled
	if ta.devEnabled && devToolsFactory != nil {
		ta.devTools = devToolsFactory()
		ta.devTools.Log("🛠️  DevTools enabled (LOTUS_DEV=true)")
	}

	// Initialize HMR if enabled
	if ta.hmrEnabled {
		// HMR initialization happens via similar factory pattern
		// For now, we just mark it as enabled
		if ta.devTools != nil {
			ta.devTools.Log("🔥 HMR enabled")
		}
	}

	// Set up rendering (React pattern: app.Render() is pure, framework handles side effects)
	term.OnRender(func() string {
		userElement := app.Render()

		// Wrap with DevTools if enabled (Chrome DevTools pattern)
		// User's app gets resized to fit remaining space (70%), DevTools takes 30%
		var element *core.Element
		if ta.devTools != nil && ta.devTools.IsEnabled() {
			devPanel := ta.devTools.Render()
			if devPanel != nil {
				// Create horizontal layout: [User App (70%)] | [DevTools Panel (30%)]
				element = &core.Element{
					Type: "box",
					Styles: map[string]string{
						"direction": "row",
						"width":     "100%",
						"height":    "100%",
					},
					Children: []*core.Element{
						// User's app - automatically resized to 70%
						{
							Type: "box",
							Styles: map[string]string{
								"direction": "column",
								"width":     "70%",
								"height":    "100%",
							},
							Children: []*core.Element{userElement},
						},
						// DevTools panel - 30% on the right with yellow border
						{
							Type: "box",
							Styles: map[string]string{
								"direction":    "column",
								"width":        "30%",
								"height":       "100%",
								"border-width": "1",
								"border-color": "#ffff00",
							},
							Children: []*core.Element{devPanel},
						},
					},
				}
			} else {
				element = userElement
			}
		} else {
			element = userElement
		}

		// Extract components from element tree (before rendering)
		components := element.ExtractComponents()

		// Use element-based rendering with Virtual DOM diffing!
		// This creates the UI context on first call
		rendered := reconciler.RenderWithElement(contextID, element, ta.width, ta.height)

		// THEN register components (context exists now)
		for id, comp := range components {
			if focusable, ok := comp.(reconciler.Focusable); ok {
				reconciler.RegisterComponent(contextID, id, focusable)
			}
		}

		// Auto-focus first FOCUSABLE component on first render
		// Only set focus if no component is currently focused
		ui := reconciler.GetContext(contextID)
		if ui != nil && ui.GetFocus() == "" {
			// Set focus to first FOCUSABLE component (skip non-focusable ones)
			for id, comp := range components {
				if _, ok := comp.(reconciler.Focusable); ok {
					reconciler.SetFocus(contextID, id)
					break
				}
			}
		}

		return rendered
	}) // Auto-wire cursor positioning
	term.OnPostRender(func() {
		if ui := reconciler.GetContext(contextID); ui != nil {
			ui.UpdateCursor(term)
		}
	})

	// Auto-wire resize handling - just update dimensions, next render will adapt
	term.OnResize(func(width, height int) {
		ta.width = width
		ta.height = height
	})

	// Auto-wire keyboard event routing to focused components
	term.OnKey(func(event tty.KeyEvent) bool {
		// Ctrl+C or Ctrl+D exits
		if event.IsCtrlC() || event.IsCtrlD() {
			return false // Exit
		}

		// Route to focused component
		if ui := reconciler.GetContext(contextID); ui != nil {
			ui.HandleKey(event)
		}

		// Return true to keep running (only Ctrl+C/D should exit)
		return true
	})

	// Filter mouse events by default
	term.SetFilterMouse(true)

	// Start (blocks until exit)
	return term.Start()
}

// RunWith provides more control over terminal app configuration
// Use this when you need custom event handlers or setup
type TerminalConfig struct {
	ContextID   string
	App         App
	FilterMouse bool
	OnKey       func(event tty.KeyEvent) bool // Custom key handler (overrides default)
	OnResize    func(width, height int)       // Custom resize handler (called after default)
}

func RunWith(config TerminalConfig) error {
	term, err := tty.New()
	if err != nil {
		return fmt.Errorf("creating terminal: %w", err)
	}

	width, height := term.Size()

	ta := &TerminalApp{
		term:      term,
		app:       config.App,
		contextID: config.ContextID,
		width:     width,
		height:    height,
	}

	// Set up rendering
	term.OnRender(func() string {
		element := config.App.Render()

		// Extract components from element tree automatically
		components := element.ExtractComponents()
		for id, comp := range components {
			if focusable, ok := comp.(reconciler.Focusable); ok {
				reconciler.RegisterComponent(config.ContextID, id, focusable)
			}
		}

		// Convert element tree to markup string
		markup := element.ToMarkup()

		// Generate CSS from inline styles (React pattern!)
		css := element.ToCSS()

		return reconciler.Render(config.ContextID, markup, css, ta.width, ta.height)
	})

	// Auto-wire cursor positioning
	term.OnPostRender(func() {
		if ui := reconciler.GetContext(config.ContextID); ui != nil {
			ui.UpdateCursor(term)
		}
	})

	// Resize handling - just update dimensions, next render will adapt
	term.OnResize(func(width, height int) {
		ta.width = width
		ta.height = height
		// Call custom handler if provided
		if config.OnResize != nil {
			config.OnResize(width, height)
		}
	})

	// Key handling
	if config.OnKey != nil {
		term.OnKey(config.OnKey)
	} else {
		// Default key handler
		term.OnKey(func(event tty.KeyEvent) bool {
			// Ctrl+C or Ctrl+D exits the app
			if event.IsCtrlC() || event.IsCtrlD() {
				return false
			}
			// Let UI handle the key event (for focused components like inputs)
			if ui := reconciler.GetContext(config.ContextID); ui != nil {
				ui.HandleKey(event)
			}
			// Always return true to keep app running (except for Ctrl+C/D above)
			return true
		})
	}

	// Mouse filtering
	term.SetFilterMouse(config.FilterMouse)

	// Start (blocks until exit)
	return term.Start()
}
