package components

import (
	"testing"
)

func TestTextInput_CursorStyles(t *testing.T) {
	tests := []struct {
		name        string
		style       CursorStyle
		expectStyle string
	}{
		{
			name:        "Block cursor",
			style:       CursorBlock,
			expectStyle: "background-color",
		},
		{
			name:        "Underline cursor",
			style:       CursorUnderline,
			expectStyle: "text-decoration",
		},
		{
			name:        "Bar cursor",
			style:       CursorBar,
			expectStyle: "bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := NewTextInput().
				WithPlaceholder("Test").
				WithCursorStyle(tt.style)
			input.Focused = true // Set focused to test cursor rendering

			// Verify cursor style is set
			if input.CursorStyle != tt.style {
				t.Errorf("Expected cursor style %v, got %v", tt.style, input.CursorStyle)
			}

			// Render and check structure
			elem := input.Render()
			if elem == nil {
				t.Fatal("Render returned nil")
			}

			hstack := elem.Children[0]
			if len(hstack.Children) < 2 {
				t.Fatalf("Expected at least 2 children, got %d", len(hstack.Children))
			}

			// Check cursor rendering based on style
			switch tt.style {
			case CursorBlock:
				// Block: should have background-color on cursor char
				cursorElem := hstack.Children[1]
				if cursorElem.Props.Styles["background-color"] == "" {
					t.Error("Block cursor missing background-color")
				}
			case CursorUnderline:
				// Underline: should have text-decoration
				cursorElem := hstack.Children[1]
				if cursorElem.Props.Styles["text-decoration"] != "underline" {
					t.Error("Underline cursor missing text-decoration")
				}
			case CursorBar:
				// Bar: should have "|" character
				barElem := hstack.Children[1]
				if barElem.Text != "|" {
					t.Errorf("Expected bar '|', got '%s'", barElem.Text)
				}
			}
		})
	}
}

func TestTextInput_CursorStyleWithText(t *testing.T) {
	input := NewTextInput().WithCursorStyle(CursorBlock)
	input.Focused = true // Set focused to test cursor rendering
	input.InsertChar("H")
	input.InsertChar("i")

	elem := input.Render()
	hstack := elem.Children[0]

	// Should have 3 children: prompt, "Hi", cursor (space) - no afterCursor since at end
	if len(hstack.Children) != 3 {
		t.Errorf("Expected 3 children, got %d", len(hstack.Children))
	}

	// Cursor char should have block styling
	cursorElem := hstack.Children[2]
	if cursorElem.Props.Styles["background-color"] != "#ffffff" {
		t.Errorf("Expected block cursor background, got '%s'", cursorElem.Props.Styles["background-color"])
	}
}

func TestTextInput_CursorStyleBar(t *testing.T) {
	input := NewTextInput().WithCursorStyle(CursorBar)
	input.Focused = true // Set focused to test cursor rendering
	input.InsertChar("H")
	input.InsertChar("i")

	elem := input.Render()
	hstack := elem.Children[0]

	// Should have 4 children for bar: prompt, "Hi", "|", cursor char (space) - no afterCursor
	if len(hstack.Children) != 4 {
		t.Errorf("Expected 4 children for bar cursor, got %d", len(hstack.Children))
	}

	// Third element should be the bar
	if hstack.Children[2].Text != "|" {
		t.Errorf("Expected '|', got '%s'", hstack.Children[2].Text)
	}
}
