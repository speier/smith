package repl

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/speier/smith/internal/config"
	"github.com/speier/smith/internal/engine"
	"github.com/speier/smith/internal/llm"
	"github.com/speier/smith/internal/safety"
)

// delayedQuitMsg is sent after showing the goodbye message
type delayedQuitMsg struct{}

// streamChunkMsg is sent when a chunk of streaming response arrives
type streamChunkMsg struct {
	chunk string
	done  bool
}

// errorMsg is sent when an error occurs during chat
type errorMsg struct {
	title   string
	message string
}

// Message represents a chat message with metadata
type Message struct {
	Content   string
	Type      string // "user", "ai", "system", "error"
	Timestamp time.Time
}

// Styles
var (
	// Sidebar styles
	sidebarStyle = lipgloss.NewStyle().
			Width(35).
			Padding(1, 2).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("240"))

	sidebarTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true)

	sidebarSectionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				MarginTop(1)

	sidebarItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("250"))

	sidebarActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")).
				Bold(true)

	sidebarIdleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("240"))

	// Main area styles
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("235"))

	autoLevelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true)

	modelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))

	aiStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("5"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

const sidebarWidth = 40

// Command represents a slash command
type Command struct {
	Name        string
	Alias       string
	Description string
}

// Available commands
var availableCommands = []Command{
	{Name: "/status", Alias: "/s", Description: "Show task counts and progress"},
	{Name: "/backlog", Alias: "/b", Description: "List all tasks"},
	{Name: "/inbox", Alias: "/i", Description: "Check agent questions/messages"},
	{Name: "/agents", Alias: "/a", Description: "Show active agents"},
	{Name: "/settings", Alias: "", Description: "Show settings"},
	{Name: "/copy", Alias: "", Description: "Copy last message to buffer"},
	{Name: "/edit-global", Alias: "", Description: "Edit global configuration in editor"},
	{Name: "/clear", Alias: "/c", Description: "Clear conversation"},
	{Name: "/help", Alias: "/h", Description: "Show help"},
	{Name: "/quit", Alias: "/q", Description: "Exit Smith"},
}

// BubbleModel is our Bubble Tea model
type BubbleModel struct {
	engine             *engine.Engine
	textarea           textarea.Model
	messages           []Message
	err                error
	width              int
	height             int
	autoLevel          string
	projectPath        string
	quitting           bool
	streamingResponse  string // Current streaming response being built
	isStreaming        bool   // Whether we're currently streaming a response
	showCommandMenu    bool
	commandMenuIndex   int
	filteredCommands   []Command
	showSettings       bool
	settingsIndex      int
	showProviderMenu   bool
	providerMenuIndex  int
	showModelMenu      bool
	modelMenuIndex     int
	historyScroll      int  // Current scroll position in message history
	needsProviderSetup bool // Whether provider needs to be selected on first run
}

// keyMap defines our key bindings
type keyMap struct {
	Send       key.Binding
	Quit       key.Binding
	CycleLevel key.Binding
	Help       key.Binding
	Clear      key.Binding
	ScrollUp   key.Binding
	ScrollDown key.Binding
	Copy       key.Binding
}

var keys = keyMap{
	Send: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "send"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "ctrl+d"),
		key.WithHelp("ctrl+c", "quit"),
	),
	CycleLevel: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "cycle auto-level"),
	),
	Help: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "help"),
	),
	Clear: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "clear"),
	),
	ScrollUp: key.NewBinding(
		key.WithKeys("pgup", "ctrl+u"),
		key.WithHelp("pgup", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("pgdown", "ctrl+d"),
		key.WithHelp("pgdown", "scroll down"),
	),
	Copy: key.NewBinding(
		key.WithKeys("ctrl+y"),
		key.WithHelp("ctrl+y", "copy last message"),
	),
}

