package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/tty"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestFormIntegration simulates the exact form example scenario
// to diagnose why typing and Tab don't work
func TestFormIntegration(t *testing.T) {
	// Create a simple app that mimics the form example
	type FormApp struct {
		name    string
		email   string
		message string
	}

	app := &FormApp{}

	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Text("Name:"),
			primitives.CreateInput("Enter your name", func(ctx primitives.Context, text string) {
				app.name = text
			}),
			vdom.Text("Email:"),
			primitives.CreateInput("Enter your email", func(ctx primitives.Context, text string) {
				app.email = text
			}),
			vdom.Text("Message:"),
			primitives.CreateInput("Enter a message", func(ctx primitives.Context, text string) {
				app.message = text
			}),
		)
	}

	fm := newFocusManager()

	// First render
	tree1 := renderFunc()
	fm.rebuild(tree1)

	t.Logf("After first render: %d focusables", len(fm.focusables))
	if len(fm.focusables) != 3 {
		t.Fatalf("Expected 3 focusables, got %d", len(fm.focusables))
	}

	// Get first input
	nameInput := tree1.Children[1].Component.(*primitives.Input)
	t.Logf("Name input focused: %v", nameInput.Focused)

	if !nameInput.Focused {
		t.Error("First input should be focused after rebuild")
	}

	// Simulate typing 'h' in first input
	event := tty.KeyEvent{Key: 'h', Char: "h"}
	ctx := Context{}
	handled := nameInput.HandleKey(ctx, event)
	t.Logf("Typing 'h' - handled: %v, value: %q, cursorPos: %d", handled, nameInput.Value, nameInput.CursorPos)

	if !handled {
		t.Error("Input should handle typing event")
	}
	if nameInput.Value != "h" {
		t.Errorf("After typing 'h', value should be 'h', got %q", nameInput.Value)
	}
	if nameInput.CursorPos != 1 {
		t.Errorf("After typing 'h', cursor should be at 1, got %d", nameInput.CursorPos)
	}

	// Second render (simulating re-render after state change)
	// This is the CRITICAL test: the new Input created by CreateInput
	// should be REPLACED by the cached one BEFORE its Render() is called
	tree2 := renderFunc()

	// BEFORE reconciliation, the tree has NEW Input instances (empty state)
	nameInputBeforeReconcile := tree2.Children[1].Component.(*primitives.Input)
	t.Logf("Before reconcile - new instance value: %q, cursor: %d",
		nameInputBeforeReconcile.Value, nameInputBeforeReconcile.CursorPos)

	// But the RENDERED tree is from the new instance!
	// Check what the actual rendered element tree contains
	nameElemBeforeReconcile := tree2.Children[1]
	t.Logf("Before reconcile - rendered element type: %d, tag: %s, children: %d",
		nameElemBeforeReconcile.Type, nameElemBeforeReconcile.Tag, len(nameElemBeforeReconcile.Children))

	if nameInputBeforeReconcile == nameInput {
		t.Error("Before reconciliation, should have NEW instance (not cached)")
	}

	// The problem: the Element tree was rendered from the NEW empty Input!
	// When we reconcile and replace Component pointer, the tree is already wrong!

	// Reconcile should replace the new instances with cached ones
	fm.reconcileComponents(tree2, "0")

	// AFTER reconciliation
	nameInput2 := tree2.Children[1].Component.(*primitives.Input)
	t.Logf("After re-render - same instance: %v, value: %q", nameInput2 == nameInput, nameInput2.Value)

	if nameInput2 != nameInput {
		t.Error("Reconciliation should preserve same instance")
	}
	if nameInput2.Value != "h" {
		t.Errorf("After re-render, value should still be 'h', got %q", nameInput2.Value)
	}

	// Simulate Tab key
	fm.next()
	t.Logf("After Tab - focusIndex: %d", fm.focusIndex)

	if fm.focusIndex != 1 {
		t.Errorf("After Tab, focusIndex should be 1, got %d", fm.focusIndex)
	}

	// Third render
	tree3 := renderFunc()
	fm.rebuild(tree3)

	emailInput := tree3.Children[3].Component.(*primitives.Input)
	t.Logf("Email input focused: %v", emailInput.Focused)

	if !emailInput.Focused {
		t.Error("Email input should be focused after Tab")
	}

	// Name input should not be focused
	nameInput3 := tree3.Children[1].Component.(*primitives.Input)
	if nameInput3.Focused {
		t.Error("Name input should not be focused after Tab to email")
	}

	// Verify name still has its value
	if nameInput3.Value != "h" {
		t.Errorf("Name value should still be 'h', got %q", nameInput3.Value)
	}
}

// TestSimpleInputIntegration tests the simple-input example scenario
func TestSimpleInputIntegration(t *testing.T) {
	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Text("Simple Input Test"),
			primitives.CreateInput("Type here...", nil),
		)
	}

	fm := newFocusManager()

	// First render
	tree1 := renderFunc()
	fm.rebuild(tree1)

	if len(fm.focusables) != 1 {
		t.Fatalf("Expected 1 focusable, got %d", len(fm.focusables))
	}

	input1 := tree1.Children[1].Component.(*primitives.Input)
	t.Logf("Input focused: %v", input1.Focused)

	if !input1.Focused {
		t.Error("Input should be auto-focused")
	}

	// Type a character
	event := tty.KeyEvent{Key: 'a', Char: "a"}
	ctx := Context{}
	handled := input1.HandleKey(ctx, event)
	t.Logf("Typing 'a' - handled: %v, value: %q", handled, input1.Value)

	if !handled {
		t.Error("Input should handle typing")
	}
	if input1.Value != "a" {
		t.Errorf("Value should be 'a', got %q", input1.Value)
	}

	// Re-render
	tree2 := renderFunc()
	fm.rebuild(tree2)

	input2 := tree2.Children[1].Component.(*primitives.Input)
	t.Logf("After re-render - same instance: %v, value: %q", input2 == input1, input2.Value)

	if input2 != input1 {
		t.Error("Reconciliation should preserve instance")
	}
	if input2.Value != "a" {
		t.Errorf("Value should still be 'a' after re-render, got %q", input2.Value)
	}
}
