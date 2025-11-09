package layout

import (
	"testing"
)

func TestBasicLayout(t *testing.T) {
	root := NewNode("box")
	child := NewNode("box")
	root.Children = append(root.Children, child)

	Layout(root, 100, 40)

	if root.Width != 100 {
		t.Errorf("expected root width 100, got %d", root.Width)
	}
	if root.Height != 40 {
		t.Errorf("expected root height 40, got %d", root.Height)
	}
}

func TestFlexColumn(t *testing.T) {
	root := NewNode("box")
	root.Styles = &ComputedStyle{
		Display: "flex",
		FlexDir: "column",
	}

	header := NewNode("box")
	header.Styles = &ComputedStyle{Height: "5"}

	content := NewNode("box")
	content.Styles = &ComputedStyle{Flex: "1"}

	footer := NewNode("box")
	footer.Styles = &ComputedStyle{Height: "3"}

	root.Children = []*Node{header, content, footer}

	Layout(root, 100, 40)

	if header.Height != 5 {
		t.Errorf("expected header height 5, got %d", header.Height)
	}
	if footer.Height != 3 {
		t.Errorf("expected footer height 3, got %d", footer.Height)
	}

	expectedHeight := 32
	if content.Height != expectedHeight {
		t.Errorf("expected content height %d, got %d", expectedHeight, content.Height)
	}
}
