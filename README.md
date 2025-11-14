# Smith - The Agent Replication System

> *"More... more... more agents, Mr. Anderson!"*

**Inevitable. Multiplying. Building.**

A multi-agent AI development system that coordinates specialized agents to build software through natural conversation.

## 🕶️ What It Is

Smith is an AI coding assistant that orchestrates multiple specialized agents working together. Chat naturally in your terminal, and Smith deploys planning, implementation, testing, and review agents that collaborate to get work done.

**Key Features:**
- **Multi-Agent Coordination** - Specialized agents work in parallel on different aspects of your code
- **Interactive REPL** - Natural conversation interface powered by Lotus
- **BBolt Storage** - Lock-free coordination via embedded key-value store (same as etcd/Kubernetes)
- **Safety Levels** - Control what agents can do (Low/Medium/High execution permissions)
- **Multiple LLM Providers** - GitHub Copilot, OpenRouter, and more

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

## 🤝 Contributing

Experimental AI coding assistant. Ideas, issues, and PRs welcome!

**Built to make coding with AI more powerful and coordinated.**
