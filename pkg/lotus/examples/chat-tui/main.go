package main

import (
	"fmt"
	"time"

	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/terminal"
)

type Message struct {
	Text      string
	Timestamp time.Time
	IsUser    bool
}

type ChatApp struct {
	messages    []Message
	input       string
	cursorPos   int // Cursor position in input
	width       int
	height      int
	cursorBlink bool
	inputScroll int // Horizontal scroll offset for long input
}

func NewChatApp(width, height int) *ChatApp {
	return &ChatApp{
		messages:    []Message{}, // Start with empty chat
		input:       "",
		cursorPos:   0,
		width:       width,
		height:      height,
		cursorBlink: true,
		inputScroll: 0,
	}
}

func (app *ChatApp) AddMessage(text string, isUser bool) {
	app.messages = append(app.messages, Message{
		Text:      text,
		Timestamp: time.Now(),
		IsUser:    isUser,
	})
}

func (app *ChatApp) UpdateSize(width, height int) {
	app.width = width
	app.height = height
}

func (app *ChatApp) insertChar(ch string) {
	// Insert character at cursor position
	app.input = app.input[:app.cursorPos] + ch + app.input[app.cursorPos:]
	app.cursorPos++
	app.adjustScroll()
}

func (app *ChatApp) deleteChar() {
	// Delete character before cursor (backspace)
	if app.cursorPos > 0 {
		app.input = app.input[:app.cursorPos-1] + app.input[app.cursorPos:]
		app.cursorPos--
		app.adjustScroll()
	}
}

func (app *ChatApp) deleteCharForward() {
	// Delete character at cursor (delete key)
	if app.cursorPos < len(app.input) {
		app.input = app.input[:app.cursorPos] + app.input[app.cursorPos+1:]
	}
}

func (app *ChatApp) moveCursorLeft() {
	if app.cursorPos > 0 {
		app.cursorPos--
		app.adjustScroll()
	}
}

func (app *ChatApp) moveCursorRight() {
	if app.cursorPos < len(app.input) {
		app.cursorPos++
		app.adjustScroll()
	}
}

func (app *ChatApp) moveCursorHome() {
	app.cursorPos = 0
	app.inputScroll = 0
}

func (app *ChatApp) moveCursorEnd() {
	app.cursorPos = len(app.input)
	app.adjustScroll()
}

func (app *ChatApp) moveCursorWordLeft() {
	if app.cursorPos == 0 {
		return
	}

	// Skip any spaces at current position
	for app.cursorPos > 0 && app.input[app.cursorPos-1] == ' ' {
		app.cursorPos--
	}

	// Move to start of current/previous word
	for app.cursorPos > 0 && app.input[app.cursorPos-1] != ' ' {
		app.cursorPos--
	}

	app.adjustScroll()
}

func (app *ChatApp) moveCursorWordRight() {
	if app.cursorPos >= len(app.input) {
		return
	}

	// Skip any spaces at current position
	for app.cursorPos < len(app.input) && app.input[app.cursorPos] == ' ' {
		app.cursorPos++
	}

	// Move to end of current/next word
	for app.cursorPos < len(app.input) && app.input[app.cursorPos] != ' ' {
		app.cursorPos++
	}

	app.adjustScroll()
}

func (app *ChatApp) adjustScroll() {
	// Adjust horizontal scroll to keep cursor visible
	// Assume input area is about width-10 characters (account for borders, prompt)
	visibleWidth := app.width - 10
	if visibleWidth < 10 {
		visibleWidth = 10
	}

	// If cursor is past the visible area, scroll right
	if app.cursorPos-app.inputScroll >= visibleWidth {
		app.inputScroll = app.cursorPos - visibleWidth + 1
	}

	// If cursor is before the visible area, scroll left
	if app.cursorPos < app.inputScroll {
		app.inputScroll = app.cursorPos
	}
}

func (app *ChatApp) getVisibleInput() (visible string, cursorOffset int) {
	visibleWidth := app.width - 10
	if visibleWidth < 10 {
		visibleWidth = 10
	}

	endPos := app.inputScroll + visibleWidth
	if endPos > len(app.input) {
		endPos = len(app.input)
	}

	visible = app.input[app.inputScroll:endPos]
	cursorOffset = app.cursorPos - app.inputScroll
	return
}

