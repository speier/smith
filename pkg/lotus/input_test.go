package lotus_test

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus/components"
	"github.com/speier/smith/pkg/lotus/layout"
	"github.com/speier/smith/pkg/lotus/parser"
	"github.com/speier/smith/pkg/lotus/reconciler"
	"github.com/speier/smith/pkg/lotus/tty"
)

// TestInputCursorPosition validates cursor positioning for input boxes
func TestInputCursorPosition(t *testing.T) {
	tests := []struct {
		name           string
		markup         string
		css            string
		expectedRow    int // 1-indexed
		expectedCol    int // 1-indexed
		expectedPrompt string
	}{
		{
			name: "input in box with border at top",
			markup: `<box border="1px solid" height="3">
				<input id="test">placeholder</input>
			</box>`,
			css:            `box { border: 1px solid; height: 3; }`,
			expectedRow:    2, // Inside border (row 1 is top border)
			expectedCol:    4, // Position 1 is left border, 2-3 is "> ", 4 is cursor
			expectedPrompt: "> ",
		},
		{
			name: "input in box with border and padding",
			markup: `<box border="1px solid" height="5" padding="1">
				<input id="test">placeholder</input>
			</box>`,
			css:            `box { border: 1px solid; height: 5; padding: 1; }`,
			expectedRow:    3, // Border(1) + Padding(1) + Content(1) = row 3
			expectedCol:    5, // Border(1) + Padding(1) + Prompt(2) = col 5
			expectedPrompt: "> ",
		},
		{
			name: "input at bottom of screen",
			markup: `<vstack height="24">
				<box flex="1"></box>
				<box id="input-box" border="1px solid" height="3">
					<input id="test">Type here...</input>
				</box>
			</vstack>`,
			css:            `#input-box { border: 1px solid; height: 3; }`,
			expectedRow:    23, // Row 22 is for messages (flex=1), 23 is input content
			expectedCol:    4,  // Inside border + prompt
			expectedPrompt: "> ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse and layout
			root := parser.Parse(tt.markup)
			if root == nil {
				t.Fatal("Parser returned nil")
			}

			styles := reconciler.GetStyles(tt.css)
			parser.ApplyStyles(root, styles)
			layout.Layout(root, 80, 24)

			// Find input node
			input := root.FindByID("test")
			if input == nil {
				t.Fatal("Input node not found")
			}

			// Calculate expected cursor position
			// Input is positioned at (input.X, input.Y) by layout
			// Cursor should be at X + len(prompt)
			actualRow := input.Y + 1                          // Convert to 1-indexed
			actualCol := input.X + len(tt.expectedPrompt) + 1 // Convert to 1-indexed

			if actualRow != tt.expectedRow {
				t.Errorf("Cursor row: got %d, want %d", actualRow, tt.expectedRow)
				t.Logf("Input node: X=%d, Y=%d, Width=%d, Height=%d", input.X, input.Y, input.Width, input.Height)
				if input.Parent != nil {
					t.Logf("Parent: X=%d, Y=%d, Border=%v, Padding=%d/%d",
						input.Parent.X, input.Parent.Y,
						input.Parent.Styles.Border,
						input.Parent.Styles.PaddingTop, input.Parent.Styles.PaddingLeft)
				}
			}

			if actualCol != tt.expectedCol {
				t.Errorf("Cursor col: got %d, want %d", actualCol, tt.expectedCol)
			}
		})
	}
}

// TestInputRendering validates that input renders with correct prompt and placeholder
func TestInputRendering(t *testing.T) {
	markup := `<box border="1px solid" height="3">
		<input id="test">Type here...</input>
	</box>`
	css := `box { border: 1px solid; height: 3; }`

	rendered := reconciler.Render("test", markup, css, 80, 3)

	// Should contain prompt and placeholder
	if !strings.Contains(rendered, "> ") {
		t.Error("Rendered output missing prompt '> '")
	}
	if !strings.Contains(rendered, "Type here...") {
		t.Error("Rendered output missing placeholder")
	}

	// Should position cursor after prompt (ESC[row;colH)
	// Expected: row 2 (inside border), col 4 (border + prompt)
	expectedCursor := "\033[2;4H"
	if !strings.Contains(rendered, expectedCursor) {
		t.Errorf("Rendered output missing cursor position %q", expectedCursor)
		// Show what cursor codes are present
		t.Logf("Rendered output:\n%s", rendered)
	}
}

