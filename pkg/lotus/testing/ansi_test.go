package testing

import (
	"testing"
)

// TestANSICapture validates the capture utility itself
func TestANSICapture(t *testing.T) {
	t.Run("Basic capture", func(t *testing.T) {
		capture := NewCapture(1024)
		n, err := capture.Write([]byte("Hello, World!"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != 13 {
			t.Errorf("Expected to write 13 bytes, got %d", n)
		}
		if capture.String() != "Hello, World!" {
			t.Errorf("Expected %q, got %q", "Hello, World!", capture.String())
		}
	})

	t.Run("Size limit", func(t *testing.T) {
		capture := NewCapture(10)
		_, _ = capture.Write([]byte("0123456789ABCDEF"))
		if len(capture.Bytes()) != 10 {
			t.Errorf("Expected 10 bytes captured, got %d", len(capture.Bytes()))
		}
	})

	t.Run("Hex output", func(t *testing.T) {
		capture := NewCapture(100)
		_, _ = capture.Write([]byte("\x1b[H"))
		hex := capture.Hex()
		if hex == "" {
			t.Error("Hex output is empty")
		}
		// Should contain the hex values for ESC[H
		// ESC = 1b, [ = 5b, H = 48
	})
}

// TestAssertions validates the assertion helper methods
func TestAssertions(t *testing.T) {
	t.Run("AssertContains", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("Hello, World!"))

		// This should pass
		mockT := &testing.T{}
		capture.AssertContains(mockT, "Hello")

		// Verify it doesn't fail when content exists
		if mockT.Failed() {
			t.Error("AssertContains should not fail when content exists")
		}
	})

	t.Run("AssertSequence", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("\x1b[2J\x1b[H\x1b[1;96mHello"))

		mockT := &testing.T{}
		capture.AssertSequence(mockT, "\x1b[2J", "\x1b[H", "Hello")

		if mockT.Failed() {
			t.Error("AssertSequence should not fail when sequences are in order")
		}
	})

	t.Run("AssertCursorAt", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("\x1b[2;5H"))

		mockT := &testing.T{}
		capture.AssertCursorAt(mockT, 2, 5)

		if mockT.Failed() {
			t.Error("AssertCursorAt should not fail when cursor position exists")
		}
	})

	t.Run("AssertClearScreen", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("\x1b[2J\x1b[H"))

		mockT := &testing.T{}
		capture.AssertClearScreen(mockT)

		if mockT.Failed() {
			t.Error("AssertClearScreen should not fail when clear sequence exists")
		}
	})

	t.Run("AssertSynchronizedOutput", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("\x1b[?2026h..content..\x1b[?2026l"))

		mockT := &testing.T{}
		capture.AssertSynchronizedOutput(mockT)

		if mockT.Failed() {
			t.Error("AssertSynchronizedOutput should not fail when sequences exist")
		}
	})
}

// TestNoLeadingNewlines validates detection of leading newlines
func TestNoLeadingNewlines(t *testing.T) {
	t.Run("Clean start with ESC sequences", func(t *testing.T) {
		capture := NewCapture(1024)
		// Start with clear screen, home, then content
		_, _ = capture.Write([]byte("\x1b[2J\x1b[HHello"))

		mockT := &testing.T{}
		capture.AssertNoLeadingNewlines(mockT)

		if mockT.Failed() {
			t.Error("Should not detect leading newlines when starting clean")
		}
	})

	t.Run("Detect leading newline", func(t *testing.T) {
		capture := NewCapture(1024)
		// Start with newline before content
		_, _ = capture.Write([]byte("\n\x1b[2J\x1b[HHello"))

		mockT := &testing.T{}
		capture.AssertNoLeadingNewlines(mockT)

		if !mockT.Failed() {
			t.Error("Should detect leading newline")
		}
	})

	t.Run("Detect carriage return + newline", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("\x1b[2J\x1b[H\r\nHello"))

		mockT := &testing.T{}
		capture.AssertNoLeadingNewlines(mockT)

		if !mockT.Failed() {
			t.Error("Should detect CRLF after escape sequences")
		}
	})
}

// TestFirstLineExtraction validates first line parsing
func TestFirstLineExtraction(t *testing.T) {
	t.Run("Simple text", func(t *testing.T) {
		capture := NewCapture(1024)
		_, _ = capture.Write([]byte("Hello\nWorld"))

		mockT := &testing.T{}
		capture.AssertFirstLine(mockT, "Hello")

		if mockT.Failed() {
			t.Error("Should extract first line correctly")
		}
	})

	t.Run("With ANSI codes", func(t *testing.T) {
		capture := NewCapture(1024)
		// Bright cyan "Hello"
		_, _ = capture.Write([]byte("\x1b[1;96mHello\x1b[0m\nWorld"))

		mockT := &testing.T{}
		capture.AssertFirstLine(mockT, "Hello")

		if mockT.Failed() {
			t.Error("Should strip ANSI codes and extract first line")
		}
	})
}
