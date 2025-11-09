package lotus

import (
	"fmt"
	"strings"
)

// VStack creates a vertical stack of children
func VStack(children ...string) string {
	return fmt.Sprintf(`<box direction="column">%s</box>`,
		strings.Join(children, "\n"))
}

// HStack creates a horizontal stack of children
func HStack(children ...string) string {
	return fmt.Sprintf(`<box direction="row">%s</box>`,
		strings.Join(children, "\n"))
}

// Text creates a text element
func Text(content string) string {
	return fmt.Sprintf(`<text>%s</text>`, content)
}

// Input creates an input element
func Input(value string) string {
	return fmt.Sprintf(`<input>%s</input>`, value)
}

// Box creates a generic box with children
func Box(children ...string) string {
	return fmt.Sprintf(`<box>%s</box>`,
		strings.Join(children, "\n"))
}

// BoxWithID creates a box with an ID
func BoxWithID(id string, children ...string) string {
	return fmt.Sprintf(`<box id="%s">%s</box>`, id,
		strings.Join(children, "\n"))
}

// BoxWithClass creates a box with a CSS class
func BoxWithClass(class string, children ...string) string {
	return fmt.Sprintf(`<box class="%s">%s</box>`, class,
		strings.Join(children, "\n"))
}

// Spacer creates a flexible spacer element
func Spacer() string {
	return `<box flex="1"></box>`
}
