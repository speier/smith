package components

import "fmt"

// Dialog is a modal dialog box
type Dialog struct {
	Title         string
	Message       string
	Buttons       []string
	Selected      int
	Width         int
	Height        int
	BorderStyle   string
	TitleColor    string
	MessageColor  string
	ButtonColor   string
	SelectedColor string
}

// NewDialog creates a new dialog
func NewDialog(title, message string, buttons []string) *Dialog {
	return &Dialog{
		Title:         title,
		Message:       message,
		Buttons:       buttons,
		Selected:      0,
		Width:         40,
		Height:        10,
		BorderStyle:   "double",
		TitleColor:    "#5af",
		MessageColor:  "#ddd",
		ButtonColor:   "#999",
		SelectedColor: "#5af",
	}
}

// SelectNext moves to next button
func (d *Dialog) SelectNext() {
	d.Selected++
	if d.Selected >= len(d.Buttons) {
		d.Selected = len(d.Buttons) - 1
	}
}

// SelectPrev moves to previous button
func (d *Dialog) SelectPrev() {
	d.Selected--
	if d.Selected < 0 {
		d.Selected = 0
	}
}

// GetSelectedButton returns the currently selected button
func (d *Dialog) GetSelectedButton() string {
	if d.Selected >= 0 && d.Selected < len(d.Buttons) {
		return d.Buttons[d.Selected]
	}
	return ""
}

// Render generates the markup for the dialog
func (d *Dialog) Render() string {
	// Build buttons
	buttonBoxes := ""
	for i, btn := range d.Buttons {
		class := "dialog-button"
		if i == d.Selected {
			class += " selected"
		}

		brackets := "[ ]"
		if i == d.Selected {
			brackets = "[*]"
		}

		buttonBoxes += fmt.Sprintf(`<box class="%s">%s %s</box>`, class, brackets, btn)
	}

	markup := fmt.Sprintf(`
		<box id="dialog-overlay">
			<box id="dialog">
				<box id="dialog-title">%s</box>
				<box id="dialog-message">%s</box>
				<box id="dialog-buttons">%s</box>
			</box>
		</box>
	`, d.Title, d.Message, buttonBoxes)

	return markup
}

// GetCSS returns the CSS for dialog styling
func (d *Dialog) GetCSS() string {
	return fmt.Sprintf(`
		#dialog-overlay {
			position: absolute;
			top: 0;
			left: 0;
			width: 100%%;
			height: 100%%;
			display: flex;
			justify-content: center;
			align-items: center;
		}
		#dialog {
			width: %d;
			height: %d;
			border: 2px solid;
			border-style: %s;
			display: flex;
			flex-direction: column;
		}
		#dialog-title {
			height: 3;
			color: %s;
			text-align: center;
			border-bottom: 1px solid;
		}
		#dialog-message {
			flex: 1;
			color: %s;
			padding: 1;
		}
		#dialog-buttons {
			height: 3;
			display: flex;
			flex-direction: row;
			justify-content: center;
		}
		.dialog-button {
			color: %s;
			margin: 0 1;
		}
		.dialog-button.selected {
			color: %s;
		}
	`, d.Width, d.Height, d.BorderStyle, d.TitleColor, d.MessageColor, d.ButtonColor, d.SelectedColor)
}
