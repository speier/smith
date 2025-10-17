package repl

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/speier/smith/internal/agent"
	"github.com/speier/smith/internal/config"
	"github.com/speier/smith/internal/coordinator"
	"github.com/speier/smith/internal/engine"
	"github.com/speier/smith/internal/llm"
	"github.com/speier/smith/internal/safety"
)

// delayedQuitMsg is sent after showing the goodbye message
type delayedQuitMsg struct{}

// tickMsg is sent periodically to update animations
type tickMsg struct{}

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

// deviceCodeMsg is sent when device code authorization starts
type deviceCodeMsg struct {
	verificationURI string
	userCode        string
	deviceCode      string
	interval        int
	expiresIn       int
}

// authCompleteMsg is sent when authentication completes
type authCompleteMsg struct {
	success bool
	error   string
}

// pollTokenMsg is sent to trigger the next poll
type pollTokenMsg struct {
	deviceCode string
	interval   int
	deadline   time.Time
}

// modelsFetchedMsg is sent when models are fetched from API
type modelsFetchedMsg struct {
	models []llm.Model
	err    error
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
	{Name: "/sessions", Alias: "", Description: "Switch to a different session"},
	{Name: "/settings", Alias: "", Description: "Change model and auto-level"},
	{Name: "/allowlist", Alias: "", Description: "Manage session command allowlist"},
	{Name: "/copy", Alias: "", Description: "Copy last message to clipboard"},
	{Name: "/clear", Alias: "/c", Description: "Clear and start new session"},
	{Name: "/help", Alias: "/h", Description: "Show help"},
	{Name: "/quit", Alias: "/q", Description: "Exit Smith"},
}

// BubbleModel is our Bubble Tea model
type BubbleModel struct {
	engine             *engine.Engine
	textarea           textarea.Model
	messages           []Message
	width              int
	height             int
	autoLevel          string
	projectPath        string
	quitting           bool
	agentCtx           context.Context
	agentCancel        context.CancelFunc
	agents             []agent.Agent // Background agents
	streamingResponse  string        // Current streaming response being built
	isStreaming        bool          // Whether we're currently streaming a response
	spinnerFrame       int           // Frame counter for loading spinner
	showCommandMenu    bool
	commandMenuIndex   int
	filteredCommands   []Command
	showSettings       bool
	settingsIndex      int
	showProviderMenu   bool
	providerMenuIndex  int
	showModelMenu      bool
	modelMenuIndex     int
	modelMenuScroll    int         // Scroll offset for model menu paging
	modelFilter        string      // Type-to-search filter for models
	fetchingModels     bool        // Whether we're currently fetching models
	cachedModels       []llm.Model // Cached models list
	historyScroll      int         // Current scroll position in message history
	needsProviderSetup bool        // Whether provider needs to be selected on first run
	showHelp           bool        // Whether help panel is visible
	showAuthPanel      bool        // Whether auth panel is visible
	authDeviceCode     string      // Device code for auth
	authUserCode       string      // User code for auth
	authVerifyURI      string      // Verification URI for auth
	// Approval state
	// Error tracking
	recentErrors []ErrorInfo // Last 3 errors for sidebar display
	// Session management
	showSessionPicker  bool                   // Whether session picker is visible
	sessionsList       []*coordinator.Session // List of available sessions
	sessionPickerIndex int                    // Selected session in picker
}

