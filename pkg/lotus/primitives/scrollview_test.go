package primitives

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestScrollViewVertical(t *testing.T) {
	// Create content with 10 lines
	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = strings.Repeat("x", 20)
	}
	content := strings.Join(lines, "\n")

	sv := NewScrollView().WithSize(20, 5) // 5 lines visible

	// Initially at top
	viewport := sv.GetViewport(content)
	visibleLines := strings.Split(viewport, "\n")
	if len(visibleLines) != 5 {
		t.Errorf("Expected 5 visible lines, got %d", len(visibleLines))
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
	// Create wide content
	content := strings.Repeat("x", 100)

	sv := NewScrollView().WithSize(20, 1) // 20 columns visible

	// Initially at left
	viewport := sv.GetViewport(content)
	if len(viewport) != 20 {
		t.Errorf("Expected 20 visible chars, got %d", len(viewport))
	}

	// Scroll right
	sv.ScrollRight(10)
	if sv.ScrollX != 10 {
		t.Errorf("ScrollX should be 10, got %d", sv.ScrollX)
	}

	viewport = sv.GetViewport(content)
	if !strings.HasPrefix(content[10:], viewport) {
		t.Error("Viewport should start at offset 10")
	}
}

func TestScrollViewAutoScroll(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5).WithAutoScroll(true)

	// Create content with 10 lines
	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")

	// Auto-scroll should jump to bottom
	_ = sv.GetViewport(content)
	if sv.ScrollY != 5 { // 10 - 5 = 5
		t.Errorf("Auto-scroll should set ScrollY to 5, got %d", sv.ScrollY)
	}
}

func TestScrollViewPageNavigation(t *testing.T) {
	sv := NewScrollView().WithSize(20, 5)

	// Create content with 20 lines
	lines := make([]string, 20)
	for i := 0; i < 20; i++ {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")
	_ = sv.GetViewport(content) // Initialize content size

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

	// Create content with 10 lines
	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")
	_ = sv.GetViewport(content) // Initialize content size

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

	// Need to initialize content size first
	lines := make([]string, 20)
	for i := 0; i < 20; i++ {
		lines[i] = "line"
	}
	content := strings.Join(lines, "\n")
	_ = sv.GetViewport(content) // Initialize content dimensions

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
		t.Errorf("Expected box element, got %q", elem.Tag)
	}
}
