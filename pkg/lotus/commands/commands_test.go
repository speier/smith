package commands

import (
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCmd   string
		wantArgs  []string
		wantIsCmd bool
	}{
		{
			name:      "simple command",
			input:     "/help",
			wantCmd:   "help",
			wantArgs:  []string{},
			wantIsCmd: true,
		},
		{
			name:      "command with args",
			input:     "/model gpt-4",
			wantCmd:   "model",
			wantArgs:  []string{"gpt-4"},
			wantIsCmd: true,
		},
		{
			name:      "command with multiple args",
			input:     "/set temperature 0.7",
			wantCmd:   "set",
			wantArgs:  []string{"temperature", "0.7"},
			wantIsCmd: true,
		},
		{
			name:      "not a command",
			input:     "hello world",
			wantCmd:   "",
			wantArgs:  nil,
			wantIsCmd: false,
		},
		{
			name:      "just slash",
			input:     "/",
			wantCmd:   "",
			wantArgs:  nil,
			wantIsCmd: false,
		},
		{
			name:      "empty string",
			input:     "",
			wantCmd:   "",
			wantArgs:  nil,
			wantIsCmd: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewCommandRegistry()
			gotCmd, gotArgs, gotIsCmd := registry.ParseCommand(tt.input)

			if gotIsCmd != tt.wantIsCmd {
				t.Errorf("ParseCommand() isCommand = %v, want %v", gotIsCmd, tt.wantIsCmd)
			}
			if gotCmd != tt.wantCmd {
				t.Errorf("ParseCommand() command = %q, want %q", gotCmd, tt.wantCmd)
			}
			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("ParseCommand() args length = %d, want %d", len(gotArgs), len(tt.wantArgs))
				return
			}
			for i := range gotArgs {
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("ParseCommand() args[%d] = %q, want %q", i, gotArgs[i], tt.wantArgs[i])
				}
			}
		})
	}
}

func TestCustomPrefix(t *testing.T) {
	// Test with ! prefix
	registry := NewCommandRegistry()
	registry.SetCommandPrefix("!")

	registry.Add("test", "Test command", func(args []string) {})

	// Should parse !test
	cmd, args, ok := registry.ParseCommand("!test arg1")
	if !ok || cmd != "test" || len(args) != 1 || args[0] != "arg1" {
		t.Errorf("!test parsing failed: got (%q, %v, %v)", cmd, args, ok)
	}

	// Should NOT parse /test
	_, _, ok = registry.ParseCommand("/test")
	if ok {
		t.Error("/test should not be recognized with ! prefix")
	}

	// Test with @bot prefix
	registry2 := NewCommandRegistry()
	registry2.SetCommandPrefix("@bot ")

	cmd, args, ok = registry2.ParseCommand("@bot help")
	if !ok || cmd != "help" {
		t.Errorf("@bot help parsing failed: got (%q, %v, %v)", cmd, args, ok)
	}
}

func TestCommandRegistry(t *testing.T) {
	registry := NewCommandRegistry()

	// Register commands
	helpCalled := false
	registry.Register(&Command{
		Name:        "help",
		Description: "Show help",
		Handler: func(args []string) {
			helpCalled = true
		},
	})

	clearCalled := false
	clearArgs := []string{}
	registry.Register(&Command{
		Name:        "clear",
		Description: "Clear screen",
		Aliases:     []string{"cls"},
		Handler: func(args []string) {
			clearCalled = true
			clearArgs = args
		},
	})

	// Test Get
	cmd, ok := registry.Get("help")
	if !ok {
		t.Error("Expected to find 'help' command")
	}
	if cmd.Name != "help" {
		t.Errorf("Command name = %q, want 'help'", cmd.Name)
	}

	// Test alias
	cmd, ok = registry.Get("cls")
	if !ok {
		t.Error("Expected to find 'cls' alias")
	}
	if cmd.Name != "clear" {
		t.Errorf("Alias resolves to %q, want 'clear'", cmd.Name)
	}

	// Test Execute
	if !registry.Execute("/help") {
		t.Error("Execute(/help) = false, want true")
	}
	if !helpCalled {
		t.Error("Help handler not called")
	}

	// Test Execute with alias
	if !registry.Execute("/cls confirm") {
		t.Error("Execute(/cls) = false, want true")
	}
	if !clearCalled {
		t.Error("Clear handler not called")
	}
	if len(clearArgs) != 1 || clearArgs[0] != "confirm" {
		t.Errorf("Clear args = %v, want ['confirm']", clearArgs)
	}

	// Test non-existent command
	if registry.Execute("/unknown") {
		t.Error("Execute(/unknown) = true, want false")
	}

	// Test non-command
	if registry.Execute("hello") {
		t.Error("Execute(hello) = true, want false")
	}

	// Test List
	commands := registry.List()
	if len(commands) != 2 {
		t.Errorf("List() length = %d, want 2", len(commands))
	}
}
