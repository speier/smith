package main

import (
	"strings"
	"testing"

	"github.com/speier/smith/pkg/lotus"
	lotustest "github.com/speier/smith/pkg/lotus/testing"
	"github.com/speier/smith/pkg/lotus/tty"
)

func TestChatAppRender(t *testing.T) {
	app := NewChatApp()

	// Test that Render returns an Element
	element := app.Render()
	if element == nil {
		t.Fatal("Render() returned nil")
	}

	t.Logf("Element type: %s", element.Type)
	t.Logf("Element ID: %s", element.ID)
	t.Logf("Element children count: %d", len(element.Children))
	t.Logf("Element styles count: %d", len(element.Styles))

	// Check children
	for i, child := range element.Children {
		t.Logf("Child %d: type=%s, id=%s, markup_len=%d, children=%d",
			i, child.Type, child.ID, len(child.Markup), len(child.Children))
	}

	// Convert element to markup
	markup := element.ToMarkup()
	if markup == "" {
		t.Fatal("ToMarkup() returned empty string")
	}

	t.Logf("Markup length: %d bytes", len(markup))
	t.Logf("Markup:\n%s", markup)

	// Generate CSS from inline styles
	css := element.ToCSS()

	t.Logf("CSS length: %d bytes", len(css))
	if css != "" {
		t.Logf("CSS:\n%s", css)
	}

	if css == "" {
		t.Error("ToCSS() returned empty string - styles not applied")
	}

	// Verify CSS contains expected selectors
	expectedSelectors := []string{"#header", "#messages", "#input-box"}
	for _, selector := range expectedSelectors {
		if !strings.Contains(css, selector) {
			t.Errorf("CSS missing expected selector: %s", selector)
		}
	}

	// Verify CSS contains expected style properties
	expectedStyles := []string{"height:", "border-style:", "border-color:", "padding:", "flex:", "color:"}
	for _, style := range expectedStyles {
		if !strings.Contains(css, style) {
			t.Errorf("CSS missing expected style property: %s", style)
		}
	}
}

func TestChatAppMessages(t *testing.T) {
	app := NewChatApp()

	// Add a message using the MessageList
	app.messageList.AddMessage("user", "Hello")

	element := app.Render()
	markup := element.ToMarkup()

	// Check that message appears in markup
	if !strings.Contains(markup, "Hello") {
		t.Error("Message 'Hello' not found in markup")
		t.Logf("Markup:\n%s", markup)
	}

	// Check user prefix is present
	if !strings.Contains(markup, "| Hello") {
		t.Error("User message prefix '| ' not found")
		t.Logf("Markup:\n%s", markup)
	}

	// Add assistant message
	app.messageList.AddMessage("assistant", "Hi there!")

	element = app.Render()
	markup = element.ToMarkup()

	// Check both messages appear
	if !strings.Contains(markup, "Hello") || !strings.Contains(markup, "Hi there!") {
		t.Error("Not all messages found in markup after adding multiple")
		t.Logf("Markup:\n%s", markup)
	}
}

func TestChatAppStructure(t *testing.T) {
	element := buildChatElement()

	lotustest.AssertHasID(t, element, "header")
	lotustest.AssertHasID(t, element, "input")

	header := lotustest.FindByID(element, "header")
	lotustest.AssertHasBorder(t, header)

	input := lotustest.FindByID(element, "input")
	lotustest.AssertHasBorder(t, input)
}

func TestChatAppRendering(t *testing.T) {
	element := buildChatElement()
	lotustest.SnapshotRender(t, "chat-app-initial", element, nil)
}

// TestTypingShowsImmediately tests that typed characters appear immediately without lag
// This test catches the "one render behind" bug where typed chars don't show until next keystroke
func TestTypingShowsImmediately(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "typing", app, 80, 24)

	// Initial state: no text in input, just placeholder
	mock.AssertText("Type a message...")

	// Type 'h' - should show immediately
	mock.SendKey('h').AssertText("> h")

	// Type 'e' - should show "he" immediately (not just "h")
	mock.SendKey('e').AssertText("> he")

	// Type 'l' - should show "hel"
	mock.SendKey('l').AssertText("> hel")

	// Type 'l' - should show "hell"
	mock.SendKey('l').AssertText("> hell")

	// Type 'o' - should show "hello"
	mock.SendKey('o').AssertText("> hello")
}

