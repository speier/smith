package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"syscall"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/render"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// App interface represents a Lotus application (like React.Component)
type App interface {
	Render() *vdom.Element
}

// DevToolsProvider is the interface for DevTools integration
type DevToolsProvider interface {
	Log(format string, args ...interface{})
	Render() *vdom.Element
	Enable()
	Disable()
	IsEnabled() bool
}

// HMRManager is the interface for HMR integration
type HMRManager interface {
	Start() error
	Stop() error
	SetCleanupHandler(func())
	SetExitHandler(func())
}

// Global factories (set by devtools package to avoid import cycle)
var devToolsFactory func() DevToolsProvider                                     //nolint:unused // Set by devtools package init()
var hmrFactory func(App, DevToolsProvider) (HMRManager, error) //nolint:unused // Set by devtools package init()

// SetDevToolsFactory registers the DevTools constructor
func SetDevToolsFactory(factory func() DevToolsProvider) {
	devToolsFactory = factory
}

// SetHMRFactory registers the HMR constructor
func SetHMRFactory(factory func(App, DevToolsProvider) (HMRManager, error)) {
	hmrFactory = factory
}

// elementApp wraps a static Element to satisfy App interface
type elementApp struct {
	element *vdom.Element
}

func (e *elementApp) Render() *vdom.Element {
	return e.element
}

// Run creates and runs a Lotus terminal app
// Accepts:
//   - App interface (with Render method)
//   - *vdom.Element (static element)
//   - string (markup string, optionally followed by data for {0}, {1}, etc.)
func Run(app any, data ...any) error {
	// Convert to App if needed
	var appInstance App
	switch v := app.(type) {
	case App:
		appInstance = v
	case *vdom.Element:
		appInstance = &elementApp{element: v}
	case string:
		// Parse markup string to element with optional data
		elem := vdom.Markup(v, data...)
		appInstance = &elementApp{element: elem}
	default:
		return fmt.Errorf("app must be App interface, *vdom.Element, or markup string, got %T", app)
	}

	// Check for state restoration from HMR
	if statePath := os.Getenv("LOTUS_STATE_PATH"); statePath != "" {
		_ = LoadAppState(appInstance, statePath)
		// Clean up state file after loading
		_ = os.Remove(statePath)
	}

	term, err := tty.New()
	if err != nil {
		return fmt.Errorf("creating terminal: %w", err)
	}

	width, height := term.Size()

	// Initialize DevTools and HMR if LOTUS_DEV=true
	var devTools DevToolsProvider
	var hmrManager HMRManager
	var hmrRestart bool // Flag to trigger restart after clean exit
	var hmrStatePath string
	
	if os.Getenv("LOTUS_DEV") == "true" && devToolsFactory != nil {
		devTools = devToolsFactory()
		
		// Set callback to trigger re-render when logs are added (BEFORE HMR starts)
		if dt, ok := devTools.(interface{ SetOnLogAdded(func()) }); ok {
			dt.SetOnLogAdded(func() {
				term.RequestRender() // Trigger re-render
			})
		}
		
		// Create HMR manager if factory exists

		if hmrFactory != nil {
			if hmr, err := hmrFactory(appInstance, devTools); err == nil {
				hmrManager = hmr
				
				// Set exit handler to trigger clean exit and restart
				hmrManager.SetExitHandler(func() {
					// Signal restart after terminal cleanup
					hmrRestart = true
					hmrStatePath = fmt.Sprintf("/tmp/lotus-state-%d.json", os.Getpid())
					// Cancel input to trigger clean exit
					term.CancelInput()
				})
				
				if err := hmrManager.Start(); err != nil {
					// Log but don't fail
					if devTools != nil {
						devTools.Log("⚠️  HMR failed to start: %v", err)
					}
				}
			}
		}
	}

	// Set up rendering using clean pipeline: vdom → style → layout → render
	term.OnRender(func() string {
		element := appInstance.Render()

		// Wrap with DevTools overlay if enabled
		if devTools != nil && devTools.IsEnabled() {
			devToolsPanel := devTools.Render()
			if devToolsPanel != nil {
				element = wrapWithDevTools(element, devToolsPanel, devTools)
			}
		}

		// 1. Resolve styles (no external CSS, just inline styles)
		resolver := style.NewResolver("")
		styled := resolver.Resolve(element)

		// 2. Compute layout
		layoutBox := layout.Compute(styled, width, height)

		// 3. Render to ANSI
		renderer := render.New()
		return renderer.Render(layoutBox)
	})

	// Auto-wire resize handling
	term.OnResize(func(w, h int) {
		width = w
		height = h
	})

	// Auto-wire keyboard event routing
	term.OnKey(func(event tty.KeyEvent) bool {
		// Ctrl+C or Ctrl+D exits
		if event.IsCtrlC() || event.IsCtrlD() {
			return false
		}

		// Ctrl+T toggles DevTools visibility
		if devTools != nil && event.Key == 20 { // Ctrl+T
			if devTools.IsEnabled() {
				devTools.Disable()
			} else {
				devTools.Enable()
			}
			return true
		}

		// Ctrl+P cycles DevTools position (right → bottom → left → right)
		if devTools != nil && devTools.IsEnabled() && event.Key == 16 { // Ctrl+P
			if dt, ok := devTools.(interface{ CyclePosition() }); ok {
				dt.CyclePosition()
			}
			return true
		}

		// First, try to route to app if it implements Focusable (backward compatibility)
		if focusable, ok := appInstance.(Focusable); ok {
			if focusable.HandleKeyEvent(event) {
				return true
			}
		}

		// Auto-discover focusable components in the element tree
		element := appInstance.Render()
		if handleEventInTree(element, event) {
			return true
		}

		return true
	})

	// Filter mouse events by default
	term.SetFilterMouse(true)

	// Start (blocks until exit)
	err = term.Start()
	
	// Cleanup HMR on exit
	if hmrManager != nil {
		_ = hmrManager.Stop()
	}
	
	// If HMR restart was requested, exec the new binary
	if hmrRestart {
		// Terminal is now cleanly restored (thanks to defers in Start())
		// Use syscall.Exec to replace this process with the new binary
		if execErr := execRestart(hmrStatePath); execErr != nil {
			return fmt.Errorf("HMR restart failed: %w (original error: %v)", execErr, err)
		}
	}
	
	return err
}

