package vdom

import "testing"

func TestTypedStyleMethods(t *testing.T) {
	t.Run("WithBorderStyle", func(t *testing.T) {
		elem := Box().WithBorderStyle(BorderStyleRounded)
		if elem.Props.Styles["border-style"] != "rounded" {
			t.Errorf("WithBorderStyle didn't set border-style correctly")
		}
	})

	t.Run("WithAlignSelf", func(t *testing.T) {
		elem := Box().WithAlignSelf(AlignSelfCenter)
		if elem.Props.Styles["align-self"] != "center" {
			t.Errorf("WithAlignSelf didn't set align-self correctly")
		}
	})

	t.Run("WithTextAlign", func(t *testing.T) {
		elem := Box().WithTextAlign(TextAlignCenter)
		if elem.Props.Styles["text-align"] != "center" {
			t.Errorf("WithTextAlign didn't set text-align correctly")
		}
	})

	t.Run("WithFlexGrow", func(t *testing.T) {
		elem := Box().WithFlexGrow(1)
		if elem.Props.Styles["flex-grow"] != "1" {
			t.Errorf("WithFlexGrow didn't set flex-grow correctly")
		}
	})

	t.Run("chained with markup", func(t *testing.T) {
		// Verify typed methods can chain with markup-parsed elements
		markupElem := Markup(`<box style="color: red">test</box>`)
		styledElem := markupElem.WithBorderStyle(BorderStyleRounded).WithFlexGrow(1)

		if styledElem.Props.Styles["color"] != "red" {
			t.Errorf("Markup style was lost after chaining typed methods")
		}
		if styledElem.Props.Styles["border-style"] != "rounded" {
			t.Errorf("WithBorderStyle didn't work after Markup")
		}
		if styledElem.Props.Styles["flex-grow"] != "1" {
			t.Errorf("WithFlexGrow didn't work after Markup")
		}
	})
}
