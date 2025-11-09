package lotus

// Component represents a reusable UI component that can render itself to markup
type Component interface {
	Render() string
}

// RenderComponent is a helper function to render a component
func RenderComponent(c Component) string {
	return c.Render()
}

// RenderComponents renders multiple components and joins them
func RenderComponents(components ...Component) string {
	rendered := make([]string, len(components))
	for i, c := range components {
		rendered[i] = c.Render()
	}
	return VStack(rendered...)
}