// ErrorInfo represents an error with context for display
type ErrorInfo struct {
	Message   string
	Timestamp time.Time
	AgentRole string
	TaskID    string
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
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new session"),
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

	// Track if we need provider setup - when no config OR when provider/model are empty
	needsProviderSetup := localCfg == nil || localCfg.Provider == "" || localCfg.Model == ""

	// Use empty provider if not yet configured (user must select in settings)
	providerID := ""
	autoLevel := safety.AutoLevelMedium

	if localCfg != nil {
		providerID = localCfg.Provider
		if localCfg.AutoLevel != "" {
			autoLevel = localCfg.AutoLevel
		}
	}

	// Only create provider if configured
	var provider llm.Provider
	if providerID != "" {
		var provErr error
		provider, provErr = llm.NewProviderByID(providerID)
		if provErr != nil {
			return nil, fmt.Errorf("creating provider: %w", provErr)
		}
	} else {
		// No provider configured - use copilot as fallback to avoid nil
		// User will be prompted to configure via settings
		provider = llm.NewCopilotProvider()
	}

	eng, err := engine.New(engine.Config{
		ProjectPath: projectPath,
		LLMProvider: provider,
		AutoLevel:   autoLevel,
	})
	if err != nil {
		return nil, fmt.Errorf("creating engine: %w", err)
	}

	// Create agent context that will be cancelled when REPL quits
	agentCtx, agentCancel := context.WithCancel(context.Background())

	// Start background agents
	coord := eng.GetCoordinator()

	// Get registry through the interface (storage-agnostic)
	reg := coord.GetRegistry()

	// Create and start agents in background goroutines
	// Stagger start times to reduce database contention
	var agents []agent.Agent

	// Implementation agent (Keymaker) - starts immediately
	implAgent := agent.NewImplementationAgent(agent.Config{
		AgentID:      "keymaker-001",
		Coordinator:  coord,
		Registry:     reg,
		Engine:       eng,
		PollInterval: 1 * time.Second, // Increased to reduce DB contention
	})
	agents = append(agents, implAgent)
	go func() {
		if err := implAgent.Start(agentCtx); err != nil && err != context.Canceled {
			log.Printf("Implementation agent error: %v", err)
		}
	}()

	// Testing agent (Sentinel) - starts after 250ms
	testAgent := agent.NewTestingAgent(agent.Config{
		AgentID:      "sentinel-001",
		Coordinator:  coord,
		Registry:     reg,
		Engine:       eng,
		PollInterval: 1 * time.Second,
	})
	agents = append(agents, testAgent)
	go func() {
		time.Sleep(250 * time.Millisecond) // Stagger start
		if err := testAgent.Start(agentCtx); err != nil && err != context.Canceled {
			log.Printf("Testing agent error: %v", err)
		}
	}()

	// Review agent (Oracle) - starts after 500ms
	reviewAgent := agent.NewReviewAgent(agent.Config{
		AgentID:      "oracle-001",
		Coordinator:  coord,
		Registry:     reg,
		Engine:       eng,
		PollInterval: 1 * time.Second,
	})
	agents = append(agents, reviewAgent)
	go func() {
		time.Sleep(500 * time.Millisecond) // Stagger start
		if err := reviewAgent.Start(agentCtx); err != nil && err != context.Canceled {
			log.Printf("Review agent error: %v", err)
		}
	}()

	// Planning agent (Architect) - starts after 750ms
	planAgent := agent.NewPlanningAgent(agent.Config{
		AgentID:      "architect-001",
		Coordinator:  coord,
		Registry:     reg,
		Engine:       eng,
		PollInterval: 1 * time.Second,
	})
	agents = append(agents, planAgent)
	go func() {
		time.Sleep(750 * time.Millisecond) // Stagger start
		if err := planAgent.Start(agentCtx); err != nil && err != context.Canceled {
			log.Printf("Planning agent error: %v", err)
		}
	}()

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

	// If provider setup needed, show setup hint (don't auto-open settings)
	if needsProviderSetup {
		messages = append(messages, Message{
			Content:   "⚙️ No configuration found. Please select your LLM provider and model:\n   → Type /settings to choose a provider and model",
			Type:      "system",
			Timestamp: time.Now(),
		})
	}

	// If initial prompt provided, add it
	if initialPrompt != "" {
		ti.SetValue(initialPrompt)
	}

	model := &BubbleModel{
		engine:             eng,
		textarea:           ti,
		messages:           messages,
		autoLevel:          autoLevel,
		projectPath:        projectPath,
		needsProviderSetup: needsProviderSetup,
		showSettings:       false, // Don't auto-open
		width:              80,    // Default width until we get WindowSizeMsg
		height:             24,    // Default height
		agentCtx:           agentCtx,
		agentCancel:        agentCancel,
		agents:             agents,
	}

	// Log agent startup
	log.Printf("🤖 Started %d background agents", len(agents))

	return model, nil
}

