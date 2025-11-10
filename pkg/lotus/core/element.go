package core

import "sort"

// ComponentRenderer is the interface that components implement to render themselves
// Extends Component with GetID() for auto-registration
type ComponentRenderer interface {
	Component
	GetID() string // Returns component ID for auto-registration
}

// Element represents a virtual DOM element (like React's ReactElement)
// Can contain markup strings, components, or children with inline styles
type Element struct {
	// Type of element: "markup", "component", "container"
	Type string

	// ID of the element (for CSS selectors)
	ID string

	// For markup elements: the raw markup string
	Markup string

	// For component elements: the component instance
	Component ComponentRenderer

	// For container elements: children
	Children []*Element

	// Inline styles (React pattern: style={{ ... }})
	Styles map[string]string
}

// NewMarkupElement creates an element from markup string
func NewMarkupElement(markup string) *Element {
	return &Element{
		Type:   "markup",
		Markup: markup,
		Styles: make(map[string]string),
	}
}

// NewComponentElement creates an element from a component
func NewComponentElement(component ComponentRenderer) *Element {
	elem := &Element{
		Type:      "component",
		Component: component,
		Styles:    make(map[string]string),
	}
	// Set element ID from component ID for diffing
	if component != nil {
		elem.ID = component.GetID()
	}
	return elem
}

// NewContainerElement creates an element with children
func NewContainerElement(children ...*Element) *Element {
	return &Element{
		Type:     "container",
		Children: children,
		Styles:   make(map[string]string),
	}
}

// SetID sets the element ID (for diffing and CSS selectors)
func (e *Element) SetID(id string) *Element {
	e.ID = id
	return e
}

// SetStyle sets an inline style property
func (e *Element) SetStyle(key, value string) *Element {
	if e.Styles == nil {
		e.Styles = make(map[string]string)
	}
	e.Styles[key] = value
	return e
}

// ToMarkup converts the element tree to markup string
func (e *Element) ToMarkup() string {
	switch e.Type {
	case "markup":
		return e.Markup
	case "component":
		return e.Component.Render()
	case "container", "container-with-markup":
		if e.Markup != "" {
			// Pre-rendered markup from builder (includes <box> wrapper with styles)
			return e.Markup
		}
		// Build from children if no pre-rendered markup
		markup := ""
		for _, child := range e.Children {
			markup += child.ToMarkup()
		}
		return markup
	default:
		return ""
	}
}

// Render implements Component interface
func (e *Element) Render() string {
	return e.ToMarkup()
}

// ToCSS generates CSS from the element tree's inline styles
func (e *Element) ToCSS() string {
	css := ""

	// Generate CSS for this element
	if e.ID != "" && len(e.Styles) > 0 {
		css += "#" + e.ID + " {\n"
		// Sort keys for deterministic output
		keys := make([]string, 0, len(e.Styles))
		for key := range e.Styles {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			css += "\t" + key + ": " + e.Styles[key] + ";\n"
		}
		css += "}\n\n"
	}

	// Recursively generate CSS for children
	for _, child := range e.Children {
		css += child.ToCSS()
	}

	return css
}

// ExtractComponents walks the element tree and extracts all components
// Returns map of component ID -> component instance
func (e *Element) ExtractComponents() map[string]interface{} {
	registry := make(map[string]interface{})
	e.extractComponentsRecursive(registry)
	return registry
}

func (e *Element) extractComponentsRecursive(registry map[string]interface{}) {
	switch e.Type {
	case "component":
		if e.Component != nil {
			registry[e.Component.GetID()] = e.Component
		}
	case "container", "container-with-markup":
		for _, child := range e.Children {
			child.extractComponentsRecursive(registry)
		}
	}
}
