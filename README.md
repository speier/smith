# Smith - The Agent Replication System

> *"More... more... more agents, Mr. Anderson!"*

**Inevitable. Multiplying. Building.**

An AI coding assistant with a powerful REPL interface and SQLite-based coordination system.

## 🕶️ What It Is

**Chat naturally with AI. Built-in coordination for future multi-agent workflows.**

A smart coding assistant powered by GitHub Copilot that helps you build software through natural conversation.

**Current features:**
- **Interactive REPL** - Bubble Tea powered terminal interface  
- **GitHub Copilot Integration** - Uses GPT-4o for intelligent responses
- **SQLite Coordination** - Event bus, file locking, agent registry for future multi-agent work
- **Safety Levels** - Control execution permissions (Low/Medium/High)

**Future:** Specialized agents (planning, implementation, testing, review) that coordinate via SQLite.

## Installation

```bash
curl -fsSL https://agentsmith.sh/install | bash
```

Or download from [releases](https://github.com/speier/smith/releases).

## Usage

```bash
# Start interactive REPL
smith

# Execute single prompt
smith exec "create a hello world program"
```

## ⚡ Quick Start

```bash
# Build from source
go build -o smith .

# Run the REPL
./smith
```

**That's it!** Start chatting and building.

## 🏗️ Architecture

### SQLite Coordination System

State managed in `.smith/` directory (auto-created on first run):

- **`smith.db`** - SQLite database with WAL mode (gitignored)
- **`kanban.md`** - Human-readable task board (committed to git)  
- **`config.yaml`** - User settings (gitignored)
- **`.gitignore`** - Auto-generated ignore rules

### Coordination Infrastructure

Built for future multi-agent workflows:

- **Event Bus** - Poll-based event streaming between agents
- **File Locking** - Transactional locks prevent file conflicts
- **Agent Registry** - Heartbeat tracking and status management
- **Task Management** - Kanban-style workflow (ready for integration)

## 🚀 Project Status

**Current - v0.1.0:**

✅ Interactive REPL with Bubble Tea  
✅ GitHub Copilot integration (GPT-4o)  
✅ SQLite coordination infrastructure  
✅ Safety levels and execution control  
✅ Event bus and file locking system  
✅ Clean, minimal codebase  

🚧 **Next Steps:**
- Kanban.md task parsing and management
- Multi-agent spawning and coordination
- Specialized agent roles (planning, implementation, testing, review)
- Agent-to-agent communication via event bus

## 🎓 Technology Stack

- **Go** - Concurrent operations and clean architecture
- **GitHub Copilot** - GPT-4o powered intelligent responses
- **Bubble Tea** - Modern, composable terminal UI
- **SQLite** - WAL mode with 25 concurrent connections
- **Markdown** - Human-readable state management (kanban.md)

## 🤝 Contributing

Experimental AI coding assistant. Ideas, issues, and PRs welcome!

---

**Built to make coding with AI more powerful and coordinated.**
