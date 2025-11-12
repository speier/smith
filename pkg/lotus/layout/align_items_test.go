package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestAlignItemsCenterInColumn(t *testing.T) {
	// Create a VStack with align-items: center
	// Child should be centered horizontally

	parent := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.BoxElement,
		},
		Style: style.ComputedStyle{
			Display:    "flex",
			FlexDir:    "column",
			AlignItems: "center",
		},
	}

	// Child with intrinsic width of 10 characters
	child := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.TextElement,
			Text: "1234567890", // 10 chars
		},
		Style: style.ComputedStyle{
			AlignSelf: "", // Should inherit parent's align-items
		},
	}

	parent.Children = []*style.StyledNode{child}

	// Layout in 100-width container
	boxes := layoutFlexColumn(parent.Children, 0, 0, 100, 50, parent.Style)

	if len(boxes) != 1 {
		t.Fatalf("Expected 1 box, got %d", len(boxes))
	}

	box := boxes[0]

	// Child should be centered: X = (100 - 10) / 2 = 45
	expectedX := 45
	if box.X != expectedX {
		t.Errorf("Child X = %d, want %d (centered in 100-width container with 10-char width)", box.X, expectedX)
	}

	// Width should be intrinsic (10)
	if box.Width != 10 {
		t.Errorf("Child Width = %d, want 10", box.Width)
	}
}

func TestAlignItemsStretchInColumn(t *testing.T) {
	// align-items: stretch should make child fill full width

	parent := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.BoxElement,
		},
		Style: style.ComputedStyle{
			Display:    "flex",
			FlexDir:    "column",
			AlignItems: "stretch", // Default
		},
	}

	child := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.TextElement,
			Text: "Hello",
		},
		Style: style.ComputedStyle{},
	}

	parent.Children = []*style.StyledNode{child}

	boxes := layoutFlexColumn(parent.Children, 0, 0, 100, 50, parent.Style)

	box := boxes[0]

	// Should stretch to full width
	if box.Width != 100 {
		t.Errorf("Child Width = %d, want 100 (stretched)", box.Width)
	}

	if box.X != 0 {
		t.Errorf("Child X = %d, want 0", box.X)
	}
}

func TestAlignItemsCenterMultiLineText(t *testing.T) {
	// Multi-line text should be centered based on its widest line

	parent := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.BoxElement,
		},
		Style: style.ComputedStyle{
			Display:    "flex",
			FlexDir:    "column",
			AlignItems: "center",
		},
	}

	// Logo-like multi-line text (3 lines, each 39 chars)
	logoText := "███████╗███╗   ███╗██╗████████╗██╗  ██╗\n" +
		"██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║\n" +
		"███████╗██╔████╔██║██║   ██║   ███████║"

	child := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.TextElement,
			Text: logoText,
		},
		Style: style.ComputedStyle{},
	}

	parent.Children = []*style.StyledNode{child}

	// Layout in 150-width container
	boxes := layoutFlexColumn(parent.Children, 0, 0, 150, 50, parent.Style)

	box := boxes[0]

	// Width should be 39 (longest line)
	if box.Width != 39 {
		t.Errorf("Child Width = %d, want 39 (widest line)", box.Width)
	}

	// Should be centered: X = (150 - 39) / 2 = 55
	expectedX := 55
	if box.X != expectedX {
		t.Errorf("Child X = %d, want %d (centered)", box.X, expectedX)
	}
}
