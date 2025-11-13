package render

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestTextStyling(t *testing.T) {
	tests := []struct {
		name          string
		style         style.ComputedStyle
		expectedCodes []string // ANSI codes we expect to see
	}{
		{
			name: "bold text",
			style: style.ComputedStyle{
				FontWeight: "bold",
				Color:      "#ffffff",
			},
			expectedCodes: []string{"\033[1m"}, // bold
		},
		{
			name: "italic text",
			style: style.ComputedStyle{
				FontStyle: "italic",
				Color:     "#ffffff",
			},
			expectedCodes: []string{"\033[3m"}, // italic
		},
		{
			name: "underlined text",
			style: style.ComputedStyle{
				TextDecoration: "underline",
				Color:          "#ffffff",
			},
			expectedCodes: []string{"\033[4m"}, // underline
		},
		{
			name: "strikethrough text",
			style: style.ComputedStyle{
				TextDecoration: "strikethrough",
				Color:          "#ffffff",
			},
			expectedCodes: []string{"\033[9m"}, // strikethrough
		},
		{
			name: "dim text (opacity 50)",
			style: style.ComputedStyle{
				Opacity: 50,
				Color:   "#ffffff",
			},
			expectedCodes: []string{"\033[2m"}, // dim
		},
		{
			name: "combined: bold + underline",
			style: style.ComputedStyle{
				FontWeight:     "bold",
				TextDecoration: "underline",
				Color:          "#ffffff",
			},
			expectedCodes: []string{"\033[1;4m"}, // bold + underline
		},
		{
			name: "combined: bold + italic + underline",
			style: style.ComputedStyle{
				FontWeight:     "bold",
				FontStyle:      "italic",
				TextDecoration: "underline",
				Color:          "#ffffff",
			},
			expectedCodes: []string{"\033[1;3;4m"}, // bold + italic + underline
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create styled node
			elem := vdom.Text("Test")
			styledNode := &style.StyledNode{
				Element:  elem,
				Style:    tt.style,
				Children: nil,
			}

			// Layout the node
			layoutBox := &layout.LayoutBox{
				Node:     styledNode,
				X:        0,
				Y:        0,
				Width:    10,
				Height:   1,
				Children: nil,
			}

			// Render to buffer
			layoutRenderer := NewLayoutRenderer()
			buffer := layoutRenderer.RenderToBuffer(layoutBox, 10, 1)
			output := RenderBufferFull(buffer)

			// Check for expected ANSI codes
			// Note: codes might be combined with color codes, e.g. "\x1b[1;97m" for bold+white
			for _, code := range tt.expectedCodes {
				// Extract the numeric codes we're looking for
				// e.g. "\x1b[1m" -> "1", "\x1b[1;4m" -> "1;4"
				codeNum := strings.TrimPrefix(code, "\033[")
				codeNum = strings.TrimPrefix(codeNum, "\x1b[")
				codeNum = strings.TrimSuffix(codeNum, "m")

				// Check if the code appears in the output
				// Can be:
				// - Exact match: \x1b[1m
				// - At start: \x1b[1;...m
				// - In middle: \x1b[...;1;...m
				// - At end: \x1b[...;1m
				found := strings.Contains(output, code) ||
					strings.Contains(output, "["+codeNum+";") ||
					strings.Contains(output, ";"+codeNum+";") ||
					strings.Contains(output, ";"+codeNum+"m")

				if !found {
					t.Errorf("Expected ANSI code %q (numeric: %s) not found in output: %q", code, codeNum, output)
				}
			}

			// Verify reset code is present
			if !strings.Contains(output, "\033[0m") {
				t.Error("Expected reset code \\033[0m not found")
			}
		})
	}
}