// TestInputTyping validates that TextInput component handles typing correctly
func TestInputTyping(t *testing.T) {
	input := components.NewTextInput("test")
	input.WithPlaceholder("Type here...")

	// Initial state
	if input.Value != "" {
		t.Errorf("Initial value: got %q, want empty", input.Value)
	}
	if input.CursorPos != 0 {
		t.Errorf("Initial cursor: got %d, want 0", input.CursorPos)
	}

	// Type "hello"
	for _, ch := range "hello" {
		input.HandleKeyEvent(tty.KeyEvent{Char: string(ch)})
	}

	if input.Value != "hello" {
		t.Errorf("After typing: got %q, want %q", input.Value, "hello")
	}
	if input.CursorPos != 5 {
		t.Errorf("Cursor after typing: got %d, want 5", input.CursorPos)
	}

	// Backspace
	input.HandleKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace})
	if input.Value != "hell" {
		t.Errorf("After backspace: got %q, want %q", input.Value, "hell")
	}
	if input.CursorPos != 4 {
		t.Errorf("Cursor after backspace: got %d, want 4", input.CursorPos)
	}

	// Arrow left
	input.HandleKeyEvent(tty.KeyEvent{Code: "[D"})
	if input.CursorPos != 3 {
		t.Errorf("Cursor after left arrow: got %d, want 3", input.CursorPos)
	}

	// Type "o" in middle (at position 3 in "hell" → "helol")
	input.HandleKeyEvent(tty.KeyEvent{Char: "o"})
	if input.Value != "helol" {
		t.Errorf("After inserting: got %q, want %q", input.Value, "helol")
	}
	if input.CursorPos != 4 {
		t.Errorf("Cursor after inserting: got %d, want 4", input.CursorPos)
	}
}

// TestInputFocus validates focus management
func TestInputFocus(t *testing.T) {
	input := components.NewTextInput("test-input")

	if !input.IsFocusable() {
		t.Error("TextInput should be focusable")
	}

	// Disabled inputs should not be focusable
	input.Disabled = true
	if input.IsFocusable() {
		t.Error("Disabled TextInput should not be focusable")
	}
}

// TestInputCursorOffset validates cursor offset calculation
func TestInputCursorOffset(t *testing.T) {
	input := components.NewTextInput("test")
	input.Value = "hello"
	input.CursorPos = 5

	offset := input.GetCursorOffset()

	// Cursor offset should account for prompt length ("> " = 2 chars)
	expectedOffset := 2 + 5 // prompt + cursor position
	if offset != expectedOffset {
		t.Errorf("Cursor offset: got %d, want %d", offset, expectedOffset)
	}

	// Test with cursor in middle
	input.CursorPos = 2
	offset = input.GetCursorOffset()
	expectedOffset = 2 + 2
	if offset != expectedOffset {
		t.Errorf("Cursor offset (middle): got %d, want %d", offset, expectedOffset)
	}
}

// TestInputInChatLayout validates input in a realistic chat app layout
func TestInputInChatLayout(t *testing.T) {
	markup := `<vstack height="24" width="80">
		<hbox id="header" height="3" border="1px solid" padding="0">
			<text>Chat TUI Demo</text>
		</hbox>
		<vstack id="messages" flex="1" padding="1">
			<text>AI: Hello!</text>
		</vstack>
		<hbox id="input-box" height="3" border="1px solid" padding="0">
			<input id="input-text">Type a message...</input>
		</hbox>
	</vstack>`

	css := `
		#header { border: 1px solid; height: 3; padding: 0; }
		#messages { flex: 1; padding: 1; }
		#input-box { border: 1px solid; height: 3; padding: 0; }
	`

	// Parse and layout
	root := parser.Parse(markup)
	styles := reconciler.GetStyles(css)
	parser.ApplyStyles(root, styles)
	layout.Layout(root, 80, 24)

	// Find input node
	input := root.FindByID("input-text")
	if input == nil {
		t.Fatal("Input node not found")
	}

	// Validate layout
	// Header: rows 1-3 (border top, content, border bottom)
	// Messages: rows 4-21 (flex=1 fills remaining)
	// Input box: rows 22-24 (3 rows: border top, content, border bottom)

	expectedY := 22 // Row 23 (0-indexed = 22)
	if input.Y != expectedY {
		t.Errorf("Input Y position: got %d, want %d", input.Y, expectedY)
	}

	// Input should be inside the box border
	expectedX := 1 // Position 1 (inside left border)
	if input.X != expectedX {
		t.Errorf("Input X position: got %d, want %d", input.X, expectedX)
	}

	// Cursor should be at row 23 (1-indexed), col 4 (1 border + 2 prompt + 1 for 1-indexed)
	expectedCursorRow := 23
	expectedCursorCol := 4

	actualRow := input.Y + 1
	actualCol := input.X + 2 + 1 // X position + prompt length + 1-indexed

	if actualRow != expectedCursorRow {
		t.Errorf("Cursor row: got %d, want %d", actualRow, expectedCursorRow)
	}
	if actualCol != expectedCursorCol {
		t.Errorf("Cursor col: got %d, want %d", actualCol, expectedCursorCol)
	}
}
