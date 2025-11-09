package terminal

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
)

func TestRenderBasic(t *testing.T) {
	root := &layout.Node{
		Type:   "box",
		Width:  20,
		Height: 5,
		X:      0,
		Y:      0,
		Styles: &layout.ComputedStyle{},
		Children: []*layout.Node{
			{
				Type:    "text",
				Content: "Hello World",
				Width:   20,
				Height:  1,
				X:       0,
				Y:       0,
				Styles:  &layout.ComputedStyle{},
			},
		},
	}

	output := Render(root)
	if output == "" {
		t.Error("Render produced empty output")
	}

	if !strings.Contains(output, "Hello World") {
		t.Error("Render output missing content")
	}
}

func TestRenderWithBorder(t *testing.T) {
	root := &layout.Node{
		Type:    "box",
		Content: "Test",
		Width:   15,
		Height:  5,
		X:       0,
		Y:       0,
		Styles: &layout.ComputedStyle{
			Border:     true,
			BorderChar: "single",
		},
	}

	output := Render(root)
	if output == "" {
		t.Error("Render produced empty output")
	}

	// Should contain box drawing characters for border
	if !strings.ContainsAny(output, "┌┐└┘─│") {
		t.Error("Render output missing border characters")
	}
}

func TestTextAlignCenter(t *testing.T) {
	parent := &layout.Node{
		Type:   "box",
		Width:  40,
		Height: 3,
		X:      0,
		Y:      0,
		Styles: &layout.ComputedStyle{
			TextAlign: "center",
		},
	}

	textNode := &layout.Node{
		Type:    "text",
		Content: "Hello",
		Width:   40,
		Height:  1,
		X:       0,
		Y:       0,
		Parent:  parent,
		Styles: &layout.ComputedStyle{
			TextAlign: "left", // Default value
		},
	}

	parent.Children = []*layout.Node{textNode}

	var buf strings.Builder
	renderText(textNode, &buf)
	output := buf.String()

	// Text should be moved to center (padding added)
	// "Hello" is 5 chars, available width is 40
	// Padding should be (40 - 5) / 2 = 17
	// So cursor should move to X=17 (18 in 1-indexed ANSI format)
	if !strings.Contains(output, "\033[1;18H") { // ESC[Y;XH format (1-indexed)
		t.Errorf("Text not centered correctly, output: %q", output)
	}
}

func TestTextAlignCenterWithANSI(t *testing.T) {
	parent := &layout.Node{
		Type:   "box",
		Width:  40,
		Height: 3,
		X:      0,
		Y:      0,
		Styles: &layout.ComputedStyle{
			TextAlign: "center",
		},
	}

	// Text with ANSI codes embedded (like from lotus.Style().Render())
	textNode := &layout.Node{
		Type:    "text",
		Content: "\033[92m\033[1mHello\033[0m", // Green bold "Hello"
		Width:   40,
		Height:  1,
		X:       0,
		Y:       0,
		Parent:  parent,
		Styles: &layout.ComputedStyle{
			TextAlign: "left",
		},
	}

	parent.Children = []*layout.Node{textNode}

	var buf strings.Builder
	renderText(textNode, &buf)
	output := buf.String()

	// ANSI codes should be stripped before calculating width
	// Plain text is "Hello" (5 chars), so padding should still be 17
	if !strings.Contains(output, "\033[1;18H") {
		t.Errorf("ANSI text not centered correctly, output: %q", output)
	}

	// Output should still contain the ANSI codes
	if !strings.Contains(output, "\033[92m") {
		t.Error("ANSI codes were removed from output")
	}
}

func TestTextAlignCenterWithPadding(t *testing.T) {
	parent := &layout.Node{
		Type:   "box",
		Width:  40,
		Height: 3,
		X:      0,
		Y:      0,
		Styles: &layout.ComputedStyle{
			TextAlign:    "center",
			PaddingLeft:  1,
			PaddingRight: 1,
		},
	}

	textNode := &layout.Node{
		Type:    "text",
		Content: "Hello",
		Width:   40,
		Height:  1,
		X:       0,
		Y:       0,
		Parent:  parent,
		Styles: &layout.ComputedStyle{
			TextAlign: "left",
		},
	}

	parent.Children = []*layout.Node{textNode}

	var buf strings.Builder
	renderText(textNode, &buf)
	output := buf.String()

	// Available width should account for parent padding: 40 - 2 = 38
	// Padding should be (38 - 5) / 2 = 16
	// Cursor: X=16 (17 in 1-indexed format)
	if !strings.Contains(output, "\033[1;17H") {
		t.Errorf("Text with parent padding not centered correctly, output: %q", output)
	}
}

func TestMultiClassSelectorWithCentering(t *testing.T) {
	// This simulates the real-world use case: .message.system { text-align: center; }
	parent := &layout.Node{
		Type:    "box",
		Classes: []string{"message", "system"},
		Width:   80,
		Height:  1,
		X:       2,
		Y:       5,
		Styles: &layout.ComputedStyle{
			TextAlign:    "center",
			PaddingLeft:  1,
			PaddingRight: 1,
		},
	}

	textNode := &layout.Node{
		Type:    "text",
		Content: "SMITH",
		Width:   80,
		Height:  1,
		X:       2,
		Y:       5,
		Parent:  parent,
		Styles: &layout.ComputedStyle{
			TextAlign: "left", // Default, should inherit from parent
		},
	}

	parent.Children = []*layout.Node{textNode}

	var buf strings.Builder
	renderText(textNode, &buf)
	output := buf.String()

	// Available width: 80 - 2 (padding) = 78
	// Text width: 5
	// Padding: (78 - 5) / 2 = 36
	// X position: 2 (parent X) + 36 = 38 (39 in 1-indexed format)
	// Y position: 5 (6 in 1-indexed format)
	if !strings.Contains(output, "\033[6;39H") { // Y=6, X=39 (1-indexed)
		t.Errorf("Multi-class centered text incorrect, output: %q", output)
	}
}
