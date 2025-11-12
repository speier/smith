package render

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestBorderColorRendering(t *testing.T) {
	// Create a box with colored border
	box := vdom.Box(
		vdom.Text("Colored border test"),
	)
	box.WithStyle("border", "single")
	box.WithStyle("border-color", "#ff00ff") // magenta
	box.WithStyle("width", "20")
	box.WithStyle("height", "3")

	// Resolve styles
	resolver := style.NewResolver("")
	styledNode := resolver.Resolve(box)

	t.Logf("Border enabled: %v", styledNode.Style.Border)
	t.Logf("Border style: %s", styledNode.Style.BorderStyle)
	t.Logf("Border color: %s", styledNode.Style.BorderColor)
	t.Logf("Text color: %s", styledNode.Style.Color)

	// Layout
	layoutBox := layout.Compute(styledNode, 100, 10)

	// Render
	renderer := New()
	output := renderer.Render(layoutBox)

	t.Logf("Output length: %d", len(output))
	t.Logf("Output:\n%s", output)

	// Check for magenta ANSI code
	magentaCode := "\033[95m"
	if !strings.Contains(output, magentaCode) {
		t.Errorf("Expected magenta ANSI code %q not found in output", magentaCode)
		t.Logf("Looking for magenta in output...")
	} else {
		t.Logf("✓ Found magenta ANSI code!")
	}
}
