package commands

import (
	"reflect"
	"strings"
)

// Context provides helpers for updating the UI from command handlers
// This is an interface so commands package doesn't depend on runtime
type Context interface {
	// Update triggers a re-render of the application
	Update()

	// Rerender is an alias for Update
	Rerender()
}

// Command represents a slash command
type Command struct {
	Name        string   // Command name (without /)
	Description string   // Help text
	Handler     any      // func([]string) or func(Context, []string)
	Aliases     []string // Alternative names
}

// CommandRegistry manages slash commands
type CommandRegistry struct {
	commands       map[string]*Command
	displayHandler func(line string) // How to display a line (e.g., add to messages)
	Prefix         string            // Command prefix (default: "/", can be "!", "@bot ", etc.)
}

// NewCommandRegistry creates a new command registry with built-in /help
func NewCommandRegistry() *CommandRegistry {
	r := &CommandRegistry{
		commands: make(map[string]*Command),
		Prefix:   "/", // Default to slash commands
	}

	// Register built-in /help command that works automatically
	r.Register(&Command{
		Name:        "help",
		Description: "Show available commands",
		Handler: func(ctx Context, args []string) {
			// Display help text using the display handler (or no-op if not set)
			for _, line := range r.GetHelpText() {
				if r.displayHandler != nil {
					r.displayHandler(line)
				}
			}
		},
	})

	return r
}

// SetDisplayHandler tells the registry how to display output (e.g., add to chat messages)
// If not set, /help and other commands that produce output will silently succeed
func (r *CommandRegistry) SetDisplayHandler(handler func(line string)) {
	r.displayHandler = handler
}

// Register adds a command to the registry
func (r *CommandRegistry) Register(cmd *Command) {
	r.commands[cmd.Name] = cmd
	// Register aliases
	for _, alias := range cmd.Aliases {
		r.commands[alias] = cmd
	}
}

// Add is a simpler way to register a command (similar to http.HandleFunc)
// Usage: registry.Add("stream", "Stream text", handler)
// Handler can be func([]string) or func(Context, []string)
func (r *CommandRegistry) Add(name, description string, handler any) {
	r.Register(&Command{
		Name:        name,
		Description: description,
		Handler:     handler,
	})
}

// Get retrieves a command by name
func (r *CommandRegistry) Get(name string) (*Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// List returns all registered commands (excluding aliases)
func (r *CommandRegistry) List() []*Command {
	seen := make(map[*Command]bool)
	var result []*Command
	for _, cmd := range r.commands {
		if !seen[cmd] {
			seen[cmd] = true
			result = append(result, cmd)
		}
	}
	return result
}

// GetHelpText returns formatted help text for all commands
// This provides a default help implementation that apps can use
func (r *CommandRegistry) GetHelpText() []string {
	var lines []string
	lines = append(lines, "") // Blank line before help
	lines = append(lines, "Available commands:")
	for _, cmd := range r.List() {
		lines = append(lines, "  /"+cmd.Name+" - "+cmd.Description)
	}
	return lines
}

// ParseCommand parses text using the registry's prefix to extract command and arguments
// Returns: command name, args, isCommand
func (r *CommandRegistry) ParseCommand(text string) (string, []string, bool) {
	if !strings.HasPrefix(text, r.Prefix) {
		return "", nil, false
	}

	// Remove prefix
	text = strings.TrimPrefix(text, r.Prefix)
	text = strings.TrimSpace(text)
	if text == "" {
		return "", nil, false
	}

	// Split into command and args
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return "", nil, false
	}

	name := parts[0]
	args := parts[1:]
	return name, args, true
}

// invokeCommandHandler invokes a command handler with signature detection
// Uses reflection to handle concrete Context types (e.g., runtime.Context)
func invokeCommandHandler(handler any, ctx Context, args []string) {
	if handler == nil {
		return
	}

	// Fast path: check known signatures first
	switch fn := handler.(type) {
	case func([]string):
		fn(args)
	case func(Context, []string):
		fn(ctx, args)
	default:
		// Use reflection to handle func(concreteContextType, []string)
		v := reflect.ValueOf(handler)
		t := v.Type()

		// Check if it's a function with 2 parameters
		if t.Kind() != reflect.Func || t.NumIn() != 2 {
			return // Ignore unsupported types
		}

		// Check if first param implements Context interface
		firstParam := t.In(0)
		contextInterface := reflect.TypeOf((*Context)(nil)).Elem()

		// Check if second param is []string
		if t.In(1).Kind() != reflect.Slice || t.In(1).Elem().Kind() != reflect.String {
			return
		}

		// If first param implements Context, call the function
		if firstParam.Implements(contextInterface) || (firstParam.Kind() == reflect.Struct && reflect.PointerTo(firstParam).Implements(contextInterface)) {
			ctxValue := reflect.ValueOf(ctx)
			if !ctxValue.IsValid() {
				ctxValue = reflect.Zero(firstParam)
			}

			v.Call([]reflect.Value{ctxValue, reflect.ValueOf(args)})
		}
	}
}

// Execute runs a command if it exists in the registry
// Returns: true if command was found and executed
func (r *CommandRegistry) Execute(text string) bool {
	return r.ExecuteWithContext(text, nil)
}

// ExecuteWithContext runs a command with context support
func (r *CommandRegistry) ExecuteWithContext(text string, ctx Context) bool {
	name, args, isCommand := r.ParseCommand(text)
	if !isCommand {
		return false
	}

	cmd, ok := r.Get(name)
	if !ok {
		return false
	}

	if cmd.Handler != nil {
		invokeCommandHandler(cmd.Handler, ctx, args)
	}
	return true
}

// Global command registry (similar to global key bindings)
var globalCommands = NewCommandRegistry()

// RegisterGlobalCommand registers a command globally (available in all apps)
// Usage: lotus.RegisterGlobalCommand("clear", "Clear screen", handler)
// Handler can be func([]string) or func(Context, []string)
func RegisterGlobalCommand(name, description string, handler any) {
	globalCommands.Add(name, description, handler)
}

// GetGlobalCommands returns the global command registry
func GetGlobalCommands() *CommandRegistry {
	return globalCommands
}

// SetGlobalCommandPrefix sets the prefix for global commands
// Examples: "/", "!", "@bot ", etc.
// Default is "/" if not set
func SetGlobalCommandPrefix(prefix string) {
	globalCommands.Prefix = prefix
}

// SetCommandPrefix sets the prefix for this registry
// Examples: "/", "!", "@bot ", etc.
func (r *CommandRegistry) SetCommandPrefix(prefix string) *CommandRegistry {
	r.Prefix = prefix
	return r
}
