package lotustest

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"unsafe"

	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/reconciler"
	"github.com/speier/smith/pkg/lotus/tty"
)

// App is the interface for applications that can be tested
// This allows both testing package users and internal code to work with the same interface
type App interface {
	Render() *lotus.Element
}

// SnapshotOptions configures snapshot testing behavior
type SnapshotOptions struct {
	// Width of terminal for rendering
	Width int
	// Height of terminal for rendering
	Height int
}

// DefaultSnapshotOptions returns default snapshot configuration
func DefaultSnapshotOptions() *SnapshotOptions {
	return &SnapshotOptions{
		Width:  80,
		Height: 24,
	}
}

// SnapshotRender validates that terminal rendering works without errors
//
// This ensures the rendering pipeline executes successfully and produces output.
// No files are written - this is purely a validation test.
//
// Usage:
//
//	func TestMyApp(t *testing.T) {
//	    app := NewMyApp()
//	    element := app.Render()
//	    lotustest.SnapshotRender(t, "my-app", element, nil)
//	}
func SnapshotRender(t *testing.T, name string, element *lotus.Element, opts *SnapshotOptions) {
	t.Helper()

	if opts == nil {
		opts = DefaultSnapshotOptions()
	} else {
		// Merge with defaults
		defaults := DefaultSnapshotOptions()
		if opts.Width == 0 {
			opts.Width = defaults.Width
		}
		if opts.Height == 0 {
			opts.Height = defaults.Height
		}
	}

	// Render the element - this validates the pipeline works
	markup := element.ToMarkup()
	css := element.ToCSS()
	rendered := reconciler.Render(name, markup, css, opts.Width, opts.Height)

	// Just validate we got output
	if len(rendered) == 0 {
		t.Fatalf("Rendering produced empty output for %s", name)
	}

	t.Logf("Rendered %s: %d bytes", name, len(rendered))
}

// SnapshotMarkup validates that markup and CSS generation works without errors
//
// This ensures the markup/CSS generation pipeline executes successfully.
// No files are written - this is purely a validation test.
//
// Usage:
//
//	func TestMarkupGeneration(t *testing.T) {
//	    elem := lotus.VStack(lotus.Text("Hello"))
//	    lotustest.SnapshotMarkup(t, "hello", elem.Render(), nil)
//	}
func SnapshotMarkup(t *testing.T, name string, element *lotus.Element, opts *SnapshotOptions) {
	t.Helper()

	// Generate markup and CSS - this validates the pipeline works
	markup := element.ToMarkup()
	css := element.ToCSS()

	// Validate we got output
	if len(markup) == 0 {
		t.Fatalf("Markup generation produced empty output for %s", name)
	}

	t.Logf("Generated markup for %s: %d bytes markup, %d bytes CSS", name, len(markup), len(css))
}

// Structural Assertion Helpers - help debug rendering issues

