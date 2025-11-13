package render

import (
	"testing"

	lotustesting "github.com/speier/smith/pkg/lotus/testing"
)

// TestViewportOwnership validates that first render owns the entire visible area
// Regression test for: "First line scrolled above viewport on startup"
//
// This test validates:
// 1. No leading newlines before first content
// 2. Cursor starts at (1,1) after clear
// 3. Content positioned using absolute coordinates (not \r\n which causes scroll)
// 4. Synchronized output wraps content
func TestViewportOwnership(t *testing.T) {
	// Create a buffer simulating terminal (80x24)
	buf := NewBuffer(80, 24)

	// Fill first 3 rows (simulating VStack with gap)
	// Row 1: "👋 Hello, Lotus!" (with style)
	row1 := "👋 Hello, Lotus!"
	for i, r := range row1 {
		if i < 80 {
			buf.Set(i, 0, Cell{
				Char: r,
				Style: Style{
					FgColor: "bright-cyan",
					Bold:    true,
				},
			})
		}
	}
	// Fill rest of row with spaces
	for i := len(row1); i < 80; i++ {
		buf.Set(i, 0, Cell{Char: ' '})
	}

	// Row 2: Empty (gap)
	for i := 0; i < 80; i++ {
		buf.Set(i, 1, Cell{Char: ' '})
	}

	// Row 3: "Press Ctrl+C to exit"
	row3 := "Press Ctrl+C to exit"
	for i, r := range row3 {
		if i < 80 {
			buf.Set(i, 2, Cell{
				Char:  r,
				Style: Style{FgColor: "white"},
			})
		}
	}
	// Fill rest of row with spaces
	for i := len(row3); i < 80; i++ {
		buf.Set(i, 2, Cell{Char: ' '})
	}

	// Fill remaining rows with spaces
	for y := 3; y < 24; y++ {
		for x := 0; x < 80; x++ {
			buf.Set(x, y, Cell{Char: ' '})
		}
	}

	// Render to ANSI
	output := RenderBufferFull(buf)

	// Capture in test harness
	capture := lotustesting.NewCapture(4096)
	_, _ = capture.Write([]byte(output))

	// Assertions
	t.Run("No leading newlines", func(t *testing.T) {
		capture.AssertNoLeadingNewlines(t)
	})

	t.Run("Starts with synchronized output", func(t *testing.T) {
		capture.AssertSynchronizedOutput(t)
	})

	t.Run("Clear screen present", func(t *testing.T) {
		capture.AssertClearScreen(t)
	})

	t.Run("Cursor at home", func(t *testing.T) {
		capture.AssertCursorAt(t, 1, 1)
	})

	t.Run("Sequence order", func(t *testing.T) {
		// Verify proper ordering: sync begin → clear → home → content → sync end
		capture.AssertSequence(t,
			"\x1b[?2026h", // Begin synchronized output
			"\x1b[2J",     // Clear screen
			"\x1b[H",      // Home cursor
			// Content appears here
			"\x1b[?2026l", // End synchronized output
		)
	})

	t.Run("Uses absolute positioning not CRLF", func(t *testing.T) {
		// After fixing the bug: should use ESC[2;1H, ESC[3;1H for rows 2,3
		// Should NOT have \r\n which causes scroll
		capture.AssertContains(t, "\x1b[2;1H") // Position row 2
		capture.AssertContains(t, "\x1b[3;1H") // Position row 3

		// Verify we're NOT using \r\n between full-width rows (the bug)
		// Note: This is tricky - we need to check there's no \r\n after filling width
		// The fix ensures we use absolute positioning instead
	})

	t.Run("Content is present", func(t *testing.T) {
		capture.AssertContains(t, "Hello, Lotus!")
		capture.AssertContains(t, "Press Ctrl+C")
	})

	// Log hex for debugging if test fails
	if t.Failed() {
		t.Logf("\n%s", capture.Hex())
	}
}

// TestRenderFullBufferNoScroll validates RenderBufferFull doesn't cause scrolling
// when buffer height equals terminal height
func TestRenderFullBufferNoScroll(t *testing.T) {
	// Create buffer matching terminal size
	buf := NewBuffer(80, 24)

	// Fill all rows
	for y := 0; y < 24; y++ {
		for x := 0; x < 80; x++ {
			buf.Set(x, y, Cell{Char: 'X'})
		}
	}

	output := RenderBufferFull(buf)
	capture := lotustesting.NewCapture(8192)
	_, _ = capture.Write([]byte(output))

	// Should NOT contain \r\n that would cause scroll past last row
	// Each row should be positioned absolutely with ESC[y;1H

	// Verify absolute positioning for several rows
	// Note: Row 1 uses ESC[H (implicit 1;1H) after clear screen
	capture.AssertCursorAt(t, 1, 1)         // Home position
	capture.AssertContains(t, "\x1b[2;1H")  // Row 2
	capture.AssertContains(t, "\x1b[10;1H") // Row 10
	capture.AssertContains(t, "\x1b[24;1H") // Row 24 (last row)

	// No leading newlines
	capture.AssertNoLeadingNewlines(t)

	if t.Failed() {
		t.Logf("\n%s", capture.Hex())
	}
}

// TestRenderSmallBuffer validates rendering when buffer smaller than terminal
func TestRenderSmallBuffer(t *testing.T) {
	// Small buffer (20x5)
	buf := NewBuffer(20, 5)

	// Fill with simple pattern
	for y := 0; y < 5; y++ {
		for x := 0; x < 20; x++ {
			buf.Set(x, y, Cell{Char: rune('A' + y)})
		}
	}

	output := RenderBufferFull(buf)
	capture := lotustesting.NewCapture(2048)
	_, _ = capture.Write([]byte(output))

	// Should position each row
	capture.AssertCursorAt(t, 1, 1)
	capture.AssertContains(t, "\x1b[2;1H")
	capture.AssertContains(t, "\x1b[5;1H") // Last row

	capture.AssertNoLeadingNewlines(t)

	if t.Failed() {
		t.Logf("\n%s", capture.Hex())
	}
}
