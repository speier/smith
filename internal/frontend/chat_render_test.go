package frontend

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/agent/session"
)

// TestInputRendering tests that typed text is visible in the rendered output
func TestInputRendering(t *testing.T) {
	// Create a mock session
	sess := &session.MockSession{}

	// Create UI with known dimensions
	ui := NewChatUI(sess, 80, 24)

	// Simulate typing "hello world"
	ui.input.Value = "hello world"
	ui.input.CursorPos = 11

	// Build the UI (this is what happens on first render)
	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)
	ui.lastMsgCount = len(history)
	ui.lastInput = ui.input.Value

	// Render to ANSI string
	output := ui.cachedUI.RenderToTerminal(true)

	// Strip ANSI codes for easier testing
	plainOutput := stripANSI(output)

	// Check that "hello world" appears in the output
	if !strings.Contains(plainOutput, "hello world") {
		t.Errorf("Expected 'hello world' in output, got:\n%s", plainOutput)
	}

	// Check that "> " prompt appears
	if !strings.Contains(plainOutput, ">") {
		t.Errorf("Expected '>' prompt in output, got:\n%s", plainOutput)
	}
}

// TestInputUpdate tests that updating input updates the rendered output
// TODO: Re-enable after updating to Element API
/*
func TestInputUpdate(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	// Initial render with "hello"
	ui.input.Value = "hello"
	ui.input.CursorPos = 5
	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)
	ui.lastMsgCount = len(history)
	ui.lastInput = ui.input.Value

	output1 := ui.cachedUI.RenderToTerminal(true)
	plain1 := stripANSI(output1)

	if !strings.Contains(plain1, "hello") {
		t.Errorf("Expected 'hello' in initial output")
	}

	// Update to "hello world" using the fast path
	ui.input.Value = "hello world"
	ui.input.CursorPos = 11
	ui.updateInputNode()
	ui.lastInput = ui.input.Value

	output2 := ui.cachedUI.RenderToTerminal(true)
	plain2 := stripANSI(output2)

	if !strings.Contains(plain2, "hello world") {
		t.Errorf("Expected 'hello world' in updated output, got:\n%s", plain2)
	}
}
*/

// TestInputNodeStructure tests that the input box has the correct child structure
func TestInputNodeStructure(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	ui.input.Value = "test"
	ui.input.CursorPos = 4
	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)

	// Find the input-text box
	inputBox := ui.cachedUI.FindByID("input-text")
	if inputBox == nil {
		t.Fatal("Could not find input-text box")
	}

	// Check it has exactly one child (the text node)
	if len(inputBox.Children) != 1 {
		t.Errorf("Expected input box to have 1 child, got %d", len(inputBox.Children))
	}

	// Check the child is a text node
	if inputBox.Children[0].Type != "text" {
		t.Errorf("Expected child to be 'text' type, got '%s'", inputBox.Children[0].Type)
	}

	// Check the content
	if inputBox.Children[0].Content != "test" {
		t.Errorf("Expected content 'test', got '%s'", inputBox.Children[0].Content)
	}

	// Check that the text node has valid layout coordinates
	textNode := inputBox.Children[0]
	if textNode.X == 0 && textNode.Y == 0 {
		t.Error("Text node has zero coordinates - layout not applied")
	}

	if textNode.Width == 0 || textNode.Height == 0 {
		t.Errorf("Text node has zero dimensions: %dx%d", textNode.Width, textNode.Height)
	}
}

// TestUpdateInputNode tests the fast update path
// TODO: Re-enable after updating to Element API
/*
func TestUpdateInputNode(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	// Initial build
	ui.input.Value = "initial"
	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)
	ui.lastInput = ui.input.Value

	inputBox := ui.cachedUI.FindByID("input-text")
	if inputBox == nil {
		t.Fatal("Could not find input-text box")
	}

	originalChild := inputBox.Children[0]
	originalX := originalChild.X
	originalY := originalChild.Y

	// Update via fast path
	ui.input.Value = "updated text"
	ui.updateInputNode()

	// Check content changed
	if inputBox.Children[0].Content != "updated text" {
		t.Errorf("Expected 'updated text', got '%s'", inputBox.Children[0].Content)
	}

	// Check it's the same child node (not recreated)
	if inputBox.Children[0] != originalChild {
		t.Error("updateInputNode created a new child instead of updating existing one")
	}

	// Check coordinates didn't change (no layout recalc needed)
	if inputBox.Children[0].X != originalX || inputBox.Children[0].Y != originalY {
		t.Error("updateInputNode changed text node coordinates")
	}
}
*/

// TestEmptyInput tests rendering with empty input
func TestEmptyInput(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	ui.input.Value = ""
	ui.input.CursorPos = 0
	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)

	output := ui.cachedUI.RenderToTerminal(true)
	plain := stripANSI(output)

	// Should have prompt even with empty input
	if !strings.Contains(plain, ">") {
		t.Error("Expected prompt '>' in output with empty input")
	}

	// Check input box has a placeholder
	inputBox := ui.cachedUI.FindByID("input-text")
	if inputBox == nil {
		t.Fatal("Could not find input-text box")
	}

	if len(inputBox.Children) > 0 && inputBox.Children[0].Content == "" {
		t.Error("Empty input should have placeholder, got empty content")
	}
}

// TestLongInput tests that long input is handled correctly
func TestLongInput(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	// Input longer than screen width
	longText := strings.Repeat("x", 200)
	ui.input.Value = longText
	ui.input.CursorPos = 100
	ui.input.Scroll = 50

	history := sess.GetHistory()
	ui.cachedUI = ui.buildFullUI(history)

	// Get visible portion
	visibleInput, _ := ui.input.GetVisible()

	// Check that visible input is a substring of the full input
	if !strings.Contains(longText, visibleInput) {
		t.Error("Visible input is not a substring of full input")
	}

	// Check that visible input is not the entire string (it should be scrolled)
	if visibleInput == longText {
		t.Error("Expected input to be scrolled for long text")
	}

	// Render should contain the visible portion
	output := ui.cachedUI.RenderToTerminal(true)
	plain := stripANSI(output)

	if !strings.Contains(plain, visibleInput) {
		t.Errorf("Expected visible portion '%s' in output", visibleInput[:20]+"...")
	}
}

// TestRenderCaching tests that caching works correctly
// TODO: Re-enable after updating to Element API
/*
func TestRenderCaching(t *testing.T) {
	sess := &session.MockSession{}
	ui := NewChatUI(sess, 80, 24)

	// First render
	ui.input.Value = "test"
	output1 := ui.render()
	cachedUI1 := ui.cachedUI

	// Second render with same input - should use cache
	_ = ui.render()
	cachedUI2 := ui.cachedUI

	if cachedUI1 != cachedUI2 {
		t.Error("Expected cached UI to be reused for same input")
	}

	// Third render with different input - should update node but keep cache
	ui.input.Value = "changed"
	output3 := ui.render()
	cachedUI3 := ui.cachedUI

	if cachedUI1 != cachedUI3 {
		t.Error("Expected cached UI to be reused when only input changes")
	}

	// Check outputs are different
	if output1 == output3 {
		t.Error("Expected different output when input changes")
	}
}
*/

// Helper: Strip ANSI escape codes for easier text comparison
func stripANSI(s string) string {
	// Remove ANSI escape sequences
	var result strings.Builder
	inEscape := false

	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
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
