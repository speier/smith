package lotus

import (
	"strings"
	"testing"
)

func TestVStack(t *testing.T) {
	result := VStack(
		Text("Hello"),
		Text("World"),
	)

	if !strings.Contains(result, `direction="column"`) {
		t.Error("VStack should create column direction")
	}
	if !strings.Contains(result, "Hello") || !strings.Contains(result, "World") {
		t.Error("VStack should contain children")
	}
}

func TestHStack(t *testing.T) {
	result := HStack(
		Text("Left"),
		Text("Right"),
	)

	if !strings.Contains(result, `direction="row"`) {
		t.Error("HStack should create row direction")
	}
	if !strings.Contains(result, "Left") || !strings.Contains(result, "Right") {
		t.Error("HStack should contain children")
	}
}

func TestText(t *testing.T) {
	result := Text("Hello World")

	if !strings.Contains(result, "<text>") {
		t.Error("Text should create text element")
	}
	if !strings.Contains(result, "Hello World") {
		t.Error("Text should contain content")
	}
}

func TestInput(t *testing.T) {
	result := Input("user input")

	if !strings.Contains(result, "<input>") {
		t.Error("Input should create input element")
	}
	if !strings.Contains(result, "user input") {
		t.Error("Input should contain value")
	}
}

func TestBoxWithID(t *testing.T) {
	result := BoxWithID("mybox", Text("content"))

	if !strings.Contains(result, `id="mybox"`) {
		t.Error("BoxWithID should set ID attribute")
	}
	if !strings.Contains(result, "content") {
		t.Error("BoxWithID should contain children")
	}
}

func TestBoxWithClass(t *testing.T) {
	result := BoxWithClass("myclass", Text("content"))

	if !strings.Contains(result, `class="myclass"`) {
		t.Error("BoxWithClass should set class attribute")
	}
	if !strings.Contains(result, "content") {
		t.Error("BoxWithClass should contain children")
	}
}

func TestSpacer(t *testing.T) {
	result := Spacer()

	if !strings.Contains(result, `flex="1"`) {
		t.Error("Spacer should have flex=1")
	}
}
