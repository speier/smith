package main

import (
	"github.com/speier/smith/pkg/lotus"
)

// ChatApp - React functional component (hooks pattern)
func ChatApp() lotus.RenderFunc {
	// Setup components (like useState/useRef hooks)
	messageList := lotus.NewMessageList("messages")
	input := lotus.NewTextInput("input-text")

	// Add initial messages
	messageList.AddMessage("assistant", "Welcome! Type a message and press Enter.")
	messageList.AddMessage("assistant", "This is a functional component (React hooks pattern)!")

	// Setup input handler (like useCallback)
	input.WithPlaceholder("Type a message...").
		WithOnSubmit(func(value string) {
			if value != "" {
				messageList.AddMessage("user", value)
				messageList.AddMessage("assistant", "Echo: "+value)
				input.Clear()
			}
		})

	// Return render function (like React functional component)
	return func() *lotus.Element {
		return lotus.VStack(
			// Header
			lotus.PanelBox("header",
				lotus.Text("💬 Functional Chat (React Hooks Pattern)"),
			).Height(3),

			// Messages
			lotus.PanelBox("messages",
				lotus.NewComponentElement(messageList),
			),

			// Input
			lotus.PanelBox("input-box",
				lotus.NewComponentElement(input),
			).Height(3),
		).Render()
	}
}

func main() {
	// ReactDOM.render(<ChatApp />, terminal)
	// DevTools + HMR auto-enabled via LOTUS_DEV=true env var
	_ = lotus.Run("app", ChatApp())
}
