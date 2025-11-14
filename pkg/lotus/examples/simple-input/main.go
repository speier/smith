package main

import "github.com/speier/smith/pkg/lotus"

type SimpleTest struct {
	message string
}

func NewSimpleTest() *SimpleTest {
	return &SimpleTest{}
}

func (app *SimpleTest) onSubmit(text string) {
	app.message = "You typed: " + text
}

func (app *SimpleTest) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Text("Simple Input Test").WithBold(),
		lotus.Text(app.message),
		lotus.Input("Type something", app.onSubmit),
	)
}

func main() {
	if err := lotus.Run(NewSimpleTest()); err != nil {
		panic(err)
	}
}
