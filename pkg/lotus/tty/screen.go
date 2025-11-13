package tty

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Screen manages terminal screen operations
type Screen struct {
	width        int
	height       int
	inAltMode    bool
	cursorHidden bool
	oldState     *term.State
}

// NewScreen creates a new screen manager
func NewScreen() (*Screen, error) {
	// Auto-detect terminal size
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		// Fallback to default
		width = 100
		height = 40
	}

	return &Screen{
		width:  width,
		height: height,
	}, nil
}

// needsAltScreenForNoScroll detects terminals that need alternate screen to prevent scrolling
// Currently: VS Code integrated terminal needs it, native terminals (Ghostty, iTerm2, Alacritty) don't

// Size returns the terminal dimensions (always checks current size)
func (s *Screen) Size() (int, int) {
	// Always get current terminal size
	width, height, err := term.GetSize(int(os.Stdin.Fd()))
	if err == nil {
		s.width = width
		s.height = height
	}
	return s.width, s.height
}

// SetRawMode enables raw terminal mode (no line buffering, no echo)
func (s *Screen) SetRawMode() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	s.oldState = oldState
	// Enable bracketed paste mode
	fmt.Print("\033[?2004h")
	return nil
}

// Restore restores the terminal to its original state
func (s *Screen) Restore() {
	// Disable bracketed paste mode
	fmt.Print("\033[?2004l")
	if s.cursorHidden {
		s.ShowCursor()
	}
	if s.inAltMode {
		s.ExitAltScreen()
	}
	// Clear the screen before restoring terminal state
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
	if s.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), s.oldState)
	}
}

// Clear clears the entire screen
func (s *Screen) Clear() {
	// Clear visible screen only (retain scrollback)
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
}

// ClearInitial clears screen AND scrollback for a pristine starting view
// Uses CSI 3J (clear scrollback) followed by full clear + cursor home.
// Avoids scroll-region manipulation which could shift content one line.
func (s *Screen) ClearInitial() {
	fmt.Print("\033[3J\033[2J\033[H")
}

// Redraw moves cursor to home position without clearing
func (s *Screen) Redraw() {
	fmt.Print("\033[H") // Move to home position
}

// EnterAltScreen switches to alternate screen buffer
func (s *Screen) EnterAltScreen() {
	// Use 1049h (alternate screen with cursor save/restore) for consistent origin
	fmt.Print("\033[?1049h") // enter alt buffer
	// Fully clear scrollback + screen and reset scroll region + home
	fmt.Print("\033[3J\033[2J\033[r\033[H")
	s.inAltMode = true
}

// ExitAltScreen switches back to normal screen buffer
func (s *Screen) ExitAltScreen() {
	// Leave alt buffer restoring previous screen
	fmt.Print("\033[?1049l")
	s.inAltMode = false
}

// HideCursor hides the terminal cursor
func (s *Screen) HideCursor() {
	fmt.Print("\033[?25l")
	s.cursorHidden = true
}

// ShowCursor shows the terminal cursor
func (s *Screen) ShowCursor() {
	fmt.Print("\033[?25h") // Show cursor
	fmt.Print("\033[1 q")  // Blinking block cursor (DECSCUSR)
	_ = os.Stdout.Sync()   // Ensure cursor command is flushed
	s.cursorHidden = false
}

// MoveCursor moves the cursor to the specified position (1-indexed)
func (s *Screen) MoveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
	_ = os.Stdout.Sync() // Ensure cursor movement is flushed
}

// DisableMouse disables mouse tracking events
func (s *Screen) DisableMouse() {
	fmt.Print("\033[?1000l\033[?1002l\033[?1003l\033[?1006l")
}

// EnableMouse enables mouse tracking events
func (s *Screen) EnableMouse() {
	fmt.Print("\033[?1000h\033[?1002h\033[?1003h\033[?1006h")
}

// Print outputs content to the screen
func (s *Screen) Print(content string) {
	fmt.Print(content)
	// Ensure output is flushed (important for interactive apps)
	_ = os.Stdout.Sync()
}
