package components

import "fmt"

// InputBox combines a label with a TextInput
type InputBox struct {
	Label       string
	Input       *TextInput
	LabelWidth  int
	LabelColor  string
	BorderStyle string
}

// NewInputBox creates a new input box with label
func NewInputBox(label string, input *TextInput) *InputBox {
	return &InputBox{
		Label:       label,
		Input:       input,
		LabelWidth:  3,
		LabelColor:  "#5af",
		BorderStyle: "single",
	}
}

// Render generates the markup for the input box
func (i *InputBox) Render() string {
	inputDisplay := i.Input.GetDisplay()

	markup := fmt.Sprintf(`
		<box id="input-container">
			<box id="input-label">%s</box>
			<box id="input-text">%s</box>
		</box>
	`, i.Label, inputDisplay)

	return markup
}

// GetCSS returns the CSS for input box styling
func (i *InputBox) GetCSS() string {
	return fmt.Sprintf(`
		#input-container {
			height: 3;
			border: 1px solid;
			border-style: %s;
			display: flex;
			flex-direction: row;
		}
		#input-label {
			width: %d;
			color: %s;
			padding: 0 0 0 1;
		}
		#input-text {
			flex: 1;
			color: #fff;
		}
	`, i.BorderStyle, i.LabelWidth, i.LabelColor)
}
