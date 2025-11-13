package main

import (
	"github.com/speier/smith/pkg/lotus"
)

// Quickstart - Your first Lotus app in 10 lines
// Demonstrates:
// - Functional components
// - Declarative UI (like React)
// - Superior rendering performance

func main() {
	_ = lotus.Run(func(ctx lotus.AppContext) *lotus.Element {
		return lotus.VStack(
			lotus.Text("👋 Hello, Lotus!").WithBold().WithColor("bright-cyan"),
			lotus.Text(""),
			lotus.Text("Press Ctrl+C to exit"),
		)
	})
}
