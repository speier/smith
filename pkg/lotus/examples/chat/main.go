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
			"Type a message below. Try 'long' to test scrolling.",
		},
	}
}

func (app *ChatApp) onSubmit(text string) {
	if text == "" {
		return
	}
	app.messages = append(app.messages, "> "+text)
	if text == "long" {
		for i := 1; i <= 50; i++ {
			app.messages = append(app.messages, fmt.Sprintf("[%02d] Auto-scroll test line", i))
		}
	} else {
		app.messages = append(app.messages, "Echo: "+text)
	}
}

func (app *ChatApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(
			lotus.Text("Lotus Chat").
				WithBold().
				WithColor("bright-cyan").
				WithPaddingX(1),
		).WithBorderStyle(lotus.BorderStyleRounded),

		lotus.Box(
			lotus.VStack(lotus.Map(app.messages, lotus.Text)...).
				WithGap(1).
				WithPaddingX(1),
		).WithBorderStyle(lotus.BorderStyleRounded).WithFlexGrow(1),

		lotus.Box(
			lotus.Input("Type a message...", app.onSubmit),
		).WithBorderStyle(lotus.BorderStyleRounded),
	)
}

func main() {
	if err := lotus.Run(NewChatApp()); err != nil {
		panic(err)
	}
}
