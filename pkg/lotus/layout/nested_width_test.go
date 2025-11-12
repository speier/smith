package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestNestedWidthConstraint(t *testing.T) {
	// Recreate the exact structure from the demo:
	// VStack (no width) containing Box (width=30)
	vstack := vdom.VStack(
		vdom.Text("Text Overflow Ellipsis:"),
		vdom.Box(
			vdom.Text("This is a very long text that will be truncated with ellipsis"),
		).WithWidth(30).WithBorderStyle(vdom.BorderStyleSingle),
	)

	// Resolve styles
	resolver := style.NewResolver("")
	styledNode := resolver.Resolve(vstack)

	// Layout with large container (like full screen width)
	layoutBox := Compute(styledNode, 200, 50)

	t.Logf("VStack width: %d", layoutBox.Width)
	t.Logf("VStack has %d children", len(layoutBox.Children))

	// Find the box child (should be second child)
	if len(layoutBox.Children) >= 2 {
		boxChild := layoutBox.Children[1]
		t.Logf("Box width: %d (should be 30)", boxChild.Width)
		t.Logf("Box style width: %s", boxChild.Node.Style.Width)

		if boxChild.Width != 30 {
			t.Errorf("Box width = %d, want 30", boxChild.Width)
		}

		// Check the text inside the box
		if len(boxChild.Children) > 0 {
			textChild := boxChild.Children[0]
			t.Logf("Text width: %d (should be 28 = 30-2 for border)", textChild.Width)

			expectedTextWidth := 28 // 30 - 2 for border
			if textChild.Width > expectedTextWidth {
				t.Errorf("Text width = %d, should be <= %d", textChild.Width, expectedTextWidth)
			}
		}
	} else {
		t.Error("VStack should have at least 2 children")
	}
}
