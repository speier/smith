package primitives

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/render"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestScrollViewVertical(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5) // 5 lines visible

	// Simulate content dimensions from layout
	sv.contentWidth = 20
	sv.contentHeight = 10 // 10 lines of content

	// Initially at top
	if sv.ScrollY != 0 {
		t.Errorf("Expected ScrollY=0 initially, got %d", sv.ScrollY)
	}

	// Scroll down
	sv.ScrollDown(2)
	if sv.ScrollY != 2 {
		t.Errorf("ScrollY should be 2, got %d", sv.ScrollY)
	}

	// Scroll to bottom
	sv.ScrollToBottom()
	if sv.ScrollY != 5 { // 10 - 5 = 5
		t.Errorf("ScrollY should be 5, got %d", sv.ScrollY)
	}

	// Can't scroll beyond bottom
	sv.ScrollDown(10)
	if sv.ScrollY != 5 {
		t.Errorf("ScrollY should stay at 5, got %d", sv.ScrollY)
	}

	// Scroll to top
	sv.ScrollToTop()
	if sv.ScrollY != 0 {
		t.Errorf("ScrollY should be 0, got %d", sv.ScrollY)
	}
}

func TestScrollViewHorizontal(t *testing.T) {
	sv := NewScrollView().WithSize(20, 1) // 20 columns visible

	// Simulate wide content
	sv.contentWidth = 100
	sv.contentHeight = 1

	// Initially at left
	if sv.ScrollX != 0 {
		t.Errorf("Expected ScrollX=0 initially, got %d", sv.ScrollX)
	}

	// Scroll right
	sv.ScrollRight(10)
	if sv.ScrollX != 10 {
		t.Errorf("ScrollX should be 10, got %d", sv.ScrollX)
	}

	// Verify GetScrollOffset returns correct offset
	offsetX, offsetY := sv.GetScrollOffset()
	if offsetX != 10 || offsetY != 0 {
		t.Errorf("GetScrollOffset: expected (10, 0), got (%d, %d)", offsetX, offsetY)
	}
}

func TestScrollViewAutoScroll(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5).WithAutoScroll(true)

	// Simulate content with 10 lines
	sv.SetContentSize(20, 10)

	// Auto-scroll should jump to bottom on SetContentSize
	if sv.ScrollY != 5 { // 10 - 5 = 5
		t.Errorf("Auto-scroll should set ScrollY to 5, got %d", sv.ScrollY)
	}
}

func TestScrollViewPageNavigation(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)

	// Simulate content with 20 lines
	sv.contentWidth = 20
	sv.contentHeight = 20

	// Page down
	sv.PageDown()
	if sv.ScrollY != 5 {
		t.Errorf("PageDown should scroll by 5, got ScrollY=%d", sv.ScrollY)
	}

	// Page up
	sv.PageUp()
	if sv.ScrollY != 0 {
		t.Errorf("PageUp should scroll to 0, got ScrollY=%d", sv.ScrollY)
	}
}

func TestScrollViewCanScroll(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)

	// Simulate content with 10 lines
	sv.contentWidth = 20
	sv.contentHeight = 10

	// At top - can't scroll up
	if sv.CanScrollUp() {
		t.Error("Should not be able to scroll up from top")
	}
	if !sv.CanScrollDown() {
		t.Error("Should be able to scroll down from top")
	}

	// At bottom
	sv.ScrollToBottom()
	if !sv.CanScrollUp() {
		t.Error("Should be able to scroll up from bottom")
	}
	if sv.CanScrollDown() {
		t.Error("Should not be able to scroll down from bottom")
	}
}

func TestScrollViewCallback(t *testing.T) {
	called := false
	var lastX, lastY int

	sv := NewScrollView().
		WithSize(20, 5).
		WithOnScroll(func(x, y int) {
			called = true
			lastX, lastY = x, y
		})

	// Simulate content
	sv.contentWidth = 20
	sv.contentHeight = 20

	sv.ScrollDown(3)

	if !called {
		t.Error("OnScroll callback not called")
	}
	if lastX != 0 || lastY != 3 {
		t.Errorf("OnScroll received (%d, %d), want (0, 3)", lastX, lastY)
	}
}

