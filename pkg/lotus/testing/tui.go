package lotustest

import (
	"strings"
	"testing"
	"unsafe"

	"github.com/speier/smith/pkg/lotus/core"
	"github.com/speier/smith/pkg/lotus/reconciler"
	"github.com/speier/smith/pkg/lotus/tty"
)

// MockTerminal provides a high-level API for testing TUI apps
// It wraps the low-level snapshot testing helpers with a cleaner interface
type MockTerminal struct {
	t          *testing.T
	app        App
	contextID  string
	width      int
	height     int
	lastOutput string
}

// NewMockTerminal creates a new mock terminal for testing TUI interactions
// It sets up the context, registers components, and performs the initial render
func NewMockTerminal(t *testing.T, name string, app App, width, height int) *MockTerminal {
	t.Helper()

	contextID := "mock-" + name

	// Initial render to set up context
	lotusElem := app.Render()
	element := (*core.Element)(unsafe.Pointer(lotusElem))

	// Extract components before rendering
	components := element.ExtractComponents()

	// DO THE INITIAL RENDER FIRST - this creates the UI context
	output := reconciler.RenderWithElement(contextID, element, width, height)

	// THEN register components (context exists now)
	for id, comp := range components {
		if focusable, ok := comp.(reconciler.Focusable); ok {
			reconciler.RegisterComponent(contextID, id, focusable)
		}
	}

	// THEN set focus to first FOCUSABLE component
	for id, comp := range components {
		if _, ok := comp.(reconciler.Focusable); ok {
			reconciler.SetFocus(contextID, id)
			break
		}
	}

	return &MockTerminal{
		t:          t,
		app:        app,
		contextID:  contextID,
		width:      width,
		height:     height,
		lastOutput: output,
	}
}

// SendKey simulates a key press and updates lastOutput with the new rendered state
// Returns the mock terminal for method chaining
func (m *MockTerminal) SendKey(key rune) *MockTerminal {
	m.t.Helper()

	event := tty.KeyEvent{
		Key:  byte(key),
		Char: string(key),
	}

	m.lastOutput = SimulateKeyPress(m.t, m.contextID, m.app, event, m.width, m.height)
	return m
}

// SendKeyEvent simulates a specific key event (for special keys like Enter, Backspace, etc.)
func (m *MockTerminal) SendKeyEvent(event tty.KeyEvent) *MockTerminal {
	m.t.Helper()

	m.lastOutput = SimulateKeyPress(m.t, m.contextID, m.app, event, m.width, m.height)
	return m
}

// AssertText validates that the given text is present in the current output
func (m *MockTerminal) AssertText(expected string) *MockTerminal {
	m.t.Helper()

	clean := stripANSI(m.lastOutput)
	if !strings.Contains(clean, expected) {
		m.t.Errorf("Expected text %q not found in output:\n%s", expected, clean)
	}

	return m
}

// AssertNotContains validates that the given text is NOT present in the current output
func (m *MockTerminal) AssertNotContains(text string) *MockTerminal {
	m.t.Helper()

	AssertNotContains(m.t, m.lastOutput, text)
	return m
}

// AssertCursor validates that the cursor is at the expected position
func (m *MockTerminal) AssertCursor(row, col int) *MockTerminal {
	m.t.Helper()

	AssertCursorAt(m.t, m.lastOutput, row, col)
	return m
}

// GetOutput returns the current rendered output (with ANSI codes)
func (m *MockTerminal) GetOutput() string {
	return m.lastOutput
}

// GetCleanOutput returns the current rendered output with ANSI codes stripped
func (m *MockTerminal) GetCleanOutput() string {
	return stripANSI(m.lastOutput)
}
