package lotus_test

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// ExampleVStack demonstrates creating a vertical stack of elements
func ExampleVStack() {
	elem := lotus.VStack(
		lotus.Text("Title"),
		lotus.Text("Subtitle"),
		lotus.Text("Content"),
	).Render()

	markup := elem.ToMarkup()
	fmt.Println(markup)
	// Output:
	// <box direction="column"><text>Title</text>
	// <text>Subtitle</text>
	// <text>Content</text></box>
}

// ExampleHStack demonstrates creating a horizontal stack of elements
func ExampleHStack() {
	elem := lotus.HStack(
		lotus.Text("Left"),
		lotus.Text("Center"),
		lotus.Text("Right"),
	).Render()

	markup := elem.ToMarkup()
	fmt.Println(markup)
	// Output:
	// <box direction="row"><text>Left</text>
	// <text>Center</text>
	// <text>Right</text></box>
}

// ExampleNewTextInput demonstrates creating a text input component
func ExampleNewTextInput() {
	input := lotus.NewTextInput("username").
		WithWidth(30).
		WithPlaceholder("Enter your username")

	// Simulate user typing
	input.InsertChar("j")
	input.InsertChar("o")
	input.InsertChar("h")
	input.InsertChar("n")

	fmt.Println(input.Value)
	// Output: john
}

// ExampleText demonstrates creating styled text
func ExampleText() {
	elem := lotus.Text("Hello, World!")
	markup := elem.ToMarkup()
	fmt.Println(markup)
	// Output: <text>Hello, World!</text>
}
