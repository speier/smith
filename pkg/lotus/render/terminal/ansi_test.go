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
