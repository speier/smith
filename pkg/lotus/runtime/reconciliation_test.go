package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestReconciliation verifies that component instances are preserved across renders
func TestReconciliation(t *testing.T) {
	fm := newFocusManager()

	// First render: create an input and set some state
	input1 := primitives.NewInput()
	input1.Value = "test value"
	input1.CursorPos = 5

	tree1 := vdom.VStack(
		vdom.Text("Label"),
		input1,
	)

	// Reconcile first time - should cache the input
	fm.reconcileComponents(tree1, "0")

	// Verify input was cached
	if len(fm.componentCache) != 1 {
		t.Fatalf("Expected 1 cached component, got %d", len(fm.componentCache))
	}

	// Second render: create a NEW input instance (simulating Render() recreating it)
	input2 := primitives.NewInput() // Fresh instance, empty state
	if input2.Value != "" {
		t.Fatal("New input should start with empty value")
	}

	tree2 := vdom.VStack(
		vdom.Text("Label"),
		input2, // NEW instance at same position
	)

	// Reconcile second time - should REPLACE input2 with cached input1
	fm.reconcileComponents(tree2, "0")

	// Get the reconciled component from the tree
	reconciledInput := tree2.Children[1].Component.(*primitives.Input)

	// Verify it's the CACHED instance (input1), not the new one (input2)
	if reconciledInput.Value != "test value" {
		t.Errorf("Expected reconciled input to have cached value 'test value', got '%s'", reconciledInput.Value)
	}
	if reconciledInput.CursorPos != 5 {
		t.Errorf("Expected reconciled input to have cached cursor pos 5, got %d", reconciledInput.CursorPos)
	}

	// Verify it's actually the same instance (pointer equality)
	if reconciledInput != input1 {
		t.Error("Expected reconciled input to be the same instance as cached input1")
	}
}

// TestReconciliationMultipleInputs verifies reconciliation works with multiple inputs
func TestReconciliationMultipleInputs(t *testing.T) {
	fm := newFocusManager()

	// First render: create two inputs with different states
	input1 := primitives.NewInput()
	input1.Value = "first"
	input2 := primitives.NewInput()
	input2.Value = "second"

	tree1 := vdom.VStack(
		input1,
		input2,
	)

	fm.reconcileComponents(tree1, "0")

	// Second render: create NEW instances
	newInput1 := primitives.NewInput()
	newInput2 := primitives.NewInput()

	tree2 := vdom.VStack(
		newInput1,
		newInput2,
	)

	fm.reconcileComponents(tree2, "0")

	// Verify both were reconciled correctly
	reconciled1 := tree2.Children[0].Component.(*primitives.Input)
	reconciled2 := tree2.Children[1].Component.(*primitives.Input)

	if reconciled1.Value != "first" {
		t.Errorf("First input: expected 'first', got '%s'", reconciled1.Value)
	}
	if reconciled2.Value != "second" {
		t.Errorf("Second input: expected 'second', got '%s'", reconciled2.Value)
	}
}

// TestReconciliationPositionMatters verifies that position is part of the cache key
func TestReconciliationPositionMatters(t *testing.T) {
	fm := newFocusManager()

	// First render: input at position 0
	input1 := primitives.NewInput()
	input1.Value = "position zero"

	tree1 := vdom.VStack(
		input1,
		vdom.Text("Other"),
	)

	fm.reconcileComponents(tree1, "0")

	// Second render: input at DIFFERENT position (1 instead of 0)
	newInput := primitives.NewInput()
	newInput.Value = "position one"

	tree2 := vdom.VStack(
		vdom.Text("Other"),
		newInput, // Now at position 1
	)

	fm.reconcileComponents(tree2, "0")

	// Verify new input at position 1 is cached separately
	reconciledInput := tree2.Children[1].Component.(*primitives.Input)
	if reconciledInput.Value != "position one" {
		t.Errorf("Input at new position should have its own state, got '%s'", reconciledInput.Value)
	}

	// Verify cache now has 2 entries (different positions)
	if len(fm.componentCache) != 2 {
		t.Errorf("Expected 2 cached components at different positions, got %d", len(fm.componentCache))
	}
}