// FindByID finds an element by ID in the tree
func FindByID(element *lotus.Element, id string) *lotus.Element {
	if element.ID == id {
		return element
	}
	for _, child := range element.Children {
		if found := FindByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// AssertHasID asserts that an element with the given ID exists
func AssertHasID(t *testing.T, element *lotus.Element, id string) *lotus.Element {
	t.Helper()
	found := FindByID(element, id)
	if found == nil {
		t.Errorf("Element with ID %q not found", id)
		return nil
	}
	return found
}

// AssertHasBorder asserts that an element has a border style
func AssertHasBorder(t *testing.T, element *lotus.Element) {
	t.Helper()
	if element == nil {
		t.Error("Element is nil")
		return
	}

	// Check if border is set in styles
	hasBorder := false
	for key := range element.Styles {
		if key == "border" || key == "border-style" {
			hasBorder = true
			break
		}
	}

	if !hasBorder {
		t.Errorf("Element %q does not have a border", element.ID)
	}
}

// AssertContainsText asserts that rendered output contains the given text
func AssertContainsText(t *testing.T, element *lotus.Element, text string, width, height int) {
	t.Helper()

	markup := element.ToMarkup()
	css := element.ToCSS()
	rendered := reconciler.Render("assert-test", markup, css, width, height)

	// Strip ANSI codes for text comparison
	clean := stripANSI(rendered)

	if !strings.Contains(clean, text) {
		t.Errorf("Rendered output does not contain %q\n\nRendered:\n%s", text, clean)
	}
}

// AssertLayout verifies the structure of an element tree
func AssertLayout(t *testing.T, element *lotus.Element, expected LayoutExpectation) {
	t.Helper()

	if expected.ID != "" && element.ID != expected.ID {
		t.Errorf("Expected ID %q, got %q", expected.ID, element.ID)
	}

	if expected.Type != "" && element.Type != expected.Type {
		t.Errorf("Expected type %q, got %q", expected.Type, element.Type)
	}

	if expected.ChildCount > 0 && len(element.Children) != expected.ChildCount {
		t.Errorf("Expected %d children, got %d", expected.ChildCount, len(element.Children))
	}

	// Recursively check children
	if len(expected.Children) > 0 {
		if len(element.Children) < len(expected.Children) {
			t.Errorf("Expected at least %d children, got %d", len(expected.Children), len(element.Children))
			return
		}

		for i, childExpectation := range expected.Children {
			AssertLayout(t, element.Children[i], childExpectation)
		}
	}
}

// LayoutExpectation defines expected element structure
type LayoutExpectation struct {
	ID         string
	Type       string
	ChildCount int
	Children   []LayoutExpectation
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			i++ // skip '['
			continue
		}
		if inEscape {
			if (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z') {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	return result.String()
}

// ExtractCursorPosition parses ANSI cursor positioning from rendered output
// Returns the row and column (1-indexed) where the cursor is positioned
// The cursor position is typically at the end of the output: "\033[{row};{col}H"
func ExtractCursorPosition(rendered string) (row, col int, found bool) {
	// Look for the last cursor positioning escape code
	// Format: ESC[{row};{col}H
	re := regexp.MustCompile(`\x1b\[(\d+);(\d+)H`)
	matches := re.FindAllStringSubmatch(rendered, -1)

	if len(matches) > 0 {
		// Get the last match (final cursor position)
		lastMatch := matches[len(matches)-1]
		if len(lastMatch) == 3 {
			row, _ = strconv.Atoi(lastMatch[1])
			col, _ = strconv.Atoi(lastMatch[2])
			return row, col, true
		}
	}

	return 0, 0, false
}

// AssertCursorAt validates that the cursor is at the expected position in rendered output
func AssertCursorAt(t *testing.T, rendered string, expectedRow, expectedCol int) {
	t.Helper()

	row, col, found := ExtractCursorPosition(rendered)
	if !found {
		t.Error("No cursor position found in rendered output")
		return
	}

	if row != expectedRow || col != expectedCol {
		t.Errorf("Cursor at (%d, %d), want (%d, %d)", row, col, expectedRow, expectedCol)
	}
}

// AssertNotContains asserts that rendered output does NOT contain the given text
func AssertNotContains(t *testing.T, rendered string, text string) {
	t.Helper()

	// Strip ANSI codes for text comparison
	clean := stripANSI(rendered)

	if strings.Contains(clean, text) {
		t.Errorf("Rendered output should not contain %q, but found it in:\n%s", text, clean)
	}
}

// SimulateKeyPress handles a key event on a context and returns the re-rendered output
// This simulates the full pipeline: key event → component update → re-render
func SimulateKeyPress(t *testing.T, contextID string, app App, event tty.KeyEvent, width, height int) string {
	t.Helper()

	// Get UI context
	ui := reconciler.GetContext(contextID)
	if ui == nil {
		t.Fatalf("Context %q not found - ensure SnapshotRender or RenderWithElement was called first", contextID)
	}

	// Handle key event (updates component state)
	ui.HandleKey(event)

	// Re-render with updated state
	lotusElem := app.Render()
	// lotus.Element is a type alias for core.Element, so this is safe
	element := (*core.Element)(unsafe.Pointer(lotusElem))
	rendered := reconciler.RenderWithElement(contextID, element, width, height)

	return rendered
}

// TestStep represents a single step in an interactive test sequence
type TestStep struct {
	KeyEvent       tty.KeyEvent // Key event to simulate
	ExpectedText   string       // Text that should be visible after this step
	NotContains    string       // Text that should NOT be visible (optional)
	ExpectedCursor CursorPos    // Expected cursor position
}

// CursorPos represents a cursor position (1-indexed, like ANSI codes)
type CursorPos struct {
	Row int
	Col int
}

// RunSequence executes a sequence of key events and validates output after each step
// This is useful for testing complex interactions like typing, editing, navigation
func RunSequence(t *testing.T, name string, app App, width, height int, steps []TestStep) {
	t.Helper()

	contextID := "test-" + name

	// Initial render to set up context
	lotusElem := app.Render()
	element := (*core.Element)(unsafe.Pointer(lotusElem))

	// Extract components before rendering
	components := element.ExtractComponents()

	// DO THE INITIAL RENDER FIRST - this creates the UI context
	_ = reconciler.RenderWithElement(contextID, element, width, height)

	// THEN register components (context exists now)
	for id, comp := range components {
		if focusable, ok := comp.(reconciler.Focusable); ok {
			reconciler.RegisterComponent(contextID, id, focusable)
		}
	}

	// THEN set focus to first component
	if len(components) > 0 {
		for id := range components {
			reconciler.SetFocus(contextID, id)
			break
		}
	}

	// Execute each step
	for i, step := range steps {
		t.Logf("Step %d: KeyEvent=%+v", i+1, step.KeyEvent)

		// Simulate key press and get new output
		rendered := SimulateKeyPress(t, contextID, app, step.KeyEvent, width, height)

		// Validate expected text is present
		if step.ExpectedText != "" {
			clean := stripANSI(rendered)
			if !strings.Contains(clean, step.ExpectedText) {
				t.Errorf("Step %d: expected text %q not found in output:\n%s", i+1, step.ExpectedText, clean)
			}
		}

		// Validate text that should NOT be present
		if step.NotContains != "" {
			AssertNotContains(t, rendered, step.NotContains)
		}

		// Validate cursor position
		if step.ExpectedCursor.Row > 0 || step.ExpectedCursor.Col > 0 {
			AssertCursorAt(t, rendered, step.ExpectedCursor.Row, step.ExpectedCursor.Col)
		}
	}
}
