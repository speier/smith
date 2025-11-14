package snapshot

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/vdom"
)

func TestSnapshot_BasicRendering(t *testing.T) {
	element := vdom.Text("Hello")
	snapshot := Render(element, 10, 5)

	snapshot.AssertContains(t, "Hello")
	snapshot.AssertFirstLine(t, "Hello")
}

func TestSnapshot_BoxAlignment(t *testing.T) {
	// This test validates the fix for vertical spacing bug
	// Content should start at top of box, not several lines down
	element := vdom.Box(
		vdom.Text("First"),
		vdom.Text("Second"),
	).WithBorderStyle(vdom.BorderStyleRounded)

	snapshot := Render(element, 20, 10)

	// Border takes 1 line at top, content should start immediately after
	// Line 0: top border
	// Line 1: First
	// Line 2: Second
	snapshot.AssertNoEmptyLinesAtTop(t)

	// Check layout: Box should have children positioned at y=1 (after border)
	if len(snapshot.LayoutBox.Children) > 0 {
		firstChild := snapshot.LayoutBox.Children[0]
		if firstChild.Y != 1 {
			t.Errorf("First child should be at y=1 (after border), got y=%d", firstChild.Y)
		}
	}
}

func TestSnapshot_VStackWithGap(t *testing.T) {
	element := vdom.VStack(
		vdom.Text("Line 1"),
		vdom.Text("Line 2"),
		vdom.Text("Line 3"),
	).WithGap(1)

	snapshot := Render(element, 20, 10)

	// With gap=1, lines should be:
	// Line 0: "Line 1"
	// Line 1: "" (gap)
	// Line 2: "Line 2"
	// Line 3: "" (gap)
	// Line 4: "Line 3"
	snapshot.AssertLineAt(t, 0, "Line 1")
	snapshot.AssertLineAt(t, 2, "Line 2")
	snapshot.AssertLineAt(t, 4, "Line 3")
}

func TestSnapshot_BoxWithBorder(t *testing.T) {
	element := vdom.Box(
		vdom.Text("Content"),
	).WithBorderStyle(vdom.BorderStyleRounded)

	snapshot := Render(element, 15, 5)

	// Should have border characters
	snapshot.AssertContains(t, "╭")
	snapshot.AssertContains(t, "╰")
	snapshot.AssertContains(t, "Content")
}

func TestSnapshot_LayoutDimensions(t *testing.T) {
	element := vdom.Box(
		vdom.Text("Test"),
	)

	snapshot := Render(element, 20, 10)
	snapshot.AssertLayout(t, 0, 0, 20, 10)
}

func TestSnapshot_ChildLayout(t *testing.T) {
	element := vdom.VStack(
		vdom.Text("First"),
		vdom.Text("Second"),
	)

	snapshot := Render(element, 20, 10)

	// First child at top
	snapshot.AssertChildLayout(t, 0, 0, 0, 20, 1)
	// Second child right after first
	snapshot.AssertChildLayout(t, 1, 0, 1, 20, 1)
}

func TestSnapshot_DumpLayout(t *testing.T) {
	element := vdom.VStack(
		vdom.Text("A"),
		vdom.Text("B"),
	)

	snapshot := Render(element, 10, 5)
	dump := snapshot.DumpLayout()

	// Should contain layout info
	if dump == "" {
		t.Error("DumpLayout should return non-empty string")
	}
}

// TestSnapshot_VerticalAlignmentRegression validates the justify-content fix
func TestSnapshot_VerticalAlignmentRegression(t *testing.T) {
	// This reproduces the chat UI structure that had vertical spacing bug
	messages := []any{
		vdom.Text("Welcome"),
		vdom.Text("Type a message"),
	}

	element := vdom.Box(
		vdom.VStack(messages...).WithGap(1),
	).
		WithBorderStyle(vdom.BorderStyleRounded).
		WithStyle("height", "20")

	snapshot := Render(element, 40, 20)

	// Content should start right after top border (at line 1)
	// Line 0: ╭─── border ───╮
	// Line 1: Welcome (first message should be here, NOT several lines down)
	visible := snapshot.Visible
	lines := snapshot.Lines

	// Find which line "Welcome" appears on
	welcomeLine := -1
	for i, line := range lines {
		if containsText(line, "Welcome") {
			welcomeLine = i
			break
		}
	}

	if welcomeLine == -1 {
		t.Fatal("Could not find 'Welcome' text in output")
	}

	// "Welcome" should be at line 1 (right after border), not 4+ lines down
	if welcomeLine > 2 {
		t.Errorf("Content appears too far down: 'Welcome' at line %d (expected line 1)\nFull output:\n%s",
			welcomeLine, visible)
	}
}

// Helper to check if line contains text (ignoring whitespace and border chars)
func containsText(line, text string) bool {
	// Remove border characters and whitespace
	cleaned := line
	for _, r := range "╭╮╰╯─│" {
		cleaned = removeRune(cleaned, r)
	}
	return contains(cleaned, text)
}

func removeRune(s string, r rune) string {
	result := ""
	for _, c := range s {
		if c != r {
			result += string(c)
		}
	}
	return result
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && indexString(s, substr) >= 0
}

func indexString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
