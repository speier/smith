package frontend

import (
	"fmt"
	"strings"
	"time"

	"github.com/speier/smith/internal/session"
	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/terminal"
)

// ChatUI represents the main chat interface
type ChatUI struct {
	session       session.Session
	input         string
	cursorPos     int
	width         int
	height        int
	cursorBlink   bool
	inputScroll   int
	messageScroll int             // Scroll offset for message history
	streaming     bool            // Agent is currently streaming a response
	streamBuf     strings.Builder // Buffer for streaming response
}

// NewChatUI creates a new chat UI
func NewChatUI(sess session.Session, width, height int) *ChatUI {
	return &ChatUI{
		session:       sess,
		input:         "",
		cursorPos:     0,
		width:         width,
		height:        height,
		cursorBlink:   true,
		inputScroll:   0,
		messageScroll: 0,
	}
}

// Run starts the chat UI event loop
func (ui *ChatUI) Run() error {
	// Create terminal
	term, err := terminal.New()
	if err != nil {
		return fmt.Errorf("failed to create terminal: %w", err)
	}

	// Get terminal size
	ui.width, ui.height = term.Size()

	// Enable mouse event filtering
	term.SetFilterMouse(true)

	// Set up render function
	term.OnRender(func() string {
		return ui.render()
	})

	// Set up tick handler for cursor blinking
	term.OnTick(500*time.Millisecond, func() {
		ui.cursorBlink = !ui.cursorBlink
	})

	// Handle terminal resize
	term.OnResize(func(width, height int) {
		ui.width = width
		ui.height = height
	})

	// Set up key handler
	term.OnKey(func(event terminal.KeyEvent) bool {
		// Handle Ctrl+C to exit
		if event.IsCtrlC() || event.IsCtrlD() {
			return false // Stop the loop
		}
		ui.handleKey(event)
		return true // Continue the loop
	})

	// Start the event loop
	return term.Start()
}

// handleKey processes keyboard input
func (ui *ChatUI) handleKey(event terminal.KeyEvent) {
	// Don't process input while streaming
	if ui.streaming {
		return
	}

	// Handle enter key
	if event.IsEnter() {
		ui.handleSubmit()
		return
	}

	if event.IsBackspace() {
		ui.deleteChar()
		return
	}

	// Handle escape sequences (special keys)
	if event.Key == terminal.KeyEscape && event.Code != "" {
		switch event.Code {
		case terminal.SeqLeft:
			ui.moveCursorLeft()
		case terminal.SeqRight:
			ui.moveCursorRight()
		case terminal.SeqUp:
			ui.scrollMessagesUp()
		case terminal.SeqDown:
			ui.scrollMessagesDown()
		case terminal.SeqHome, terminal.SeqHome2:
			ui.moveCursorHome()
		case terminal.SeqEnd, terminal.SeqEnd2:
			ui.moveCursorEnd()
		case terminal.SeqDelete:
			ui.deleteCharForward()
		case terminal.SeqCtrlLeft, terminal.SeqAltLeft:
			ui.moveCursorWordLeft()
		case terminal.SeqCtrlRight, terminal.SeqAltRight:
			ui.moveCursorWordRight()
		case terminal.SeqCmdLeft:
			ui.moveCursorHome()
		case terminal.SeqCmdRight:
			ui.moveCursorEnd()
		}
		return
	}

	// Handle printable characters
	if event.IsPrintable() {
		ui.insertChar(event.Char)
	}
}

// handleSubmit sends the message to the agent
func (ui *ChatUI) handleSubmit() {
	if ui.input == "" {
		return
	}

	message := ui.input
	ui.input = ""
	ui.cursorPos = 0
	ui.inputScroll = 0

	// Start streaming
	ui.streaming = true
	ui.streamBuf.Reset()

	// Reset scroll to show new message
	ui.messageScroll = 999999 // Will auto-clamp to bottom

	// Send message asynchronously
	go func() {
		stream, err := ui.session.SendMessage(message)
		if err != nil {
			ui.streamBuf.WriteString(fmt.Sprintf("Error: %v", err))
			ui.streaming = false
			return
		}

		// Consume stream
		for chunk := range stream {
			ui.streamBuf.WriteString(chunk)
		}

		ui.streaming = false
	}()
}

