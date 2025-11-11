package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestComputeIntrinsicWidth tests that text elements get correct intrinsic width
func TestComputeIntrinsicWidth(t *testing.T) {
	tests := []struct {
		name     string
		element  *vdom.Element
		expected int
	}{
		{
			name:     "prompt",
			element:  vdom.Text("> "),
			expected: 2,
		},
		{
			name:     "short_text",
			element:  vdom.Text("Hello"),
			expected: 5,
		},
		{
			name:     "long_text",
			element:  vdom.Text("Say something..."),
			expected: 16,
		},
		{
			name:     "multiline_text",
			element:  vdom.Text("Line 1\nLonger line 2\nShort"),
			expected: 13, // Length of longest line
		},
		{
			name:     "empty_text",
			element:  vdom.Text(""),
			expected: 1, // Minimum width
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Resolve styles
			resolver := style.NewResolver("")
			styled := resolver.Resolve(tt.element)

			// Compute intrinsic width
			width := computeIntrinsicWidth(styled)

			if width != tt.expected {
				t.Errorf("Expected width %d, got %d for text %q", tt.expected, width, tt.element.Text)
			}
		})
	}
}

// TestHStackWithTextElements tests that HStack correctly layouts text children
func TestHStackWithTextElements(t *testing.T) {
	// Create HStack with prompt + placeholder (like TextInput with placeholder)
	root := vdom.HStack(
		vdom.Text("> "),                                             // prompt (2 chars)
		vdom.Text("Say something...").WithStyle("color", "#808080"), // placeholder (16 chars)
	)

	// Resolve styles
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)

	// Compute layout
	layoutBox := Compute(styled, 80, 3)

	// Verify layout has 2 children
	if len(layoutBox.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layoutBox.Children))
	}

	// First child (prompt) should have width = 2
	prompt := layoutBox.Children[0]
	if prompt.Width != 2 {
		t.Errorf("Expected prompt width 2, got %d", prompt.Width)
	}

	// Second child (placeholder) should have width = 16
	placeholder := layoutBox.Children[1]
	if placeholder.Width != 16 {
		t.Errorf("Expected placeholder width 16, got %d", placeholder.Width)
	}

	// They should be positioned next to each other
	if placeholder.X != prompt.X+prompt.Width {
		t.Errorf("Placeholder should be positioned after prompt: prompt.X=%d prompt.Width=%d placeholder.X=%d",
			prompt.X, prompt.Width, placeholder.X)
	}

	t.Logf("✓ HStack layout correct: prompt w=%d x=%d, placeholder w=%d x=%d",
		prompt.Width, prompt.X, placeholder.Width, placeholder.X)
}

// TestHStackWithMixedContent tests HStack with fixed + flexible children
func TestHStackWithMixedContent(t *testing.T) {
	// HStack with text + flexible box
	root := vdom.HStack(
		vdom.Text("Label: "),                                   // fixed width (7 chars)
		vdom.Box(vdom.Text("content")).WithStyle("flex-grow", "1"), // flexible
	)

	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := Compute(styled, 100, 3)

	if len(layoutBox.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layoutBox.Children))
	}

	label := layoutBox.Children[0]
	content := layoutBox.Children[1]

	// Label should have intrinsic width
	if label.Width != 7 {
		t.Errorf("Expected label width 7, got %d", label.Width)
	}

	// Content should take remaining space
	expectedContentWidth := 100 - 7
	if content.Width != expectedContentWidth {
		t.Errorf("Expected content width %d, got %d", expectedContentWidth, content.Width)
	}

	t.Logf("✓ Mixed layout correct: label w=%d, content w=%d", label.Width, content.Width)
}
