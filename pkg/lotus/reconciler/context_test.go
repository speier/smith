package reconciler

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/components"
)

func TestRender_FirstCall(t *testing.T) {
	// Clean up any existing context
	DestroyContext("test-first")

	markup := `<box id="root">Test</box>`
	css := `#root { color: #fff; }`

	output := Render("test-first", markup, css, 100, 40)

	if output == "" {
		t.Error("Render() returned empty string")
	}

	// Verify context was created
	ui := GetContext("test-first")
	if ui == nil {
		t.Fatal("Context was not created")
	}

	if ui.Width != 100 || ui.Height != 40 {
		t.Errorf("Expected dimensions 100x40, got %dx%d", ui.Width, ui.Height)
	}

	DestroyContext("test-first")
}

func TestRender_Reconciliation(t *testing.T) {
	DestroyContext("test-reconcile")

	// First render
	markup1 := `<box id="root">First</box>`
	css1 := `#root { color: #fff; }`
	Render("test-reconcile", markup1, css1, 100, 40)

	ui1 := GetContext("test-reconcile")

	// Second render with different markup (should reconcile, not create new)
	markup2 := `<box id="root">Second</box>`
	css2 := `#root { color: #0f0; }`
	Render("test-reconcile", markup2, css2, 100, 40)

	ui2 := GetContext("test-reconcile")

	// Should be same UI instance (pointer equality)
	if ui1 != ui2 {
		t.Error("Render() created new UI instead of reconciling existing one")
	}

	// Content should be updated
	root := ui2.FindByID("root")
	if root == nil {
		t.Fatal("Root not found")
	}

	// Check that root has children (content was updated)
	if len(root.Children) == 0 {
		t.Error("Root should have children after reconciliation")
	}

	DestroyContext("test-reconcile")
}

func TestRender_DimensionChange(t *testing.T) {
	DestroyContext("test-reflow")

	// First render
	Render("test-reflow", `<box id="root">Test</box>`, ``, 100, 40)
	ui1 := GetContext("test-reflow")

	// Render with different dimensions
	Render("test-reflow", `<box id="root">Test</box>`, ``, 80, 30)
	ui2 := GetContext("test-reflow")

	// Same instance
	if ui1 != ui2 {
		t.Error("Expected same UI instance")
	}

	// Dimensions updated
	if ui2.Width != 80 || ui2.Height != 30 {
		t.Errorf("Expected dimensions 80x30, got %dx%d", ui2.Width, ui2.Height)
	}

	DestroyContext("test-reflow")
}

func TestGetContext_NonExistent(t *testing.T) {
	ui := GetContext("non-existent-context")
	if ui != nil {
		t.Error("GetContext() should return nil for non-existent context")
	}
}

func TestDestroyContext(t *testing.T) {
	DestroyContext("test-destroy")

	Render("test-destroy", `<box>Test</box>`, ``, 100, 40)

	if GetContext("test-destroy") == nil {
		t.Fatal("Context should exist after Render()")
	}

	DestroyContext("test-destroy")

	if GetContext("test-destroy") != nil {
		t.Error("Context should be destroyed")
	}
}

func TestRegisterComponent_Convenience(t *testing.T) {
	DestroyContext("test-register")

	Render("test-register", `<box>Test</box>`, ``, 100, 40)

	// Create a mock component
	input := components.NewTextInput("test-input").WithWidth(50)

	// Use convenience wrapper
	RegisterComponent("test-register", "input", input)

	// Verify it was registered
	ui := GetContext("test-register")
	component := ui.GetComponent("input")

	if component == nil {
		t.Error("Component was not registered")
	}

	if component != input {
		t.Error("Registered component is not the same instance")
	}

	DestroyContext("test-register")
}

func TestSetFocus_Convenience(t *testing.T) {
	DestroyContext("test-focus")

	Render("test-focus", `<box>Test</box>`, ``, 100, 40)
	input := components.NewTextInput("test-input").WithWidth(50)
	RegisterComponent("test-focus", "input", input)

	// Use convenience wrapper
	SetFocus("test-focus", "input")

	// Verify focus was set
	ui := GetContext("test-focus")
	if ui.GetFocus() != "input" {
		t.Errorf("Expected focus on 'input', got '%s'", ui.GetFocus())
	}

	DestroyContext("test-focus")
}

func TestRender_MultipleContexts(t *testing.T) {
	DestroyContext("ctx1")
	DestroyContext("ctx2")

	// Render two independent contexts
	output1 := Render("ctx1", `<box>Context 1</box>`, ``, 100, 40)
	output2 := Render("ctx2", `<box>Context 2</box>`, ``, 80, 30)

	if output1 == output2 {
		t.Error("Different contexts should produce different output")
	}

	ui1 := GetContext("ctx1")
	ui2 := GetContext("ctx2")

	if ui1 == ui2 {
		t.Error("Different contexts should have different UI instances")
	}

	if ui1.Width != 100 || ui2.Width != 80 {
		t.Error("Contexts have incorrect dimensions")
	}

	DestroyContext("ctx1")
	DestroyContext("ctx2")
}

func TestRender_PreservesComponentRegistrations(t *testing.T) {
	DestroyContext("test-preserve")

	// First render and register component
	Render("test-preserve", `<box>First</box>`, ``, 100, 40)
	input := components.NewTextInput("test-input").WithWidth(50)
	RegisterComponent("test-preserve", "input", input)
	SetFocus("test-preserve", "input")

	ui1 := GetContext("test-preserve")
	focus1 := ui1.GetFocus()

	// Second render (reconciliation)
	Render("test-preserve", `<box>Second</box>`, ``, 100, 40)

	ui2 := GetContext("test-preserve")
	focus2 := ui2.GetFocus()
	component2 := ui2.GetComponent("input")

	// Component registration and focus should be preserved
	if focus1 != focus2 {
		t.Errorf("Focus was not preserved: expected '%s', got '%s'", focus1, focus2)
	}

	if component2 != input {
		t.Error("Component registration was not preserved")
	}

	DestroyContext("test-preserve")
}

func TestRender_ThreadSafety(t *testing.T) {
	DestroyContext("test-concurrent")

	// Render from multiple goroutines
	done := make(chan bool)

	for i := 0; i < 10; i++ {
		go func() {
			Render("test-concurrent", `<box>Test</box>`, ``, 100, 40)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Should only have one UI instance
	ui := GetContext("test-concurrent")
	if ui == nil {
		t.Error("Context should exist after concurrent renders")
	}

	DestroyContext("test-concurrent")
}

func TestRender_ActualMarkupChanges(t *testing.T) {
	DestroyContext("test-markup-change")

	// First render with one child
	markup1 := `
		<box id="root">
			<box id="child1">First</box>
		</box>
	`
	output1 := Render("test-markup-change", markup1, ``, 100, 40)

	// Second render with different children
	markup2 := `
		<box id="root">
			<box id="child1">First</box>
			<box id="child2">Second</box>
		</box>
	`
	output2 := Render("test-markup-change", markup2, ``, 100, 40)

	// Output should be different
	if output1 == output2 {
		t.Error("Output should change when markup changes")
	}

	// Should be able to find new child
	ui := GetContext("test-markup-change")
	child2 := ui.FindByID("child2")
	if child2 == nil {
		t.Fatal("New child should exist after reconciliation")
	}

	DestroyContext("test-markup-change")
}
