package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/primitives"
	"github.com/speier/smith/pkg/lotus/vdom"
)

// TestComponentReferencePreservation tests that components maintain their reference
// when converted to elements, enabling event routing to work
func TestComponentReferencePreservation(t *testing.T) {
	// Create a component
	input := primitives.CreateInput("", nil)

	// Convert to element (this is what Box() does internally)
	element := vdom.ToElement(input)

	if element == nil {
		t.Fatal("ToElement returned nil")
	}

	// Verify component reference is preserved
	if element.Component == nil {
		t.Error("Component reference was lost during ToElement conversion")
	}

	if element.Component != input {
		t.Error("Component reference doesn't match original component")
	}

	// Verify type is set correctly
	if element.Type != vdom.ComponentElement {
		t.Errorf("Expected ComponentElement type, got %v", element.Type)
	}

	// Verify we can find focusable component in tree
	if focusable, ok := element.Component.(Focusable); !ok {
		t.Error("Component should be Focusable")
	} else if !focusable.IsFocusable() {
		t.Error("TextInput should be focusable")
	}
}

// TestEventRoutingInTree tests that handleEventInTree can find components
func TestEventRoutingInTree(t *testing.T) {
	// Create a simple app with a text input
	type TestApp struct {
		input *primitives.Input
	}

	app := &TestApp{
		input: primitives.CreateInput("", nil),
	}

	// Render method that boxes the input
	renderFunc := func() *vdom.Element {
		return vdom.VStack(
			vdom.Box(vdom.Text("Header")),
			vdom.Box(app.input), // Component should be preserved here
		)
	}

	element := renderFunc()

	// Traverse tree looking for the component
	var foundComponent vdom.Component
	var traverse func(*vdom.Element)
	traverse = func(e *vdom.Element) {
		if e == nil {
			return
		}
		if e.Component != nil {
			foundComponent = e.Component
		}
		for _, child := range e.Children {
			traverse(child)
		}
	}

	traverse(element)

	if foundComponent == nil {
		t.Error("Component not found in element tree")
	}

	if foundComponent != app.input {
		t.Error("Found component doesn't match original input")
	}
}

// Note: Focus integration tests with lotusui components (Tabs, RadioGroup) should be
// in pkg/lotusui tests. Runtime tests focus on core focus management with primitives.
