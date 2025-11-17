package vdom

// Builder provides React-like helper functions for creating elements
// This makes the API nicer: Box(...) instead of NewElement("box", ...)

// Box creates a box container element (like <div>)
// Accepts string (auto-wrapped as Text), *Element, or Component
func Box(children ...any) *Element {
	elements := make([]*Element, 0, len(children))
	for _, child := range children {
		if elem := toElement(child); elem != nil {
			elements = append(elements, elem)
		}
	}
	// Box should be display:flex with flex-direction:column and align-items:stretch by default
	// This ensures children stretch to fill the container (CSS flexbox default behavior)
	return NewElement("box", Props{
		Styles: map[string]string{
			"display":         "flex",
			"flex-direction":  "column",
			"align-items":     "stretch",
			"justify-content": "flex-start", // Align children to top (main-axis)
		},
	}, elements...)
}

// Text creates a text element
func Text(text string) *Element {
	return NewText(text)
}

// VStack creates a vertical stack (flex-direction: column)
// Accepts string (auto-wrapped as Text), *Element, or Component
// Special case: single []string argument is automatically converted
func VStack(children ...any) *Element {
	// Special case: if single argument is a []string, convert it
	if len(children) == 1 {
		if msgs, ok := children[0].([]string); ok {
			children = make([]any, len(msgs))
			for i, msg := range msgs {
				children[i] = msg
			}
		}
	}

	elements := make([]*Element, 0, len(children))
	for _, child := range children {
		if elem := toElement(child); elem != nil {
			elements = append(elements, elem)
		}
	}
	return NewElement("box", Props{
		Styles: map[string]string{
			"display":         "flex",
			"flex-direction":  "column",
			"align-items":     "stretch",    // CSS default - children stretch to fill cross-axis
			"justify-content": "flex-start", // Align children to top (main-axis)
		},
	}, elements...)
}

// HStack creates a horizontal stack (flex-direction: row)
// Accepts string (auto-wrapped as Text), *Element, or Component
// Special case: single []string argument is automatically converted
func HStack(children ...any) *Element {
	// Special case: if single argument is a []string, convert it
	if len(children) == 1 {
		if msgs, ok := children[0].([]string); ok {
			children = make([]any, len(msgs))
			for i, msg := range msgs {
				children[i] = msg
			}
		}
	}

	elements := make([]*Element, 0, len(children))
	for _, child := range children {
		if elem := toElement(child); elem != nil {
			elements = append(elements, elem)
		}
	}
	return NewElement("box", Props{
		Styles: map[string]string{
			"display":        "flex",
			"flex-direction": "row",
		},
	}, elements...)
}

// toElement converts any value to an Element
// Handles: string (auto-wrap as Text), Node (Element or Component), *Element
func toElement(v any) *Element {
	switch val := v.(type) {
	case string:
		// Auto-wrap strings as Text nodes
		return Text(val)
	case *Element:
		return val
	case Node:
		return ToElement(val)
	default:
		return nil
	}
}

// Map transforms a slice of items into Elements using a mapping function (like React's map)
// Usage: lotus.VStack(lotus.Map(messages, lotus.Text)...)
func Map[T any](items []T, fn func(T) *Element) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

// Strings converts []string to []any for use in VStack/HStack
// Strings are auto-converted to Text elements by toElement()
// Usage: lotus.VStack(lotus.Strings(messages)...)
func Strings(items []string) []any {
	result := make([]any, len(items))
	for i, item := range items {
		result[i] = item
	}
	return result
}
