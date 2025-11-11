package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestFlexboxZeroSize(t *testing.T) {
	// Container with zero size should not crash
	container := vdom.VStack(
		vdom.Text("test"),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 0, 0)
	if layout == nil {
		t.Fatal("Layout returned nil")
	}

	// Should have zero dimensions
	if layout.Width != 0 || layout.Height != 0 {
		t.Errorf("Expected zero dimensions, got %dx%d", layout.Width, layout.Height)
	}
}

func TestFlexboxSingleChild(t *testing.T) {
	container := vdom.VStack(
		vdom.Box(vdom.Text("single")).WithFlexGrow(1),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 24)
	if layout == nil {
		t.Fatal("Layout returned nil")
	}

	// Single child with flex-grow should fill container
	if len(layout.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(layout.Children))
	}

	child := layout.Children[0]
	if child.Height != 24 {
		t.Errorf("Child height = %d, want 24", child.Height)
	}
}

func TestFlexboxMultipleFlexGrow(t *testing.T) {
	// Three children with flex-grow: 1, 2, 1
	container := vdom.VStack(
		vdom.Box(vdom.Text("1")).WithFlexGrow(1),
		vdom.Box(vdom.Text("2")).WithFlexGrow(2),
		vdom.Box(vdom.Text("3")).WithFlexGrow(1),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 40)
	if len(layout.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(layout.Children))
	}

	// Total flex-grow = 4, height = 40
	// Child 0: 40 * (1/4) = 10
	// Child 1: 40 * (2/4) = 20
	// Child 2: 40 * (1/4) = 10
	if layout.Children[0].Height != 10 {
		t.Errorf("Child 0 height = %d, want 10", layout.Children[0].Height)
	}
	if layout.Children[1].Height != 20 {
		t.Errorf("Child 1 height = %d, want 20", layout.Children[1].Height)
	}
	if layout.Children[2].Height != 10 {
		t.Errorf("Child 2 height = %d, want 10", layout.Children[2].Height)
	}
}

func TestFlexboxNestedContainers(t *testing.T) {
	// Nested VStack inside HStack
	container := vdom.HStack(
		vdom.VStack(
			vdom.Text("1"),
			vdom.Text("2"),
		).WithFlexGrow(1),
		vdom.VStack(
			vdom.Text("3"),
			vdom.Text("4"),
		).WithFlexGrow(1),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 24)
	if len(layout.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(layout.Children))
	}

	// Each VStack should get half the width (40)
	if layout.Children[0].Width != 40 {
		t.Errorf("Child 0 width = %d, want 40", layout.Children[0].Width)
	}
	if layout.Children[1].Width != 40 {
		t.Errorf("Child 1 width = %d, want 40", layout.Children[1].Width)
	}
}

func TestFlexboxWithBorders(t *testing.T) {
	container := vdom.VStack(
		vdom.Box(vdom.Text("test")).
			WithBorderStyle(vdom.BorderStyleRounded).
			WithFlexGrow(1),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 24)
	if len(layout.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(layout.Children))
	}

	child := layout.Children[0]
	// Border takes 2 lines (top + bottom)
	expectedHeight := 24
	if child.Height != expectedHeight {
		t.Errorf("Child height = %d, want %d", child.Height, expectedHeight)
	}
}

func TestFlexboxEmptyContainer(t *testing.T) {
	container := vdom.VStack()

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 24)
	if layout == nil {
		t.Fatal("Layout returned nil")
	}

	if len(layout.Children) != 0 {
		t.Errorf("Expected 0 children, got %d", len(layout.Children))
	}
}

func TestFlexboxMixedFlexAndFixed(t *testing.T) {
	// Mix of flex-grow and fixed height
	container := vdom.VStack(
		vdom.Box(vdom.Text("fixed")).WithStyle("height", "5"),
		vdom.Box(vdom.Text("flex")).WithFlexGrow(1),
		vdom.Box(vdom.Text("fixed")).WithStyle("height", "3"),
	)

	resolver := style.NewResolver("")
	styledTree := resolver.Resolve(container)

	layout := Compute(styledTree, 80, 24)
	if len(layout.Children) != 3 {
		t.Fatalf("Expected 3 children, got %d", len(layout.Children))
	}

	// Fixed: 5 + 3 = 8, Remaining: 24 - 8 = 16
	if layout.Children[0].Height != 5 {
		t.Errorf("Child 0 height = %d, want 5", layout.Children[0].Height)
	}
	if layout.Children[1].Height != 16 {
		t.Errorf("Child 1 height = %d, want 16", layout.Children[1].Height)
	}
	if layout.Children[2].Height != 3 {
		t.Errorf("Child 2 height = %d, want 3", layout.Children[2].Height)
	}
}
