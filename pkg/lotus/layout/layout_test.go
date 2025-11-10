package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/core"
)

func TestBasicLayout(t *testing.T) {
	root := core.NewNode("box")
	child := core.NewNode("box")
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
	root := core.NewNode("box")
	root.Styles = &core.ComputedStyle{
		Display: "flex",
		FlexDir: "column",
	}

	header := core.NewNode("box")
	header.Styles = &core.ComputedStyle{Height: "5"}

	content := core.NewNode("box")
	content.Styles = &core.ComputedStyle{Flex: "1"}

	footer := core.NewNode("box")
	footer.Styles = &core.ComputedStyle{Height: "3"}

	root.Children = []*core.Node{header, content, footer}

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
