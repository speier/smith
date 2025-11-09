package terminal

import (
	"os"
)

// KeyEvent represents a keyboard input event
type KeyEvent struct {
	Key  byte   // ASCII value of the key (for simple keys)
	Char string // Character representation (empty for control keys)
	Code string // ANSI escape sequence code for special keys
}

// Special key codes
const (
	KeyCtrlC      = 3
	KeyCtrlD      = 4
	KeyBackspace  = 127
	KeyBackspace2 = 8 // Some terminals use 8
	KeyEnter      = 13
	KeyEnter2     = 10 // Some terminals use 10
	KeyEscape     = 27
	KeyDelete     = 0 // Delete key is sent as escape sequence
	KeyLeft       = 0 // Arrow keys are sent as escape sequences
	KeyRight      = 0
	KeyUp         = 0
	KeyDown       = 0
	KeyHome       = 0
	KeyEnd        = 0
)

// Special key escape sequences
const (
	SeqDelete     = "[3~"
	SeqLeft       = "[D"
	SeqRight      = "[C"
	SeqUp         = "[A"
	SeqDown       = "[B"
	SeqHome       = "[H"
	SeqHome2      = "[1~"
	SeqEnd        = "[F"
	SeqEnd2       = "[4~"
	SeqCtrlLeft   = "[1;5D" // Ctrl+Left (word jump left)
	SeqCtrlRight  = "[1;5C" // Ctrl+Right (word jump right)
	SeqAltLeft    = "[1;3D" // Alt+Left (word jump left on some terminals)
	SeqAltRight   = "[1;3C" // Alt+Right (word jump right on some terminals)
	SeqCmdLeft    = "[1;9D" // Cmd+Left on Mac (beginning of line)
	SeqCmdRight   = "[1;9C" // Cmd+Right on Mac (end of line)
	SeqShiftLeft  = "[1;2D" // Shift+Left (for future selection support)
	SeqShiftRight = "[1;2C" // Shift+Right (for future selection support)
)

// IsCtrlC checks if the key is Ctrl+C
func (k KeyEvent) IsCtrlC() bool {
	return k.Key == KeyCtrlC
}

// IsCtrlD checks if the key is Ctrl+D
func (k KeyEvent) IsCtrlD() bool {
	return k.Key == KeyCtrlD
}

// IsEnter checks if the key is Enter
func (k KeyEvent) IsEnter() bool {
	return k.Key == KeyEnter || k.Key == KeyEnter2
}

// IsBackspace checks if the key is Backspace
func (k KeyEvent) IsBackspace() bool {
	return k.Key == KeyBackspace || k.Key == KeyBackspace2
}

// IsPrintable checks if the key is a printable character
func (k KeyEvent) IsPrintable() bool {
	return k.Key >= 32 && k.Key < 127
}

// InputReader handles keyboard input from stdin
type InputReader struct {
	filterMouse  bool
	escapeBuffer []byte
	inEscapeSeq  bool
}

// NewInputReader creates a new input reader
func NewInputReader() *InputReader {
	return &InputReader{
		filterMouse:  false,
		escapeBuffer: make([]byte, 0, 10),
		inEscapeSeq:  false,
	}
}

// SetFilterMouse enables/disables mouse event filtering
func (r *InputReader) SetFilterMouse(filter bool) {
	r.filterMouse = filter
}

// ReadKey reads a single key event from stdin
// Returns nil if an escape sequence should be filtered
func (r *InputReader) ReadKey() (*KeyEvent, error) {
	buf := make([]byte, 1)
	_, err := os.Stdin.Read(buf)
	if err != nil {
		return nil, err
	}

	b := buf[0]

	// CRITICAL: Always handle control keys first (Ctrl+C, Ctrl+D)
	// These should NEVER be filtered, even during escape sequences
	if b == KeyCtrlC || b == KeyCtrlD {
		r.inEscapeSeq = false
		r.escapeBuffer = r.escapeBuffer[:0]
		event := &KeyEvent{Key: b}
		return event, nil
	}

	// Handle escape sequences
	if b == KeyEscape {
		r.inEscapeSeq = true
		r.escapeBuffer = r.escapeBuffer[:0]
		r.escapeBuffer = append(r.escapeBuffer, b)
		return nil, nil
	}

	// Build up escape sequence
	if r.inEscapeSeq {
		r.escapeBuffer = append(r.escapeBuffer, b)

		// Check if sequence is complete
		seq := string(r.escapeBuffer[1:]) // Skip ESC byte

		// Word/line navigation (check longer sequences first)
		if seq == SeqCtrlLeft || seq == SeqAltLeft {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqCtrlLeft}, nil
		}
		if seq == SeqCtrlRight || seq == SeqAltRight {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqCtrlRight}, nil
		}
		if seq == SeqCmdLeft {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqCmdLeft}, nil
		}
		if seq == SeqCmdRight {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqCmdRight}, nil
		}

		// Arrow keys and Home/End
		if seq == SeqLeft {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqLeft}, nil
		}
		if seq == SeqRight {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqRight}, nil
		}
		if seq == SeqUp {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqUp}, nil
		}
		if seq == SeqDown {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqDown}, nil
		}
		if seq == SeqHome || seq == SeqHome2 {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqHome}, nil
		}
		if seq == SeqEnd || seq == SeqEnd2 {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqEnd}, nil
		}
		if seq == SeqDelete {
			r.inEscapeSeq = false
			return &KeyEvent{Code: SeqDelete}, nil
		}

		// If mouse filtering is enabled, filter mouse events
		if r.filterMouse {
			// Mouse events typically look like [<0;x;yM or similar
			if b == 'M' || b == 'm' || (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == '~' {
				r.inEscapeSeq = false
				r.escapeBuffer = r.escapeBuffer[:0]
				return nil, nil // Filter
			}
		}

		// Not complete yet, wait for more bytes
		if len(r.escapeBuffer) > 10 {
			// Sequence too long, reset
			r.inEscapeSeq = false
			r.escapeBuffer = r.escapeBuffer[:0]
		}
		return nil, nil
	}

	// Regular key
	event := &KeyEvent{Key: b}
	if b >= 32 && b < 127 {
		event.Char = string(b)
	}

	return event, nil
}
