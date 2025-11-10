package frontend

import (
	"fmt"
	"strings"
	"testing"

	"github.com/speier/smith/internal/session"
)

// TestCursorPositioning tests cursor column calculation
func TestCursorPositioning(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	testCases := []struct {
		input      string
		cursorPos  int
		wantColumn int
	}{
		{"", 0, 4},      // Empty input, cursor at start
		{"h", 0, 4},     // Cursor before 'h'
		{"h", 1, 5},     // Cursor after 'h'
		{"hello", 0, 4}, // Cursor at start
		{"hello", 5, 9}, // Cursor at end (after 'o')
		{"hello", 3, 7}, // Cursor in middle (after 'l', before second 'l')
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input=%q,pos=%d", tc.input, tc.cursorPos), func(t *testing.T) {
			ui.input.Value = tc.input
			ui.input.CursorPos = tc.cursorPos
			ui.input.Scroll = 0

			_, cursorOffset := ui.input.GetVisible()
			actualColumn := 4 + cursorOffset

			if actualColumn != tc.wantColumn {
				t.Errorf("Expected cursor at column %d, got %d (offset=%d)",
					tc.wantColumn, actualColumn, cursorOffset)
			}
		})
	}
}

// TestCursorInRenderedOutput manually inspects the rendered output
func TestCursorInRenderedOutput(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	// Type "hello"
	ui.input.Value = "hello"
	ui.input.CursorPos = 5

	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)

	output := ui.cachedUI.RenderToTerminal(true)
	plain := stripANSI(output)

	// Print the output for manual inspection
	t.Logf("Rendered output:\n%s", plain)

	// Find the input line
	lines := strings.Split(plain, "\n")
	for i, line := range lines {
		if strings.Contains(line, ">") && strings.Contains(line, "hello") {
			t.Logf("Input line %d: %q", i, line)

			// Show column positions
			ruler := ""
			for j := 0; j < len(line); j++ {
				ruler += fmt.Sprintf("%d", (j+1)%10)
			}
			t.Logf("Columns:    %s", ruler)

			// Find where "hello" starts
			helloIdx := strings.Index(line, "hello")
			if helloIdx >= 0 {
				t.Logf("'hello' starts at column %d (0-indexed: %d)", helloIdx+1, helloIdx)
				t.Logf("Cursor should be after 'o' at column %d", helloIdx+6)
			}
		}
	}
}