// NewBubbleModel creates a new Bubble Tea model
func NewBubbleModel(projectPath string, initialPrompt string) (*BubbleModel, error) {
	// Ensure config directory exists and copy rules if needed
	configDir, err := config.EnsureConfigDir()
	if err != nil {
		return nil, fmt.Errorf("ensuring config directory: %w", err)
	}

	// Ensure safety rules are in config directory (user can customize)
	if err := safety.EnsureRulesInConfigDir(configDir); err != nil {
		// Just log warning, don't fail - we have embedded fallback
		fmt.Printf("Warning: Could not setup safety rules: %v\n", err)
	}

	// Load or create local config
	localCfg, err := config.LoadLocal(projectPath)
	if err != nil {
		return nil, fmt.Errorf("loading local config: %w", err)
	}

	// Track if we need provider setup (will show in settings)
	needsProviderSetup := localCfg == nil

	// Use default provider if not yet configured
	providerID := "copilot"
	autoLevel := safety.AutoLevelMedium

	if localCfg != nil {
		providerID = localCfg.Provider
		if localCfg.AutoLevel != "" {
			autoLevel = localCfg.AutoLevel
		}
	}

	// Create LLM provider based on config
	provider, err := llm.NewProviderByID(providerID)
	if err != nil {
		return nil, fmt.Errorf("creating provider: %w", err)
	}

	eng, err := engine.New(engine.Config{
		ProjectPath: projectPath,
		LLMProvider: provider,
	})
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	// Create textarea
	ti := textarea.New()
	ti.Placeholder = "What would you like to build?"
	ti.Focus()
	ti.Prompt = "> "
	ti.CharLimit = 4000
	ti.SetWidth(80)
	ti.SetHeight(1)
	ti.ShowLineNumbers = false
	ti.FocusedStyle.CursorLine = lipgloss.NewStyle() // Remove background from focused line
	ti.BlurredStyle.CursorLine = lipgloss.NewStyle() // Remove background from blurred line

	// Key bindings for textarea
	ti.KeyMap.InsertNewline.SetEnabled(false) // Disable enter for newline

	messages := []Message{}

	// Add welcome message
	welcome := renderWelcome()
	messages = append(messages, Message{
		Content:   welcome,
		Type:      "system",
		Timestamp: time.Now(),
	})

	// If initial prompt provided, add it
	if initialPrompt != "" {
		ti.SetValue(initialPrompt)
	}

	return &BubbleModel{
		engine:             eng,
		textarea:           ti,
		messages:           messages,
		autoLevel:          autoLevel,
		projectPath:        projectPath,
		needsProviderSetup: needsProviderSetup,
		width:              80, // Default width until we get WindowSizeMsg
		height:             24, // Default height
	}, nil
}

func (m BubbleModel) Init() tea.Cmd {
	return textarea.Blink
}

// delayedQuit returns a command that waits a moment then quits
func delayedQuit() tea.Cmd {
	return tea.Tick(time.Millisecond*2000, func(t time.Time) tea.Msg {
		return delayedQuitMsg{}
	})
}

