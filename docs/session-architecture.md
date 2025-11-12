# Session Architecture

Clean separation between **Session** (core agent system) and **UI** (interface layer).

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    User Interfaces                       │
│  (How users interact with Smith)                        │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  CLI              TUI              HTTP                  │
│  (stdin/stdout)   (Bubble Tea)     (REST API)           │
│                                                           │
│  WebSocket        gRPC             Desktop App           │
│                                                           │
└────────────────────┬────────────────────────────────────┘
                     │ implements ui.UI
                     │ uses session.Session
                     ▼
┌─────────────────────────────────────────────────────────┐
│                      Session                             │
│  (Interactive coding session with agent system)         │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  • SendMessage(msg) -> <-chan string  (streaming)        │
│  • GetHistory() -> []Message                             │
│  • Reset()                                               │
│                                                           │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Agent System                            │
│  (Planning, Implementation, Testing, Review)            │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  Coordinator  →  Agents  →  Engine  →  LLM + Tools      │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

## Key Concepts

### Session
Represents an interactive coding session backed by the multi-agent system.
- Maintains conversation history
- Coordinates agent workflow
- Streams responses in real-time
- Manages context and state

**Implementations:**
- `MockSession` - For testing UIs (fake responses)
- `AgentSession` - Real implementation (Coming soon)

### UI (User Interface)
How users interact with sessions. Lives in `/internal/ui/`.

| Interface | Use Case | Package |
|-----------|----------|---------|
| `CLI` | Terminal, scripts, piping | `ui/cli` |
| `TUI` | Full-screen interactive | `ui/tui` |
| `HTTP` | REST API for web/mobile | `ui/http` |
| `WebSocket` | Real-time web chat | `ui/websocket` |

## Usage

### Basic Example
```go
// Create session
sess := session.NewMockSession()

// Create UI (Lotus TUI component)
ui := frontend.NewChatUI(sess)

// Run with Lotus runtime
lotus.Run(ui)
```

### With Real Agent System (Coming Soon)
```go
// Create agent-backed session
sess := session.NewAgentSession(
    coordinator: coord,
    agents: []agent.Agent{planning, implementation, testing, review},
)

// Use with Lotus TUI
ui := frontend.NewChatUI(sess)
lotus.Run(ui)
```

## Package Structure

```
pkg/agent/session/     # Session interface & implementations
├── interface.go      # Session interface, Message type
├── mock_session.go   # Mock for testing
└── agent_session.go  # Real agent system (TODO)

internal/
├── frontend/         # UI implementations (Lotus-based TUI)
│   ├── chat.go      # Main chat UI component
│   ├── messagelist.go  # Message list component
│   └── branding.go  # Welcome/goodbye banners
│
└── cli/             # CLI commands (Cobra)
    ├── root.go      # Main command, runs Lotus TUI
    └── exec.go      # Exec subcommand
```

## Testing

```bash
# Test the new structure
go build ./pkg/agent/session ./internal/frontend

# Run the main app (Lotus TUI)
go run .
```

## Migration Notes

**Archived old attempts:**
- `archive/ui-attempts/repl/` - Old REPL (mixed concerns)
- `archive/ui-attempts/tui/` - Old TUI (broken v1/v2 issues)
- `archive/ui-attempts/chat-experiments/` - Experimental chat UIs

**Renamed for clarity:**
- `internal/chat/` → `pkg/agent/session/` (session = agent-backed conversation)
- `Engine` → `Session` (session = conversation + agent system)
- `Frontend/Interface` → `frontend/` (Lotus TUI components)
- Moved session to `pkg/` (can be imported by other packages)

## Design Principles

1. **Separation of Concerns**: Session doesn't know about UI, UI doesn't know about agents
2. **Streaming First**: All responses stream for better UX
3. **UI Agnostic**: Session works with any UI (CLI, TUI, HTTP, etc.)
4. **Testable**: MockSession for testing UIs, mock UIs for testing sessions
5. **Extensible**: Easy to add new UIs without changing session

## Next Steps

1. ✅ Session interface defined (`pkg/agent/session/`)
2. ✅ Lotus TUI working (`internal/frontend/`)
3. 🔨 Create AgentSession (connect to real agent system)
4. 🔨 Add CLI text interface (stdin/stdout for piping)
5. 🔨 Add HTTP REST API (future)
6. 🔨 Add WebSocket interface (future)