// execRestart replaces the current process with the HMR-rebuilt binary
func execRestart(statePath string) error {
	// This is a bit hacky but we need to import devtools package
	// For now, inline the exec logic here
	binaryPath := "/tmp/lotus-hmr-app"
	
	// Verify the new binary exists
	if _, err := os.Stat(binaryPath); err != nil {
		return fmt.Errorf("rebuilt binary not found: %w", err)
	}

	// Prepare environment with state path
	env := os.Environ()
	if statePath != "" {
		env = append(env, fmt.Sprintf("LOTUS_STATE_PATH=%s", statePath))
	}

	// Use syscall.Exec to replace current process with new binary
	// This is Unix-specific but works on macOS and Linux
	// The current process is replaced - no new process spawned!
	return syscall.Exec(binaryPath, []string{binaryPath}, env)
}

// handleEventInTree traverses the element tree to find and route keyboard events to focusable components
func handleEventInTree(element *vdom.Element, event tty.KeyEvent) bool {
	if element == nil {
		return false
	}

	// Check if this element wraps a focusable component
	if element.Component != nil {
		if focusable, ok := element.Component.(Focusable); ok {
			if focusable.IsFocusable() {
				// Route event to this component
				if focusable.HandleKeyEvent(event) {
					return true
				}
			}
		}
	}

	// Recurse into children
	for _, child := range element.Children {
		if handleEventInTree(child, event) {
			return true
		}
	}

	return false
}

// Stateful is an interface for components that can save/restore state
type Stateful interface {
	SaveState() map[string]interface{}
	LoadState(map[string]interface{}) error
	GetID() string
}

// SaveAppState saves app state to a JSON file for HMR
func SaveAppState(app App, path string) error {
	state := map[string]interface{}{
		"version": "1.0",
	}

	// Traverse the element tree and collect state from stateful components
	element := app.Render()
	components := collectStatefulComponents(element)
	
	if len(components) > 0 {
		componentStates := make(map[string]interface{})
		for _, comp := range components {
			if comp.GetID() != "" {
				componentStates[comp.GetID()] = comp.SaveState()
			}
		}
		state["components"] = componentStates
	}

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadAppState loads app state from a JSON file for HMR
func LoadAppState(app App, path string) error {
	// Check if state file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil // No state to restore
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		return err
	}

	// Restore component state
	if componentStates, ok := state["components"].(map[string]interface{}); ok {
		element := app.Render()
		components := collectStatefulComponents(element)
		
		for _, comp := range components {
			if comp.GetID() != "" {
				if compState, exists := componentStates[comp.GetID()]; exists {
					if stateMap, ok := compState.(map[string]interface{}); ok {
						_ = comp.LoadState(stateMap)
					}
				}
			}
		}
	}

	return nil
}

// CollectStatefulComponents traverses the element tree and collects stateful components
// Exported for DevTools to check for missing IDs
func CollectStatefulComponents(element *vdom.Element) []Stateful {
	return collectStatefulComponents(element)
}

// collectStatefulComponents traverses the element tree and collects stateful components
func collectStatefulComponents(element *vdom.Element) []Stateful {
	var components []Stateful
	
	if element == nil {
		return components
	}

	// Check if this element wraps a stateful component
	if element.Component != nil {
		if stateful, ok := element.Component.(Stateful); ok {
			components = append(components, stateful)
		}
	}

	// Recurse into children
	for _, child := range element.Children {
		components = append(components, collectStatefulComponents(child)...)
	}

	return components
}

// wrapWithDevTools wraps the app element with DevTools panel based on position
func wrapWithDevTools(app *vdom.Element, devToolsPanel *vdom.Element, devTools DevToolsProvider) *vdom.Element {
	// Type assert to get position (if available)
	type positionGetter interface {
		GetPosition() string
	}

	position := "right" // Default
	if pg, ok := devTools.(positionGetter); ok {
		position = pg.GetPosition()
	}

	// Style DevTools panel with border and dark background
	styledDevTools := vdom.Box(devToolsPanel).
		WithStyle("background-color", "#1a1a1a").
		WithBorderStyle(vdom.BorderStyleRounded)

	switch position {
	case "right":
		// Right side: HStack with app (60%) + devtools (40%)
		// Apply flex-grow directly to preserve app's internal flex behavior
		return vdom.HStack(
			app.Clone().WithFlexGrow(3),    // 60% (3/5) - app maintains internal flex
			styledDevTools.WithFlexGrow(2), // 40% (2/5)
		)

	case "bottom":
		// Bottom: VStack with app (70%) + devtools (30%)
		return vdom.VStack(
			app.Clone().WithFlexGrow(7),    // 70% - app maintains internal flex
			styledDevTools.WithFlexGrow(3), // 30%
		)

	case "left":
		// Left side: HStack with devtools (40%) + app (60%)
		return vdom.HStack(
			styledDevTools.WithFlexGrow(2), // 40% (2/5)
			app.Clone().WithFlexGrow(3),    // 60% (3/5) - app maintains internal flex
		)

	default:
		// Fallback to right
		return vdom.HStack(
			app.Clone().WithFlexGrow(3),
			styledDevTools.WithFlexGrow(2),
		)
	}
}