func (m BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case streamChunkMsg:
		if !m.isStreaming {
			return m, nil
		}

		m.streamingResponse += msg.chunk

		if msg.done {
			// Streaming complete
			m.isStreaming = false
			m.messages = append(m.messages, Message{Content: aiStyle.Render("🤖 AI: ") + m.streamingResponse, Type: "ai", Timestamp: time.Now()})
			m.streamingResponse = ""
		}

		return m, nil

	case delayedQuitMsg:
		// Time to quit
		return m, tea.Quit

	case tea.KeyMsg:
		// Global key handlers - these work EVERYWHERE

		// Ctrl+C or Ctrl+D always quits with goodbye message
		if key.Matches(msg, keys.Quit) || msg.String() == "ctrl+c" || msg.String() == "ctrl+d" {
			// Show goodbye message and quit after a short delay
			m.quitting = true
			return m, tea.Tick(time.Millisecond*800, func(t time.Time) tea.Msg {
				return delayedQuitMsg{}
			})
		}

		// ESC closes any open panel/menu
		if msg.Type == tea.KeyEsc {
			if m.showModelMenu {
				m.showModelMenu = false
				m.modelMenuIndex = 0
				return m, nil
			}
			if m.showProviderMenu {
				m.showProviderMenu = false
				m.providerMenuIndex = 0
				return m, nil
			}
			if m.showSettings {
				m.showSettings = false
				m.settingsIndex = 0
				return m, nil
			}
			if m.showCommandMenu {
				m.showCommandMenu = false
				m.filteredCommands = nil
				m.commandMenuIndex = 0
				return m, nil
			}
			// If nothing is open, ESC does nothing
			return m, nil
		}

		// Handle provider menu
		if m.showProviderMenu {
			switch msg.String() {
			case "up", "k":
				if m.providerMenuIndex > 0 {
					m.providerMenuIndex--
				}
				return m, nil
			case "down", "j":
				providers := llm.GetAvailableProviders()
				if m.providerMenuIndex < len(providers)-1 {
					m.providerMenuIndex++
				}
				return m, nil
			case "enter":
				// Select provider
				providers := llm.GetAvailableProviders()
				if m.providerMenuIndex < len(providers) {
					selectedProvider := providers[m.providerMenuIndex]
					cfg, _ := config.Load()
					cfg.Provider = selectedProvider.ID
					cfg.Save()
					m.messages = append(m.messages, Message{Content: fmt.Sprintf("✓ Provider changed to %s", selectedProvider.Name), Type: "system", Timestamp: time.Now()})
					m.showProviderMenu = false
					m.providerMenuIndex = 0
				}
				return m, nil
			case "q":
				// q also closes menu
				m.showProviderMenu = false
				m.providerMenuIndex = 0
				return m, nil
			}
			return m, nil
		}

		// Handle model menu
		if m.showModelMenu {
			switch msg.String() {
			case "up", "k":
				if m.modelMenuIndex > 0 {
					m.modelMenuIndex--
				}
				return m, nil
			case "down", "j":
				// Get models for current provider
				cfg, _ := config.Load()
				provider, err := llm.NewProviderByID(cfg.Provider)
				if err == nil {
					models, err := provider.GetModels()
					if err == nil && m.modelMenuIndex < len(models)-1 {
						m.modelMenuIndex++
					}
				}
				return m, nil
			case "enter":
				// Select model
				cfg, _ := config.Load()
				provider, err := llm.NewProviderByID(cfg.Provider)
				if err == nil {
					models, err := provider.GetModels()
					if err == nil && m.modelMenuIndex < len(models) {
						selectedModel := models[m.modelMenuIndex]
						cfg.Model = selectedModel.ID
						cfg.Save()
						m.messages = append(m.messages, Message{Content: fmt.Sprintf("✓ Model changed to %s", selectedModel.Name), Type: "system", Timestamp: time.Now()})
						m.showModelMenu = false
						m.modelMenuIndex = 0
					}
				}
				return m, nil
			case "q":
				// q also closes menu
				m.showModelMenu = false
				m.modelMenuIndex = 0
				return m, nil
			}
			return m, nil
		}

		// Handle settings panel navigation
		if m.showSettings {
			switch msg.String() {
			case "up", "k":
				if m.settingsIndex > 0 {
					m.settingsIndex--
				}
				return m, nil
			case "down", "j":
				if m.settingsIndex < 6 {
					m.settingsIndex++
				}
				return m, nil
			case "enter":
				// Handle selecting setting based on index
				switch m.settingsIndex {
				case 0: // Provider (global)
					m.showProviderMenu = true
					m.providerMenuIndex = 0
				case 1: // Model (global)
					m.showModelMenu = true
					m.modelMenuIndex = 0
				case 2: // Reasoning (local)
					m.cycleAutoLevel()
				case 4: // Edit global config
					m.editGlobalConfig()
				case 5: // Edit safety rules
					// TODO: Open editor
				case 6: // Open config directory
					// TODO: Open directory
				}
				return m, nil
			case "q":
				// q also closes settings
				m.showSettings = false
				return m, nil
			}
			return m, nil
		}

		// Handle command menu navigation
		if m.showCommandMenu {
			switch msg.String() {
			case "up", "ctrl+p":
				if m.commandMenuIndex > 0 {
					m.commandMenuIndex--
				}
				return m, nil
			case "down", "ctrl+n":
				if m.commandMenuIndex < len(m.filteredCommands)-1 {
					m.commandMenuIndex++
				}
				return m, nil
			case "enter":
				// Select and execute the command
				if len(m.filteredCommands) > 0 {
					selected := m.filteredCommands[m.commandMenuIndex]

					// Hide menu first
					m.showCommandMenu = false
					m.filteredCommands = nil
					m.commandMenuIndex = 0

					// Execute the command immediately
					quitCmd := m.handleSlashCommand(selected.Name)
					m.textarea.Reset()

					// If it's a quit command, return it
					if quitCmd != nil {
						return m, quitCmd
					}
				}
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keys.Help):
			m.messages = append(m.messages, Message{Content: renderHelp(), Type: "system", Timestamp: time.Now()})
			return m, nil

		case key.Matches(msg, keys.CycleLevel):
			m.cycleAutoLevel()
			return m, nil

		case key.Matches(msg, keys.Clear):
			m.messages = []Message{}
			m.messages = append(m.messages, Message{Content: renderWelcome(), Type: "system", Timestamp: time.Now()})
			m.textarea.Reset()
			return m, nil

		case key.Matches(msg, keys.ScrollUp):
			m.historyScroll++
			return m, nil

		case key.Matches(msg, keys.ScrollDown):
			if m.historyScroll > 0 {
				m.historyScroll--
			}
			return m, nil

		case key.Matches(msg, keys.Copy):
			m.copyLastMessage()
			return m, nil

		case key.Matches(msg, keys.Send):
			// Get message and send it
			userMsg := strings.TrimSpace(m.textarea.Value())
			if userMsg == "" {
				return m, nil
			}

			// Handle slash commands
			if strings.HasPrefix(userMsg, "/") {
				quitCmd := m.handleSlashCommand(userMsg)
				m.textarea.Reset()
				if quitCmd != nil {
					return m, quitCmd
				}
				return m, nil
			}

			// Add user message to display
			m.messages = append(m.messages, Message{Content: userMsg, Type: "user", Timestamp: time.Now()})

			// Start streaming response
			m.isStreaming = true
			m.streamingResponse = ""
			m.messages = append(m.messages, Message{Content: "", Type: "ai", Timestamp: time.Now()})

			// Send to engine with streaming
			go func() {
				var response strings.Builder
				err := m.engine.ChatStream(userMsg, func(chunk string) error {
					response.WriteString(chunk)
					m.streamingResponse = response.String()
					// Update the last message in place
					if len(m.messages) > 0 {
						lastIdx := len(m.messages) - 1
						m.messages[lastIdx].Content = m.streamingResponse
					}
					return nil
				})

				if err != nil {
					// Handle error with better messages and recovery suggestions
					errMsg := err.Error()
					var errorContent string

					// Check for specific error types and provide helpful messages
					if strings.Contains(errMsg, "not authenticated") || strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "unauthorized") {
						errorContent = "Authentication Required\n   → Run /settings to configure your provider and model\n   → Make sure your API keys are set up correctly"
					} else if strings.Contains(errMsg, "network") || strings.Contains(errMsg, "connection") || strings.Contains(errMsg, "timeout") {
						errorContent = "Network Error\n   → Check your internet connection\n   → Try again in a moment\n   → If the problem persists, check your provider's status"
					} else if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "quota") {
						errorContent = "Rate Limit Exceeded\n   → Wait a few minutes before trying again\n   → Consider upgrading your plan or switching providers"
					} else if strings.Contains(errMsg, "model not found") || strings.Contains(errMsg, "invalid model") {
						errorContent = "Model Not Available\n   → Run /settings to select a different model\n   → Check if your selected model is still supported"
					} else {
						errorContent = fmt.Sprintf("Error: %s\n   → Try rephrasing your message\n   → Check /settings if the problem persists", errMsg)
					}

					m.messages = append(m.messages, Message{Content: errorContent, Type: "error", Timestamp: time.Now()})
				}

				// Mark streaming as complete
				m.isStreaming = false
				// Finalize the message
				if len(m.messages) > 0 {
					lastIdx := len(m.messages) - 1
					m.messages[lastIdx].Content = m.streamingResponse
				}
			}()

			// Clear textarea
			m.textarea.Reset()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
	}

	// Don't update textarea if quitting or if any overlay is active
	if m.quitting {
		return m, tea.Quit
	}

	// Skip textarea updates when settings/menus are shown
	if !m.showSettings && !m.showProviderMenu && !m.showModelMenu {
		// Update textarea
		m.textarea, cmd = m.textarea.Update(msg)
		cmds = append(cmds, cmd)

		// Check if we should show command menu
		currentValue := m.textarea.Value()
		if strings.HasPrefix(currentValue, "/") && !m.showCommandMenu {
			// Show command menu and filter commands
			m.showCommandMenu = true
			m.filterCommands(currentValue)
			m.commandMenuIndex = 0
		} else if m.showCommandMenu && !strings.HasPrefix(currentValue, "/") {
			// Hide menu if "/" was deleted
			m.showCommandMenu = false
			m.filteredCommands = nil
			m.commandMenuIndex = 0
		} else if m.showCommandMenu {
			// Update filtered commands as user types
			m.filterCommands(currentValue)
			// Keep index in bounds
			if m.commandMenuIndex >= len(m.filteredCommands) {
				m.commandMenuIndex = len(m.filteredCommands) - 1
			}
			if m.commandMenuIndex < 0 {
				m.commandMenuIndex = 0
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m BubbleModel) View() string {
	if m.quitting {
		// Show goodbye message centered on screen
		goodbye := lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
			Bold(true).
			Width(m.width).
			Height(m.height).
			Align(lipgloss.Center, lipgloss.Center).
			Render("👋 Goodbye, Mr. Anderson...")
		return goodbye
	}

	// Calculate dimensions
	mainWidth := m.width - sidebarWidth - 2
	if mainWidth < 40 {
		mainWidth = 40 // Minimum width
	}

	// Build sidebar
	sidebar := m.renderSidebar()

	// Build main chat area
	mainArea := m.renderMainArea(mainWidth)

	// Join horizontally: main area on left, sidebar on right
	base := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainArea,
		sidebar,
	)

	// Overlay settings panel if active
	if m.showSettings {
		return m.renderSettingsPanel(base)
	}

	return base
}

func (m BubbleModel) renderSidebar() string {
	sections := []string{}

	// Logo/Branding
	logo := sidebarTitleStyle.Render("SMITH")
	sections = append(sections, logo)

	// Session info
	sessionInfo := sidebarSectionStyle.Render("New Session")
	pathInfo := sidebarItemStyle.Render(m.projectPath)
	sections = append(sections, sessionInfo, pathInfo)

	// Model status
	sections = append(sections, "")
	modelInfo := sidebarItemStyle.Render("◇ Claude Sonnet 4")
	thinkingStatus := sidebarIdleStyle.Render("  Thinking Off")
	if m.isStreaming {
		thinkingStatus = sidebarActiveStyle.Render("  Thinking...")
	}
	autoLevel := sidebarItemStyle.Render(fmt.Sprintf("  Auto: %s", m.getAutoLevelDisplay()))
	sections = append(sections, modelInfo, thinkingStatus, autoLevel)

	// Modified Files
	sections = append(sections, "")
	sections = append(sections, sidebarSectionStyle.Render("🗂️ Modified Files"))
	modifiedFiles := m.getModifiedFiles()
	if len(modifiedFiles) == 0 {
		sections = append(sections, sidebarIdleStyle.Render("None"))
	} else {
		for _, file := range modifiedFiles {
			sections = append(sections, sidebarItemStyle.Render("• "+file))
		}
	}

	// Active Agents
	sections = append(sections, "")
	sections = append(sections, sidebarSectionStyle.Render("🤖 Active Agents"))
	agents := m.getActiveAgents()
	if len(agents) == 0 {
		sections = append(sections, sidebarIdleStyle.Render("None"))
	} else {
		for _, agent := range agents {
			if agent.active {
				sections = append(sections, sidebarActiveStyle.Render("● "+agent.name))
			} else {
				sections = append(sections, sidebarIdleStyle.Render("○ "+agent.name))
			}
		}
	}

	// Task counts
	sections = append(sections, "")
	sections = append(sections, sidebarSectionStyle.Render("📋 Tasks"))
	// TODO: Parse task counts from m.engine.GetStatus()
	// For now, show placeholder
	sections = append(sections, sidebarItemStyle.Render("Todo:   0"))
	sections = append(sections, sidebarItemStyle.Render("WIP:    0"))
	sections = append(sections, sidebarItemStyle.Render("Done:   0"))

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return sidebarStyle.
		Height(m.height - 2).
		Render(content)
}

func (m BubbleModel) renderMainArea(width int) string {
	// Message history
	historyHeight := m.height - 6 // Reserve space for input and help
	history := m.renderHistory(historyHeight, width)

	// Input box
	m.textarea.SetWidth(width - 4)
	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 1).
		Width(width - 2).
		Render(m.textarea.View())

	// Command menu (if visible)
	var commandMenu string
	if m.showCommandMenu {
		commandMenu = m.renderCommandMenu(width)
	}

	// Bottom help line
	helpText := helpStyle.Render("esc cancel · tab focus · ctrl+l clear · shift+tab cycle · ctrl+c quit")

	components := []string{history, ""}

	// Add command menu before input if visible
	if m.showCommandMenu {
		components = append(components, commandMenu)
	}

	components = append(components, inputBox, helpText)

	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		components...,
	)

	return lipgloss.NewStyle().
		Width(width).
		Height(m.height).
		Render(mainContent)
}

