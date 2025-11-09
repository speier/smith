package terminal

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
func needsAltScreenForNoScroll() bool {
	termProgram := os.Getenv("TERM_PROGRAM")
	// VS Code terminal scrolls without alt screen
	return termProgram == "vscode"
}

// Size returns the terminal dimensions
func (s *Screen) Size() (int, int) {
	return s.width, s.height
}

// SetRawMode enables raw terminal mode (no line buffering, no echo)
func (s *Screen) SetRawMode() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	s.oldState = oldState
	return nil
}

// Restore restores the terminal to its original state
func (s *Screen) Restore() {
	if s.cursorHidden {
		s.ShowCursor()
	}
	if s.inAltMode {
		s.ExitAltScreen()
	}
	if s.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), s.oldState)
	}
}

// Clear clears the entire screen
func (s *Screen) Clear() {
	// Disable scrolling by setting scroll region to full screen
	fmt.Printf("\033[1;%dr", s.height) // Set scroll region to entire screen
	fmt.Print("\033[2J\033[H")         // Clear screen and move cursor to top
}

// Redraw moves cursor to home position without clearing
func (s *Screen) Redraw() {
	fmt.Print("\033[H") // Move to home position
}

// EnterAltScreen switches to alternate screen buffer
func (s *Screen) EnterAltScreen() {
	// Some terminals (VS Code) need alt screen to prevent scrolling
	// Others (Ghostty, iTerm2, Alacritty) work better without it
	if needsAltScreenForNoScroll() {
		fmt.Print("\033[?47h") // Use simpler alt screen (better Ctrl+C handling)
		fmt.Print("\033[2J")   // Clear the alternate screen
	} else {
		// Native terminals: Just clear screen, no alt buffer needed
		fmt.Print("\033[2J\033[H")
	}
	s.inAltMode = true
}

// ExitAltScreen switches back to normal screen buffer
func (s *Screen) ExitAltScreen() {
	if needsAltScreenForNoScroll() {
		fmt.Print("\033[?47l") // Restore screen
	}
	s.inAltMode = false
}

// HideCursor hides the terminal cursor
func (s *Screen) HideCursor() {
	fmt.Print("\033[?25l")
	s.cursorHidden = true
}

// ShowCursor shows the terminal cursor
func (s *Screen) ShowCursor() {
	fmt.Print("\033[?25h")
	s.cursorHidden = false
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