func TestVisibility(t *testing.T) {
	// Create element
	elem := vdom.Text("Hidden Text")

	// Test visible
	styledVisible := &style.StyledNode{
		Element: elem,
		Style: style.ComputedStyle{
			Visibility: "visible",
			Color:      "#ffffff",
		},
	}

	layoutBox := &layout.LayoutBox{
		Node:   styledVisible,
		X:      0,
		Y:      0,
		Width:  20,
		Height: 1,
	}

	layoutRenderer := NewLayoutRenderer()
	buffer := layoutRenderer.RenderToBuffer(layoutBox, 20, 1)
	output := RenderBufferFull(buffer)

	if !strings.Contains(output, "Hidden Text") {
		t.Error("Visible element should render text")
	}

	// Test hidden
	styledHidden := &style.StyledNode{
		Element: elem,
		Style: style.ComputedStyle{
			Visibility: "hidden",
			Color:      "#ffffff",
		},
	}

	layoutBox.Node = styledHidden
	buffer = layoutRenderer.RenderToBuffer(layoutBox, 20, 1)
	output = RenderBufferFull(buffer)

	// Hidden text should not be rendered (but layout space is preserved)
	if strings.Contains(output, "Hidden Text") {
		t.Error("Hidden element should not render text")
	}
}

func TestBorderColor(t *testing.T) {
	elem := vdom.Box()

	// Test with border-color
	styledNode := &style.StyledNode{
		Element: elem,
		Style: style.ComputedStyle{
			Border:      true,
			BorderStyle: "single",
			BorderColor: "#ff0000", // red
			Color:       "#ffffff", // white (text color)
		},
	}

	layoutBox := &layout.LayoutBox{
		Node:   styledNode,
		X:      0,
		Y:      0,
		Width:  10,
		Height: 5,
	}

	layoutRenderer := NewLayoutRenderer()
	buffer := layoutRenderer.RenderToBuffer(layoutBox, 10, 5)
	output := RenderBufferFull(buffer)

	// Should contain red color code for border
	if !strings.Contains(output, "\033[91m") { // bright red
		t.Error("Border should use border-color (red)")
	}
}

func TestMaxLines(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLines int
		expected string
	}{
		{
			name:     "3 lines clamped to 2",
			text:     "Line 1\nLine 2\nLine 3",
			maxLines: 2,
			expected: "Line 2...", // Last line should have ellipsis
		},
		{
			name:     "5 lines clamped to 3",
			text:     "First\nSecond\nThird\nFourth\nFifth",
			maxLines: 3,
			expected: "Third...",
		},
		{
			name:     "no clamping when lines < maxLines",
			text:     "Line 1\nLine 2",
			maxLines: 5,
			expected: "Line 2", // No ellipsis
		},
		{
			name:     "maxLines = 0 means unlimited",
			text:     "Line 1\nLine 2\nLine 3\nLine 4",
			maxLines: 0,
			expected: "Line 4", // No ellipsis, all lines shown
		},
		{
			name:     "maxLines = 1 shows only first line",
			text:     "First line\nSecond line\nThird line",
			maxLines: 1,
			expected: "First line...",
		},
		{
			name:     "Unicode handling",
			text:     "Hello 世界\nSecond line\nThird line",
			maxLines: 2,
			expected: "Second line...",
		},
		{
			name:     "short last line gets ellipsis appended",
			text:     "Line 1\nAB\nLine 3",
			maxLines: 2,
			expected: "AB...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := vdom.Text(tt.text)

			styledNode := &style.StyledNode{
				Element: elem,
				Style: style.ComputedStyle{
					MaxLines: tt.maxLines,
					Color:    "#ffffff",
				},
			}

			layoutBox := &layout.LayoutBox{
				Node:   styledNode,
				X:      0,
				Y:      0,
				Width:  50,
				Height: 10,
			}

			layoutRenderer := NewLayoutRenderer()
			buffer := layoutRenderer.RenderToBuffer(layoutBox, 50, 10)
			output := RenderBufferFull(buffer)

			// Strip ANSI codes for easier testing
			cleaned := stripANSI(output)

			if !strings.Contains(cleaned, tt.expected) {
				t.Errorf("Expected output to contain %q, got:\n%s", tt.expected, cleaned)
			}
		})
	}
}

func stripANSI(s string) string {
	// Simple ANSI code stripper for testing
	result := strings.Builder{}
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
