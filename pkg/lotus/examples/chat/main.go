package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// ChatApp - Simple chat UI using only primitives
type ChatApp struct {
	title      string
	messages   []string // Store message lines
	input      *lotus.Input
	scrollView *lotus.ScrollView
}

// NewChatApp creates a new chat application
func NewChatApp() *ChatApp {
	app := &ChatApp{
		title:    "💬 Chat Example (Primitives Only)",
		messages: []string{"Welcome! Type a message below.", "Try typing 'long' to test scrolling."},
	}

	// Setup scroll view with auto-scroll
	app.scrollView = lotus.NewScrollView().
		WithID("messages-scroll").
		WithAutoScroll(true)

	// Setup input with inline handler
	app.input = lotus.NewInput().
		WithID("chat-input").
		WithPlaceholder("Say something...").
		WithOnSubmit(func(value string) {
			if value != "" {
				// Add user message
				app.messages = append(app.messages, fmt.Sprintf("> %s", value))

				// Generate response
				if value == "long" {
					// Add many lines to test scrolling
					for i := 1; i <= 30; i++ {
						app.messages = append(app.messages, fmt.Sprintf("Line %d: This is a test line to verify scrolling works properly.", i))
					}
				} else {
					app.messages = append(app.messages, fmt.Sprintf("You said: %s", value))
				}

				app.input.Clear()
			}
		})

	return app
}

// Render - 3-panel layout: header, messages, input
func (app *ChatApp) Render() *lotus.Element {
	// Build message elements from strings
	messageElements := make([]any, len(app.messages))
	for i, msg := range app.messages {
		messageElements[i] = lotus.Text(msg)
	}

	// Update scroll view content
	app.scrollView.WithContent(lotus.VStack(messageElements...))

	// VStack = flex-direction: column
	return lotus.VStack(
		// Header
		lotus.Box(lotus.Text(app.title)).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Messages (fills remaining space with scrolling)
		lotus.Box(app.scrollView).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Input
		lotus.Box(app.input).
			WithBorderStyle(lotus.BorderStyleRounded),
	)
}

func main() {
	// ReactDOM.render(<ChatApp />)
	// DevTools + HMR auto-enabled via LOTUS_DEV=true env var
	if err := lotus.Run(NewChatApp()); err != nil {
		panic(err)
	}
}
