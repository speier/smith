package main

import (
	"github.com/speier/smith/pkg/lotus"
)

// Demo showcasing practical TUI styling features
func main() {
	app := lotus.VStack(
		// Header with bold text
		lotus.Text("🎨 Lotus TUI Styling Demo").
			WithBold().
			WithTextAlign(lotus.TextAlignCenter).
			WithColor("#ffff00"), // yellow

		lotus.Text(""), // spacer

		// Text styling
		lotus.Text("Regular Text"),
		lotus.Text("Bold Text").WithBold(),
		lotus.Text("Italic Text").WithItalic(),
		lotus.Text("Underlined Text").WithUnderline(),
		lotus.Text("Strikethrough Text").WithStrikethrough(),
		lotus.Text("Dimmed Text (50% opacity)").WithDim(),

		lotus.Text(""), // spacer

		// Combined styles
		lotus.Text("Bold + Underline").WithBold().WithUnderline(),
		lotus.Text("Bold + Italic + Underline").
			WithBold().
			WithItalic().
			WithUnderline().
			WithColor("#00ff00"), // green

		lotus.Text(""), // spacer

		// Reverse video (for selections)
		lotus.Text(" SELECTED ").WithReverse(),

		lotus.Text(""), // spacer

		// Text overflow
		lotus.VStack(
			lotus.Text("Text Overflow Ellipsis:"),
			lotus.Box(
				lotus.Text("This is a very long text that will be truncated with ellipsis").
					WithTextOverflow("ellipsis"),
			).WithWidth(30).WithBorderStyle(lotus.BorderStyleSingle),
		),

		lotus.Text(""), // spacer

		// Border colors
		lotus.Text("Border Colors:"),
		lotus.HStack(
			lotus.Box(lotus.Text("Red").WithTextAlign(lotus.TextAlignCenter)).
				WithWidth(12).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#ff0000"),
			lotus.Box(lotus.Text("Green").WithTextAlign(lotus.TextAlignCenter)).
				WithWidth(12).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#00ff00"),
			lotus.Box(lotus.Text("Cyan").WithTextAlign(lotus.TextAlignCenter)).
				WithWidth(12).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#00ffff"),
			lotus.Box(lotus.Text("Yellow").WithTextAlign(lotus.TextAlignCenter)).
				WithWidth(12).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#ffff00"),
		).WithGap(1),

		lotus.Text(""), // spacer

		// Line clamping (MaxLines)
		lotus.Text("Line Clamping (MaxLines):"),
		lotus.HStack(
			// Preview card: 3 lines max
			lotus.Box(
				lotus.VStack(
					lotus.Text("Task #1").WithBold(),
					lotus.Text("Implement authentication\nAdd user login\nCreate session management\nHash passwords\nAdd OAuth support").
						WithMaxLines(3),
				).WithGap(0),
			).
				WithWidth(25).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#00ff00"),

			// File preview: 4 lines max
			lotus.Box(
				lotus.VStack(
					lotus.Text("config.yaml").WithBold(),
					lotus.Text("server:\n  port: 8080\n  host: localhost\ndatabase:\n  url: postgres://...\n  pool: 10").
						WithMaxLines(4),
				).WithGap(0),
			).
				WithWidth(25).
				WithBorderStyle(lotus.BorderStyleSingle).
				WithBorderColor("#00ffff"),
		).WithGap(1),

		lotus.Text(""), // spacer

		// Help text
		lotus.Text("Press Ctrl+C to exit").
			WithDim().
			WithTextAlign(lotus.TextAlignCenter),
	).
		WithGap(0).
		WithPaddingY(1).
		WithAlignItems(lotus.AlignItemsCenter)

	if err := lotus.RunElement(app); err != nil {
		panic(err)
	}
}
