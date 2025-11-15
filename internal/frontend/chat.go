package frontend

import (
	"fmt"
	"strings"

	"github.com/speier/smith/pkg/agent/session"
	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotusui"
)

// ChatUI represents the main chat interface (React-like component)
type ChatUI struct {
	session        session.Session
	input          *lotus.InputComponent
	messageList    *MessageList
	renderCallback func() // Callback to trigger re-renders from async operations (streaming)
	commands       *lotus.CommandRegistry
	modal          *lotusui.Modal
}

// NewChatUI creates a new chat application
func NewChatUI(sess session.Session) *ChatUI {
	app := &ChatUI{
		session:     sess,
		messageList: NewMessageList(),
		commands:    lotus.NewCommandRegistry(),
	}

	// Register slash commands
	app.registerCommands()

	// Setup input with inline handler
	app.input = lotus.CreateInput("Type your message...", func(value string) {
		if value != "" {
			app.handleSubmit(value)
		}
	})

	// Load existing history
	app.loadHistory()

	return app
}

// SetRenderCallback sets the callback to trigger re-renders (called by Lotus runtime)
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
	// Split logo lines for proper rendering
	logoLines := strings.Split(GetLogoLines(), "\n")
	logoElements := make([]any, len(logoLines))
	for i, line := range logoLines {
		logoElements[i] = lotus.Text(line).WithTextAlign(lotus.TextAlignCenter)
	}

	// Split welcome text for proper rendering
	welcomeLines := strings.Split(GetWelcomeText(), "\n")
	welcomeElements := make([]any, len(welcomeLines))
	for i, line := range welcomeLines {
		welcomeElements[i] = lotus.Text(line).WithTextAlign(lotus.TextAlignCenter)
	}

	return lotus.VStack(
		// Logo section
		lotus.VStack(logoElements...).WithAlignItems(lotus.AlignItemsCenter),
		// Welcome section
		lotus.VStack(welcomeElements...).WithAlignItems(lotus.AlignItemsCenter),
	).
		WithAlignItems(lotus.AlignItemsCenter).
		WithGap(1).     // gap: 1rem - space between logo and welcome
		WithPaddingY(1) // padding: 1rem 0 - space before first and after last child
}

// handleSubmit sends the message to the agent or executes slash command
func (app *ChatUI) handleSubmit(message string) {
	// Check for slash command
	if strings.HasPrefix(message, "/") {
		if app.commands.Execute(message) {
			// Command executed, clear input
			app.input.Clear()
			return
		}
		// Unknown command - show error
		app.messageList.AddMessage("system", fmt.Sprintf("Unknown command: %s\nType /help for available commands", message))
		app.input.Clear()
		return
	}

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
			// Async update needs manual re-render trigger
			if app.renderCallback != nil {
				app.renderCallback()
			}
			return
		}

		// Stream response with live updates
		var responseBuf string
		for chunk := range stream {
			responseBuf += chunk
			// Update streaming buffer and trigger re-render (async operation)
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
		// Final re-render after streaming completes
		if app.renderCallback != nil {
			app.renderCallback()
		}
	}()
}

// Render - 3-panel layout: header, messages, input (React render pattern)
func (app *ChatUI) Render() *lotus.Element {
	content := lotus.VStack(
		// Messages (fills remaining space with scrolling)
		lotus.Box(app.messageList.Render()).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),

		// Input
		lotus.Box(app.input).
			WithBorderStyle(lotus.BorderStyleRounded),
	)

	// If modal is open, render it on top
	if app.modal != nil && app.modal.Open {
		return lotus.VStack(
			content,
			app.modal.Render(),
		)
	}

	return content
}

// registerCommands registers all available slash commands
func (app *ChatUI) registerCommands() {
	// /help - Show available commands
	app.commands.Register(&lotus.Command{
		Name:        "help",
		Description: "Show available commands",
		Handler: func(args []string) {
			helpText := "Available commands:\n"
			for _, cmd := range app.commands.List() {
				helpText += fmt.Sprintf("  /%s - %s\n", cmd.Name, cmd.Description)
			}
			app.messageList.AddMessage("system", helpText)
		},
	})

	// /clear - Clear conversation
	app.commands.Register(&lotus.Command{
		Name:        "clear",
		Description: "Clear conversation history",
		Aliases:     []string{"cls"},
		Handler: func(args []string) {
			app.showClearConfirmation()
		},
	})

	// /model - Change model (placeholder)
	app.commands.Register(&lotus.Command{
		Name:        "model",
		Description: "Change LLM model",
		Handler: func(args []string) {
			app.showModelPicker(args)
		},
	})
}

// showClearConfirmation shows a modal to confirm clearing conversation
func (app *ChatUI) showClearConfirmation() {
	app.modal = lotusui.NewModal().
		WithTitle("Clear Conversation").
		WithContent(lotus.Text("Are you sure you want to clear all messages?\nThis cannot be undone.")).
		WithButtons([]lotusui.ModalButton{
			{
				Label:   "Clear",
				Variant: "danger",
				OnClick: func() {
					app.messageList.Clear()
					app.messageList.SetHeader(app.buildHeaderV2())
					app.modal.Close()
				},
			},
			{
				Label:   "Cancel",
				Variant: "secondary",
				OnClick: func() {
					app.modal.Close()
				},
			},
		}).
		WithOnClose(func() {
			// Modal closed
		})

	app.modal.Show()
}

// showModelPicker shows a modal to select a model
func (app *ChatUI) showModelPicker(args []string) {
	// If model specified in args, use it directly
	if len(args) > 0 {
		modelName := strings.Join(args, " ")
		app.messageList.AddMessage("system", fmt.Sprintf("Model changed to: %s\n(Model switching not yet implemented)", modelName))
		return
	}

	// Otherwise show picker modal
	app.modal = lotusui.NewModal().
		WithTitle("Select Model").
		WithContent(lotus.VStack(
			lotus.Text("Available models:"),
			lotus.Text("• GPT-4"),
			lotus.Text("• GPT-3.5-turbo"),
			lotus.Text("• Claude-3-opus"),
			lotus.Text(""),
			lotus.Text("(Model selection UI coming soon)"),
		)).
		WithButtons([]lotusui.ModalButton{
			{
				Label:   "Close",
				Variant: "primary",
				OnClick: func() {
					app.modal.Close()
				},
			},
		}).
		WithOnClose(func() {
			// Modal closed
		})

	app.modal.Show()
}
