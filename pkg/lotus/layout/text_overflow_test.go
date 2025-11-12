package layout

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestWidthConstraintWithTextOverflow(t *testing.T) {
	// Create a box with explicit width=30 containing long text
	box := vdom.Box(
		vdom.Text("This is a very long text that will be truncated with ellipsis"),
	)
	box.WithStyle("width", "30")
	box.WithStyle("border", "single")
	box.WithStyle("text-overflow", "ellipsis")
	
	// The text child should get the overflow style
	if len(box.Children) > 0 {
		box.Children[0].WithStyle("text-overflow", "ellipsis")
	}

	// Resolve styles
	resolver := style.NewResolver("")
	styledNode := resolver.Resolve(box)

	// Layout with large container
	layoutBox := Compute(styledNode, 200, 50)

	// Box should be exactly 30 wide
	if layoutBox.Width != 30 {
		t.Errorf("Box width = %d, want 30", layoutBox.Width)
	}

	// Text child should inherit the constrained width
	if len(layoutBox.Children) > 0 {
		textBox := layoutBox.Children[0]
		t.Logf("Text box width: %d", textBox.Width)
		t.Logf("Text style width: %s", textBox.Node.Style.Width)
		t.Logf("Parent align-items: %s", styledNode.Style.AlignItems)
		
		// With border, available width is 30 - 2 = 28
		// Text should stretch to fill: 28 (since align-items defaults to stretch)
		expectedTextWidth := 28 // 30 - 2 for border
		if textBox.Width > expectedTextWidth {
			t.Errorf("Text box width = %d, should be <= %d (constrained by parent)", 
				textBox.Width, expectedTextWidth)
		}
	} else {
		t.Error("Box should have text child")
	}
}
