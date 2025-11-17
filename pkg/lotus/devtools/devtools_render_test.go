package devtools

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/render"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestDevToolsRenderWithBorder tests that the DevTools panel renders correctly with border
func TestDevToolsRenderWithBorder(t *testing.T) {
	dt := New()
	dt.Log("Test message 1")
	dt.Log("Test message 2")

	// Render the panel
	panel := dt.Render()
	if panel == nil {
		t.Fatal("Expected panel to render, got nil")
	}

	// Wrap it in a Box with border (simulating wrapWithDevTools)
	wrapped := vdom.Box(panel).
		WithStyle("background-color", "#1a1a1a").
		WithBorderStyle(vdom.BorderStyleRounded).
		WithFlexGrow(1)

	// Resolve styles
	resolver := style.NewResolver("")
	styled := resolver.Resolve(wrapped)

	// Layout at fixed size
	width, height := 60, 10
	layoutTree := layout.Compute(styled, width, height)

	// Render to buffer
	renderer := render.NewLayoutRenderer()
	buf := renderer.RenderToBuffer(layoutTree, width, height)

	// Convert to ANSI string
	output := render.RenderBufferFull(buf)

	t.Log("DevTools panel output (raw ANSI):")
	t.Logf("%q", output)

	// Verify border characters are present in the output
	if !strings.Contains(output, "╭") || !strings.Contains(output, "╮") {
		t.Errorf("Expected top border with ╭ and ╮")
	}

	if !strings.Contains(output, "╰") || !strings.Contains(output, "╯") {
		t.Errorf("Expected bottom border with ╰ and ╯")
	}

	// Check for vertical borders
	if !strings.Contains(output, "│") {
		t.Error("Expected vertical borders (│)")
	}

	// Count each border corner - should appear exactly once
	if count := strings.Count(output, "╭"); count != 1 {
		t.Errorf("Expected exactly 1 '╭', got %d", count)
	}
	if count := strings.Count(output, "╮"); count != 1 {
		t.Errorf("Expected exactly 1 '╮', got %d", count)
	}
	if count := strings.Count(output, "╰"); count != 1 {
		t.Errorf("Expected exactly 1 '╰', got %d", count)
	}
	if count := strings.Count(output, "╯"); count != 1 {
		t.Errorf("Expected exactly 1 '╯', got %d", count)
	}
}

// TestDevToolsToggle tests that toggling DevTools doesn't cause border artifacts
func TestDevToolsToggle(t *testing.T) {
	dt := New()
	dt.Log("Initial message")

	// Render 1: Enabled
	panel1 := dt.Render()
	if panel1 == nil {
		t.Fatal("Expected panel to render when enabled")
	}

	wrapped1 := vdom.Box(panel1).
		WithStyle("background-color", "#1a1a1a").
		WithBorderStyle(vdom.BorderStyleRounded)

	resolver := style.NewResolver("")
	styled1 := resolver.Resolve(wrapped1)
	layoutTree1 := layout.Compute(styled1, 60, 10)
	renderer := render.NewLayoutRenderer()
	buf1 := renderer.RenderToBuffer(layoutTree1, 60, 10)
	output1 := render.RenderBufferFull(buf1)

	// Disable
	dt.Disable()
	panel2 := dt.Render()
	if panel2 != nil {
		t.Error("Expected nil panel when disabled")
	}

	// Re-enable
	dt.Enable()
	dt.Log("After re-enable")

	// Render 2: Re-enabled
	panel3 := dt.Render()
	if panel3 == nil {
		t.Fatal("Expected panel to render when re-enabled")
	}

	wrapped3 := vdom.Box(panel3).
		WithStyle("background-color", "#1a1a1a").
		WithBorderStyle(vdom.BorderStyleRounded)

	styled3 := resolver.Resolve(wrapped3)
	layoutTree3 := layout.Compute(styled3, 60, 10)
	buf3 := renderer.RenderToBuffer(layoutTree3, 60, 10)
	output3 := render.RenderBufferFull(buf3)

	t.Log("After re-enable (raw ANSI):")
	t.Logf("%q", output3)

	// Check that borders are clean (no duplicates)
	// Count border characters
	topLeftCount := strings.Count(output3, "╭")
	topRightCount := strings.Count(output3, "╮")
	bottomLeftCount := strings.Count(output3, "╰")
	bottomRightCount := strings.Count(output3, "╯")

	if topLeftCount != 1 {
		t.Errorf("Expected exactly 1 '╭' in border, got %d", topLeftCount)
	}
	if topRightCount != 1 {
		t.Errorf("Expected exactly 1 '╮' in border, got %d", topRightCount)
	}
	if bottomLeftCount != 1 {
		t.Errorf("Expected exactly 1 '╰' in border, got %d", bottomLeftCount)
	}
	if bottomRightCount != 1 {
		t.Errorf("Expected exactly 1 '╯' in border, got %d", bottomRightCount)
	}

	// Suppress unused warning
	_ = output1
}
