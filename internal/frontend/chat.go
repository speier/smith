package frontend

import (
	"fmt"

	"github.com/speier/smith/pkg/agent/session"
	"github.com/speier/smith/pkg/lotus"
)

// ChatUI represents the main chat interface (React-like component)
type ChatUI struct {
	session        session.Session
	input          *lotus.Input
	messageList    *MessageList
	renderCallback func() // Callback to trigger re-renders from async operations
}

// NewChatUI creates a new chat application
func NewChatUI(sess session.Session) *ChatUI {
	app := &ChatUI{
		session:     sess,
		messageList: NewMessageList(),
	}

	// Setup input with inline handler
	app.input = lotus.NewInput().
		WithID("chat-input").
		WithPlaceholder("Type your message...").
		WithOnSubmit(func(value string) {
			if value != "" {
				app.handleSubmit(value)
			}
		})

	// Load existing history
	app.loadHistory()

	return app
}

// SetRenderCallback sets the callback to trigger re-renders
func (app *ChatUI) SetRenderCallback(cb func()) {
	app.renderCallback = cb
}

// loadHistory loads existing messages from session
func (app *ChatUI) loadHistory() {
	history := app.session.GetHistory()

	// If no history, set welcome banner header
	if len(history) == 0 {
		app.messageList.SetHeader(app.buildHeaderV2())
		return
	}

	// Add history messages
	for _, msg := range history {
		app.messageList.AddMessage(msg.Role, msg.Content)
	}
}

// // buildHeaderV1 - Original approach with empty Text lines for spacing
// func (app *ChatUI) buildHeaderV1() *lotus.Element {
// 	return lotus.VStack(
// 		lotus.Text(""), // Empty line
// 		lotus.Text(GetLogoLines()).WithTextAlign(lotus.TextAlignCenter),
// 		lotus.Text(""), // Empty line
// 		lotus.Text(GetWelcomeText()).WithTextAlign(lotus.TextAlignCenter),
// 		lotus.Text(""), // Empty line
// 	).WithAlignItems(lotus.AlignItemsCenter)
// }

// buildHeaderV2 - CSS flexbox approach with gap and padding
func (app *ChatUI) buildHeaderV2() *lotus.Element {
	return lotus.VStack(
		lotus.Text(GetLogoLines()).WithTextAlign(lotus.TextAlignCenter),
		lotus.Text(GetWelcomeText()).WithTextAlign(lotus.TextAlignCenter),
	).
		WithAlignItems(lotus.AlignItemsCenter).
		WithGap(1).     // gap: 1rem - space between logo and welcome
		WithPaddingY(1) // padding: 1rem 0 - space before first and after last child
}

// handleSubmit sends the message to the agent
func (app *ChatUI) handleSubmit(message string) {
	// Add user message immediately
	app.messageList.AddMessage("user", message)

	// Clear input
	app.input.Clear()

	// Show "thinking" indicator
	app.messageList.SetStreaming(true, "")

	// Send message asynchronously
	go func() {
		stream, err := app.session.SendMessage(message)
		if err != nil {
			app.messageList.SetStreaming(false, "")
			app.messageList.AddMessage("system", fmt.Sprintf("Error: %v", err))
			if app.renderCallback != nil {
				app.renderCallback()
			}
			return
		}

		// Stream response with live updates
		var responseBuf string
		for chunk := range stream {
			responseBuf += chunk
			// Update streaming buffer and trigger re-render
			app.messageList.SetStreaming(true, responseBuf)
			if app.renderCallback != nil {
				app.renderCallback()
			}
		}

		// Add complete assistant response
		app.messageList.SetStreaming(false, "")
		if responseBuf != "" {
			app.messageList.AddMessage("assistant", responseBuf)
		}
		if app.renderCallback != nil {
			app.renderCallback()
		}
	}()
}

// Render - 3-panel layout: header, messages, input (React render pattern)
func (app *ChatUI) Render() *lotus.Element {
	return lotus.VStack(
		// Header - SMITH ASCII logo + version
		// lotus.Box(lotus.Text("💬 SMITH - Multi-Agent System")).
		// 	WithBorderStyle(lotus.BorderStyleRounded).
		// 	WithColor("10"), // Bright green

		// Messages (fills remaining space with scrolling)
		lotus.Box(
			lotus.NewScrollView().
				WithID("messages-scroll").
				WithContent(app.messageList.Render()),
		).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Input
		lotus.Box(app.input).
			WithBorderStyle(lotus.BorderStyleRounded),
	)
}
