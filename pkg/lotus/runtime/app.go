package runtime

import (
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
//   - FunctionalComponent (func(Context) *vdom.Element)
//   - *vdom.Element (static element)
//   - string (markup string, optionally followed by data for {0}, {1}, etc.)
func Run(app any, data ...any) error {
	// Convert to App if needed
	var appInstance App
	switch v := app.(type) {
	case App:
		appInstance = v
	case FunctionalComponent:
		// Wrap functional component to satisfy App interface
		appInstance = &functionalApp{renderFn: v}
	case func(Context) *vdom.Element:
		// Support bare function type
		appInstance = &functionalApp{renderFn: FunctionalComponent(v)}
	case *vdom.Element:
		appInstance = &elementApp{element: v}
	case string:
		// Parse markup string to element with optional data
		elem := vdom.Markup(v, data...)
		appInstance = &elementApp{element: elem}
	default:
		return fmt.Errorf("app must be App interface, FunctionalComponent, *vdom.Element, or markup string, got %T", app)
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

	// If app has SetRenderCallback, provide it with term.RequestRender
	if setter, ok := appInstance.(interface{ SetRenderCallback(func()) }); ok {
		setter.SetRenderCallback(term.RequestRender)
	}

	// Initialize focus manager
	focusMgr := newFocusManager()

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

		// Enable render stats logging to DevTools
		term.SetStatsLogger(devTools.Log)

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

	// Store previous buffer for differential rendering
	var previousBuffer *render.Buffer

	// Set up rendering using clean pipeline: vdom → style → layout → buffer → diff → ansi
	term.OnRender(func() string {
		// Render once and update focus on that tree
		element := appInstance.Render()

		// Rebuild focus list and update component states on the rendered tree
		focusMgr.rebuild(element)

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

		// 3. Render layout to buffer
		layoutRenderer := render.NewLayoutRenderer()
		currentBuffer := layoutRenderer.RenderToBuffer(layoutBox, width, height)

		// 4. Compute diff and render ANSI
		var output string
		if previousBuffer == nil {
			// First render - full render
			output = render.RenderBufferFull(currentBuffer)
		} else {
			// Differential rendering
			diff := render.ComputeDiff(previousBuffer, currentBuffer)
			output = render.RenderBufferDiff(previousBuffer, currentBuffer, diff)
		}

		// Store current buffer for next render
		previousBuffer = currentBuffer

		return output
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

		// Tab key - cycle focus to next component
		if event.Key == '\t' {
			focusMgr.next()
			term.RequestRender() // Trigger re-render to show focus change
			return true
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

		// Route event to the currently focused component
		focused := focusMgr.getFocused()
		if focused != nil {
			if focused.HandleKeyEvent(event) {
				term.RequestRender() // Trigger re-render after focused component handles event
				return true
			}
		}

		// If no focused component handled it, try global handlers in the tree
		element := appInstance.Render()
		if handleEventInTreeGlobal(element, event, focusMgr) {
			term.RequestRender() // Trigger re-render after global handler handles event
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
