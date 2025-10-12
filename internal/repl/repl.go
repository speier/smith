package repl

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// REPL is a CLI frontend for the Smith engine
type REPL struct {
	model       *BubbleModel
	projectPath string
}

func New(projectPath string) (*REPL, error) {
	model, err := NewBubbleModel(projectPath, "")
	if err != nil {
		return nil, err
	}

	return &REPL{
		model:       model,
		projectPath: projectPath,
	}, nil
}

func (r *REPL) Start(initialPrompt string) error {
	// If initial prompt provided, update the model
	if initialPrompt != "" {
		model, err := NewBubbleModel(r.projectPath, initialPrompt)
		if err != nil {
			return err
		}
		r.model = model
	}

	p := tea.NewProgram(
		r.model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Note: Bubble Tea catches SIGINT (Ctrl+C) by default and converts it to a KeyMsg
	// Our Update handler catches tea.KeyCtrlC to show goodbye message before quitting

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running bubble tea: %w", err)
	}

	return nil
}
