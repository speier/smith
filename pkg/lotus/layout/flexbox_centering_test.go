package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "plain text",
			input:    "HELLO",
			expected: 5,
		},
		{
			name:     "text with green bold ANSI",
			input:    "\x1b[38;5;10;1mHELLO\x1b[0m",
			expected: 5,
		},
		{
			name:     "logo line with ANSI",
			input:    "\x1b[38;5;10;1m███████╗███╗   ███╗██╗████████╗██╗  ██╗\x1b[0m",
			expected: 39, // Actual character count
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "only ANSI codes",
			input:    "\x1b[38;5;10;1m\x1b[0m",
			expected: 0,
		},
		{
			name:     "multiple ANSI sequences",
			input:    "\x1b[31mRED\x1b[0m \x1b[32mGREEN\x1b[0m",
			expected: 9, // "RED GREEN" = 9 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := visibleLen(tt.input)
			if got != tt.expected {
				t.Errorf("visibleLen(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestComputeIntrinsicWidthWithANSI(t *testing.T) {
	// Test that ComputeIntrinsicWidth correctly handles ANSI-styled text
	// by using visibleLen to exclude escape codes from width calculation

	tests := []struct {
		name          string
		element       *vdom.Element
		expectedWidth int
	}{
		{
			name:          "plain text",
			element:       vdom.Text("Hello"),
			expectedWidth: 5,
		},
		{
			name:          "text with color",
			element:       vdom.Text("Hello").WithColor("red"),
			expectedWidth: 5, // ANSI codes should not count
		},
		{
			name:          "text with bold",
			element:       vdom.Text("Hello").WithBold(),
			expectedWidth: 5,
		},
		{
			name:          "text with color and bold",
			element:       vdom.Text("World").WithColor("bright-cyan").WithBold(),
			expectedWidth: 5,
		},
		{
			name:          "multi-line text with ANSI",
			element:       vdom.Text("Line 1\nLine 2").WithColor("green"),
			expectedWidth: 6, // Longest line is "Line 1" or "Line 2" = 6 chars
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply styles to generate ANSI codes
			resolver := style.NewResolver("")
			styledNode := resolver.Resolve(tt.element)

			// Compute intrinsic width
			width := ComputeIntrinsicWidth(styledNode)

			if width != tt.expectedWidth {
				t.Errorf("Expected width %d, got %d", tt.expectedWidth, width)
			}
		})
	}
}