// filterModels filters models by name or ID based on the filter string
func (m BubbleModel) filterModels(models []llm.Model, filter string) []llm.Model {
	if filter == "" || models == nil {
		return models
	}

	filter = strings.ToLower(filter)
	var filtered []llm.Model
	for _, model := range models {
		if strings.Contains(strings.ToLower(model.Name), filter) ||
			strings.Contains(strings.ToLower(model.ID), filter) {
			filtered = append(filtered, model)
		}
	}
	return filtered
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

// tick returns a command that sends periodic tick messages for animations
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

func (m BubbleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tickMsg:
		// Update spinner animation if streaming
		if m.isStreaming {
			m.spinnerFrame++
			return m, tick()
		}
		return m, nil

	case streamChunkMsg:
		if !m.isStreaming {
			return m, nil
		}

		m.streamingResponse += msg.chunk

		if msg.done {
			// Streaming complete
			m.isStreaming = false
			m.messages = append(m.messages, Message{Content: aiStyle.Render("🕶️ Smith: ") + m.streamingResponse, Type: "ai", Timestamp: time.Now()})
			m.streamingResponse = ""
		}

		return m, nil

	case deviceCodeMsg:
		// Store auth info and show auth panel
		m.authDeviceCode = msg.deviceCode
		m.authUserCode = msg.userCode
		m.authVerifyURI = msg.verificationURI
		m.showAuthPanel = true

		// Close provider menu now that auth panel is showing
		m.showProviderMenu = false
		m.providerMenuIndex = 0

		// Copy auth code to clipboard
		go func() {
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("pbcopy")
			case "linux":
				cmd = exec.Command("xclip", "-selection", "clipboard")
			case "windows":
				cmd = exec.Command("clip")
			}
			if cmd != nil {
				stdin, err := cmd.StdinPipe()
				if err == nil {
					_ = cmd.Start()
					_, _ = stdin.Write([]byte(msg.userCode))
					_ = stdin.Close()
					_ = cmd.Wait()
				}
			}
		}()

		// Open browser automatically
		go func() {
			// Try to open browser on macOS/Linux/Windows
			var cmd *exec.Cmd
			switch runtime.GOOS {
			case "darwin":
				cmd = exec.Command("open", msg.verificationURI)
			case "linux":
				cmd = exec.Command("xdg-open", msg.verificationURI)
			case "windows":
				cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", msg.verificationURI)
			}
			if cmd != nil {
				_ = cmd.Run()
			}
		}()

		// Calculate deadline for timeout
		timeout := time.Duration(msg.expiresIn) * time.Second
		deadline := time.Now().Add(timeout)

		// Start polling loop with first poll after interval
		return m, tea.Tick(time.Duration(msg.interval)*time.Second, func(t time.Time) tea.Msg {
			return pollTokenMsg{
				deviceCode: msg.deviceCode,
				interval:   msg.interval,
				deadline:   deadline,
			}
		})

	case pollTokenMsg:
		// Check if we've timed out
		if time.Now().After(msg.deadline) {
			m.showAuthPanel = false
			m.messages = append(m.messages, Message{Content: "❌ Authentication timed out", Type: "error", Timestamp: time.Now()})
			return m, nil
		}

		// Poll for token
		return m, func() tea.Msg {
			provider := llm.NewCopilotProvider()
			token, err := provider.PollForToken(msg.deviceCode)
			if err != nil {
				return authCompleteMsg{
					success: false,
					error:   fmt.Sprintf("polling error: %v", err),
				}
			}

			if token == "pending" {
				// Still waiting, poll again after interval
				return tea.Tick(time.Duration(msg.interval)*time.Second, func(t time.Time) tea.Msg {
					return pollTokenMsg{
						deviceCode: msg.deviceCode,
						interval:   msg.interval,
						deadline:   msg.deadline,
					}
				})()
			}

			// Success! Save the token
			if err := provider.SetAuth(token); err != nil {
				return authCompleteMsg{
					success: false,
					error:   fmt.Sprintf("failed to save auth: %v", err),
				}
			}

			return authCompleteMsg{success: true}
		}

	case authCompleteMsg:
		// Close auth panel
		m.showAuthPanel = false

		if msg.success {
			m.messages = append(m.messages, Message{Content: "✅ Successfully authenticated with GitHub Copilot!", Type: "system", Timestamp: time.Now()})

			// Auto-open model menu and start fetching models
			m.showModelMenu = true
			m.modelMenuIndex = 0
			m.fetchingModels = true
			return m, m.fetchModels()
		} else {
			m.messages = append(m.messages, Message{Content: fmt.Sprintf("❌ Authentication failed: %s", msg.error), Type: "error", Timestamp: time.Now()})
		}
		return m, nil

	case modelsFetchedMsg:
		// Models fetched
		m.fetchingModels = false
		if msg.err != nil {
			// Keep error in cache to show in UI
			m.cachedModels = nil
		} else {
			m.cachedModels = msg.models
		}
		return m, nil

	case delayedQuitMsg:
		// Stop background agents before quitting
		if m.agentCancel != nil {
			m.agentCancel()
			log.Printf("🛑 Stopped %d background agents", len(m.agents))
		}
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
			if m.showAuthPanel {
				m.showAuthPanel = false
				return m, nil
			}
			if m.showSessionPicker {
				m.showSessionPicker = false
				m.sessionPickerIndex = 0
				return m, nil
			}
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
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
				return m, textarea.Blink
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

		// Toggle help with ?
		if msg.String() == "?" && !m.showSettings && !m.showProviderMenu && !m.showModelMenu {
			m.showHelp = !m.showHelp
			return m, nil
		}

		// Handle session picker
		if m.showSessionPicker {
			switch msg.String() {
			case "up", "k":
				if m.sessionPickerIndex > 0 {
					m.sessionPickerIndex--
				}
				return m, nil
			case "down", "j":
				if m.sessionPickerIndex < len(m.sessionsList)-1 {
					m.sessionPickerIndex++
				}
				return m, nil
			case "enter":
				// Switch to selected session
				if m.sessionPickerIndex < len(m.sessionsList) {
					selectedSession := m.sessionsList[m.sessionPickerIndex]
					ctx := context.Background()
					coord := m.engine.GetCoordinator()

					err := coord.SwitchSession(ctx, selectedSession.SessionID)
					if err != nil {
						m.messages = append(m.messages, Message{
							Content:   fmt.Sprintf("❌ Failed to switch session: %v", err),
							Type:      "error",
							Timestamp: time.Now(),
						})
					} else {
						// Clear messages and show switch confirmation
						m.messages = []Message{}
						m.messages = append(m.messages, Message{
							Content:   fmt.Sprintf("📂 Switched to session: %s", selectedSession.Title),
							Type:      "system",
							Timestamp: time.Now(),
						})
					}

					// Close picker
					m.showSessionPicker = false
					m.sessionPickerIndex = 0
					m.sessionsList = nil
				}
				return m, nil
			}
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
					_ = cfg.Save()

					// If provider requires auth, start auth flow and keep settings visible until auth starts
					if selectedProvider.RequiresAuth && selectedProvider.ID == "copilot" {
						// Don't close provider menu yet - auth panel will take over
						return m, m.startCopilotAuth()
					}

					// Close menu and show provider changed message
					m.showProviderMenu = false
					m.providerMenuIndex = 0
					m.messages = append(m.messages, Message{Content: fmt.Sprintf("✓ Provider changed to %s", selectedProvider.Name), Type: "system", Timestamp: time.Now()})
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
			// Use cached models if available
			allModels := m.cachedModels
			if allModels == nil {
				// Try to get from provider
				cfg, _ := config.Load()
				provider, err := llm.NewProviderByID(cfg.Provider)
				if err == nil {
					allModels, _ = provider.GetModels()
				}
			}

			// Apply filter
			models := m.filterModels(allModels, m.modelFilter)

			const pageSize = 10 // Fixed page size

			switch msg.String() {
			case "backspace":
				// Remove last character from filter
				if len(m.modelFilter) > 0 {
					m.modelFilter = m.modelFilter[:len(m.modelFilter)-1]
					m.modelMenuIndex = 0
					m.modelMenuScroll = 0
				}
				return m, nil
			case "up", "k":
				if len(models) > 0 {
					if m.modelMenuIndex > 0 {
						m.modelMenuIndex--
					} else {
						// Cycle to end
						m.modelMenuIndex = len(models) - 1
					}
					// Adjust scroll to show current item
					m.modelMenuScroll = (m.modelMenuIndex / pageSize) * pageSize
				}
				return m, nil
			case "down", "j":
				if len(models) > 0 {
					if m.modelMenuIndex < len(models)-1 {
						m.modelMenuIndex++
					} else {
						// Cycle to beginning
						m.modelMenuIndex = 0
					}
					// Adjust scroll to show current item
					m.modelMenuScroll = (m.modelMenuIndex / pageSize) * pageSize
				}
				return m, nil
			case "enter":
				// Select model and save to local project config
				if len(models) > 0 && m.modelMenuIndex < len(models) {
					localCfg, _ := config.LoadLocal(m.projectPath)
					if localCfg == nil {
						// Create minimal config
						cfg, _ := config.Load()
						localCfg = &config.LocalConfig{
							Provider:  cfg.Provider,
							Model:     "",
							AutoLevel: "medium",
							Version:   1,
						}
					}
					selectedModel := models[m.modelMenuIndex]
					localCfg.Model = selectedModel.ID
					_ = localCfg.SaveLocal(m.projectPath)
					m.messages = append(m.messages, Message{Content: fmt.Sprintf("✓ Model changed to %s", selectedModel.Name), Type: "system", Timestamp: time.Now()})
					m.showModelMenu = false
					m.modelMenuIndex = 0
					m.modelMenuScroll = 0
					m.modelFilter = ""
				}
				return m, nil
			case "q", "esc":
				// Close menu
				m.showModelMenu = false
				m.modelMenuIndex = 0
				m.modelMenuScroll = 0
				m.modelFilter = ""
				return m, nil
			default:
				// Type to filter - only single printable characters
				if len(msg.String()) == 1 {
					char := msg.String()
					// Check if it's a printable character (not a control char)
					if char >= " " && char <= "~" {
						m.modelFilter += strings.ToLower(char)
						m.modelMenuIndex = 0
						m.modelMenuScroll = 0
					}
				}
				return m, nil
			}
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
					m.fetchingModels = true
					return m, m.fetchModels()
				case 2: // Reasoning (local)
					m.cycleAutoLevel()
				case 4: // Edit safety rules
					// TODO: Open safety rules editor
				case 5: // Open config directory
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
			// Start new session (Ctrl+N)
			ctx := context.Background()
			coord := m.engine.GetCoordinator()
			sessionID, err := coord.CreateNewSession(ctx)
			if err != nil {
				m.messages = append(m.messages, Message{
					Content:   fmt.Sprintf("❌ Failed to create new session: %v", err),
					Type:      "error",
					Timestamp: time.Now(),
				})
				return m, nil
			}
			m.messages = []Message{}
			m.messages = append(m.messages, Message{
				Content:   fmt.Sprintf("🆕 Started new session: %s", sessionID),
				Type:      "system",
				Timestamp: time.Now(),
			})
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

			// Check if model is configured before sending
			localCfg, _ := config.LoadLocal(m.projectPath)
			if localCfg == nil || localCfg.Model == "" {
				m.messages = append(m.messages, Message{
					Content:   "🕶️ I'd love to help, but I need a brain first!\n   → Type /settings to pick a model\n   → Then we can chat about building amazing things ✨",
					Type:      "error",
					Timestamp: time.Now(),
				})
				m.textarea.Reset()
				return m, nil
			} // Add user message to display
			m.messages = append(m.messages, Message{Content: userMsg, Type: "user", Timestamp: time.Now()})

			// Start streaming response with loading indicator
			m.isStreaming = true
			m.streamingResponse = ""
			m.spinnerFrame = 0
			m.messages = append(m.messages, Message{Content: "⏳ Thinking...", Type: "ai", Timestamp: time.Now()})

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

			// Clear textarea and start tick animation
			m.textarea.Reset()
			return m, tick()
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
	if !m.showSettings && !m.showProviderMenu && !m.showModelMenu && !m.showSessionPicker {
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
			Render("💊 Goodbye, Mr. Anderson...")
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

	// Overlay auth panel if active
	if m.showAuthPanel {
		return m.renderAuthPanel(base)
	}

	// Overlay session picker if active
	if m.showSessionPicker {
		return m.renderSessionPicker(base)
	}

	// Overlay help panel if active
	if m.showHelp {
		return m.renderHelpPanel(base)
	}

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

	// Current Session info
	ctx := context.Background()
	coord := m.engine.GetCoordinator()
	session, err := coord.GetCurrentSession(ctx)

	if err == nil && session != nil {
		sessionTitle := session.Title
		if len(sessionTitle) > 25 {
			sessionTitle = sessionTitle[:22] + "..."
		}

		// Calculate elapsed time since session started
		elapsed := time.Since(session.StartedAt)
		var timeStr string
		if elapsed < time.Minute {
			timeStr = "just now"
		} else if elapsed < time.Hour {
			mins := int(elapsed.Minutes())
			timeStr = fmt.Sprintf("%dm ago", mins)
		} else if elapsed < 24*time.Hour {
			hours := int(elapsed.Hours())
			timeStr = fmt.Sprintf("%dh ago", hours)
		} else {
			days := int(elapsed.Hours() / 24)
			timeStr = fmt.Sprintf("%dd ago", days)
		}

		sessionInfo := sidebarSectionStyle.Render("📂 " + sessionTitle)
		sessionDetails := sidebarItemStyle.Render(fmt.Sprintf("Started %s • %d tasks", timeStr, session.TaskCount))
		sections = append(sections, sessionInfo, sessionDetails)

		// Show token usage for this session
		if usage, err := coord.GetSessionUsage(ctx, session.SessionID); err == nil && usage != nil && usage.TotalTokens > 0 {
			var tokenStr string
			if usage.TotalTokens < 1000 {
				tokenStr = fmt.Sprintf("%d", usage.TotalTokens)
			} else {
				tokenStr = fmt.Sprintf("%.1fk", float64(usage.TotalTokens)/1000)
			}
			usageInfo := sidebarItemStyle.Render(fmt.Sprintf("💰 Tokens: %s", tokenStr))
			sections = append(sections, usageInfo)
		}
	} else {
		// Fallback if session not available
		sessionInfo := sidebarSectionStyle.Render("New Session")
		sections = append(sections, sessionInfo)
	}

	pathInfo := sidebarItemStyle.Render(m.projectPath)
	sections = append(sections, pathInfo)

	// Model status
	sections = append(sections, "")

	// Get current model from local config
	modelName := "not configured"
	localCfg, _ := config.LoadLocal(m.projectPath)
	if localCfg != nil && localCfg.Model != "" {
		modelName = localCfg.Model
	}

	modelInfo := sidebarItemStyle.Render(fmt.Sprintf("◇ %s", modelName))
	thinkingStatus := sidebarIdleStyle.Render("  Thinking Off")
	if m.isStreaming {
		thinkingStatus = sidebarActiveStyle.Render("  Thinking...")
	}
	autoLevel := sidebarItemStyle.Render(fmt.Sprintf("  Auto: %s", m.getAutoLevelDisplay()))
	sections = append(sections, modelInfo, thinkingStatus, autoLevel)

	// Modified Files
	sections = append(sections, "")
	sections = append(sections, sidebarSectionStyle.Render("🗂️  Modified Files"))
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
	sections = append(sections, sidebarSectionStyle.Render("🕶️  Active Agents"))
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

	// Get real-time task stats from engine
	taskStats, err := m.engine.GetCoordinator().GetTaskStats()
	if err != nil || taskStats == nil {
		sections = append(sections, sidebarIdleStyle.Render("No active tasks"))
	} else {
		// Show task counts with visual indicators
		if taskStats.Backlog > 0 {
			sections = append(sections, sidebarItemStyle.Render(fmt.Sprintf("📥 Backlog: %d", taskStats.Backlog)))
		}
		if taskStats.WIP > 0 {
			sections = append(sections, sidebarActiveStyle.Render(fmt.Sprintf("🔄 Active:  %d", taskStats.WIP)))
		}
		if taskStats.Review > 0 {
			sections = append(sections, sidebarItemStyle.Render(fmt.Sprintf("👀 Review:  %d", taskStats.Review)))
		}
		if taskStats.Done > 0 {
			sections = append(sections, sidebarItemStyle.Render(fmt.Sprintf("✅ Done:    %d", taskStats.Done)))
		}

		// If no tasks at all
		if taskStats.Backlog == 0 && taskStats.WIP == 0 && taskStats.Review == 0 && taskStats.Done == 0 {
			sections = append(sections, sidebarIdleStyle.Render("No tasks yet"))
		}
	}

	// Active Tasks Details (show what agents are working on)
	if taskStats != nil && taskStats.WIP > 0 {
		sections = append(sections, "")
		sections = append(sections, sidebarSectionStyle.Render("⚡  Working On"))

		// Get coordinator to list active tasks
		coord := m.engine.GetCoordinator()
		tasks, err := coord.GetTasksByStatus("wip")

		if err == nil && len(tasks) > 0 {
			// Show up to 3 active tasks
			maxTasks := 3
			if len(tasks) > maxTasks {
				tasks = tasks[:maxTasks]
			}

			for _, task := range tasks {
				// Get priority indicator
				priorityIcon := "🟡" // medium (default)
				switch task.Priority {
				case 2:
					priorityIcon = "🔴" // high
				case 0:
					priorityIcon = "🟢" // low
				}

				// Get agent icon based on role
				agentIcon := "🤖"
				switch task.Role {
				case "architect", "planning":
					agentIcon = "🏛️"
				case "keymaker", "implementation":
					agentIcon = "🔑"
				case "sentinel", "testing":
					agentIcon = "🦑"
				case "oracle", "review":
					agentIcon = "🔮"
				}

				taskTitle := task.Title
				if len(taskTitle) > 23 { // Shortened to make room for priority
					taskTitle = taskTitle[:20] + "..."
				}

				// Calculate elapsed time
				elapsed := time.Since(task.StartedAt)
				var elapsedStr string
				if elapsed < time.Minute {
					elapsedStr = fmt.Sprintf("%.0fs", elapsed.Seconds())
				} else if elapsed < time.Hour {
					elapsedStr = fmt.Sprintf("%.0fm", elapsed.Minutes())
				} else {
					elapsedStr = fmt.Sprintf("%.1fh", elapsed.Hours())
				}

				sections = append(sections, sidebarActiveStyle.Render(fmt.Sprintf("%s %s %s", priorityIcon, agentIcon, taskTitle)))
				sections = append(sections, sidebarIdleStyle.Render(fmt.Sprintf("  ⏱️  %s", elapsedStr)))
			}
		}
	}

	// Recent Errors
	recentErrors := m.getRecentErrors()
	if len(recentErrors) > 0 {
		sections = append(sections, "")
		sections = append(sections, sidebarSectionStyle.Render("⚠️  Recent Errors"))

		for _, errInfo := range recentErrors {
			// Format timestamp as relative time
			elapsed := time.Since(errInfo.Timestamp)
			var timeStr string
			if elapsed < time.Minute {
				timeStr = "just now"
			} else if elapsed < time.Hour {
				timeStr = fmt.Sprintf("%.0fm ago", elapsed.Minutes())
			} else if elapsed < 24*time.Hour {
				timeStr = fmt.Sprintf("%.0fh ago", elapsed.Hours())
			} else {
				timeStr = fmt.Sprintf("%.0fd ago", elapsed.Hours()/24)
			}

			// Truncate error message
			errMsg := errInfo.Message
			if len(errMsg) > 30 {
				errMsg = errMsg[:27] + "..."
			}

			// Get agent icon
			agentIcon := "🤖"
			switch errInfo.AgentRole {
			case "architect", "planning":
				agentIcon = "🏛️"
			case "keymaker", "implementation":
				agentIcon = "🔑"
			case "sentinel", "testing":
				agentIcon = "🦑"
			case "oracle", "review":
				agentIcon = "🔮"
			}

			sections = append(sections, lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")). // Red for errors
				Render(fmt.Sprintf("%s %s", agentIcon, errMsg)))
			sections = append(sections, sidebarIdleStyle.Render(fmt.Sprintf("  %s", timeStr)))
		}
	}

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

	// Bottom help line - toggle with ?
	var helpText string
	if m.showHelp {
		helpText = helpStyle.Render("? to hide help")
	} else {
		helpText = helpStyle.Render("? for help")
	}

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

	// Update engine auto-level
	m.engine.SetAutoLevel(m.autoLevel)

	// Save to config
	localCfg, _ := config.LoadLocal(m.projectPath)
	if localCfg == nil {
		localCfg = &config.LocalConfig{}
	}
	localCfg.AutoLevel = m.autoLevel
	_ = localCfg.SaveLocal(m.projectPath)

	// No notification needed - visible in status bar
}

func (m *BubbleModel) fetchModels() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return modelsFetchedMsg{err: fmt.Errorf("loading config: %w", err)}
		}

		provider, err := llm.NewProviderByID(cfg.Provider)
		if err != nil {
			return modelsFetchedMsg{err: fmt.Errorf("creating provider: %w", err)}
		}

		models, err := provider.GetModels()
		if err != nil {
			return modelsFetchedMsg{err: err}
		}

		return modelsFetchedMsg{models: models}
	}
}