func (m *BubbleModel) cycleAutoLevel() {
	switch m.autoLevel {
	case safety.AutoLevelLow:
		m.autoLevel = safety.AutoLevelMedium
	case safety.AutoLevelMedium:
		m.autoLevel = safety.AutoLevelHigh
	case safety.AutoLevelHigh:
		m.autoLevel = safety.AutoLevelLow
	default:
		m.autoLevel = safety.AutoLevelMedium
	}

	// Save to config
	localCfg, _ := config.LoadLocal(m.projectPath)
	if localCfg == nil {
		localCfg = &config.LocalConfig{}
	}
	localCfg.AutoLevel = m.autoLevel
	_ = localCfg.SaveLocal(m.projectPath)

	// No notification needed - visible in status bar
}

func (m *BubbleModel) getAutoLevelDisplay() string {
	switch m.autoLevel {
	case safety.AutoLevelLow:
		return "Low"
	case safety.AutoLevelMedium:
		return "Medium"
	case safety.AutoLevelHigh:
		return "High"
	default:
		return "Medium"
	}
}

// Helper methods for sidebar data
type agentInfo struct {
	name   string
	active bool
}

func (m *BubbleModel) getModifiedFiles() []string {
	// TODO: Get from file watcher
	// For now, return empty or mock data
	return []string{}
}