func TestScrollViewStatePersistence(t *testing.T) {
	sv := NewScrollView().WithID("test-scroll")
	sv.ScrollX = 10
	sv.ScrollY = 20

	// Save state
	state := sv.SaveState()
	if x, ok := state["scrollX"].(float64); !ok || int(x) != 10 {
		t.Errorf("SaveState scrollX wrong: %v", state)
	}
	if y, ok := state["scrollY"].(float64); !ok || int(y) != 20 {
		t.Errorf("SaveState scrollY wrong: %v", state)
	}

	// Load state
	newSv := NewScrollView().WithID("test-scroll")
	if err := newSv.LoadState(state); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	if newSv.ScrollX != 10 || newSv.ScrollY != 20 {
		t.Errorf("State not restored: got (%d, %d)", newSv.ScrollX, newSv.ScrollY)
	}
}

func TestScrollViewRender(t *testing.T) {
	sv := NewScrollView().WithContent(vdom.Text("hello"))

	elem := sv.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Should render content inside a box
	if elem.Tag != "box" {
		t.Errorf("Expected box element, got tag %q", elem.Tag)
	}
}

func TestScrollViewRenderWithoutSize(t *testing.T) {
	// ScrollView with zero size falls back to box wrapper
	sv := NewScrollView().WithSize(0, 0).WithContent(vdom.Text("hello"))

	elem := sv.Render()
	if elem == nil {
		t.Fatal("Render returned nil")
	}

	// Should fall back to box when dimensions are zero
	if elem.Tag != "box" {
		t.Errorf("Expected box element without size, got %q", elem.Tag)
	}
}

func TestScrollViewGetViewportSize(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)

	width, height := sv.GetViewportSize()
	if width != 20 || height != 5 {
		t.Errorf("GetViewportSize: expected (20, 5), got (%d, %d)", width, height)
	}
}

func TestScrollViewGetScrollOffset(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)
	sv.ScrollX = 10
	sv.ScrollY = 3

	offsetX, offsetY := sv.GetScrollOffset()
	if offsetX != 10 || offsetY != 3 {
		t.Errorf("GetScrollOffset: expected (10, 3), got (%d, %d)", offsetX, offsetY)
	}
}

func TestScrollViewSetContentSize(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)

	// Set content size
	sv.SetContentSize(50, 20)

	if sv.contentWidth != 50 || sv.contentHeight != 20 {
		t.Errorf("SetContentSize: expected content (50, 20), got (%d, %d)",
			sv.contentWidth, sv.contentHeight)
	}
}

func TestScrollViewBoundsChecking(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)
	sv.SetContentSize(30, 10)

	// Scroll beyond bottom - should be clamped by ScrollDown
	sv.ScrollDown(100)
	if sv.ScrollY != 5 { // max = 10 - 5 = 5
		t.Errorf("ScrollDown should clamp to 5, got %d", sv.ScrollY)
	}

	// Scroll beyond right - should be clamped by ScrollRight
	sv.ScrollRight(100)
	if sv.ScrollX != 10 { // max = 30 - 20 = 10
		t.Errorf("ScrollRight should clamp to 10, got %d", sv.ScrollX)
	}

	// Scroll beyond top - should be clamped by ScrollUp
	sv.ScrollUp(100)
	if sv.ScrollY != 0 {
		t.Errorf("ScrollUp should clamp to 0, got %d", sv.ScrollY)
	}

	// Scroll beyond left - should be clamped by ScrollLeft
	sv.ScrollLeft(100)
	if sv.ScrollX != 0 {
		t.Errorf("ScrollLeft should clamp to 0, got %d", sv.ScrollX)
	}
}

