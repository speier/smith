package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestDevTools70_30Split tests the exact DevTools layout scenario
func TestDevTools70_30Split(t *testing.T) {
	// Create element tree (what user writes)
	root := vdom.HStack(
		vdom.Box(vdom.Text("App Content")).
			WithStyle("width", "70%").
			WithStyle("height", "100%"),
		vdom.Box(vdom.Text("DevTools Panel")).
			WithStyle("width", "30%").
			WithStyle("height", "100%").
			WithStyle("border", "single"),
	)
	// Resolve styles
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)

	// Compute layout (PURE MATH)
	layout := Compute(styled, 160, 40)

	// Verify root
	if layout.Width != 160 || layout.Height != 40 {
		t.Errorf("Root: expected 160x40, got %dx%d", layout.Width, layout.Height)
	}

	// Verify we have 2 children
	if len(layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layout.Children))
	}

	// Verify app container (70%)
	app := layout.Children[0]
	expectedAppWidth := 112 // 160 * 70 / 100
	if app.Width != expectedAppWidth {
		t.Errorf("App container: expected width=%d, got %d", expectedAppWidth, app.Width)
	}
	if app.Height != 40 {
		t.Errorf("App container: expected height=40, got %d", app.Height)
	}
	if app.X != 0 {
		t.Errorf("App container: expected X=0, got %d", app.X)
	}
	if app.Y != 0 {
		t.Errorf("App container: expected Y=0, got %d", app.Y)
	}

	// Verify devtools container (30%)
	dev := layout.Children[1]
	expectedDevWidth := 48 // 160 * 30 / 100
	if dev.Width != expectedDevWidth {
		t.Errorf("DevTools container: expected width=%d, got %d", expectedDevWidth, dev.Width)
	}
	if dev.Height != 40 {
		t.Errorf("DevTools container: expected height=40, got %d", dev.Height)
	}
	if dev.X != expectedAppWidth {
		t.Errorf("DevTools container: expected X=%d, got %d", expectedAppWidth, dev.X)
	}
	if dev.Y != 0 {
		t.Errorf("DevTools container: expected Y=0, got %d", dev.Y)
	}

	// Verify they touch (no gap, no overlap)
	appRightEdge := app.X + app.Width
	devLeftEdge := dev.X
	if appRightEdge != devLeftEdge {
		t.Errorf("Gap between containers: app ends at %d, dev starts at %d", appRightEdge, devLeftEdge)
	}

	// Verify total width
	totalWidth := app.Width + dev.Width
	if totalWidth != 160 {
		t.Errorf("Total width: expected 160, got %d", totalWidth)
	}

	t.Logf("✓ Layout correct!")
	t.Logf("  Root: %dx%d", layout.Width, layout.Height)
	t.Logf("  App:  X=%d Y=%d W=%d H=%d", app.X, app.Y, app.Width, app.Height)
	t.Logf("  Dev:  X=%d Y=%d W=%d H=%d", dev.X, dev.Y, dev.Width, dev.Height)
}

// TestFlexGrow tests flex-grow behavior
func TestFlexGrow(t *testing.T) {
	// Create: HStack with two children, one flex=1, one flex=2
	root := vdom.HStack(
		vdom.Box(vdom.Text("Small")).WithStyle("flex", "1"),
		vdom.Box(vdom.Text("Large")).WithStyle("flex", "2"),
	)

	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layout := Compute(styled, 300, 100)

	if len(layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layout.Children))
	}

	// flex=1 gets 1/3 of space, flex=2 gets 2/3
	small := layout.Children[0]
	large := layout.Children[1]

	expectedSmall := 100 // 300 * 1/3
	expectedLarge := 200 // 300 * 2/3

	if small.Width != expectedSmall {
		t.Errorf("Small (flex=1): expected width=%d, got %d", expectedSmall, small.Width)
	}
	if large.Width != expectedLarge {
		t.Errorf("Large (flex=2): expected width=%d, got %d", expectedLarge, large.Width)
	}

	t.Logf("✓ Flex grow correct!")
	t.Logf("  Small (flex=1): W=%d", small.Width)
	t.Logf("  Large (flex=2): W=%d", large.Width)
}

// TestMixedFixedAndFlex tests fixed width + flex children
func TestMixedFixedAndFlex(t *testing.T) {
	// Fixed 100px + flex=1 (takes remaining space)
	root := vdom.HStack(
		vdom.Box(vdom.Text("Fixed")).WithStyle("width", "100"),
		vdom.Box(vdom.Text("Flex")).WithStyle("flex", "1"),
	)

	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layout := Compute(styled, 300, 100)

	if len(layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layout.Children))
	}

	fixed := layout.Children[0]
	flex := layout.Children[1]

	if fixed.Width != 100 {
		t.Errorf("Fixed: expected width=100, got %d", fixed.Width)
	}
	if flex.Width != 200 { // 300 - 100
		t.Errorf("Flex: expected width=200, got %d", flex.Width)
	}

	t.Logf("✓ Mixed layout correct!")
	t.Logf("  Fixed: W=%d", fixed.Width)
	t.Logf("  Flex:  W=%d", flex.Width)
}
