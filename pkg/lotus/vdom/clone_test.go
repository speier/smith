package vdom

import "testing"

func TestElementClone(t *testing.T) {
	original := Box(Text("hello")).
		WithStyle("color", "red").
		WithStyle("padding", "2").
		WithFlexGrow(1).
		WithBorderStyle(BorderStyleRounded)

	cloned := original.Clone()

	// Verify clone is not the same object
	if original == cloned {
		t.Error("Clone returned same instance")
	}

	// Verify properties are copied
	if cloned.Tag != original.Tag {
		t.Errorf("Tag not cloned: got %q, want %q", cloned.Tag, original.Tag)
	}

	// Verify styles are copied
	if cloned.Props.Styles["color"] != "red" {
		t.Errorf("Style not cloned: color = %q", cloned.Props.Styles["color"])
	}
	if cloned.Props.Styles["padding"] != "2" {
		t.Errorf("Style not cloned: padding = %q", cloned.Props.Styles["padding"])
	}
}

func TestElementCloneModification(t *testing.T) {
	original := Box(Text("hello")).WithStyle("color", "red")
	cloned := original.Clone().WithStyle("color", "blue")

	// Verify original unchanged
	if original.Props.Styles["color"] != "red" {
		t.Errorf("Original modified: color = %q", original.Props.Styles["color"])
	}

	// Verify clone changed
	if cloned.Props.Styles["color"] != "blue" {
		t.Errorf("Clone not modified: color = %q", cloned.Props.Styles["color"])
	}
}

func TestVStackChildren(t *testing.T) {
	stack := VStack(
		Text("line1"),
		Text("line2"),
		Text("line3"),
	)

	if len(stack.Children) != 3 {
		t.Errorf("VStack children count = %d, want 3", len(stack.Children))
	}

	// Verify flex-direction
	if stack.Props.Styles["flex-direction"] != "column" {
		t.Errorf("VStack flex-direction = %q, want 'column'", stack.Props.Styles["flex-direction"])
	}
}

func TestHStackChildren(t *testing.T) {
	stack := HStack(
		Text("col1"),
		Text("col2"),
		Text("col3"),
	)

	if len(stack.Children) != 3 {
		t.Errorf("HStack children count = %d, want 3", len(stack.Children))
	}

	// Verify flex-direction
	if stack.Props.Styles["flex-direction"] != "row" {
		t.Errorf("HStack flex-direction = %q, want 'row'", stack.Props.Styles["flex-direction"])
	}
}

func TestBoxWithComplexChildren(t *testing.T) {
	box := Box(
		VStack(
			Text("nested1"),
			HStack(
				Text("deep1"),
				Text("deep2"),
			),
		),
	)

	if len(box.Children) != 1 {
		t.Errorf("Box children count = %d, want 1", len(box.Children))
	}

	vstack := box.Children[0]
	if len(vstack.Children) != 2 {
		t.Errorf("VStack children count = %d, want 2", len(vstack.Children))
	}

	hstack := vstack.Children[1]
	if len(hstack.Children) != 2 {
		t.Errorf("HStack children count = %d, want 2", len(hstack.Children))
	}
}

func TestWithFlexGrow(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"zero", 0, "0"},
		{"positive", 1, "1"},
		{"large", 10, "10"},
		{"negative clamped", -5, "0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elem := Box(Text("test")).WithFlexGrow(tt.input)
			if elem.Props.Styles["flex-grow"] != tt.expected {
				t.Errorf("flex-grow = %q, want %q", elem.Props.Styles["flex-grow"], tt.expected)
			}
		})
	}
}

func TestBorderStyles(t *testing.T) {
	tests := []struct {
		style    BorderStyle
		expected string
	}{
		{BorderStyleSingle, "single"},
		{BorderStyleRounded, "rounded"},
		{BorderStyleDouble, "double"},
		{BorderStyleNone, "none"},
	}

	for _, tt := range tests {
		t.Run(string(tt.style), func(t *testing.T) {
			elem := Box(Text("test")).WithBorderStyle(tt.style)
			if elem.Props.Styles["border-style"] != tt.expected {
				t.Errorf("border-style = %q, want %q", elem.Props.Styles["border-style"], tt.expected)
			}
		})
	}
}

func TestAlignSelf(t *testing.T) {
	tests := []struct {
		align    AlignSelf
		expected string
	}{
		{AlignSelfStretch, "stretch"},
		{AlignSelfFlexStart, "flex-start"},
		{AlignSelfFlexEnd, "flex-end"},
		{AlignSelfCenter, "center"},
	}

	for _, tt := range tests {
		t.Run(string(tt.align), func(t *testing.T) {
			elem := Box(Text("test")).WithAlignSelf(tt.align)
			if elem.Props.Styles["align-self"] != tt.expected {
				t.Errorf("align-self = %q, want %q", elem.Props.Styles["align-self"], tt.expected)
			}
		})
	}
}
