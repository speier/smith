package main

import (
	"fmt"

	"github.com/charmbracelet/glamour"
	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/components"
)

// ChatApp - Simple 3-panel chat UI
type ChatApp struct {
	title    string
	messages *components.TextBox
	input    *components.TextInput
	renderer *glamour.TermRenderer
}

// NewChatApp creates a new chat application
func NewChatApp() *ChatApp {
	// Create markdown renderer (for later use)
	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(0),
		glamour.WithPreservedNewLines(),
	)

	app := &ChatApp{
		title:    "💬 Chat Example",
		messages: components.NewTextBox().WithAutoScroll(true),
		renderer: renderer,
	}

	// Add welcome message
	app.messages.AppendLine("assistant: Welcome! Type a message below.")

	// Setup input with inline handler
	app.input = components.NewTextInput().
		WithID("chat-input"). // ID required for HMR state persistence
		WithPlaceholder("Say something...").
		WithOnSubmit(func(value string) {
			if value != "" {
				// Add user message
				app.messages.AppendLine(fmt.Sprintf("user: %s", value))

				// Echo response
				response := "You said: " + value
				app.messages.AppendLine(fmt.Sprintf("assistant: %s", response))

				app.input.Clear()
			}
		})

	return app
}

// Render - 3-panel layout matching flexbox: header (auto), messages (flex-grow: 1), input (auto)
func (app *ChatApp) Render() *lotus.Element {
	// VStack = flex-direction: column
	return lotus.VStack(
		// Item 0: Header (auto height based on content + border)
		lotus.Box(app.title).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Item 1: Messages (flex-grow: 1, fills remaining space)
		lotus.Box(app.messages).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Item 2: Input (auto height based on content + border)
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
