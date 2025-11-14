package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// FormApp tests multiple inputs with Tab navigation
type FormApp struct {
	name    string
	email   string
	message string
	results []string
}

func NewFormApp() *FormApp {
	return &FormApp{
		results: []string{
			"📝 Form Test - Multiple Inputs",
			"",
			"Instructions:",
			"- Type in any field",
			"- Press Tab to move between fields",
			"- Press Enter to submit current field",
			"",
		},
	}
}

func (app *FormApp) onNameSubmit(text string) {
	app.name = text
	app.results = append(app.results, fmt.Sprintf("Name: %s", text))
}

func (app *FormApp) onEmailSubmit(text string) {
	app.email = text
	app.results = append(app.results, fmt.Sprintf("Email: %s", text))
}

func (app *FormApp) onMessageSubmit(text string) {
	app.message = text
	app.results = append(app.results, fmt.Sprintf("Message: %s", text))
}

// in the form app we might showcase input types like name text, password password, age number, email validation

func (app *FormApp) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(
			lotus.Text("Multi-Input Form Test").
				WithBold().
				WithColor("bright-cyan"),
		).WithBorderStyle(lotus.BorderStyleRounded).
			WithPaddingX(1),

		lotus.VStack(
			lotus.Text("Name:").WithBold(),
			lotus.Input("Enter your name", app.onNameSubmit),
			lotus.Text("Email:").WithBold(),
			lotus.Input("Enter your email", app.onEmailSubmit),
			lotus.Text("Message:").WithBold(),
			lotus.Input("Enter a message", app.onMessageSubmit),
		).WithGap(1).WithPaddingX(1).WithFlexGrow(1),

		lotus.Box(
			lotus.VStack(lotus.Map(app.results, lotus.Text)...).
				WithGap(0).
				WithPaddingX(1),
		).WithBorderStyle(lotus.BorderStyleRounded).
			WithPaddingY(1),

		lotus.Box(
			lotus.Text("Press Tab to navigate • Enter to submit • Ctrl+C to quit").
				WithColor("bright-black"),
		).WithPaddingX(1),
	)
}

func main() {
	if err := lotus.Run(NewFormApp()); err != nil {
		panic(err)
	}
}
