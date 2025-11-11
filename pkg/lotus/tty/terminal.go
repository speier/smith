// Package terminal provides a high-level API for terminal I/O operations.
//
// Architecture:
//   - tty.Terminal: High-level event loop with callbacks
//   - tty.Screen: Screen management (alt buffer, cursor, clear, size)
//   - tty.InputReader: Keyboard input handling with escape filtering
//
// This package separates terminal I/O concerns from application logic.
// Applications should use Terminal with OnRender/OnKey/OnTick callbacks.
//
// Example:
//
//	term, _ := tty.New()
//	term.OnRender(func() string { return "Hello!" })
//	term.OnKey(func(e KeyEvent) bool { return !e.IsCtrlC() })
//	term.Start()
package tty

import (
	"os"
	"time"
)

// RenderFunc is called to render content to the screen
type RenderFunc func() string

// PostRenderFunc is called after content has been printed (for cursor positioning)
type PostRenderFunc func()

// KeyHandler is called when a key is pressed
type KeyHandler func(event KeyEvent) bool // return false to stop the loop

// TickHandler is called on each tick (for animations/blinking cursor)
type TickHandler func()

// ResizeHandler is called when the terminal size changes
type ResizeHandler func(width, height int)

// Terminal provides a high-level API for terminal I/O
type Terminal struct {
	screen         *Screen
	input          *InputReader
	renderFunc     RenderFunc
	postRenderFunc PostRenderFunc
	keyHandler     KeyHandler
	tickHandler    TickHandler
	resizeHandler  ResizeHandler
	tickRate       time.Duration
	useAltScreen   bool
	lastWidth      int
	lastHeight     int
	renderChan     chan bool // Channel to request renders externally
}

// New creates a new Terminal instance
func New() (*Terminal, error) {
	screen, err := NewScreen()
	if err != nil {
		return nil, err
	}

	width, height := screen.Size()

	input, err := NewInputReader()
	if err != nil {
		return nil, err
	}

	return &Terminal{
		screen:       screen,
		input:        input,
		tickRate:     0,    // No ticking by default
		useAltScreen: true, // Try with simpler alt screen sequence
		lastWidth:    width,
		lastHeight:   height,
		renderChan:   make(chan bool, 10), // Buffered to avoid blocking
	}, nil
}

// Size returns the terminal dimensions
func (t *Terminal) Size() (int, int) {
	return t.screen.Size()
}

// OnRender sets the render function
func (t *Terminal) OnRender(fn RenderFunc) {
	t.renderFunc = fn
}

// OnPostRender sets the post-render function (called after content is printed)
// This is useful for positioning the cursor after the content has been displayed
func (t *Terminal) OnPostRender(fn PostRenderFunc) {
	t.postRenderFunc = fn
}

// OnKey sets the key handler
func (t *Terminal) OnKey(fn KeyHandler) {
	t.keyHandler = fn
}

// OnTick sets the tick handler and tick rate
func (t *Terminal) OnTick(rate time.Duration, fn TickHandler) {
	t.tickRate = rate
	t.tickHandler = fn
}

// OnResize sets the resize handler
func (t *Terminal) OnResize(fn ResizeHandler) {
	t.resizeHandler = fn
}

// SetFilterMouse enables/disables mouse event filtering
func (t *Terminal) SetFilterMouse(filter bool) {
	t.input.SetFilterMouse(filter)
}

// SetUseAltScreen enables/disables alternate screen buffer
// Note: Some terminals have issues with Ctrl+C when alt screen is enabled
func (t *Terminal) SetUseAltScreen(use bool) {
	t.useAltScreen = use
}

// ShowCursor shows the terminal cursor
func (t *Terminal) ShowCursor() {
	t.screen.ShowCursor()
}

// HideCursor hides the terminal cursor
func (t *Terminal) HideCursor() {
	t.screen.HideCursor()
}

// MoveCursor moves the cursor to the specified position (1-indexed)
func (t *Terminal) MoveCursor(row, col int) {
	t.screen.MoveCursor(row, col)
}

// ForceCleanup forcefully restores terminal to normal mode (for HMR restart)
func (t *Terminal) ForceCleanup() {
	t.screen.Restore()
	_ = t.input.Close()
}

// CancelInput cancels the input reader, causing Start() to exit cleanly
func (t *Terminal) CancelInput() {
	if t.input != nil {
		t.input.Cancel()
	}
}

// Start initializes the terminal and starts the event loop
func (t *Terminal) Start() error {
	// Setup terminal
	if err := t.screen.SetRawMode(); err != nil {
		return err
	}
	defer t.screen.Restore()
	defer func() { _ = t.input.Close() }()

	if t.useAltScreen {
		t.screen.EnterAltScreen()
	} else {
		t.screen.Clear()
	}
	// Hide terminal cursor - we render our own cursor characters
	t.screen.HideCursor()

	// Initial render
	t.render()

	// Start ticker if configured
	var ticker *time.Ticker
	var tickChan <-chan time.Time
	if t.tickRate > 0 && t.tickHandler != nil {
		ticker = time.NewTicker(t.tickRate)
		tickChan = ticker.C
		defer ticker.Stop()
	}

	// Event loop channels
	keyChan := make(chan *KeyEvent, 100) // Larger buffer for fast typing
	errChan := make(chan error, 1)
	resizeChan := make(chan struct{}, 1)

	// Keyboard reader goroutine
	go func() {
		for {
			event, err := t.input.ReadKey()
			if err != nil {
				errChan <- err
				return
			}
			if event != nil {
				keyChan <- event
			}
		}
	}()

	// Resize checker goroutine (check periodically, signal if changed)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for range ticker.C {
			width, height := t.screen.Size()
			if width != t.lastWidth || height != t.lastHeight {
				select {
				case resizeChan <- struct{}{}:
				default:
					// Already pending
				}
			}
		}
	}()

	// Main event loop - NO busy waiting, pure event-driven
	for {
		select {
		case event := <-keyChan:
			// Handle key event
			var shouldContinue bool
			if t.keyHandler != nil {
				shouldContinue = t.keyHandler(*event)
			} else {
				// Default: stop on Ctrl+C or Ctrl+D
				shouldContinue = !event.IsCtrlC() && !event.IsCtrlD()
			}

			if !shouldContinue {
				t.input.Cancel()
				return nil
			}

			// Immediate render for instant feedback
			t.render()

		case <-tickChan:
			// Handle tick
			if t.tickHandler != nil {
				t.tickHandler()
			}
			t.render()

		case <-resizeChan:
			// Handle resize
			width, height := t.screen.Size()
			t.lastWidth = width
			t.lastHeight = height
			if t.resizeHandler != nil {
				t.resizeHandler(width, height)
			}
			t.render()

		case <-t.renderChan:
			// External render request (e.g., from HMR/DevTools)
			t.render()

		case err := <-errChan:
			return err
		}
	}
}

// RequestRender requests a render from external sources (non-blocking)
func (t *Terminal) RequestRender() {
	select {
	case t.renderChan <- true:
	default:
		// Channel full, render already pending
	}
}

// render calls the render function and updates the screen
func (t *Terminal) render() {
	if t.renderFunc != nil {
		t.screen.Redraw()
		output := t.renderFunc()
		t.screen.Print(output)
		// Flush to ensure output is displayed before cursor positioning
		_ = os.Stdout.Sync()
	}
	// Position cursor AFTER content is printed
	if t.postRenderFunc != nil {
		t.postRenderFunc()
	}
}
