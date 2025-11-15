package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/speier/smith/pkg/lotus"
)

// ChatApp demonstrates auto-scrolling and streaming text in a chat UI
type ChatApp struct {
	messages []string
}

func NewChatApp() *ChatApp {
	app := &ChatApp{
		messages: []string{
			"💬 Lotus Chat Demo",
			"",
			"Type /help to see available commands",
			"",
		},
	}

	// Register commands globally - they work in all inputs
	lotus.RegisterGlobalCommand("long", "Add 50 messages to test auto-scroll", func(ctx lotus.Context, args []string) {
		for i := 1; i <= 50; i++ {
			app.messages = append(app.messages, fmt.Sprintf("[%02d] Message - scroll with ↑↓", i))
		}
		app.messages = append(app.messages, "")
		app.addMessage(assistantPrefix, "✓ Auto-scrolled to bottom")
	})

	lotus.RegisterGlobalCommand("stream", "Stream text word-by-word like an LLM", func(ctx lotus.Context, args []string) {
		app.startStreaming(ctx, "This is a streaming response! Watch as text appears word by word, simulating how an LLM streams tokens in real-time. The text automatically wraps to fit the window width and auto-scrolls as new content arrives. Pretty cool, right?")
	})

	lotus.RegisterGlobalCommand("wrap", "Show text wrapping with a long message", func(ctx lotus.Context, args []string) {
		app.addMessage(systemPrefix, "This is a very long message that demonstrates automatic text wrapping in Lotus. When text exceeds the window width, it wraps to multiple lines at word boundaries. ANSI colors are preserved, and the layout engine correctly calculates the height needed for wrapped content.")
	})

	return app
}

// Message prefixes
const (
	userPrefix      = "\x1b[36m> %s\x1b[0m"
	assistantPrefix = "\x1b[32m%s\x1b[0m"
	systemPrefix    = "\x1b[33m%s\x1b[0m"
)

// addMessage adds a message with the specified color prefix
func (app *ChatApp) addMessage(prefix, text string) {
	app.messages = append(app.messages, fmt.Sprintf(prefix, text))
}

func (app *ChatApp) startStreaming(ctx lotus.Context, fullText string) {
	// Add initial message with cursor
	app.messages = append(app.messages, "\x1b[32m▌\x1b[0m")
	msgIndex := len(app.messages) - 1

	go func() {
		var current string
		for i, word := range strings.Fields(fullText) {
			time.Sleep(50 * time.Millisecond)
			if i > 0 {
				current += " "
			}
			current += word
			// Update message in-place with cursor
			app.messages[msgIndex] = fmt.Sprintf("\x1b[32m%s ▌\x1b[0m", current)
			ctx.Update() // Use context from parameter
		}
		// Final message without cursor
		app.messages[msgIndex] = fmt.Sprintf("\x1b[32m%s\x1b[0m", current)
		ctx.Update() // Use context from parameter
	}()
}

func (app *ChatApp) onSubmit(ctx lotus.Context, text string) {
	if text == "" {
		return
	}

	app.addMessage(userPrefix, text)
	app.addMessage(assistantPrefix, "Roger that!")
}

func (app *ChatApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(
			lotus.Text("Lotus Chat").WithBold().WithColor("bright-cyan"),
		).WithBorderStyle(lotus.BorderStyleRounded),

		lotus.Box(
			lotus.VStack(app.messages),
		).WithBorderStyle(lotus.BorderStyleRounded).WithFlexGrow(1),

		lotus.Box(
			lotus.Input("Type a message...", app.onSubmit),
		).WithBorderStyle(lotus.BorderStyleRounded),
	)
}

func main() {
	if err := lotus.Run(NewChatApp); err != nil {
		panic(err)
	}
}