func (m *BubbleModel) getActiveAgents() []agentInfo {
	// TODO: Get from coordinator
	// For now, return mock data showing the planned agents
	return []agentInfo{
		{name: "planning", active: false},
		{name: "implementation", active: false},
		{name: "testing", active: false},
		{name: "review", active: false},
	}
}

func (m *BubbleModel) filterCommands(input string) {
	input = strings.ToLower(strings.TrimSpace(input))
	m.filteredCommands = []Command{}

	for _, cmd := range availableCommands {
		// Match by name or alias
		if strings.HasPrefix(strings.ToLower(cmd.Name), input) ||
			(cmd.Alias != "" && strings.HasPrefix(strings.ToLower(cmd.Alias), input)) {
			m.filteredCommands = append(m.filteredCommands, cmd)
		}
	}
}

func (m *BubbleModel) renderCommandMenu(width int) string {
	if len(m.filteredCommands) == 0 {
		return ""
	}

	menuStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("6")).
		Padding(0, 1).
		Width(width - 2)

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true).
		Background(lipgloss.Color("240"))

	var lines []string
	for i, cmd := range m.filteredCommands {
		var line string
		if cmd.Alias != "" {
			line = fmt.Sprintf("%-10s %-6s %s", cmd.Name, cmd.Alias, cmd.Description)
		} else {
			line = fmt.Sprintf("%-10s        %s", cmd.Name, cmd.Description)
		}

		if i == m.commandMenuIndex {
			line = selectedStyle.Render("> " + line)
		} else {
			line = commandStyle.Render("  " + line)
		}
		lines = append(lines, line)
	}

	navHelp := helpStyle.Render("Use ↑↓ to navigate, Tab/Enter to select, Esc to cancel")
	lines = append(lines, "", navHelp)

	content := strings.Join(lines, "\n")
	return menuStyle.Render(content)
}