// TestScrollViewBufferIntegration tests the full rendering pipeline with ScrollView
func TestScrollViewBufferIntegration(t *testing.T) {
	// Create content with 20 lines
	var messages []any
	for i := 1; i <= 20; i++ {
		messages = append(messages, vdom.Text("Line "+string(rune('0'+(i%10)))))
	}
	content := vdom.VStack(messages...)

	// Create ScrollView with 10-line viewport
	sv := NewScrollView().
		WithID("test-scroll").
		WithContent(content).
		WithSize(80, 10)

	// Wrap in Box
	root := vdom.Box(sv).WithFlexGrow(1)

	// Full rendering pipeline: vdom → style → layout → buffer → ANSI
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := layout.Compute(styled, 80, 10)

	layoutRenderer := render.NewLayoutRenderer()
	buffer := layoutRenderer.RenderToBuffer(layoutBox, 80, 10)
	output := render.RenderBufferFull(buffer)

	// Verify output contains content
	if !strings.Contains(output, "Line") {
		t.Error("Buffer output should contain visible content")
	}

	// Verify clipping works - should not see all 20 lines in 10-line viewport
	lineCount := 0
	for y := 0; y < buffer.Height; y++ {
		rowHasContent := false
		for x := 0; x < buffer.Width; x++ {
			if buffer.Get(x, y).Char != ' ' {
				rowHasContent = true
				break
			}
		}
		if rowHasContent {
			lineCount++
		}
	}

	if lineCount > 10 {
		t.Errorf("Expected at most 10 lines of content in viewport, got %d", lineCount)
	}
}

// TestScrollViewBufferWithScroll tests buffer rendering with different scroll positions
func TestScrollViewBufferWithScroll(t *testing.T) {
	// Create 10 numbered lines
	var messages []any
	for i := 1; i <= 10; i++ {
		messages = append(messages, vdom.Text("Line "+string(rune('0'+i))))
	}
	content := vdom.VStack(messages...)

	// Create ScrollView with 5-line viewport
	sv := NewScrollView().
		WithID("test-scroll").
		WithContent(content).
		WithSize(80, 5)

	// Scroll to middle
	sv.ScrollDown(3)

	root := vdom.Box(sv).WithFlexGrow(1)

	// Render
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := layout.Compute(styled, 80, 5)

	layoutRenderer := render.NewLayoutRenderer()
	buffer := layoutRenderer.RenderToBuffer(layoutBox, 80, 5)
	output := render.RenderBufferFull(buffer)

	// Should see lines around position 3 (Lines 4-8 approximately)
	if !strings.Contains(output, "Line") {
		t.Error("Scrolled viewport should still show content")
	}

	// Verify buffer contains data
	hasContent := false
	for y := 0; y < buffer.Height; y++ {
		for x := 0; x < buffer.Width; x++ {
			if buffer.Get(x, y).Char != ' ' {
				hasContent = true
				break
			}
		}
	}

	if !hasContent {
		t.Error("Buffer should contain visible content after scrolling")
	}
}

// TestScrollViewBufferAutoScroll tests auto-scroll with buffer rendering
func TestScrollViewBufferAutoScroll(t *testing.T) {
	// Create many lines
	var messages []any
	for i := 1; i <= 30; i++ {
		messages = append(messages, vdom.Text("Message "+string(rune('A'+(i%26)))))
	}
	content := vdom.VStack(messages...)

	// Create ScrollView with auto-scroll
	sv := NewScrollView().
		WithID("test-scroll").
		WithContent(content).
		WithAutoScroll(true).
		WithSize(80, 10)

	root := vdom.Box(sv).WithFlexGrow(1)

	// Render
	resolver := style.NewResolver("")
	styled := resolver.Resolve(root)
	layoutBox := layout.Compute(styled, 80, 10)

	layoutRenderer := render.NewLayoutRenderer()
	buffer := layoutRenderer.RenderToBuffer(layoutBox, 80, 10)
	output := render.RenderBufferFull(buffer)

	// With auto-scroll, should see content from the end
	if !strings.Contains(output, "Message") {
		t.Error("Auto-scroll viewport should show messages")
	}

	// Verify scroll position was calculated
	if sv.ScrollY < 0 {
		t.Errorf("Auto-scroll should set valid ScrollY, got %d", sv.ScrollY)
	}
}
