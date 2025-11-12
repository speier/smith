package layout

import (
	"testing"
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
	// This tests that computeIntrinsicWidth correctly uses visibleLen
	// We'll need to create a styled node with ANSI text

	// TODO: Add test cases for actual styled nodes with ANSI codes
	// This would require creating vdom elements and style nodes
	t.Skip("Need to add integration test with actual styled nodes")
}
