package frontend

import (
	"fmt"
	"strings"
	"time"

	"github.com/speier/smith/internal/session"
	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/components"

	// "github.com/speier/smith/pkg/lotus/engine" // TODO: Update to use Element API
	"github.com/speier/smith/pkg/lotus/tty"
)

// ChatUI represents the main chat interface
type ChatUI struct {
	session   session.Session
	input     *components.TextInput
	inputBox  *components.InputBox
	messages  *components.MessageList
	width     int
	height    int
	streaming bool            // Agent is currently streaming a response
	streamBuf strings.Builder // Buffer for streaming response
	term      *tty.Terminal

	// Rendering optimization: cache the Lotus UI and only rebuild when content changes
	cachedUI        *lotus.UI
	lastMsgCount    int    // Track when messages change
	lastInput       string // Track when input changes (for full re-render)
	needsFullRender bool   // Force full render on next frame
	didFullRebuild  bool   // Track if last render was a full rebuild (needs screen clear)
}

// NewChatUI creates a new chat UI
func NewChatUI(sess session.Session, width, height int) *ChatUI {
	input := components.NewTextInput("chat-input") // TODO: Update to new API
	msgList := components.NewMessageList("chat-messages")
	msgList.SetDimensions(width-2, height-6)

	return &ChatUI{
		session:  sess,
		input:    input,
		inputBox: components.NewInputBox("> ", input),
		messages: msgList,
		width:    width,
		height:   height,
	}
}