func (app *ChatApp) Render() string {
	// Build message history (last N messages that fit)
	maxMessages := app.height - 8 // Reserve space for header and input
	if maxMessages < 3 {
		maxMessages = 3
	}
	startIdx := 0
	if len(app.messages) > maxMessages {
		startIdx = len(app.messages) - maxMessages
	}

	// Create individual box for each message
	messageBoxes := ""
	for _, msg := range app.messages[startIdx:] {
		var line string
		if msg.IsUser {
			// User input with | prefix
			line = fmt.Sprintf("| %s", msg.Text)
			messageBoxes += fmt.Sprintf(`<box class="message">%s</box>`, line)
		} else {
			// System response without prefix, plus blank line after
			line = msg.Text
			messageBoxes += fmt.Sprintf(`<box class="message">%s</box>`, line)
			messageBoxes += `<box class="message"></box>` // Blank line
		}
	}

	// Get visible portion of input and cursor position
	visibleInput, cursorOffset := app.getVisibleInput()

	// Build input with cursor
	inputDisplay := ""
	if len(visibleInput) == 0 {
		if app.cursorBlink {
			inputDisplay = "|"
		} else {
			inputDisplay = " "
		}
	} else {
		for i, ch := range visibleInput {
			if i == cursorOffset && app.cursorBlink {
				inputDisplay += "|"
			}
			inputDisplay += string(ch)
		}
		// Cursor at end
		if cursorOffset >= len(visibleInput) && app.cursorBlink {
			inputDisplay += "|"
		}
	}

	markup := fmt.Sprintf(`
		<box id="root">
			<box id="header">Lotus Chat - Press Ctrl+C to exit</box>
			<box id="messages">%s</box>
			<box id="input-container">
				<box id="input-label">> </box>
				<box id="input-text">%s</box>
			</box>
		</box>
	`, messageBoxes, inputDisplay)

	css := `
		#root {
			display: flex;
			flex-direction: column;
			height: 100%;
			width: 100%;
		}

		#header {
			height: 3;
			color: #5af;
			border: 1px solid;
			border-style: rounded;
			text-align: center;
		}

		#messages {
			flex: 1;
			padding: 0;
			color: #ddd;
			display: flex;
			flex-direction: column;
			border: 1px solid;
			border-style: single;
		}

		.message {
			height: 1;
			margin: 0;
			padding: 0 1 0 1;
		}

		#input-container {
			height: 3;
			border: 1px solid;
			border-style: single;
			display: flex;
			flex-direction: row;
		}

		#input-label {
			width: 3;
			color: #5af;
			padding: 0 0 0 1;
		}

		#input-text {
			flex: 1;
			color: #fff;
		}
	`

	ui := lotus.New(markup, css, app.width, app.height)
	return ui.RenderToTerminal()
}

func main() {
	// Create terminal
	term, err := terminal.New()
	if err != nil {
		fmt.Printf("Error creating terminal: %v\n", err)
		return
	}

	// Get terminal size
	width, height := term.Size()

	// Create app
	app := NewChatApp(width, height)

	// Enable mouse event filtering
	term.SetFilterMouse(true)

	// Set up render function
	term.OnRender(func() string {
		return app.Render()
	})

	// Set up tick handler for cursor blinking
	term.OnTick(500*time.Millisecond, func() {
		app.cursorBlink = !app.cursorBlink
	})

	// Handle terminal resize
	term.OnResize(func(width, height int) {
		app.UpdateSize(width, height)
	})

	// Set up key handler
	term.OnKey(func(event terminal.KeyEvent) bool {
		// Ctrl+C or Ctrl+D exits
		if event.IsCtrlC() || event.IsCtrlD() {
			return false
		}

		// Enter sends message
		if event.IsEnter() {
			if app.input != "" {
				app.AddMessage(app.input, true)
				// Simulate system response
				go func() {
					time.Sleep(500 * time.Millisecond)
					app.AddMessage("Message received!", false)
				}()
				app.input = ""
				app.cursorPos = 0
				app.inputScroll = 0
			}
			return true
		}

		// Backspace
		if event.IsBackspace() {
			app.deleteChar()
			return true
		}

		// Delete key
		if event.Code == terminal.SeqDelete {
			app.deleteCharForward()
			return true
		}

		// Arrow keys for cursor movement
		if event.Code == terminal.SeqLeft {
			app.moveCursorLeft()
			return true
		}

		if event.Code == terminal.SeqRight {
			app.moveCursorRight()
			return true
		}

		// Word navigation (Ctrl+Left/Right or Alt+Left/Right)
		if event.Code == terminal.SeqCtrlLeft {
			app.moveCursorWordLeft()
			return true
		}

		if event.Code == terminal.SeqCtrlRight {
			app.moveCursorWordRight()
			return true
		}

		// Cmd+Left/Right on Mac (beginning/end of line)
		if event.Code == terminal.SeqCmdLeft {
			app.moveCursorHome()
			return true
		}

		if event.Code == terminal.SeqCmdRight {
			app.moveCursorEnd()
			return true
		}

		// Home key
		if event.Code == terminal.SeqHome {
			app.moveCursorHome()
			return true
		}

		// End key
		if event.Code == terminal.SeqEnd {
			app.moveCursorEnd()
			return true
		}

		// Printable characters
		if event.IsPrintable() {
			app.insertChar(event.Char)
		}

		return true
	})

	// Start the terminal (blocks until exit)
	if err := term.Start(); err != nil {
		fmt.Printf("Terminal error: %v\n", err)
		return
	}

	// Clean exit
	fmt.Print("\033[2J\033[H") // Clear screen
	fmt.Println("Goodbye from Lotus Chat!")
}