func (m *BubbleModel) startCopilotAuth() tea.Cmd {
	return func() tea.Msg {
		provider := llm.NewCopilotProvider()

		// Start device flow
		deviceCode, err := provider.Authorize()
		if err != nil {
			return errorMsg{
				title:   "Auth Error",
				message: fmt.Sprintf("Failed to start authorization: %v", err),
			}
		}

		// Return device code info to display to user
		return deviceCodeMsg{
			verificationURI: deviceCode.VerificationURI,
			userCode:        deviceCode.UserCode,
			deviceCode:      deviceCode.DeviceCode,
			interval:        deviceCode.Interval,
			expiresIn:       deviceCode.ExpiresIn,
		}
	}
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
	coord := m.engine.GetCoordinator()
	reg := coord.GetRegistry()

	agents, err := reg.GetActiveAgents(context.Background())
	if err != nil {
		// Return empty list on error
		return []agentInfo{}
	}

	result := []agentInfo{}
	for _, agent := range agents {
		result = append(result, agentInfo{
			name:   string(agent.Role),
			active: agent.Status == "active",
		})
	}
	return result
}

// getRecentErrors returns errors from failed tasks
func (m *BubbleModel) getRecentErrors() []ErrorInfo {
	// Get failed tasks from coordinator
	coord := m.engine.GetCoordinator()
	tasks, err := coord.GetTasksByStatus("failed")
	if err != nil {
		return m.recentErrors // Return any manually added errors
	}

	// Convert failed tasks to ErrorInfo (up to 3 most recent)
	var errors []ErrorInfo
	for i, task := range tasks {
		if i >= 3 {
			break
		}
		if task.Error != "" {
			errors = append(errors, ErrorInfo{
				Message:   task.Error,
				Timestamp: task.UpdatedAt,
				AgentRole: task.Role,
				TaskID:    task.ID,
			})
		}
	}

	// If no failed tasks, return manually added errors
	if len(errors) == 0 {
		return m.recentErrors
	}

	return errors
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
	case "/allowlist":
		m.handleAllowlistCommand(parts[1:])
	case "/copy":
		m.copyLastMessage()
	case "/sessions":
		// Load and show session picker
		ctx := context.Background()
		coord := m.engine.GetCoordinator()
		sessions, err := coord.ListSessions(ctx, 10) // Show last 10 sessions
		if err != nil {
			m.messages = append(m.messages, Message{
				Content:   fmt.Sprintf("❌ Failed to load sessions: %v", err),
				Type:      "error",
				Timestamp: time.Now(),
			})
			return nil
		}

		m.sessionsList = sessions
		m.sessionPickerIndex = 0
		m.showSessionPicker = true
	case "/settings":
		m.showSettings = true
		m.settingsIndex = 0
	case "/new":
		// Archive current session and start new one
		ctx := context.Background()
		coord := m.engine.GetCoordinator()
		sessionID, err := coord.CreateNewSession(ctx)
		if err != nil {
			m.messages = append(m.messages, Message{
				Content:   fmt.Sprintf("❌ Failed to create new session: %v", err),
				Type:      "error",
				Timestamp: time.Now(),
			})
			return nil
		}

		// Clear messages
		m.messages = []Message{}
		m.messages = append(m.messages, Message{
			Content:   fmt.Sprintf("🆕 Started new session: %s", sessionID),
			Type:      "system",
			Timestamp: time.Now(),
		})
	case "/quit", "/q", "/exit":
		m.quitting = true
		return delayedQuit()
	default:
		m.messages = append(m.messages, Message{Content: "Unknown command: " + parts[0], Type: "error", Timestamp: time.Now()})
		m.messages = append(m.messages, Message{Content: "💡 Type /help to see available commands", Type: "system", Timestamp: time.Now()})
	}

	return nil
}

