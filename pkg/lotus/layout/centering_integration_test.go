package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestActualLogoCentering tests centering multi-line text with VStack align-items
func TestActualLogoCentering(t *testing.T) {
	// Lotus logo with ANSI codes (testing the Lotus framework itself)
	magentaBold := "\x1b[38;5;13;1m"
	reset := "\x1b[0m"

	logoLines := []string{
		"██╗      ██████╗ ████████╗██╗   ██╗███████╗",
		"██║     ██╔═══██╗╚══██╔══╝██║   ██║██╔════╝",
		"██║     ██║   ██║   ██║   ██║   ██║███████╗",
		"██║     ██║   ██║   ██║   ██║   ██║╚════██║",
		"███████╗╚██████╔╝   ██║   ╚██████╔╝███████║",
		"╚══════╝ ╚═════╝    ╚═╝    ╚═════╝ ╚══════╝",
	}

	// Build colored logo like a typical multi-line banner
	var logoText string
	for i, line := range logoLines {
		if i > 0 {
			logoText += "\n"
		}
		logoText += magentaBold + line + reset
	}

	// Create the ACTUAL structure from MessageList.Render():
	// MessageList VStack (with align-items:stretch default) -> Header VStack -> Logo Text

	messageListVStack := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.BoxElement,
		},
		Style: style.ComputedStyle{
			Display:    "flex",
			FlexDir:    "column",
			AlignItems: "stretch", // CSS default for flexbox - should stretch header to full width
		},
	}

	headerVStack := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.BoxElement,
		},
		Style: style.ComputedStyle{
			Display:    "flex",
			FlexDir:    "column",
			AlignItems: "center", // Should center logo horizontally
		},
	}

	logoText_node := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.TextElement,
			Text: logoText,
		},
		Style: style.ComputedStyle{},
	}

	// Build hierarchy
	headerVStack.Children = []*style.StyledNode{logoText_node}
	messageListVStack.Children = []*style.StyledNode{headerVStack}

	// Layout in 150-width container (like terminal)
	containerWidth := 150
	boxes := layoutFlexColumn(messageListVStack.Children, 0, 0, containerWidth, 100, messageListVStack.Style)

	if len(boxes) != 1 {
		t.Fatalf("Expected 1 box (header), got %d", len(boxes))
	}

	headerBox := boxes[0]

	// Header should be stretched to full width (150) because parent has align-items:stretch
	t.Logf("Container Width: %d", containerWidth)
	t.Logf("Header Width: %d", headerBox.Width)
	t.Logf("Header X Position: %d", headerBox.X)

	if headerBox.Width != containerWidth {
		t.Errorf("Header Width = %d, want %d (should be stretched by parent align-items:stretch)", headerBox.Width, containerWidth)
	}

	if headerBox.X != 0 {
		t.Errorf("Header X = %d, want 0 (should be at left edge)", headerBox.X)
	}

	// Now check the logo inside the header
	if len(headerBox.Children) != 1 {
		t.Fatalf("Expected 1 child (logo) in header, got %d", len(headerBox.Children))
	}

	logoBox := headerBox.Children[0]

	// Logo intrinsic width should be 43 (widest line of Lotus logo)
	const expectedLogoWidth = 43
	if logoBox.Width != expectedLogoWidth {
		t.Errorf("Logo Width = %d, want %d (intrinsic width of widest line)", logoBox.Width, expectedLogoWidth)
	}

	// Logo centering depends on header width
	// Header is 150 chars wide (stretched by parent), logo should be centered at X=55

	expectedLogoX := (containerWidth - expectedLogoWidth) / 2

	// Always log the actual values
	t.Logf("Logo Width: %d", logoBox.Width)
	t.Logf("Logo X Position: %d", logoBox.X)
	t.Logf("Expected X Position (centered in %d-width container): %d", containerWidth, expectedLogoX)

	if logoBox.X != expectedLogoX {
		t.Errorf("Logo X = %d, want %d (should be centered in %d-width container)", logoBox.X, expectedLogoX, containerWidth)
		t.Logf("Logo is offset by %d chars from center", expectedLogoX-logoBox.X)
	}
}

// TestLogoIntrinsicWidth verifies the logo width calculation
func TestLogoIntrinsicWidth(t *testing.T) {
	magentaBold := "\x1b[38;5;13;1m"
	reset := "\x1b[0m"

	logoLines := []string{
		"██╗      ██████╗ ████████╗██╗   ██╗███████╗",
		"██║     ██╔═══██╗╚══██╔══╝██║   ██║██╔════╝",
		"██║     ██║   ██║   ██║   ██║   ██║███████╗",
		"██║     ██║   ██║   ██║   ██║   ██║╚════██║",
		"███████╗╚██████╔╝   ██║   ╚██████╔╝███████║",
		"╚══════╝ ╚═════╝    ╚═╝    ╚═════╝ ╚══════╝",
	}

	var logoText string
	for i, line := range logoLines {
		if i > 0 {
			logoText += "\n"
		}
		logoText += magentaBold + line + reset
	}

	logoNode := &style.StyledNode{
		Element: &vdom.Element{
			Type: vdom.TextElement,
			Text: logoText,
		},
		Style: style.ComputedStyle{},
	}

	width := ComputeIntrinsicWidth(logoNode)

	// Each line is 43 visible characters (Lotus logo)
	expectedWidth := 43
	if width != expectedWidth {
		t.Errorf("Logo intrinsic width = %d, want %d", width, expectedWidth)
		t.Logf("Logo text (first line): %q", logoLines[0])
		t.Logf("First line visible length: %d", visibleLen(magentaBold+logoLines[0]+reset))
	}
}