// TestReconciliationWithTabNavigation simulates the form example scenario:
// - Multiple inputs created inline in Render()
// - Typing in inputs should preserve state
// - Tab navigation should work correctly
func TestReconciliationWithTabNavigation(t *testing.T) {
	fm := newFocusManager()

	// Helper to simulate a form's Render() method
	// This creates NEW input instances every time (inline creation)
	renderForm := func() *vdom.Element {
		return vdom.VStack(
			vdom.Text("Name:"),
			primitives.CreateInput("Enter your name", nil),
			vdom.Text("Email:"),
			primitives.CreateInput("Enter your email", nil),
			vdom.Text("Message:"),
			primitives.CreateInput("Enter a message", nil),
		)
	}

	// First render
	tree1 := renderForm()
	fm.rebuild(tree1)

	// Should have 3 focusables
	if len(fm.focusables) != 3 {
		t.Fatalf("Expected 3 focusables, got %d", len(fm.focusables))
	}

	// Get the inputs from the tree (after reconciliation)
	nameInput := tree1.Children[1].Component.(*primitives.Input)
	emailInput := tree1.Children[3].Component.(*primitives.Input)
	messageInput := tree1.Children[5].Component.(*primitives.Input)

	// Simulate typing in first input
	nameInput.Value = "John"
	nameInput.CursorPos = 4

	// Second render (NEW instances created inline)
	tree2 := renderForm()
	fm.rebuild(tree2)

	// Verify the SAME instances were preserved
	nameInput2 := tree2.Children[1].Component.(*primitives.Input)
	if nameInput2 != nameInput {
		t.Error("Name input should be the same instance after reconciliation")
	}
	if nameInput2.Value != "John" {
		t.Errorf("Name input value should be preserved, got '%s'", nameInput2.Value)
	}
	if nameInput2.CursorPos != 4 {
		t.Errorf("Name input cursor should be preserved, got %d", nameInput2.CursorPos)
	}

	// Tab to next input
	fm.next()

	// Third render (NEW instances again)
	tree3 := renderForm()
	fm.rebuild(tree3)

	// Email input should now be focused
	emailInput3 := tree3.Children[3].Component.(*primitives.Input)
	if !emailInput3.Focused {
		t.Error("Email input should be focused after Tab")
	}

	// Name input state should still be preserved
	nameInput3 := tree3.Children[1].Component.(*primitives.Input)
	if nameInput3.Value != "John" {
		t.Errorf("Name input value should still be preserved after Tab, got '%s'", nameInput3.Value)
	}

	// Type in email input
	emailInput3.Value = "john@example.com"
	emailInput3.CursorPos = 16

	// Fourth render
	tree4 := renderForm()
	fm.rebuild(tree4)

	// Both inputs should preserve state
	nameInput4 := tree4.Children[1].Component.(*primitives.Input)
	emailInput4 := tree4.Children[3].Component.(*primitives.Input)

	if nameInput4.Value != "John" {
		t.Errorf("Name input value lost, got '%s'", nameInput4.Value)
	}
	if emailInput4.Value != "john@example.com" {
		t.Errorf("Email input value lost, got '%s'", emailInput4.Value)
	}

	// Tab to third input
	fm.next()

	// Fifth render
	tree5 := renderForm()
	fm.rebuild(tree5)

	messageInput5 := tree5.Children[5].Component.(*primitives.Input)
	if !messageInput5.Focused {
		t.Error("Message input should be focused after second Tab")
	}

	// All previous state should be preserved
	nameInput5 := tree5.Children[1].Component.(*primitives.Input)
	emailInput5 := tree5.Children[3].Component.(*primitives.Input)

	if nameInput5.Value != "John" {
		t.Errorf("Name input value lost after Tab to third input, got '%s'", nameInput5.Value)
	}
	if emailInput5.Value != "john@example.com" {
		t.Errorf("Email input value lost after Tab to third input, got '%s'", emailInput5.Value)
	}

	// Verify all instances are preserved throughout
	if nameInput5 != nameInput {
		t.Error("Name input instance should be preserved through all renders")
	}
	if emailInput5 != emailInput {
		t.Error("Email input instance should be preserved through all renders")
	}
	if messageInput5 != messageInput {
		t.Error("Message input instance should be preserved through all renders")
	}
}