// Input manipulation methods
func (ui *ChatUI) insertChar(ch string) {
	ui.input = ui.input[:ui.cursorPos] + ch + ui.input[ui.cursorPos:]
	ui.cursorPos++
	ui.adjustScroll()
}

func (ui *ChatUI) deleteChar() {
	if ui.cursorPos > 0 {
		ui.input = ui.input[:ui.cursorPos-1] + ui.input[ui.cursorPos:]
		ui.cursorPos--
		ui.adjustScroll()
	}
}

func (ui *ChatUI) deleteCharForward() {
	if ui.cursorPos < len(ui.input) {
		ui.input = ui.input[:ui.cursorPos] + ui.input[ui.cursorPos+1:]
	}
}

func (ui *ChatUI) moveCursorLeft() {
	if ui.cursorPos > 0 {
		ui.cursorPos--
		ui.adjustScroll()
	}
}

func (ui *ChatUI) moveCursorRight() {
	if ui.cursorPos < len(ui.input) {
		ui.cursorPos++
		ui.adjustScroll()
	}
}

func (ui *ChatUI) moveCursorHome() {
	ui.cursorPos = 0
	ui.inputScroll = 0
}

func (ui *ChatUI) moveCursorEnd() {
	ui.cursorPos = len(ui.input)
	ui.adjustScroll()
}

func (ui *ChatUI) moveCursorWordLeft() {
	if ui.cursorPos == 0 {
		return
	}

	// Skip spaces
	for ui.cursorPos > 0 && ui.input[ui.cursorPos-1] == ' ' {
		ui.cursorPos--
	}

	// Move to word start
	for ui.cursorPos > 0 && ui.input[ui.cursorPos-1] != ' ' {
		ui.cursorPos--
	}

	ui.adjustScroll()
}

func (ui *ChatUI) moveCursorWordRight() {
	if ui.cursorPos >= len(ui.input) {
		return
	}

	// Skip spaces
	for ui.cursorPos < len(ui.input) && ui.input[ui.cursorPos] == ' ' {
		ui.cursorPos++
	}

	// Move to word end
	for ui.cursorPos < len(ui.input) && ui.input[ui.cursorPos] != ' ' {
		ui.cursorPos++
	}

	ui.adjustScroll()
}

func (ui *ChatUI) adjustScroll() {
	visibleWidth := ui.width - 10
	if visibleWidth < 10 {
		visibleWidth = 10
	}

	if ui.cursorPos-ui.inputScroll >= visibleWidth {
		ui.inputScroll = ui.cursorPos - visibleWidth + 1
	}

	if ui.cursorPos < ui.inputScroll {
		ui.inputScroll = ui.cursorPos
	}
}

// Message scrolling methods
func (ui *ChatUI) scrollMessagesUp() {
	if ui.messageScroll > 0 {
		ui.messageScroll--
	}
}

func (ui *ChatUI) scrollMessagesDown() {
	ui.messageScroll++
	// Will be clamped in render
}

func (ui *ChatUI) getVisibleInput() (visible string, cursorOffset int) {
	visibleWidth := ui.width - 10
	if visibleWidth < 10 {
		visibleWidth = 10
	}

	endPos := ui.inputScroll + visibleWidth
	if endPos > len(ui.input) {
		endPos = len(ui.input)
	}

	visible = ui.input[ui.inputScroll:endPos]
	cursorOffset = ui.cursorPos - ui.inputScroll
	return
}