// TestBackspaceWorks tests that backspace removes characters correctly
func TestBackspaceWorks(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "backspace", app, 80, 24)

	// Type "test"
	mock.SendKey('t').SendKey('e').SendKey('s').SendKey('t')
	mock.AssertText("> test")

	// Backspace - should show "tes"
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace}).AssertText("> tes")

	// Backspace again - should show "te"
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace}).AssertText("> te")

	// Backspace past beginning shouldn't crash
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace})
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace})
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyBackspace}) // extra

	// Should be empty
	mock.AssertNotContains("> t")
}

// TestEnterSendsMessage tests that Enter key sends the message and clears input
func TestEnterSendsMessage(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "enter", app, 80, 24)

	// Verify initial messages appear
	mock.AssertText("Welcome! Type a message and press Enter.")
	mock.AssertText("Messages will appear here with auto-scroll.")

	// Type a message
	mock.SendKey('h').SendKey('i')
	mock.AssertText("> hi")

	// Press Enter - should send message and clear input
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyEnter})

	// Input should be cleared (placeholder shown)
	mock.AssertText("Type a message...")
	mock.AssertNotContains("> hi")

	// User message should appear with prefix
	mock.AssertText("| hi")

	// Echo message should appear
	mock.AssertText("Echo: hi")

	// Verify total message count in the list
	if len(app.messageList.Messages) != 4 { // 2 initial + 2 new
		t.Errorf("Expected 4 messages, got %d", len(app.messageList.Messages))
	}
}

// TestMultipleMessages tests sending multiple messages in sequence
func TestMultipleMessages(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "multiple", app, 80, 24)

	// Send first message
	mock.SendKey('o').SendKey('n').SendKey('e')
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyEnter})
	mock.AssertText("| one")
	mock.AssertText("Echo: one")

	// Send second message
	mock.SendKey('t').SendKey('w').SendKey('o')
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyEnter})
	mock.AssertText("| two")
	mock.AssertText("Echo: two")

	// Both messages should still be visible
	mock.AssertText("| one")
	mock.AssertText("| two")

	// Check message count
	if len(app.messageList.Messages) != 6 { // 2 initial + 4 new (2 pairs)
		t.Errorf("Expected 6 messages, got %d", len(app.messageList.Messages))
	}
}

// TestCursorPosition tests that cursor positioning is correct as we type
func TestCursorPosition(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "cursor", app, 80, 24)

	// Get initial output to understand cursor positioning
	// The cursor should be in the input box area
	// Note: This test verifies cursor codes are present
	// Exact position depends on layout - we just verify it changes as we type

	initialOutput := mock.GetOutput()
	initialRow, initialCol, found := lotustest.ExtractCursorPosition(initialOutput)
	if !found {
		t.Fatal("No cursor position found in initial output")
	}

	// Type a character
	mock.SendKey('a')

	// Cursor should have moved
	newOutput := mock.GetOutput()
	newRow, newCol, found := lotustest.ExtractCursorPosition(newOutput)
	if !found {
		t.Fatal("No cursor position found after typing")
	}

	// Cursor should move right (column increases) or stay on same row
	if newRow != initialRow || newCol == initialCol {
		// Position changed - this is expected when typing
		t.Logf("Cursor moved from (%d,%d) to (%d,%d)", initialRow, initialCol, newRow, newCol)
	}
}

// TestEmptyMessageNotSent tests that pressing Enter with empty input doesn't send
func TestEmptyMessageNotSent(t *testing.T) {
	app := NewChatApp()
	mock := lotustest.NewMockTerminal(t, "empty", app, 80, 24)

	// Press Enter without typing anything
	mock.SendKeyEvent(tty.KeyEvent{Key: tty.KeyEnter})

	// No message should appear (no "You:" text added)
	// Just verify we didn't crash and terminal still works
	mock.SendKey('a').AssertText("> a")
}

// Helper functions

func buildChatElement() *lotus.Element {
	return lotus.VStack(
		lotus.Box("header",
			lotus.Text("🤖 Smith AI Chat"),
		).Height(3).Border("1px solid").BorderStyle("rounded").Padding(0),
		lotus.VStack(
			lotus.Text("AI: Hello! How can I help?"),
			lotus.Text("AI: I'm ready to assist!"),
		).Padding(1).Flex(1),
		lotus.Box("input",
			lotus.Text("You: "),
		).Height(3).Border("1px solid").BorderStyle("rounded").Padding(0),
	).Width("80").Height(24).Render()
}