func (m *BubbleModel) handleAllowlistCommand(args []string) {
	if len(args) == 0 {
		// Show current allowlist
		allowlist := safety.GetSessionAllowlist()
		if len(allowlist) == 0 {
			m.messages = append(m.messages, Message{
				Content:   "📋 Session Allowlist (empty)\n\nNo commands have been added to the session allowlist yet.",
				Type:      "system",
				Timestamp: time.Now(),
			})
		} else {
			content := "📋 Session Allowlist\n\nCommands that will bypass safety checks this session:\n\n"
			for i, cmd := range allowlist {
				content += fmt.Sprintf("%d. %s\n", i+1, cmd)
			}
			content += fmt.Sprintf("\nTotal: %d command(s)", len(allowlist))
			m.messages = append(m.messages, Message{
				Content:   content,
				Type:      "system",
				Timestamp: time.Now(),
			})
		}
		return
	}

	subcommand := args[0]
	switch subcommand {
	case "clear":
		safety.ClearSessionAllowlist()
		m.messages = append(m.messages, Message{
			Content:   "✅ Session allowlist cleared",
			Type:      "system",
			Timestamp: time.Now(),
		})
	case "add":
		if len(args) < 2 {
			m.messages = append(m.messages, Message{
				Content:   "❌ Usage: /allowlist add <command>",
				Type:      "error",
				Timestamp: time.Now(),
			})
			return
		}
		command := strings.Join(args[1:], " ")
		safety.AddToSessionAllowlist(command)
		m.messages = append(m.messages, Message{
			Content:   fmt.Sprintf("✅ Added to session allowlist: %s", command),
			Type:      "system",
			Timestamp: time.Now(),
		})
	default:
		m.messages = append(m.messages, Message{
			Content:   "❌ Unknown subcommand. Usage:\n  /allowlist          - Show current allowlist\n  /allowlist add <cmd> - Add command\n  /allowlist clear    - Clear all",
			Type:      "error",
			Timestamp: time.Now(),
		})
	}
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

	// Format messages with styling and add spacing between messages
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
			return centered + "\n\n" + strings.Join(formatted[1:], "\n\n")
		}
		return centered
	}

	return strings.Join(formatted, "\n\n")
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

  /status    /s  - Show task counts and progress
  /settings      - Change model and auto-level
  /allowlist     - Manage session command allowlist
                   /allowlist          - Show current allowlist
                   /allowlist add <cmd> - Add command to allowlist
                   /allowlist clear    - Clear allowlist
  /copy          - Copy last message to clipboard
  /new            - Start new session
  /help      /h  - Show this help
  /quit      /q  - Exit Smith

