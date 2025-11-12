package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// ChatApp - Simple chat UI using only primitives
type ChatApp struct {
	title    string
	messages []string // Store message lines
	input    *lotus.Input
}

// NewChatApp creates a new chat application
func NewChatApp() *ChatApp {
	app := &ChatApp{
		title:    "💬 Chat Example (Primitives Only)",
		messages: []string{"Welcome! Type a message below."},
	}

	// Setup input with inline handler
	app.input = lotus.NewInput().
		WithID("chat-input").
		WithPlaceholder("Say something...").
		WithOnSubmit(func(value string) {
			if value != "" {
				// Add user message
				app.messages = append(app.messages, fmt.Sprintf("> %s", value))

				// Echo response
				app.messages = append(app.messages, fmt.Sprintf("You said: %s", value))

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

	// VStack = flex-direction: column
	return lotus.VStack(
		// Header
		lotus.Box(lotus.Text(app.title)).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Messages (fills remaining space)
		lotus.Box(
			lotus.VStack(messageElements...),
		).
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
