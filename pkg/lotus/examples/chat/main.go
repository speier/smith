package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// Minimal chat UI - demonstrates Lotus's outstanding DX
// Build a production-ready chat in ~30 lines with:
// - Functional components
// - Auto-scrolling message history
// - Responsive layout
// - Superior rendering performance

func main() {
	err := lotus.Run(func(ctx lotus.AppContext) *lotus.Element {
		// State lives in closure - simple and clean
		messages := []string{
			"💬 Welcome to Lotus Chat!",
			"Type a message below. Try 'long' to test scrolling.",
		}

		// Build message list
		messageElements := make([]any, len(messages))
		for i, msg := range messages {
			messageElements[i] = lotus.Text(msg)
		}

		// 3-panel layout: header, scrollable messages, input
		return lotus.VStack(
			// Header
			lotus.Box(
				lotus.Text("Lotus Chat").
					WithBold().
					WithColor("bright-cyan"),
			).WithBorderStyle(lotus.BorderStyleRounded),

			// Messages with auto-scroll
			lotus.Box(
				lotus.NewScrollView().
					WithAutoScroll(true).
					WithContent(lotus.VStack(messageElements...)),
			).
				WithFlexGrow(1).
				WithBorderStyle(lotus.BorderStyleRounded),

			// Input - simplified API
			lotus.Box(
				lotus.CreateInput("Type a message...", func(text string) {
					if text == "" {
						return
					}

					// Add user message
					messages = append(messages, fmt.Sprintf("> %s", text))

					// Simulate response
					if text == "long" {
						// Stress test: 50 lines to verify auto-scroll
						for i := 1; i <= 50; i++ {
							messages = append(messages, fmt.Sprintf("[%02d] Auto-scroll test line", i))
						}
					} else {
						messages = append(messages, fmt.Sprintf("Echo: %s", text))
					}

					// Trigger re-render to show new messages
					ctx.Rerender()
				}),
			).WithBorderStyle(lotus.BorderStyleRounded),
		)
	})

	if err != nil {
		panic(err)
	}
}
