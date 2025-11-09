// Package terminal provides a high-level API for terminal I/O operations.
//
// Architecture:
//   - terminal.Terminal: High-level event loop with callbacks
//   - terminal.Screen: Screen management (alt buffer, cursor, clear, size)
//   - terminal.InputReader: Keyboard input handling with escape filtering
//
// This package separates terminal I/O concerns from application logic.
// Applications should use Terminal with OnRender/OnKey/OnTick callbacks.
//
// Example:
//
//	term, _ := terminal.New()
//	term.OnRender(func() string { return "Hello!" })
//	term.OnKey(func(e KeyEvent) bool { return !e.IsCtrlC() })
//	term.Start()
package terminal

import (
	"time"
)

// RenderFunc is called to render content to the screen
type RenderFunc func() string

// KeyHandler is called when a key is pressed
type KeyHandler func(event KeyEvent) bool // return false to stop the loop

// TickHandler is called on each tick (for animations/blinking cursor)
type TickHandler func()

// ResizeHandler is called when the terminal size changes
type ResizeHandler func(width, height int)

// Terminal provides a high-level API for terminal I/O
type Terminal struct {
	screen        *Screen
	input         *InputReader
	renderFunc    RenderFunc
	keyHandler    KeyHandler
	tickHandler   TickHandler
	resizeHandler ResizeHandler
	tickRate      time.Duration
	useAltScreen  bool
	lastWidth     int
	lastHeight    int
}

// New creates a new Terminal instance
func New() (*Terminal, error) {
	screen, err := NewScreen()
	if err != nil {
		return nil, err
	}

	width, height := screen.Size()

	return &Terminal{
		screen:       screen,
		input:        NewInputReader(),
		tickRate:     0,    // No ticking by default
		useAltScreen: true, // Try with simpler alt screen sequence
		lastWidth:    width,
		lastHeight:   height,
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

// Start initializes the terminal and starts the event loop
func (t *Terminal) Start() error {
	// Setup terminal
	if err := t.screen.SetRawMode(); err != nil {
		return err
	}
	defer t.screen.Restore()

	if t.useAltScreen {
		t.screen.EnterAltScreen()
	} else {
		t.screen.Clear()
	}
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

	// Event loop
	keyChan := make(chan *KeyEvent, 10)
	errChan := make(chan error, 1)

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

	// Main loop
	for {
		// Check for terminal resize
		width, height := t.screen.Size()
		if width != t.lastWidth || height != t.lastHeight {
			t.lastWidth = width
			t.lastHeight = height
			if t.resizeHandler != nil {
				t.resizeHandler(width, height)
			}
			t.render()
		}

		select {
		case event := <-keyChan:
			// Handle key event
			if t.keyHandler != nil {
				if !t.keyHandler(*event) {
					return nil
				}
			}
			// Default: stop on Ctrl+C or Ctrl+D
			if event.IsCtrlC() || event.IsCtrlD() {
				return nil
			}
			t.render()

		case <-tickChan:
			// Handle tick
			if t.tickHandler != nil {
				t.tickHandler()
			}
			t.render()

		case err := <-errChan:
			return err

		default:
			// Small sleep to prevent busy loop when checking resize
			time.Sleep(50 * time.Millisecond)
		}
	}
}

// render calls the render function and updates the screen
func (t *Terminal) render() {
	if t.renderFunc != nil {
		t.screen.Redraw()
		t.screen.Print(t.renderFunc())
	}
}
