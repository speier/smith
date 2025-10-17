# Smith - The Agent Replication System

> *"More... more... more agents, Mr. Anderson!"*

**Inevitable. Multiplying. Building.**

A multi-agent AI development system that coordinates specialized agents to build software through natural conversation.

## 🕶️ What It Is

Smith is an AI coding assistant that orchestrates multiple specialized agents working together. Chat naturally in your terminal, and Smith deploys planning, implementation, testing, and review agents that collaborate to get work done.

**Key Features:**
- **Multi-Agent Coordination** - Specialized agents work in parallel on different aspects of your code
- **Interactive REPL** - Natural conversation interface powered by Bubble Tea
- **BBolt Storage** - Lock-free coordination via embedded key-value store (same as etcd/Kubernetes)
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

**Current - v0.2.0:**

✅ Interactive REPL with Bubble Tea  
✅ GitHub Copilot integration (GPT-4o)  
✅ BBolt-based coordination (lock-free, concurrent)  
✅ Multi-agent spawning and coordination  
✅ Specialized agent roles (Architect, Keymaker, Sentinel, Oracle)  
✅ Agent-to-agent communication via event bus  
✅ Task queue and parallel execution  
✅ File locking system for safe concurrent edits  
✅ Safety levels and execution control  

🚧 **In Progress:**
- Agent status dashboard in REPL
- Task visualization and progress tracking
- Agent memory/context sharing
- Task dependencies and prioritization
- Cost tracking for LLM API calls

🎯 **Roadmap:**
- Agent performance metrics
- Configuration file support
- Better error messages and diagnostics
- Stress testing with 10+ concurrent agents

## 🎓 Technology Stack

- **Go** - Performance and clean architecture
- **GitHub Copilot** - GPT-4o powered intelligence
- **Bubble Tea** - Modern terminal UI framework
- **BBolt** - Embedded key-value database (powers etcd/Kubernetes), enables lock-free concurrent agent coordination

## 🤝 Contributing

Experimental AI coding assistant. Ideas, issues, and PRs welcome!

---

**Built to make coding with AI more powerful and coordinated.**
