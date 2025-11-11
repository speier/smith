package runtime

import (
	"testing"

	"github.com/speier/smith/pkg/lotus/vdom"
)

// Mock stateful component for testing
type mockStateful struct {
	id    string
	state map[string]interface{}
}

func (m *mockStateful) GetID() string {
	return m.id
}

func (m *mockStateful) SaveState() map[string]interface{} {
	return m.state
}

func (m *mockStateful) LoadState(state map[string]interface{}) error {
	m.state = state
	return nil
}

func (m *mockStateful) Render() *vdom.Element {
	return vdom.Text("mock")
}

func TestCollectStatefulComponents(t *testing.T) {
	tests := []struct {
		name     string
		element  *vdom.Element
		expected int
	}{
		{
			name:     "nil element",
			element:  nil,
			expected: 0,
		},
		{
			name:     "element without component",
			element:  vdom.Box(vdom.Text("hello")),
			expected: 0,
		},
		{
			name: "single stateful component",
			element: &vdom.Element{
				Type:      vdom.ComponentElement,
				Component: &mockStateful{id: "test1", state: map[string]interface{}{"value": "foo"}},
			},
			expected: 1,
		},
		{
			name: "nested stateful components",
			element: vdom.VStack(
				&vdom.Element{
					Type:      vdom.ComponentElement,
					Component: &mockStateful{id: "test1", state: map[string]interface{}{"value": "foo"}},
				},
				vdom.Box(&vdom.Element{
					Type:      vdom.ComponentElement,
					Component: &mockStateful{id: "test2", state: map[string]interface{}{"value": "bar"}},
				}),
			),
			expected: 2,
		},
		{
			name: "mixed components (stateful and non-stateful)",
			element: vdom.VStack(
				&vdom.Element{
					Type:      vdom.ComponentElement,
					Component: &mockStateful{id: "test1", state: map[string]interface{}{"value": "foo"}},
				},
				vdom.Text("regular text"),
				vdom.Box(vdom.Text("more text")),
			),
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			components := CollectStatefulComponents(tt.element)
			if len(components) != tt.expected {
				t.Errorf("CollectStatefulComponents() = %d components, want %d", len(components), tt.expected)
			}
		})
	}
}

func TestStatefulComponentsWithoutID(t *testing.T) {
	// Component without ID should still be collected but won't persist
	comp := &mockStateful{id: "", state: map[string]interface{}{"value": "test"}}
	element := &vdom.Element{
		Type:      vdom.ComponentElement,
		Component: comp,
	}

	components := CollectStatefulComponents(element)
	if len(components) != 1 {
		t.Errorf("Expected 1 component, got %d", len(components))
	}

	if components[0].GetID() != "" {
		t.Errorf("Expected empty ID, got %q", components[0].GetID())
	}
}

func TestStateSaveAndLoad(t *testing.T) {
	comp1 := &mockStateful{
		id:    "comp1",
		state: map[string]interface{}{"value": "hello", "count": float64(42)},
	}
	comp2 := &mockStateful{
		id:    "comp2",
		state: map[string]interface{}{"checked": true},
	}

	// Save state
	saved1 := comp1.SaveState()
	saved2 := comp2.SaveState()

	// Create new components
	newComp1 := &mockStateful{id: "comp1"}
	newComp2 := &mockStateful{id: "comp2"}

	// Load state
	if err := newComp1.LoadState(saved1); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}
	if err := newComp2.LoadState(saved2); err != nil {
		t.Fatalf("LoadState failed: %v", err)
	}

	// Verify state restored
	if v, ok := newComp1.state["value"].(string); !ok || v != "hello" {
		t.Errorf("State not restored correctly: value = %v", newComp1.state["value"])
	}
	if v, ok := newComp1.state["count"].(float64); !ok || v != 42 {
		t.Errorf("State not restored correctly: count = %v", newComp1.state["count"])
	}
	if v, ok := newComp2.state["checked"].(bool); !ok || v != true {
		t.Errorf("State not restored correctly: checked = %v", newComp2.state["checked"])
	}
}
