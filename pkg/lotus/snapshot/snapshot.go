package snapshot

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/render"
	"github.com/speier/smith/pkg/lotus/style"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// RenderSnapshot captures the visual output of a Lotus element for testing
type RenderSnapshot struct {
	Width     int
	Height    int
	Raw       string // Raw ANSI output
	Visible   string // Stripped of ANSI codes for visual comparison
	Lines     []string
	Element   *vdom.Element
	LayoutBox *layout.LayoutBox
}

// Render creates a snapshot of how an element renders at given dimensions
func Render(element *vdom.Element, width, height int) *RenderSnapshot {
	// Style the element
	resolver := style.NewResolver("") // No custom CSS for testing
	styledNode := resolver.Resolve(element)

	// Layout it
	layoutBox := layout.Compute(styledNode, width, height)

	// Render to buffer
	renderer := render.NewLayoutRenderer()
	buf := renderer.RenderToBuffer(layoutBox, width, height)

	// Convert buffer to string
	raw := render.RenderBufferFull(buf)
	visible := stripANSI(raw)

	// Split into lines based on buffer dimensions, not newlines
	// The buffer uses absolute cursor positioning, so we need to extract lines from the grid
	lines := extractLinesFromBuffer(buf, width, height)

	return &RenderSnapshot{
		Width:     width,
		Height:    height,
		Raw:       raw,
		Visible:   visible,
		Lines:     lines,
		Element:   element,
		LayoutBox: layoutBox,
	}
}

// extractLinesFromBuffer extracts visual lines from the render buffer
func extractLinesFromBuffer(buf *render.Buffer, width, height int) []string {
	lines := make([]string, height)
	for y := 0; y < height; y++ {
		line := ""
		for x := 0; x < width; x++ {
			cell := buf.Get(x, y)
			line += string(cell.Char)
		}
		// Trim trailing spaces for cleaner comparison
		lines[y] = strings.TrimRight(line, " ")
	}
	return lines
}

// AssertVisible checks that visible text matches expected output (ignoring ANSI)
func (s *RenderSnapshot) AssertVisible(t *testing.T, expected string) {
	t.Helper()
	if s.Visible != expected {
		t.Errorf("Visible output mismatch:\nExpected:\n%s\n\nGot:\n%s", expected, s.Visible)
	}
}

// AssertContains checks that visible output contains the given text
func (s *RenderSnapshot) AssertContains(t *testing.T, text string) {
	t.Helper()
	if !strings.Contains(s.Visible, text) {
		t.Errorf("Expected visible output to contain %q\nGot:\n%s", text, s.Visible)
	}
}

// AssertLineCount checks the number of rendered lines
func (s *RenderSnapshot) AssertLineCount(t *testing.T, expected int) {
	t.Helper()
	actual := len(s.Lines)
	if actual != expected {
		t.Errorf("Expected %d lines, got %d:\n%s", expected, actual, s.Visible)
	}
}

// AssertLineAt checks specific line content (0-indexed)
func (s *RenderSnapshot) AssertLineAt(t *testing.T, lineNum int, expected string) {
	t.Helper()
	if lineNum < 0 || lineNum >= len(s.Lines) {
		t.Fatalf("Line %d out of range (have %d lines)", lineNum, len(s.Lines))
	}
	actual := s.Lines[lineNum]
	if actual != expected {
		t.Errorf("Line %d mismatch:\nExpected: %q\nGot:      %q", lineNum, expected, actual)
	}
}

// AssertFirstLine checks that first line matches (common test case)
func (s *RenderSnapshot) AssertFirstLine(t *testing.T, expected string) {
	t.Helper()
	s.AssertLineAt(t, 0, expected)
}

// AssertLayout checks layout box dimensions
func (s *RenderSnapshot) AssertLayout(t *testing.T, x, y, width, height int) {
	t.Helper()
	if s.LayoutBox.X != x || s.LayoutBox.Y != y ||
		s.LayoutBox.Width != width || s.LayoutBox.Height != height {
		t.Errorf("Layout mismatch:\nExpected: {x:%d y:%d w:%d h:%d}\nGot:      {x:%d y:%d w:%d h:%d}",
			x, y, width, height,
			s.LayoutBox.X, s.LayoutBox.Y, s.LayoutBox.Width, s.LayoutBox.Height)
	}
}

// AssertChildLayout checks a specific child's layout
func (s *RenderSnapshot) AssertChildLayout(t *testing.T, childIndex, x, y, width, height int) {
	t.Helper()
	if childIndex < 0 || childIndex >= len(s.LayoutBox.Children) {
		t.Fatalf("Child index %d out of range (have %d children)", childIndex, len(s.LayoutBox.Children))
	}
	child := s.LayoutBox.Children[childIndex]
	if child.X != x || child.Y != y || child.Width != width || child.Height != height {
		t.Errorf("Child %d layout mismatch:\nExpected: {x:%d y:%d w:%d h:%d}\nGot:      {x:%d y:%d w:%d h:%d}",
			childIndex, x, y, width, height,
			child.X, child.Y, child.Width, child.Height)
	}
}

// AssertNoEmptyLinesAtTop checks that content starts at top (no vertical spacing bug)
func (s *RenderSnapshot) AssertNoEmptyLinesAtTop(t *testing.T) {
	t.Helper()
	if len(s.Lines) == 0 {
		return // Empty is OK
	}
	// Check first line isn't empty (or just whitespace)
	firstLine := strings.TrimSpace(s.Lines[0])
	if firstLine == "" {
		t.Errorf("First line is empty - content not aligned to top:\n%s", s.Visible)
	}
}

// DumpLayout prints layout tree for debugging
func (s *RenderSnapshot) DumpLayout() string {
	return dumpLayoutBox(s.LayoutBox, 0)
}

func dumpLayoutBox(box *layout.LayoutBox, indent int) string {
	prefix := strings.Repeat("  ", indent)
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s{x:%d y:%d w:%d h:%d", prefix, box.X, box.Y, box.Width, box.Height))
	if box.Node != nil && box.Node.Element != nil {
		if box.Node.Element.Type == 1 { // TextElement
			text := box.Node.Element.Text
			if len(text) > 20 {
				text = text[:20] + "..."
			}
			sb.WriteString(fmt.Sprintf(" text:%q", text))
		}
	}
	sb.WriteString("}\n")

	for _, child := range box.Children {
		sb.WriteString(dumpLayoutBox(child, indent+1))
	}

	return sb.String()
}

// stripANSI removes ANSI escape sequences from text
func stripANSI(s string) string {
	// Match ESC sequences: ESC [ ... m (SGR), ESC [ ... H (cursor), etc.
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(s, "")
}