func (m *BubbleModel) handleSlashCommand(cmd string) tea.Cmd {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return nil
	}

	switch parts[0] {
	case "/help", "/h":
		m.messages = append(m.messages, Message{Content: renderHelp(), Type: "system", Timestamp: time.Now()})
	case "/status", "/s":
		status := m.engine.GetStatus()
		m.messages = append(m.messages, Message{Content: status, Type: "system", Timestamp: time.Now()})
	case "/copy":
		m.copyLastMessage()
	case "/edit-global":
		m.editGlobalConfig()
	case "/settings":
		m.showSettings = true
		m.settingsIndex = 0
	case "/clear", "/c":
		// Handled in Update
	case "/quit", "/q", "/exit":
		m.quitting = true
		return delayedQuit()
	default:
		m.messages = append(m.messages, Message{Content: "Unknown command: " + parts[0], Type: "error", Timestamp: time.Now()})
		m.messages = append(m.messages, Message{Content: "💡 Type /help to see available commands", Type: "system", Timestamp: time.Now()})
	}

	return nil
}

func (m *BubbleModel) renderHistory(maxHeight int, width int) string {
	if len(m.messages) == 0 {
		return ""
	}

	// Calculate how many messages we can show
	totalMessages := len(m.messages)
	scrollOffset := m.historyScroll

	// Ensure scroll offset is valid
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	if scrollOffset > totalMessages {
		scrollOffset = totalMessages
	}

	// Get visible messages (from scroll position, up to maxHeight)
	start := 0
	if totalMessages > maxHeight {
		start = totalMessages - maxHeight - scrollOffset
		if start < 0 {
			start = 0
		}
	}

	end := totalMessages - scrollOffset
	if end < 0 {
		end = 0
	}
	if end > totalMessages {
		end = totalMessages
	}

	visible := m.messages[start:end]

	// Format messages with timestamps and styling
	var formatted []string
	for _, msg := range visible {
		formattedMsg := m.formatMessage(msg, width)
		formatted = append(formatted, formattedMsg)
	}

	// If the first message is the welcome screen, center it horizontally
	if len(formatted) > 0 && strings.Contains(formatted[0], "███") {
		// Use PlaceHorizontal to center the welcome message in the available width
		centered := lipgloss.PlaceHorizontal(width, lipgloss.Center, formatted[0])

		if len(formatted) > 1 {
			return centered + "\n" + strings.Join(formatted[1:], "\n")
		}
		return centered
	}

	return strings.Join(formatted, "\n")
}