// Run starts the chat UI event loop
func (ui *ChatUI) Run() error {
	// Create terminal
	term, err := tty.New()
	if err != nil {
		return fmt.Errorf("failed to create terminal: %w", err)
	}
	ui.term = term

	// Get terminal size
	ui.width, ui.height = term.Size()

	// Enable mouse event filtering
	term.SetFilterMouse(true)

	// Set up render function
	term.OnRender(func() string {
		return ui.render()
	})

	// Position cursor after rendering - now automatic via focus management!
	term.OnPostRender(func() {
		if ui.cachedUI != nil {
			ui.cachedUI.UpdateCursor(term)
		}
	})

	// Set up ticker for streaming updates (refresh while agent is responding)
	term.OnTick(50*time.Millisecond, func() {
		// Only trigger renders when actively streaming
		// When not streaming, renders happen on keypress/events only
		if !ui.streaming {
			return // Skip tick when idle - saves CPU and allows cursor to blink
		}
		// Tick handler is called every 50ms to update streaming content smoothly
		// The render will be triggered automatically after this callback
	})

	// Handle terminal resize
	term.OnResize(func(width, height int) {
		ui.width = width
		ui.height = height
		// Use Lotus's fast Reflow instead of full rebuild
		if ui.cachedUI != nil {
			ui.cachedUI.Reflow(width, height)
		} else {
			// First render or cache was cleared
			ui.needsFullRender = true
		}
	})

	// Set up key handler
	term.OnKey(func(event tty.KeyEvent) bool {
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
func (ui *ChatUI) handleKey(event tty.KeyEvent) {
	// Don't process input while streaming
	if ui.streaming {
		return
	}

	// Handle enter key (app-level)
	if event.IsEnter() {
		ui.handleSubmit()
		return
	}

	// Handle up/down for message scrolling (app-level)
	if event.Key == tty.KeyEscape && event.Code != "" {
		switch event.Code {
		case tty.SeqUp:
			ui.scrollMessagesUp()
			return
		case tty.SeqDown:
			ui.scrollMessagesDown()
			return
		}
	}

	// Route all other keys to focused component automatically!
	if ui.cachedUI != nil {
		ui.cachedUI.HandleKey(event)
	}
}

// handleSubmit sends the message to the agent
func (ui *ChatUI) handleSubmit() {
	if ui.input.Value == "" {
		return
	}

	message := ui.input.Value
	ui.input.Clear()

	// Start streaming
	ui.streaming = true
	ui.streamBuf.Reset()

	// Reset scroll to show new message (auto-scroll)
	ui.messages.ScrollToBottom()

	// Send message asynchronously
	go func() {
		stream, err := ui.session.SendMessage(message)
		if err != nil {
			ui.streamBuf.WriteString(fmt.Sprintf("Error: %v", err))
			ui.streaming = false
			ui.needsFullRender = true // Force rebuild to show error
			return
		}

		// Consume stream
		for chunk := range stream {
			ui.streamBuf.WriteString(chunk)
		}

		ui.streaming = false
		ui.needsFullRender = true // Force rebuild when streaming completes
	}()
}

// Message scrolling methods
func (ui *ChatUI) scrollMessagesUp() {
	ui.messages.ScrollUp()
}

func (ui *ChatUI) scrollMessagesDown() {
	ui.messages.ScrollDown()
}

// UpdateSize updates the UI dimensions and input width
func (ui *ChatUI) UpdateSize(width, height int) {
	ui.width = width
	ui.height = height
	ui.input.Width = width - 10 // Update input width
	ui.messages.Width = width - 2
	ui.messages.Height = height - 6
	ui.needsFullRender = true
}

// render generates the UI markup and renders it
// Smart caching: Only re-parse HTML when messages change, not on every keystroke
func (ui *ChatUI) render() string {
	// Build message history with wrapping
	history := ui.session.GetHistory()
	currentMsgCount := len(history)
	if ui.streaming {
		currentMsgCount++ // Count streaming message
	}

	// Check if we need a full re-render
	needsRebuild := ui.cachedUI == nil ||
		ui.needsFullRender ||
		currentMsgCount != ui.lastMsgCount ||
		ui.streaming // ALWAYS rebuild while streaming (content changes every tick)

	if needsRebuild {
		// Full rebuild - parse HTML/CSS, build layout
		ui.cachedUI = ui.buildFullUI(history)
		ui.lastMsgCount = currentMsgCount
		ui.lastInput = ui.input.Value
		ui.needsFullRender = false
		ui.didFullRebuild = true // Mark that we did a full rebuild
	} else if ui.input.Value != ui.lastInput {
		// Input changed but messages didn't - update just the input node
		// TODO: ui.updateInputNode() - needs Element API update
		ui.needsFullRender = true // Force full render for now
		ui.lastInput = ui.input.Value
		ui.didFullRebuild = false // Fast path - no rebuild
	} else {
		// Nothing changed - reusing cached UI
		ui.didFullRebuild = false
	}

	// Render the (possibly cached) UI to ANSI
	return ui.cachedUI.RenderToTerminal(ui.didFullRebuild)
}

// buildFullUI creates a new Lotus UI from scratch (slow but complete)
func (ui *ChatUI) buildFullUI(history []session.Message) *lotus.UI {
	// Clear and rebuild message list
	ui.messages.Clear()
	ui.messages.IsStreaming = ui.streaming

	// If no history, show welcome banner
	if len(history) == 0 {
		welcomeText := GetWelcomeBanner()
		ui.messages.AddMessage("system", welcomeText)
	}

	// Add history messages
	for _, msg := range history {
		ui.messages.AddMessage(msg.Role, msg.Content)
	}

	// If streaming, add partial response
	if ui.streaming {
		partial := ui.streamBuf.String()
		if partial != "" {
			ui.messages.AddMessage("assistant", partial)
		}
	}

	// Build markup using components
	markup := fmt.Sprintf(`
		<box id="root">
			<box id="messages">%s</box>
			%s
		</box>
	`, ui.messages.Render(), ui.inputBox.Render())

	css := `
		#root {
			display: flex;
			flex-direction: column;
			height: 100%;
			width: 100%;
		}

		#messages {
			flex: 1;
			padding: 0;
			display: flex;
			flex-direction: column;
			border: 1px solid;
			border-style: single;
		}
	` + ui.messages.GetCSS() + ui.inputBox.GetCSS()

	lotusUI := lotus.New(markup, css, ui.width, ui.height)

	// Register input component and set focus for automatic cursor management
	lotusUI.RegisterComponent("input-text", ui.input)
	lotusUI.SetFocus("input-text")

	return lotusUI
}

// updateInputNode updates just the input text node (fast - no HTML parsing)
// TODO: Update to use Element API instead of legacy engine.Node
/*
func (ui *ChatUI) updateInputNode() {
	if ui.cachedUI == nil {
		return
	}

	// Find the input text box by ID
	inputBox := ui.cachedUI.FindByID("input-text")
	if inputBox == nil {
		return
	}

	// Update its content using component
	inputDisplay := ui.input.GetDisplay()

	// The actual text is in the first child (type="text" node)
	if len(inputBox.Children) > 0 && inputBox.Children[0].Type == "text" {
		// Just update the existing text node's content
		inputBox.Children[0].Content = inputDisplay
	} else {
		// No child text node - create one with proper layout
		textNode := engine.NewNode("text")
		textNode.Content = inputDisplay
		textNode.Parent = inputBox

		// Copy layout from parent (position text inside the box)
		textNode.X = inputBox.X
		textNode.Y = inputBox.Y
		textNode.Width = inputBox.Width
		textNode.Height = inputBox.Height

		inputBox.Children = []*engine.Node{textNode}
	}
}
*/

// PositionCursor positions the cursor after rendering is complete
