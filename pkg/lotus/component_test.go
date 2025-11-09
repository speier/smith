package lotus

import (
	"strings"
	"testing"
)

// MockComponent is a test component implementation
type MockComponent struct {
	content string
}

func (m *MockComponent) Render() string {
	return Text(m.content)
}

func TestComponent(t *testing.T) {
	comp := &MockComponent{content: "Test Content"}
	result := RenderComponent(comp)

	if !strings.Contains(result, "Test Content") {
		t.Error("Component should render its content")
	}
}

func TestRenderComponents(t *testing.T) {
	comp1 := &MockComponent{content: "First"}
	comp2 := &MockComponent{content: "Second"}

	result := RenderComponents(comp1, comp2)

	if !strings.Contains(result, "First") || !strings.Contains(result, "Second") {
		t.Error("RenderComponents should render all components")
	}
	if !strings.Contains(result, `direction="column"`) {
		t.Error("RenderComponents should stack vertically")
	}
}