func renderWelcome() string {
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // Matrix green
		Bold(true)

	versionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Align(lipgloss.Center)

	welcomeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Align(lipgloss.Center)

	logo := logoStyle.Render(`
███████╗███╗   ███╗██╗████████╗██╗  ██╗
██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║
███████╗██╔████╔██║██║   ██║   ███████║
╚════██║██║╚██╔╝██║██║   ██║   ██╔══██║
███████║██║ ╚═╝ ██║██║   ██║   ██║  ██║
╚══════╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝`)

	version := versionStyle.Width(46).Render("v0.1.0")
	welcome := welcomeStyle.Width(46).Render("Mr. Anderson... Welcome back.\nI've been expecting you.")

	return lipgloss.JoinVertical(
		lipgloss.Center,
		"",
		logo,
		"",
		version,
		"",
		welcome,
		"",
	)
}

func renderHelp() string {
	help := `
📚 Commands:

  /status   /s  - Show task counts and progress
  /backlog  /b  - List all tasks
  /inbox    /i  - Check agent questions/messages
  /agents   /a  - Show active agents
  /settings     - Show settings
  /copy         - Copy last message to buffer
  /edit-global  - Edit global configuration in editor
  /clear    /c  - Clear conversation
  /help     /h  - Show this help
  /quit     /q  - Exit Smith

⌨️  Keyboard Shortcuts:

  Enter           - Send message
  Tab             - Focus input area
  Shift+Tab       - Cycle reasoning level (Low/Medium/High)
  Ctrl+L          - Clear conversation
  Ctrl+H          - Show this help
  Ctrl+Y          - Copy last message
  Ctrl+C/Ctrl+D   - Quit Smith
  Page Up/Ctrl+U  - Scroll message history up
  Page Down       - Scroll message history down
  ESC             - Cancel/close menus

💡 Just chat naturally - no commands needed to build features!
More agents, Mr. Anderson... always more agents.
`
	return helpStyle.Render(help)
}

func (m *BubbleModel) renderSettings() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true)

	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	// Get current config
	cfg, err := config.Load()
	if err != nil {
		return errorStyle.Render(fmt.Sprintf("Error loading config: %v", err))
	}

	// Get config directory
	configDir, _ := config.GetConfigDir()

	var settings strings.Builder

	settings.WriteString(titleStyle.Render("Settings") + "\n\n")

	// Model & Reasoning section
	settings.WriteString(sectionStyle.Render("Model & Reasoning") + "\n")
	settings.WriteString(labelStyle.Render("> Model: ") + valueStyle.Render(cfg.Model) + "\n")
	settings.WriteString(labelStyle.Render("  Reasoning level: ") + m.autoLevel + "\n\n")

	// Preferences section
	settings.WriteString(sectionStyle.Render("Preferences") + "\n")
	settings.WriteString(labelStyle.Render("  Completion bell: Off\n"))
	settings.WriteString(labelStyle.Render("  Edit AllowList & DenyList: Open in editor\n\n"))

	// Config location
	settings.WriteString(sectionStyle.Render("Configuration") + "\n")
	settings.WriteString(labelStyle.Render("  Config directory: ") + configDir + "\n\n")

	settings.WriteString(footerStyle.Render("↑/↓ to navigate, Enter to select, ESC to exit"))

	return settings.String()
}

