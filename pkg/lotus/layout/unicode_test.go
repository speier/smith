package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestEmojiWidth tests that emojis and wide Unicode characters are measured correctly
func TestEmojiWidth(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected int // Expected display width
	}{
		{"ASCII only", "Hello", 5},
		{"Single emoji", "📝", 2},
		{"Emoji + text", "📝 Test", 7}, // 2 (emoji) + 1 (space) + 4 (Test)
		{"Multiple emojis", "📝📝", 4},
		{"CJK characters", "你好", 4},                         // Chinese characters are wide (2 each)
		{"Mixed content", "Hi 📝 你好", 10},                    // 2 + 1 + 2 + 1 + 4
		{"Form title", "📝 Form Test - Multiple Inputs", 30}, // 2 (emoji) + 28 (rest)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a text element
			root := vdom.Text(tt.text)

			// Resolve styles
			resolver := style.NewResolver("")
			styled := resolver.Resolve(root)

			// Compute intrinsic width using our Unicode-aware function
			width := ComputeIntrinsicWidth(styled)

			if width != tt.expected {
				t.Errorf("Width mismatch for %q: got %d, want %d", tt.text, width, tt.expected)
			}
		})
	}
}

// TestEmojiInBorderedBox tests that emoji text doesn't overflow box borders
func TestEmojiInBorderedBox(t *testing.T) {
	// Create a box with border containing emoji - this was the actual bug
	root := vdom.Box(
		vdom.Text("📝 Form Test - Multiple Inputs"),
	).WithStyle("border", "rounded").
		WithStyle("padding", "0 1")

	// Resolve styles
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)

	// Compute layout
	layoutBox := Compute(styled, 80, 10)

	// The text width should account for emoji being 2 columns
	// Text: "📝 Form Test - Multiple Inputs" = 30 columns (2 for emoji + 28 for rest)
	// With padding-x: 1 (left) + 30 (text) + 1 (right) = 32
	// With border: 1 (left) + 32 (content) + 1 (right) = 34

	// The box should be wide enough to contain the text without overflow
	expectedMinWidth := 34
	if layoutBox.Width < expectedMinWidth {
		t.Errorf("Box width %d is too narrow for emoji text (expected >= %d)",
			layoutBox.Width, expectedMinWidth)
	}

	t.Logf("✓ Box width %d correctly accommodates emoji text (min %d)",
		layoutBox.Width, expectedMinWidth)
}
