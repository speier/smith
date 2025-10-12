package repl

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/speier/smith/internal/config"
	"github.com/speier/smith/internal/engine"
	"github.com/speier/smith/internal/safety"
)

// delayedQuitMsg is sent after showing the goodbye message
type delayedQuitMsg struct{} // Styles
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
	{Name: "/auth", Alias: "", Description: "Authenticate with GitHub Copilot"},
	{Name: "/clear", Alias: "/c", Description: "Clear conversation"},
	{Name: "/help", Alias: "/h", Description: "Show help"},
	{Name: "/quit", Alias: "/q", Description: "Exit Smith"},
}

// BubbleModel is our Bubble Tea model
type BubbleModel struct {
	engine           *engine.Engine
	textarea         textarea.Model
	messages         []string
	err              error
	width            int
	height           int
	autoLevel        string
	projectPath      string
	quitting         bool
	showCommandMenu  bool
	commandMenuIndex int
	filteredCommands []Command
}

// keyMap defines our key bindings
type keyMap struct {
	Send       key.Binding
	Quit       key.Binding
	CycleLevel key.Binding
	Help       key.Binding
	Clear      key.Binding
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

	eng, err := engine.New(engine.Config{
		ProjectPath: projectPath,
	})
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	// Load auto-level from config
	localCfg, _ := config.LoadLocal(projectPath)
	autoLevel := safety.AutoLevelMedium
	if localCfg != nil && localCfg.AutoLevel != "" {
		autoLevel = localCfg.AutoLevel
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

	messages := []string{}

	// Add welcome message
	welcome := renderWelcome()
	messages = append(messages, welcome)

	// If initial prompt provided, add it
	if initialPrompt != "" {
		ti.SetValue(initialPrompt)
	}

	return &BubbleModel{
		engine:      eng,
		textarea:    ti,
		messages:    messages,
		autoLevel:   autoLevel,
		projectPath: projectPath,
		width:       80, // Default width until we get WindowSizeMsg
		height:      24, // Default height
	}, nil
}

func (m BubbleModel) Init() tea.Cmd {
	return textarea.Blink
}

// delayedQuit returns a command that waits a moment then quits
func delayedQuit() tea.Cmd {
	return tea.Tick(time.Millisecond*1500, func(t time.Time) tea.Msg {
		return delayedQuitMsg{}
	})
}

func (m BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case delayedQuitMsg:
		// Now actually quit after showing the goodbye message
		return m, tea.Quit

	case tea.KeyMsg:
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
			case "esc":
				// Cancel command menu
				m.showCommandMenu = false
				m.filteredCommands = nil
				m.commandMenuIndex = 0
				return m, nil
			}
		}

		switch {
		case key.Matches(msg, keys.Quit):
			m.quitting = true
			return m, delayedQuit()

		case key.Matches(msg, keys.CycleLevel):
			m.cycleAutoLevel()
			return m, nil

		case key.Matches(msg, keys.Clear):
			m.messages = []string{}
			m.messages = append(m.messages, renderWelcome())
			m.textarea.Reset()
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
			m.messages = append(m.messages, promptStyle.Render("💬 You: ")+userMsg)

			// Send to engine
			response, err := m.engine.Chat(userMsg)
			if err != nil {
				m.messages = append(m.messages, errorStyle.Render("❌ Error: ")+err.Error())
			} else {
				m.messages = append(m.messages, aiStyle.Render("🤖 AI: ")+response)
			}

			// Clear textarea
			m.textarea.Reset()
			return m, nil
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textarea.SetWidth(msg.Width - 4)
	}

	// Don't update textarea if quitting
	if m.quitting {
		return m, tea.Quit
	}

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

	return m, tea.Batch(cmds...)
}

func (m BubbleModel) View() string {
	if m.quitting {
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
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		mainArea,
		sidebar,
	)
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
		{name: "Planning", active: false},
		{name: "Implementation", active: false},
		{name: "Testing", active: false},
		{name: "Review", active: false},
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
		m.messages = append(m.messages, renderHelp())
	case "/status", "/s":
		status := m.engine.GetStatus()
		m.messages = append(m.messages, status)
	case "/settings":
		m.messages = append(m.messages, m.renderSettings())
	case "/clear", "/c":
		// Handled in Update
	case "/auth":
		m.messages = append(m.messages, "🔐 Authentication: Use /help for auth commands (TODO)")
	case "/quit", "/q", "/exit":
		m.quitting = true
		return delayedQuit()
	default:
		m.messages = append(m.messages, errorStyle.Render("❌ Unknown command: ")+parts[0])
		m.messages = append(m.messages, helpStyle.Render("💡 Type /help to see available commands"))
	}

	return nil
}

func (m *BubbleModel) renderHistory(maxHeight int, width int) string {
	if len(m.messages) == 0 {
		return ""
	}

	// Get last N messages that fit in height
	start := 0
	if len(m.messages) > maxHeight {
		start = len(m.messages) - maxHeight
	}

	visible := m.messages[start:]

	// If the first message is the welcome screen, center it horizontally
	if len(visible) > 0 && strings.Contains(visible[0], "███") {
		// Use PlaceHorizontal to center the welcome message in the available width
		// The height will remain as the welcome message's natural height
		centered := lipgloss.PlaceHorizontal(width, lipgloss.Center, visible[0])

		if len(visible) > 1 {
			return centered + "\n" + strings.Join(visible[1:], "\n")
		}
		return centered
	}

	return strings.Join(visible, "\n")
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
  /auth         - Authenticate with GitHub Copilot
  /clear    /c  - Clear conversation
  /help     /h  - Show this help
  /quit     /q  - Exit Smith

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
