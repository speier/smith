package runtime

import (
	"fmt"
	"os"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/render"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// App interface represents a Lotus application (like React.Component)
type App interface {
	Render(ctx Context) *vdom.Element
}

// elementApp wraps a static Element to satisfy App interface
type elementApp struct {
	element *vdom.Element
}

func (e *elementApp) Render(ctx Context) *vdom.Element {
	return e.element
}

// Run creates and runs a Lotus terminal app
// Accepts App interface implementations (types with Render(Context) *Element method)
func Run(app App) error {
	appInstance := app

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

	// Initialize scroll manager for overflow:auto elements
	scrollMgr := newScrollManager()
	var scrollableElementID string // Track which element should receive scroll events

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
		// Create context for this render cycle
		renderCtx := Context{RenderCallback: term.RequestRender}
		// Render once and update focus on that tree
		element := appInstance.Render(renderCtx)

		// Rebuild focus list and update component states on the rendered tree
		focusMgr.rebuild(element)

		// Find scrollable element (element with overflow:auto or flex-grow > 0)
		scrollableElementID = findScrollableElement(element)

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
		layoutRenderer.ScrollManager = scrollMgr // Enable overflow:auto support
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

		// ESC key - check for open modals first (framework-level modal handling)
		if event.Key == tty.KeyEscape {
			renderCtx := Context{RenderCallback: term.RequestRender}
			element := appInstance.Render(renderCtx)
			if handleModalEscape(element) {
				term.RequestRender()
				return true
			}
		}

		// Create context for event handling
		ctx := Context{RenderCallback: term.RequestRender}

		// First, check global key handlers (highest priority)
		if handleGlobalKeysWithContext(event, ctx) {
			term.RequestRender()
			return true
		}

		// Tab key - cycle focus to next component
		if event.Key == '\t' {
			focusMgr.next()
			term.RequestRender() // Trigger re-render to show focus change
			return true
		}

		// Arrow keys scroll the scrollable element if one is set
		if scrollableElementID != "" {
			var dx, dy int
			switch event.Code {
			case tty.SeqUp:
				dy = -1
			case tty.SeqDown:
				dy = 1
			case tty.SeqLeft:
				dx = -1
			case tty.SeqRight:
				dx = 1
			}

			if dx != 0 || dy != 0 {
				if scrollMgr.ScrollBy(scrollableElementID, dx, dy) {
					term.RequestRender()
					return true
				}
			}
		}

		// Ctrl+T toggles DevTools visibility
		if devTools != nil && event.Key == 20 { // Ctrl+T
			if devTools.IsEnabled() {
				devTools.Disable()
			} else {
				devTools.Enable()
			}
			// Force full re-render when toggling DevTools to avoid artifacts
			previousBuffer = nil
			return true
		}

		// Ctrl+P cycles DevTools position (right → bottom → left → right)
		if devTools != nil && devTools.IsEnabled() && event.Key == 16 { // Ctrl+P
			if dt, ok := devTools.(interface{ CyclePosition() }); ok {
				dt.CyclePosition()
				// Force full re-render when changing position to avoid artifacts
				previousBuffer = nil
			}
			return true
		}

		// Route event to the currently focused component
		focused := focusMgr.getFocused()
		if focused != nil {
			if handler, ok := focused.(interface {
				HandleKey(Context, tty.KeyEvent) bool
			}); ok {
				handled := handler.HandleKey(ctx, event)
				if devTools != nil {
					devTools.Log("Key handled=%v, requesting render", handled)
				}
				term.RequestRender()
				if handled {
					return true
				}
			}
		}

		// If no focused component handled it, try global handlers in the tree
		renderCtx2 := Context{RenderCallback: term.RequestRender}
		element := appInstance.Render(renderCtx2)
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
		if execErr := hmrManager.ExecRestart(hmrStatePath); execErr != nil {
			return fmt.Errorf("HMR restart failed: %w (original error: %v)", execErr, err)
		}
	}

	return err
}

// RunFunc runs a functional component as a Lotus app
// Accepts: func(Context) *vdom.Element
func RunFunc(component FunctionalComponent) error {
	return Run(&functionalApp{renderFn: component})
}

// RunElement runs a static element as a Lotus app
// Accepts: *vdom.Element
func RunElement(element *vdom.Element) error {
	return Run(&elementApp{element: element})
}

// handleModalEscape searches for open modals in the tree and closes them on ESC
// Returns true if a modal was found and closed
func handleModalEscape(elem *vdom.Element) bool {
	if elem == nil {
		return false
	}

	// Check if this element's component implements the Modal interface
	// Modal interface: IsOpen() bool, ShouldCloseOnEscape() bool, Close()
	type EscapeableModal interface {
		IsOpen() bool
		ShouldCloseOnEscape() bool
		Close()
	}

	if component := elem.Component; component != nil {
		if modal, ok := component.(EscapeableModal); ok {
			if modal.IsOpen() && modal.ShouldCloseOnEscape() {
				modal.Close()
				return true
			}
		}
	}

	// Recurse into children to find nested modals
	for _, child := range elem.Children {
		if handleModalEscape(child) {
			return true
		}
	}

	return false
}

// findScrollableElement searches for the first element with overflow:auto or flex-grow > 0
// Returns element path or empty string if none found
func findScrollableElement(elem *vdom.Element) string {
	if elem == nil {
		return ""
	}

	// Check current element
	if elem.Props.Styles != nil {
		// Explicit overflow:auto takes priority
		if elem.Props.Styles["overflow"] == "auto" {
			return elem.Path
		}
		// Smart default: flex-grow > 0
		if flexGrow := elem.Props.Styles["flex-grow"]; flexGrow != "" && flexGrow != "0" {
			return elem.Path
		}
	}

	// Recurse into children
	for _, child := range elem.Children {
		if path := findScrollableElement(child); path != "" {
			return path
		}
	}

	return ""
}
