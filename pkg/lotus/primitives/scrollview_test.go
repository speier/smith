package primitives

import (
	"testing"

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
