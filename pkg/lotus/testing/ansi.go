// Package testing provides testing utilities for Lotus framework.
//
// The hex capture approach allows deterministic testing of ANSI sequences
// without requiring actual terminal emulation or file I/O.
package testing

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
)

// ANSICapture captures terminal output for testing and validation
type ANSICapture struct {
	buf     *bytes.Buffer
	maxSize int
}

// NewCapture creates a new ANSI output capture with size limit
func NewCapture(maxSize int) *ANSICapture {
	return &ANSICapture{
		buf:     &bytes.Buffer{},
		maxSize: maxSize,
	}
}

// Write implements io.Writer interface
func (c *ANSICapture) Write(p []byte) (n int, err error) {
	remaining := c.maxSize - c.buf.Len()
	if remaining <= 0 {
		return 0, nil // Silently ignore writes beyond limit
	}
	toWrite := p
	if len(toWrite) > remaining {
		toWrite = toWrite[:remaining]
	}
	return c.buf.Write(toWrite)
}

// String returns the captured output as a string
func (c *ANSICapture) String() string {
	return c.buf.String()
}

// Bytes returns the captured output as bytes
func (c *ANSICapture) Bytes() []byte {
	return c.buf.Bytes()
}

// Hex returns a formatted hex dump of captured bytes
func (c *ANSICapture) Hex() string {
	var out strings.Builder
	data := c.buf.Bytes()

	out.WriteString("=== HEX VIEW ===\n")
	for i := 0; i < len(data); i++ {
		if i > 0 && i%16 == 0 {
			out.WriteString("\n")
		}
		fmt.Fprintf(&out, "%02x ", data[i])
	}
	out.WriteString("\n")

	return out.String()
}

// Reset clears the capture buffer
func (c *ANSICapture) Reset() {
	c.buf.Reset()
}

// AssertContains checks if captured output contains the expected string
func (c *ANSICapture) AssertContains(t *testing.T, expected string) {
	t.Helper()
	if !strings.Contains(c.String(), expected) {
		t.Errorf("Expected output to contain %q, got:\n%s\n\nHex:\n%s", expected, c.String(), c.Hex())
	}
}

// AssertNotContains checks if captured output does NOT contain the string
func (c *ANSICapture) AssertNotContains(t *testing.T, unexpected string) {
	t.Helper()
	if strings.Contains(c.String(), unexpected) {
		t.Errorf("Expected output to NOT contain %q, but it does.\nHex:\n%s", unexpected, c.Hex())
	}
}

// AssertSequence checks if ANSI sequences appear in order
func (c *ANSICapture) AssertSequence(t *testing.T, sequences ...string) {
	t.Helper()
	output := c.String()
	lastPos := -1

	for _, seq := range sequences {
		pos := strings.Index(output, seq)
		if pos == -1 {
			t.Errorf("Expected sequence %q not found in output.\nHex:\n%s", seq, c.Hex())
			return
		}
		if pos < lastPos {
			t.Errorf("Sequence %q found at position %d, but should appear after position %d.\nHex:\n%s",
				seq, pos, lastPos, c.Hex())
			return
		}
		lastPos = pos
	}
}

// AssertNoLeadingNewlines verifies no \r\n or \n appears before first printable content
func (c *ANSICapture) AssertNoLeadingNewlines(t *testing.T) {
	t.Helper()
	data := c.Bytes()

	// Skip ANSI escape sequences
	inEscape := false
	for i := 0; i < len(data); i++ {
		b := data[i]

		// Start of escape sequence
		if b == 0x1b {
			inEscape = true
			continue
		}

		// Inside escape sequence
		if inEscape {
			// End of CSI sequence (typically a letter or specific char)
			if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || b == 'h' || b == 'l' || b == 'H' {
				inEscape = false
			}
			continue
		}

		// First non-escape character
		if b == '\r' || b == '\n' {
			t.Errorf("Found leading newline (byte %02x) at position %d before first content.\nHex:\n%s", b, i, c.Hex())
			return
		}

		// Found first printable content, we're done
		if b >= 0x20 {
			return
		}
	}
}

// AssertFirstLine extracts first visible line and compares to expected
// Strips ANSI color codes but preserves content
func (c *ANSICapture) AssertFirstLine(t *testing.T, expected string) {
	t.Helper()

	// Strip ANSI sequences from captured output
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)
	stripped := ansiRegex.ReplaceAllString(c.String(), "")

	// Get first line
	lines := strings.Split(stripped, "\n")
	if len(lines) == 0 {
		t.Fatalf("No lines found in output.\nHex:\n%s", c.Hex())
	}

	firstLine := strings.TrimSpace(lines[0])
	if firstLine != expected {
		t.Errorf("First line mismatch.\nExpected: %q\nGot: %q\n\nHex:\n%s", expected, firstLine, c.Hex())
	}
}

// AssertCursorAt checks if cursor positioning sequence ESC[row;colH appears
func (c *ANSICapture) AssertCursorAt(t *testing.T, row, col int) {
	t.Helper()

	// Look for ESC[row;colH or ESC[H (home = 1;1)
	expected := fmt.Sprintf("\x1b[%d;%dH", row, col)
	if row == 1 && col == 1 {
		// Also accept ESC[H as home
		if strings.Contains(c.String(), "\x1b[H") || strings.Contains(c.String(), expected) {
			return
		}
	} else if strings.Contains(c.String(), expected) {
		return
	}

	t.Errorf("Expected cursor position ESC[%d;%dH not found.\nHex:\n%s", row, col, c.Hex())
}

// AssertClearScreen verifies screen clear sequence is present
func (c *ANSICapture) AssertClearScreen(t *testing.T) {
	t.Helper()
	if !strings.Contains(c.String(), "\x1b[2J") {
		t.Errorf("Expected screen clear sequence ESC[2J not found.\nHex:\n%s", c.Hex())
	}
}

// AssertSynchronizedOutput verifies synchronized output mode is used
func (c *ANSICapture) AssertSynchronizedOutput(t *testing.T) {
	t.Helper()
	output := c.String()

	hasBegin := strings.Contains(output, "\x1b[?2026h")
	hasEnd := strings.Contains(output, "\x1b[?2026l")

	if !hasBegin || !hasEnd {
		t.Errorf("Expected synchronized output sequences (ESC[?2026h and ESC[?2026l).\nFound begin: %v, end: %v\nHex:\n%s",
			hasBegin, hasEnd, c.Hex())
	}
}