func (m *BubbleModel) renderSettingsPanel(base string) string {
	// Show provider menu if active
	if m.showProviderMenu {
		return m.renderProviderMenu(base)
	}

	// Show model menu if active
	if m.showModelMenu {
		return m.renderModelMenu(base)
	}

	// Get current configs
	cfg, err := config.Load()
	if err != nil {
		return base
	}

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("236")).
		Bold(true)

	globalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("11")) // Yellow for global

	localStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("14")) // Cyan for local

	// Settings items - now showing global vs local
	items := []string{
		"Provider: " + cfg.Provider + " (global)",
		"Model: " + cfg.Model + " (global)",
		"Reasoning: " + m.autoLevel + " (local)",
		"",
		"[Edit global config in editor]",
		"[Edit safety rules]",
		"[Open config directory]",
	}

	var content strings.Builder
	content.WriteString(titleStyle.Render("⚙️ Settings") + "\n\n")

	content.WriteString(sectionStyle.Render("Configuration") + "\n")
	for i := 0; i < 3; i++ {
		var displayText string
		if i == m.settingsIndex {
			if strings.Contains(items[i], "(global)") {
				displayText = selectedStyle.Render("> " + items[i])
			} else {
				displayText = selectedStyle.Render("> " + items[i])
			}
		} else {
			if strings.Contains(items[i], "(global)") {
				displayText = globalStyle.Render("  " + items[i])
			} else {
				displayText = localStyle.Render("  " + items[i])
			}
		}
		content.WriteString(displayText + "\n")
	}

	content.WriteString("\n" + sectionStyle.Render("Actions") + "\n")
	for i := 4; i < len(items); i++ {
		if i == m.settingsIndex {
			content.WriteString(selectedStyle.Render("> "+items[i]) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  "+items[i]) + "\n")
		}
	}

	content.WriteString("\n")
	content.WriteString(helpStyle.Render("↑/↓ navigate · enter select · esc/q close"))

	// Create panel
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(60).
		Render(content.String())

	// Overlay on center of screen
	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		panel,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m *BubbleModel) renderProviderMenu(base string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("236")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	providers := llm.GetAvailableProviders()

	var content strings.Builder
	content.WriteString(titleStyle.Render("Select Provider") + "\n\n")

	for i, provider := range providers {
		if i == m.providerMenuIndex {
			content.WriteString(selectedStyle.Render("> "+provider.Name) + "\n")
			content.WriteString(descStyle.Render("  "+provider.Description) + "\n\n")
		} else {
			content.WriteString(itemStyle.Render("  "+provider.Name) + "\n")
			content.WriteString(descStyle.Render("  "+provider.Description) + "\n\n")
		}
	}

	content.WriteString(helpStyle.Render("↑/↓ navigate · enter select · esc cancel"))

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(70).
		Render(content.String())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		panel,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

func (m *BubbleModel) renderModelMenu(base string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("236")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	cfg, _ := config.Load()
	provider, err := llm.NewProviderByID(cfg.Provider)
	if err != nil {
		return base
	}

	models, err := provider.GetModels()
	if err != nil {
		return base
	}

	var content strings.Builder
	content.WriteString(titleStyle.Render("Select Model") + "\n")
	content.WriteString(descStyle.Render("Provider: "+provider.GetName()) + "\n\n")

	for i, model := range models {
		if i == m.modelMenuIndex {
			content.WriteString(selectedStyle.Render("> "+model.Name) + "\n")
			content.WriteString(descStyle.Render("  "+model.Description) + "\n\n")
		} else {
			content.WriteString(itemStyle.Render("  "+model.Name) + "\n")
			content.WriteString(descStyle.Render("  "+model.Description) + "\n\n")
		}
	}

	content.WriteString(helpStyle.Render("↑/↓ navigate · enter select · esc cancel"))

	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(70).
		Render(content.String())

	return lipgloss.Place(
		m.width,
		m.height,
		lipgloss.Center,
		lipgloss.Center,
		panel,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("0")),
	)
}

// formatMessage formats a message with timestamp and appropriate styling
func (m *BubbleModel) formatMessage(msg Message, width int) string {
	// Format timestamp
	timeStr := msg.Timestamp.Format("15:04:05")

	// Create timestamp style
	timeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)

	// Format based on message type
	switch msg.Type {
	case "user":
		return promptStyle.Render("💬 You: ") + msg.Content + " " + timeStyle.Render("["+timeStr+"]")
	case "ai":
		return aiStyle.Render("🤖 AI: ") + msg.Content + " " + timeStyle.Render("["+timeStr+"]")
	case "system":
		return msg.Content // System messages (like welcome) don't need timestamp
	case "error":
		return errorStyle.Render("❌ "+msg.Content) + " " + timeStyle.Render("["+timeStr+"]")
	default:
		return msg.Content + " " + timeStyle.Render("["+timeStr+"]")
	}
}

// copyLastMessage shows the last message in a copy-friendly format
func (m *BubbleModel) copyLastMessage() {
	if len(m.messages) == 0 {
		return
	}

	// Find the last non-system message
	for i := len(m.messages) - 1; i >= 0; i-- {
		msg := m.messages[i]
		if msg.Type != "system" {
			// Show message in a format easy to copy
			copyMsg := fmt.Sprintf("\n--- Copy Buffer ---\n%s\n--- End Copy ---\n", msg.Content)
			m.messages = append(m.messages, Message{
				Content:   copyMsg,
				Type:      "system",
				Timestamp: time.Now(),
			})
			return
		}
	}
}

// editGlobalConfig opens the global config file in the user's editor
func (m *BubbleModel) editGlobalConfig() {
	configDir, err := config.GetConfigDir()
	if err != nil {
		m.messages = append(m.messages, Message{Content: fmt.Sprintf("Error getting config directory: %v", err), Type: "error", Timestamp: time.Now()})
		return
	}

	configPath := filepath.Join(configDir, "config.json")

	// Try to open in editor
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano" // fallback
	}

	cmd := exec.Command(editor, configPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		m.messages = append(m.messages, Message{Content: fmt.Sprintf("Error opening editor: %v", err), Type: "error", Timestamp: time.Now()})
		return
	}

	m.messages = append(m.messages, Message{Content: "Global config edited. Restart Smith to apply changes.", Type: "system", Timestamp: time.Now()})
}
