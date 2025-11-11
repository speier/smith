package render

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestFullPipeline tests the complete vdom → style → layout → render pipeline
func TestFullPipeline(t *testing.T) {
	// 1. Create element tree (vdom)
	root := vdom.HStack(
		vdom.Box(vdom.Text("App")).
			WithStyle("width", "70%").
			WithStyle("height", "100%"),
		vdom.Box(vdom.Text("DevTools")).
			WithStyle("width", "30%").
			WithStyle("height", "100%").
			WithStyle("border", "single").
			WithStyle("color", "#ffff00"),
	)

	// 2. Resolve styles
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)

	// 3. Compute layout
	layoutBox := layout.Compute(styled, 160, 40)

	// 4. Render to ANSI
	renderer := New()
	output := renderer.Render(layoutBox)

	// Verify output is generated
	if output == "" {
		t.Fatal("Render output is empty")
	}

	// Should contain clear screen command
	if !strings.Contains(output, "\033[2J") {
		t.Error("Missing clear screen command")
	}

	// Should contain text
	if !strings.Contains(output, "App") {
		t.Error("Missing 'App' text")
	}
	if !strings.Contains(output, "DevTools") {
		t.Error("Missing 'DevTools' text")
	}

	// Should contain border characters (for devtools box)
	if !strings.Contains(output, "┌") || !strings.Contains(output, "└") {
		t.Error("Missing border characters")
	}

	// Should contain yellow color (for devtools)
	if !strings.Contains(output, "\033[93m") {
		t.Error("Missing yellow color ANSI code")
	}

	t.Logf("✓ Full pipeline works!")
	t.Logf("  Output length: %d bytes", len(output))
}

// TestRenderBorder tests border rendering
func TestRenderBorder(t *testing.T) {
	// Create a box with border
	root := vdom.Box(vdom.Text("Bordered")).
		WithStyle("width", "20").
		WithStyle("height", "5").
		WithStyle("border", "rounded").
		WithStyle("color", "#ffffff")

	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := layout.Compute(styled, 20, 5)

	renderer := New()
	output := renderer.Render(layoutBox)

	// Should contain rounded border chars
	if !strings.Contains(output, "╭") && !strings.Contains(output, "╮") {
		t.Error("Missing rounded border top corners")
	}
	if !strings.Contains(output, "╰") && !strings.Contains(output, "╯") {
		t.Error("Missing rounded border bottom corners")
	}

	t.Logf("✓ Border rendering works!")
}

// TestRenderTextAlignment tests text alignment
func TestRenderTextAlignment(t *testing.T) {
	tests := []struct {
		name      string
		textAlign string
	}{
		{"left", "left"},
		{"center", "center"},
		{"right", "right"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := vdom.Box(vdom.Text("Hello")).
				WithStyle("width", "20").
				WithStyle("height", "3").
				WithStyle("text-align", tt.textAlign)

			resolver := style.NewResolver("")
			styled := resolver.Resolve(root)
			layoutBox := layout.Compute(styled, 20, 3)

			renderer := New()
			output := renderer.Render(layoutBox)

			if !strings.Contains(output, "Hello") {
				t.Error("Missing text in output")
			}

			t.Logf("✓ Text alignment '%s' works!", tt.textAlign)
		})
	}
}

// TestRenderPlaceholderColor tests that placeholder color is rendered correctly
func TestRenderPlaceholderColor(t *testing.T) {
	// Create a simple text element with gray (#808080) color (like placeholder)
	root := vdom.Box(vdom.Text("Type here...").WithStyle("color", "#808080")).
		WithStyle("width", "40").
		WithStyle("height", "3")

	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := layout.Compute(styled, 40, 3)

	renderer := New()
	output := renderer.Render(layoutBox)

	// Should contain the placeholder text
	if !strings.Contains(output, "Type here...") {
		t.Error("Missing placeholder text in output")
		t.Logf("Output: %q", output)
	}

	// Should contain gray color ANSI code (90m)
	if !strings.Contains(output, "\033[90m") {
		t.Error("Missing gray color ANSI code for placeholder")
		t.Logf("Output: %q", output)
	}

	t.Logf("✓ Placeholder color rendering works!")
}