⌨️  Keyboard Shortcuts:

  Enter           - Send message
  Tab             - Focus input area
  Shift+Tab       - Cycle reasoning level (Low/Medium/High)
  Ctrl+N          - Start new session
  Ctrl+H          - Show this help
  Ctrl+Y          - Copy last message
  Ctrl+C/Ctrl+D   - Quit Smith
  Page Up/Ctrl+U  - Scroll message history up
  Page Down       - Scroll message history down
  ESC             - Cancel/close menus

💡 Just chat naturally - no commands needed to build features!
`
	return helpStyle.Render(help)
}

func (m *BubbleModel) renderAuthPanel(base string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	codeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render("🔐 GitHub Copilot Authentication") + "\n\n")

	content.WriteString(labelStyle.Render("Please visit:") + "\n")
	content.WriteString(codeStyle.Render(m.authVerifyURI) + "\n\n")

	content.WriteString(labelStyle.Render("Enter code (copied to clipboard):") + "\n")
	content.WriteString(codeStyle.Render(m.authUserCode) + "\n\n")

	content.WriteString(labelStyle.Render("⏳ Waiting for authorization...") + "\n\n")

	content.WriteString(helpStyle.Render("Browser opened • Code copied to clipboard\nPress Esc to cancel"))

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

func (m *BubbleModel) renderSessionPicker(base string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	itemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Background(lipgloss.Color("236")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("📂 Session History") + "\n\n")

	if len(m.sessionsList) == 0 {
		content.WriteString(dimStyle.Render("No previous sessions found.") + "\n")
	} else {
		for i, session := range m.sessionsList {
			// Calculate relative time
			elapsed := time.Since(session.LastActive)
			var timeStr string
			if elapsed < time.Minute {
				timeStr = "just now"
			} else if elapsed < time.Hour {
				mins := int(elapsed.Minutes())
				timeStr = fmt.Sprintf("%dm ago", mins)
			} else if elapsed < 24*time.Hour {
				hours := int(elapsed.Hours())
				timeStr = fmt.Sprintf("%dh ago", hours)
			} else {
				days := int(elapsed.Hours() / 24)
				if days == 1 {
					timeStr = "1 day ago"
				} else {
					timeStr = fmt.Sprintf("%d days ago", days)
				}
			}

			// Format: Title (5 tasks) • Started 2h ago
			line := fmt.Sprintf("%s (%d tasks) • %s", session.Title, session.TaskCount, timeStr)

			// Truncate if too long
			if len(line) > 60 {
				line = line[:57] + "..."
			}

			if i == m.sessionPickerIndex {
				content.WriteString(selectedStyle.Render("> "+line) + "\n")
			} else {
				content.WriteString(itemStyle.Render("  "+line) + "\n")
			}
		}
	}

	content.WriteString("\n" + helpStyle.Render("↑/↓: Navigate  Enter: Select  Esc: Cancel"))

	// Create panel
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(70).
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

func (m *BubbleModel) renderHelpPanel(base string) string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("208")).
		Bold(true)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("252"))

	var content strings.Builder
	content.WriteString(titleStyle.Render("Keyboard Shortcuts & Commands") + "\n\n")

	// Basics section
	content.WriteString(sectionStyle.Render("Basics") + "\n")
	content.WriteString(keyStyle.Render("  Enter          ") + descStyle.Render("Send message") + "\n")
	content.WriteString(keyStyle.Render("  Ctrl + C       ") + descStyle.Render("Quit") + "\n")
	content.WriteString(keyStyle.Render("  Ctrl + N       ") + descStyle.Render("Start new session") + "\n")
	content.WriteString(keyStyle.Render("  Shift + Tab    ") + descStyle.Render("Cycle auto-level") + "\n")
	content.WriteString(keyStyle.Render("  Esc            ") + descStyle.Render("Cancel/close menus") + "\n")
	content.WriteString(keyStyle.Render("  ?              ") + descStyle.Render("Toggle this help") + "\n\n")

	// Commands section
	content.WriteString(sectionStyle.Render("Commands") + "\n")
	for _, cmd := range availableCommands {
		cmdText := cmd.Name
		if cmd.Alias != "" {
			cmdText += ", " + cmd.Alias
		}
		content.WriteString(keyStyle.Render(fmt.Sprintf("  %-15s", cmdText)) + descStyle.Render(cmd.Description) + "\n")
	}

	content.WriteString("\n" + helpStyle.Render("Press ? or Esc to close"))

	// Create panel
	panel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 2).
		Width(70).
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

func (m *BubbleModel) renderSettingsPanel(base string) string {
	// Show provider menu if active
	if m.showProviderMenu {
		return m.renderProviderMenu(base)
	}

	// Show model menu if active
	if m.showModelMenu {
		return m.renderModelMenu(base)
	}

	// Get merged config (global + local)
	merged, err := config.LoadWithMerge(m.projectPath)
	if err != nil {
		return base
	}
	cfg := &merged.Config

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

	// Add placeholders for empty values
	providerDisplay := cfg.Provider
	if providerDisplay == "" {
		providerDisplay = "(not configured)"
	}
	modelDisplay := cfg.Model
	if modelDisplay == "" {
		modelDisplay = "(not configured)"
	}

	// Settings items - clean display without (global)/(local) labels
	items := []string{
		"Provider: " + providerDisplay,
		"Model: " + modelDisplay,
		"Reasoning: " + m.autoLevel,
		"",
		"[Edit safety rules]",
		"[Open config directory]",
	}

	var content strings.Builder
	content.WriteString(titleStyle.Render("⚙️ Settings") + "\n\n")

	content.WriteString(sectionStyle.Render("Configuration") + "\n")
	for i := 0; i < 3; i++ {
		if i == m.settingsIndex {
			content.WriteString(selectedStyle.Render("> "+items[i]) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  "+items[i]) + "\n")
		}
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

	// Show loading indicator if fetching
	if m.fetchingModels {
		loadingContent := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Render("⏳ Fetching models...")

		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("11")).
			Padding(1, 2).
			Width(70).
			Render(titleStyle.Render("Select Model") + "\n\n" + loadingContent)

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	// Use cached models and apply filter
	allModels := m.cachedModels
	if allModels == nil {
		// Show error if we don't have cached models
		errorContent := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Render("No models available. Please try again.")

		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("1")).
			Padding(1, 2).
			Width(70).
			Render(titleStyle.Render("Select Model") + "\n\n" + errorContent + "\n\n" + helpStyle.Render("esc to close"))

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	// Apply filter
	models := m.filterModels(allModels, m.modelFilter)

	cfg, _ := config.Load()
	provider, _ := llm.NewProviderByID(cfg.Provider)

	var content strings.Builder
	content.WriteString(titleStyle.Render("Select Model") + "\n")
	content.WriteString(descStyle.Render("Provider: "+provider.GetName()) + "\n")

	// Show filter if active
	if m.modelFilter != "" {
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Bold(true)
		content.WriteString(filterStyle.Render("Filter: "+m.modelFilter) + "\n")
	}

	// Check if no matches
	if len(models) == 0 {
		noMatchStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Italic(true)
		content.WriteString("\n" + noMatchStyle.Render("No matching models") + "\n\n")
		content.WriteString(helpStyle.Render("backspace to clear filter · esc cancel"))

		panel := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("11")).
			Padding(1, 2).
			Width(70).
			Render(content.String())

		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, panel)
	}

	// Fixed page size
	const pageSize = 10
	totalModels := len(models)
	currentPage := (m.modelMenuIndex / pageSize) + 1
	totalPages := (totalModels + pageSize - 1) / pageSize

	// Show page indicator
	content.WriteString(descStyle.Render(fmt.Sprintf("Page %d/%d (%d models total)", currentPage, totalPages, totalModels)) + "\n\n")

	startIdx := m.modelMenuScroll
	endIdx := startIdx + pageSize
	if endIdx > totalModels {
		endIdx = totalModels
	}

	// Render visible models (max 10 per page)
	for i := startIdx; i < endIdx; i++ {
		model := models[i]
		if i == m.modelMenuIndex {
			content.WriteString(selectedStyle.Render("> "+model.Name) + "\n")
			if model.Description != "" {
				content.WriteString(descStyle.Render("  "+model.Description) + "\n")
			}
		} else {
			content.WriteString(itemStyle.Render("  "+model.Name) + "\n")
		}
	}

	helpText := "↑/↓ navigate (cycles) · enter select · "
	if m.modelFilter != "" {
		helpText += "backspace clear filter · "
	} else {
		helpText += "type to filter · "
	}
	helpText += "esc cancel"
	content.WriteString("\n" + helpStyle.Render(helpText))

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

// formatMessage formats a message with appropriate styling
func (m *BubbleModel) formatMessage(msg Message, width int) string {
	// Format based on message type
	switch msg.Type {
	case "user":
		return promptStyle.Render("💬 You: ") + msg.Content
	case "ai":
		// Show animated spinner if this is the loading message during streaming
		if m.isStreaming && msg.Content == "⏳ Thinking..." {
			spinners := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinner := spinners[m.spinnerFrame%len(spinners)]
			return aiStyle.Render("🕶️ Smith: ") + spinner + " Thinking..."
		}
		return aiStyle.Render("🕶️ Smith: ") + msg.Content
	case "system":
		return msg.Content
	case "error":
		return errorStyle.Render(msg.Content) // No extra icon, error style is enough
	default:
		return msg.Content
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
