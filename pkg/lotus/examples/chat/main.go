package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

type ChatApp struct {
	messages []string
}

func NewChatApp() *ChatApp {
	return &ChatApp{
		messages: []string{
			"💬 Welcome to Lotus Chat!",
			"",
			"This demonstrates auto-scrolling chat UX:",
			"• Type messages and they appear at the bottom",
			"• New messages auto-scroll to stay visible",
			"• Use ↑↓ arrow keys to scroll through history",
			"• Scroll to bottom automatically when typing",
			"",
			"Try 'long' to add 50 messages and see auto-scroll in action!",
			"Try 'demo' to simulate a conversation.",
			"",
			"────────────────────────────────────────",
		},
	}
}

func (app *ChatApp) onSubmit(text string) {
	if text == "" {
		return
	}
	app.messages = append(app.messages, "\x1b[36m> "+text+"\x1b[0m") // Cyan user message

	switch text {
	case "long":
		// Add 50 messages to demonstrate auto-scroll
		for i := 1; i <= 50; i++ {
			app.messages = append(app.messages, fmt.Sprintf("[%02d] Auto-scroll test message - scroll up with ↑ to see earlier messages", i))
		}
		app.messages = append(app.messages, "")
		app.messages = append(app.messages, "\x1b[32m✓ Added 50 messages! Notice how it auto-scrolled to the bottom.\x1b[0m")
		app.messages = append(app.messages, "\x1b[32m  Try using ↑ arrow key to scroll up through history.\x1b[0m")
	case "demo":
		// Simulate a conversation
		conversation := []string{
			"\x1b[32mAssistant: Hello! How can I help you today?\x1b[0m",
			"",
			"\x1b[36m> I need help with scrolling\x1b[0m",
			"",
			"\x1b[32mAssistant: Great question! In Lotus, scrolling works automatically:\x1b[0m",
			"\x1b[32m• Messages stay at the bottom (like VS Code Chat)\x1b[0m",
			"\x1b[32m• New content auto-scrolls into view\x1b[0m",
			"\x1b[32m• Use arrow keys to scroll through history\x1b[0m",
			"",
			"\x1b[36m> That's exactly what I needed!\x1b[0m",
			"",
			"\x1b[32mAssistant: Perfect! Try the 'long' command to see it in action.\x1b[0m",
		}
		app.messages = append(app.messages, conversation...)
	default:
		app.messages = append(app.messages, "\x1b[32mEcho: "+text+"\x1b[0m") // Green echo
	}
}

func (app *ChatApp) Render() *lotus.Element {
	return lotus.VStack(
		// Header with instructions
		lotus.Box(
			lotus.VStack(
				lotus.Text("Lotus Chat - Auto-Scroll Demo").WithBold().WithColor("bright-cyan"),
				lotus.Text("Use ↑↓ to scroll • Try 'long' or 'demo' commands").WithColor("8"),
			).WithPaddingX(1),
		).WithBorderStyle(lotus.BorderStyleRounded),

		// Messages area - auto-scrolls to bottom, flex-grow enables overflow:auto
		lotus.Box(
			lotus.VStack(lotus.Map(app.messages, lotus.Text)...).
				WithPaddingX(1),
		).WithBorderStyle(lotus.BorderStyleRounded).WithFlexGrow(1),

		// Input
		lotus.Box(
			lotus.Input("Type a message (or 'long', 'demo')...", app.onSubmit),
		).WithBorderStyle(lotus.BorderStyleRounded),
	)
}

func main() {
	if err := lotus.Run(NewChatApp()); err != nil {
		panic(err)
	}
}
