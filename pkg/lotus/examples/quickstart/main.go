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
	if err := lotus.RunFunc(func(ctx lotus.Context) *lotus.Element {
		return lotus.VStack(
			lotus.Text("👋 Hello, Lotus!").WithBold().WithColor("bright-cyan"),
			lotus.Text("Press Ctrl+C to exit"),
		).WithGap(1)
	}); err != nil {
		panic(err)
	}
}