// render generates the UI markup and renders it
func (ui *ChatUI) render() string {
	// Build message history with wrapping
	history := ui.session.GetHistory()

	// Calculate message area dimensions
	messageWidth := ui.width - 4 // Account for borders and padding
	if messageWidth < 20 {
		messageWidth = 20
	}

	// Wrap all messages and count total lines
	type wrappedMsg struct {
		role  string
		lines []string
	}
	var wrapped []wrappedMsg

	// If no history, show welcome banner (like original REPL)
	if len(history) == 0 {
		welcomeText := GetWelcomeBanner()
		// Don't wrap - welcome banner is already formatted with ASCII art and centering
		lines := strings.Split(welcomeText, "\n")
		wrapped = append(wrapped, wrappedMsg{
			role:  "system",
			lines: lines,
		})
	}

	for _, msg := range history {
		lines := wrapText(msg.Content, messageWidth)
		wrapped = append(wrapped, wrappedMsg{
			role:  msg.Role,
			lines: lines,
		})
	}

	// If streaming, add partial response
	if ui.streaming {
		partial := ui.streamBuf.String()
		if partial != "" {
			lines := wrapText(partial, messageWidth)
			wrapped = append(wrapped, wrappedMsg{
				role:  "assistant",
				lines: lines,
			})
		}
	}

	// Calculate total lines needed
	totalLines := 0
	for _, msg := range wrapped {
		totalLines += len(msg.lines) + 1 // +1 for spacing
	}

	// Calculate visible area
	maxVisibleLines := ui.height - 8
	if maxVisibleLines < 3 {
		maxVisibleLines = 3
	}

	// Clamp scroll
	maxScroll := totalLines - maxVisibleLines
	if maxScroll < 0 {
		maxScroll = 0
	}
	if ui.messageScroll > maxScroll {
		ui.messageScroll = maxScroll
	}
	if ui.messageScroll < 0 {
		ui.messageScroll = 0
	}

	// Auto-scroll to bottom if not manually scrolled
	atBottom := ui.messageScroll >= maxScroll
	if atBottom && !ui.streaming {
		ui.messageScroll = maxScroll
	}

	// Build visible messages
	messageBoxes := ""
	currentLine := 0
	visibleStart := ui.messageScroll
	visibleEnd := ui.messageScroll + maxVisibleLines

	for _, msg := range wrapped {
		for _, line := range msg.lines {
			if currentLine >= visibleStart && currentLine < visibleEnd {
				class := "message " + msg.role
				if ui.streaming && msg.role == "assistant" {
					class += " streaming"
				}
				// Prefix user messages with "| "
				prefix := ""
				if msg.role == "user" {
					prefix = "| "
				}
				messageBoxes += fmt.Sprintf(`<box class="%s">%s%s</box>`, class, prefix, line)
			}
			currentLine++
		}
		// Add spacing line
		if currentLine >= visibleStart && currentLine < visibleEnd {
			messageBoxes += `<box class="message"></box>`
		}
		currentLine++
	}

	// Build input with cursor
	visibleInput, cursorOffset := ui.getVisibleInput()
	inputDisplay := ""
	if len(visibleInput) == 0 {
		if ui.cursorBlink {
			inputDisplay = "|"
		} else {
			inputDisplay = " "
		}
	} else {
		for i, ch := range visibleInput {
			if i == cursorOffset && ui.cursorBlink {
				inputDisplay = "|"
			}
			inputDisplay += string(ch)
		}
		if cursorOffset >= len(visibleInput) && ui.cursorBlink {
			inputDisplay += "|"
		}
	}

	// Build markup
	markup := fmt.Sprintf(`
		<box id="root">
			<box id="messages">%s</box>
			<box id="input-container">
				<box id="input-label">> </box>
				<box id="input-text">%s</box>
			</box>
		</box>
	`, messageBoxes, inputDisplay)

	return lotus.New(markup, smithCSS, ui.width, ui.height).RenderToTerminal()
}

// wrapText wraps text to fit within the given width
func wrapText(text string, width int) []string {
	if width < 10 {
		width = 10
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	currentLine := ""
	for _, word := range words {
		// If word itself is longer than width, split it
		if len(word) > width {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			// Split long word
			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}
			if len(word) > 0 {
				currentLine = word
			}
			continue
		}

		// Try adding word to current line
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= width {
			currentLine = testLine
		} else {
			// Start new line
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// smithCSS defines the Matrix-themed styling
const smithCSS = `
	#root {
		display: flex;
		flex-direction: column;
		height: 100%;
		width: 100%;
	}

	#messages {
		flex: 1;
		padding: 0;
		color: 10;
		display: flex;
		flex-direction: column;
		border: 1px solid;
		border-style: single;
	}

	.message {
		height: 1;
		margin: 0;
		padding: 0 1 0 1;
		width: 100%;
	}

	.message.system {
		text-align: center;
		color: 10;
	}

	.message.user {
		color: 10;
	}

	.message.assistant {
		color: 10;
	}

	.message.streaming {
		color: 10;
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
		color: #0f0;
		padding: 0 0 0 1;
	}

	#input-text {
		flex: 1;
		color: #0f0;
	}
`
