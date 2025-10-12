package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/speier/smith/internal/cli"
	"github.com/speier/smith/internal/repl"
)

func main() {
	// If args provided and first arg is not a known command or flag, treat as initial prompt
	if len(os.Args) > 1 {
		first := os.Args[1]
		// Check if it's a known subcommand or flag
		if first != "exec" && !strings.HasPrefix(first, "-") {
			// Treat everything as initial prompt
			prompt := strings.Join(os.Args[1:], " ")
			r, err := repl.New(".")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			if err := r.Start(prompt); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	// Otherwise use normal CLI
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
