package core

import (
	"strings"
	"testing"
)

func TestBoxBuilder(t *testing.T) {
	box := NewBox().
		ID("testbox").
		Class("test-class").
		Direction(Column).
		Flex(1).
		Width("100").
		Height("50").
		Color("#0f0").
		Children(
			"<text>Child 1</text>",
			"<text>Child 2</text>",
		)

	result := box.ToMarkup()

	tests := []string{
		`id="testbox"`,
		`class="test-class"`,
		`direction="column"`,
		`flex="1"`,
		`width="100"`,
		`height="50"`,
		`color="#0f0"`,
		"Child 1",
		"Child 2",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("BoxBuilder result missing: %s", test)
		}
	}
}

func TestBoxBuilderNested(t *testing.T) {
	inner := NewBox().
		ID("inner").
		Children("<text>Nested</text>")

	outer := NewBox().
		ID("outer").
		Children(inner)

	result := outer.ToMarkup()

	if !strings.Contains(result, `id="outer"`) {
		t.Error("Should contain outer box")
	}
	if !strings.Contains(result, `id="inner"`) {
		t.Error("Should contain nested box")
	}
	if !strings.Contains(result, "Nested") {
		t.Error("Should contain nested content")
	}
}

func TestTextBuilder(t *testing.T) {
	text := NewText("Hello").
		Color("#ff0000").
		Background("#000000").
		Bold()

	result := text.ToMarkup()

	tests := []string{
		`color="#ff0000"`,
		`background="#000000"`,
		`bold="true"`,
		"Hello",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("TextBuilder result missing: %s", test)
		}
	}
}

func TestBuilderWithComponent(t *testing.T) {
	// MockComponent for testing component integration
	type testComponent struct {
		content string
	}
	// Implement Component interface
	impl := func(tc *testComponent) Component {
		return componentFunc(func() string {
			return "<text>" + tc.content + "</text>"
		})
	}

	comp := impl(&testComponent{content: "Component Content"})

	box := NewBox().
		Children(
			"<text>Text</text>",
			comp,
		)

	result := box.ToMarkup()

	if !strings.Contains(result, "Text") {
		t.Error("Should contain text child")
	}
	if !strings.Contains(result, "Component Content") {
		t.Error("Should contain component content")
	}
}

// componentFunc is a helper to create Component from a function
type componentFunc func() string

func (f componentFunc) Render() string {
	return f()
}

func TestMixedAPI(t *testing.T) {
	// MockComponent for testing mixed API
	comp := componentFunc(func() string {
		return "<text>From Component</text>"
	})

	builder := NewBox().
		ID("mixed").
		Children(
			`<box direction="column"><text>From Helper</text>
Raw string</box>`,
			comp,
		)

	result := builder.ToMarkup()

	tests := []string{
		`id="mixed"`,
		"From Helper",
		"Raw string",
		"From Component",
	}

	for _, test := range tests {
		if !strings.Contains(result, test) {
			t.Errorf("Mixed API result missing: %s", test)
		}
	}
}
