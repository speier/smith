package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/tty"
)

type ScrollTestApp struct {
	items         []string
	count         int
	renderRequest func()
}

func NewScrollTestApp() *ScrollTestApp {
	return &ScrollTestApp{
		items: []string{
			"Initial item 1",
			"Initial item 2",
		},
		count: 2,
	}
}

func (app *ScrollTestApp) SetRenderCallback(cb func()) {
	app.renderRequest = cb
}

func (app *ScrollTestApp) addItems() {
	for i := 0; i < 10; i++ {
		app.count++
		app.items = append(app.items, fmt.Sprintf("Item %d", app.count))
	}
}

func (app *ScrollTestApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Text(fmt.Sprintf("Total items: %d", len(app.items))).
			WithBold().
			WithColor("bright-cyan"),

		lotus.Box(
			lotus.VStack(lotus.Map(app.items, lotus.Text)...).
				WithGap(1).
				WithPaddingX(1),
		).
			WithBorderStyle(lotus.BorderStyleRounded).
			WithFlexGrow(1),

		lotus.Text("Press 'a' to add 10 items, Ctrl+C to quit").
			WithColor("yellow"),
	)
}

func (app *ScrollTestApp) HandleKeyEvent(event tty.KeyEvent) bool {
	if event.Key == 'a' || event.Key == 'A' {
		fmt.Printf("Adding 10 items... (total will be %d)\n", len(app.items)+10)
		app.addItems()
		if app.renderRequest != nil {
			app.renderRequest()
		}
		return true
	}
	return false
}

func main() {
	app := NewScrollTestApp()
	if err := lotus.Run(app); err != nil {
		panic(err)
	}
}
