# Smith - The Agent Replication System

> *"More... more... more agents, Mr. Anderson!"*

**Inevitable. Multiplying. Building.**

A multi-agent AI development system that coordinates specialized agents to build software through natural conversation.

## 🕶️ What It Is

Smith is an AI coding assistant that orchestrates multiple specialized agents working together. Chat naturally in your terminal, and Smith deploys planning, implementation, testing, and review agents that collaborate to get work done.

**Key Features:**
- **Multi-Agent Coordination** - Specialized agents work in parallel on different aspects of your code
- **Interactive REPL** - Natural conversation interface powered by Bubble Tea
- **SQLite-Based State** - Agents coordinate through a local database in `.smith/`
- **Safety Levels** - Control what agents can do (Low/Medium/High execution permissions)
- **GitHub Copilot Integration** - Powered by GPT-4o for intelligent responses

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

## 🚀 Project Status

**Current - v0.1.0:**

✅ Interactive REPL with Bubble Tea  
✅ GitHub Copilot integration (GPT-4o)  
✅ SQLite coordination infrastructure  
✅ Safety levels and execution control  
✅ Event bus and file locking system  

🚧 **Next Steps:**
- Multi-agent spawning and coordination
- Specialized agent roles (planning, implementation, testing, review)
- Agent-to-agent communication via event bus
- Task queue and parallel execution

## 🎓 Technology Stack

- **Go** - Performance and clean architecture
- **GitHub Copilot** - GPT-4o powered intelligence
- **Bubble Tea** - Modern terminal UI framework
- **SQLite** - WAL mode coordination with concurrent access

## 🤝 Contributing

Experimental AI coding assistant. Ideas, issues, and PRs welcome!

---

**Built to make coding with AI more powerful and coordinated.**
