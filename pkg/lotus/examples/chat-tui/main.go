package main

import (
	"github.com/speier/smith/pkg/lotus"
)

// ChatApp - React functional component
type ChatApp struct {
	messageList *lotus.MessageList // useState with MessageList for scrolling
	input       *lotus.TextInput   // useRef()
}

func NewChatApp() *ChatApp {
	app := &ChatApp{
		messageList: lotus.NewMessageList("messages"), // No hardcoded sizes!
	}

	// Add initial messages
	app.messageList.AddMessage("assistant", "Welcome! Type a message and press Enter.")
	app.messageList.AddMessage("assistant", "Messages will appear here with auto-scroll.")

	// Setup input with inline handler (like React)
	app.input = lotus.NewTextInput("input-text").
		WithPlaceholder("Type a message...").
		WithOnSubmit(func(value string) {
			if value != "" {
				// Add messages (component handles scrolling)
				app.messageList.AddMessage("user", value)
				app.messageList.AddMessage("assistant", "Echo: "+value)
				app.input.Clear()
			}
		})

	return app
}

// Render - JSX-like declarative UI
func (app *ChatApp) Render() *lotus.Element {
	// JSX-like tree structure
	return lotus.VStack(
		// Header
		lotus.PanelBox("header",
			lotus.Text("💬 Chat TUI - Press Ctrl+C to exit"),
		).Height(3).BorderStyle("rounded").BorderColor("blue").Padding(0),

		// Messages with auto-scrolling (full width)
		lotus.Box("messages-container",
			lotus.NewComponentElement(app.messageList),
		).Flex(1).Border("1px solid").Padding(0).Color("#ddd"),

		// Input
		lotus.PanelBox("input-box",
			lotus.NewComponentElement(app.input),
		).Height(3),
	).Render()
}

func main() {
	// ReactDOM.render(<ChatApp />, terminal)
	// DevTools + HMR auto-enabled via LOTUS_DEV=true env var
	_ = lotus.Run("app", NewChatApp())
}
