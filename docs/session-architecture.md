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

// Create UI
ui, _ := cli.New()

// Run
ui.Run(sess)
```

### With Real Agent System (Coming Soon)
```go
// Create agent-backed session
sess := session.NewAgentSession(
    coordinator: coord,
    agents: []agent.Agent{planning, implementation, testing, review},
)

// Use ANY interface
ui, _ := cli.New()  // or tui.New(), http.NewServer(), etc.
ui.Run(sess)
```

## Package Structure

```
internal/
├── session/           # Session interface & implementations
│   ├── interface.go  # Session interface, Message type
│   ├── mock_session.go   # Mock for testing
│   └── agent_session.go  # Real agent system (TODO)
│
└── ui/               # All UI implementations
    ├── interface.go  # UI interface
    ├── cli/         # CLI interface
    │   └── cli.go
    └── tui/         # TUI interface (TODO)
        └── tui.go
```

## Testing

```bash
# Test the new structure
go build ./internal/session ./internal/ui/...

# Run CLI demo
go run ./cmd/smith-cli-demo
```

## Migration Notes

**Archived old attempts:**
- `archive/ui-attempts/repl/` - Old REPL (mixed concerns)
- `archive/ui-attempts/tui/` - Old TUI (broken v1/v2 issues)
- `archive/ui-attempts/chat-experiments/` - Experimental chat UIs

**Renamed for clarity:**
- `internal/chat/` → `internal/session/` (better name for agent coding tool)
- `Engine` → `Session` (session = conversation + agent system)
- `Frontend/Interface` → `UI` (clearer, no confusion with web APIs)

## Design Principles

1. **Separation of Concerns**: Session doesn't know about UI, UI doesn't know about agents
2. **Streaming First**: All responses stream for better UX
3. **UI Agnostic**: Session works with any UI (CLI, TUI, HTTP, etc.)
4. **Testable**: MockSession for testing UIs, mock UIs for testing sessions
5. **Extensible**: Easy to add new UIs without changing session

## Next Steps

1. ✅ CLI interface working
2. 🔨 Build TUI interface (Bubble Tea v2)
3. 🔨 Create AgentSession (connect to real agent system)
4. 🔨 Add HTTP REST API
5. 🔨 Add WebSocket interface


Clean separation between **Engine** (core logic) and **Interface** (UI layer).

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    Chat Interfaces                       │
│  (UI Layer - How users interact with Smith)             │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  CLIInterface        TUIInterface      HTTPInterface     │
│  (stdin/stdout)      (Bubble Tea)      (REST API)        │
│                                                           │
│  WebSocketInterface  gRPCInterface     ...               │
│                                                           │
└────────────────────┬────────────────────────────────────┘
                     │ implements Interface
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                     Chat Engine                          │
│  (Core Logic - Agent system, LLM, memory, tools)        │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  • SendMessage(msg) -> <-chan string  (streaming)        │
│  • GetHistory() -> []Message                             │
│  • Reset()                                               │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

## Key Concepts

### Engine (Core)
The "backend" that powers all interfaces. Contains:
- Agent orchestration
- LLM integration
- Memory/context management
- Tool execution
- Safety checks

**Why "Engine" not "Backend"?**
- When we add REST API, that's also a "backend"
- Engine = the core chat logic
- Interface = how users access it (CLI, TUI, HTTP, WebSocket, etc.)

### Interface (UI Layer)
How users interact with the engine. Examples:

| Interface | Use Case | Technology |
|-----------|----------|------------|
| `CLIInterface` | Local terminal, piping, scripts | stdin/stdout + Glamour |
| `TUIInterface` | Rich terminal UI with mouse | Bubble Tea v2 + viewport |
| `HTTPInterface` | REST API for web/mobile | HTTP server |
| `WebSocketInterface` | Real-time web chat | WebSocket server |
| `gRPCInterface` | Service-to-service | gRPC |

## Current Implementation

### MockEngine (Demo)
Simple mock for testing. In production, this would be your actual agent system.

```go
engine := chat.NewMockEngine()
ch, _ := engine.SendMessage("Hello!")
for chunk := range ch {
    fmt.Print(chunk) // Streams response word-by-word
}
```

### CLIInterface (inquirer.js style)
Simple readline-style interface with markdown rendering.

```go
ui, _ := chat.NewCLIInterface()
ui.Run(engine) // Blocks until user quits
```

**Features:**
- ✅ Markdown rendering (Glamour)
- ✅ Syntax highlighting in code blocks
- ✅ Colored output (Lipgloss)
- ✅ Streaming responses
- ✅ History display

## Future Interfaces

### TUIInterface (Coming Soon)
Full-screen Bubble Tea interface with:
- Scrollable chat history (viewport)
- Multi-line input (textarea)
- Mouse support
- Animations
- Split views (chat + inspector)

### HTTPInterface (Planned)
REST API for web/mobile clients:

```
POST /api/chat/send
GET  /api/chat/history
POST /api/chat/reset
GET  /api/chat/stream (SSE)
```

### WebSocketInterface (Planned)
Real-time bidirectional communication:

```javascript
ws://localhost:8080/ws
// Send: {"type": "message", "content": "Hello"}
// Receive: {"type": "chunk", "content": "Hi"}
```

## Usage

### Basic Example
```go
// Create engine
engine := chat.NewMockEngine()

// Create interface
ui, _ := chat.NewCLIInterface()

// Run
ui.Run(engine)
```

### With Real Agent System (Future)
```go
// Create engine with real agent system
engine := chat.NewAgentEngine(
    chat.WithLLM(llm.NewCopilot()),
    chat.WithMemory(memory.NewBBolt("chat.db")),
    chat.WithTools(tools.All()...),
)

// Use ANY interface
ui, _ := chat.NewCLIInterface()  // or NewTUIInterface(), NewHTTPInterface(), etc.
ui.Run(engine)
```

## Testing

```bash
# Test the architecture
go test ./internal/chat ./cmd/chat-basic

# Run the CLI demo
go run ./cmd/chat-basic
```

## Design Principles

1. **Separation of Concerns**: Engine doesn't know about UI, UI doesn't know about agents
2. **Streaming First**: All responses stream for better UX
3. **Interface Agnostic**: Engine works with any interface (CLI, TUI, HTTP, etc.)
4. **Testable**: Mock engine for testing UIs, mock interfaces for testing engine
5. **Extensible**: Easy to add new interfaces without changing engine

## Why This Matters

**Multiple frontends from one engine:**
```
smith chat              # CLI interface
smith tui               # TUI interface
smith serve --http      # HTTP API
smith serve --grpc      # gRPC service
```

All powered by the **same agent engine**.
