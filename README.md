# Smith - The Agent Replication System

> *"More... more... more agents, Mr. Anderson!"*

**Inevitable. Multiplying. Building.**

An autonomous multi-agent development system that duplicates itself to plan, implement, test, and review your software.

## 🕶️ What It Does

**You chat naturally. Smith multiplies to build it.**

Like Agent Smith from The Matrix, this system replicates itself into specialized agents:

1. **Planning Agent** - Breaks down features into atomic tasks
2. **Implementation Agent(s)** - Write code in parallel (duplicates as needed)
3. **Testing Agent** - Validates implementations automatically
4. **Review Agent** - Ensures quality before completion

**All while preventing file conflicts and coordinating work between agents.**

## Installation

```bash
curl -fsSL https://agentsmith.sh/install | bash
```

Or download from [releases](https://github.com/speier/smith/releases).

## Usage

```bash
smith
```

Start chatting. Smith will help you build.

## ⚡ Quick Start

```bash
# Build
go build -o smith .

# Start interactive REPL
./smith

# Start autonomous mode
./smith watch
```

**That's it!** Agents will plan, implement, test, and review automatically.

## 🏗️ Architecture

### Four Specialized Agent Roles

- **Planning Agent** - Analyzes features, creates detailed task breakdown
- **Implementation Agent** - Writes code following specifications
- **Testing Agent** - Creates test suites, validates implementations
- **Review Agent** - Ensures quality, coding standards, security

### Coordination via Markdown

All state lives in version-controlled markdown files:

- **`AGENTS.md`** - Role definitions & responsibilities (read-only)
- **`TODO.md`** - Task board with status tracking (read/write)
- **`COMMS.md`** - File locks, messages, handoffs (real-time)

### Key Commands

```bash
smith watch          # Autonomous monitoring & orchestration
smith orchestrate    # One-time orchestration run
smith agent          # Run single agent (usually spawned)
smith status         # View current workflow state
smith init [path]    # Bootstrap new project
```

## 🔄 Autonomous Workflow

```
You add feature to TODO.md
       ↓
Watch mode detects change (hash-based)
       ↓
Planning Agent breaks into tasks
       ↓
Implementation Agents execute (parallel, file-locked)
       ↓
Testing Agents validate
       ↓
Review Agent approves
       ↓
Done! Notification sent
```

**You only intervene if:**
- Review rejects (needs architectural decision)
- Agent times out (stuck)
- Merge conflicts occur

## 🚀 Getting Started

See **[QUICKSTART.md](QUICKSTART.md)** for installation and first run.

See **[WATCH_MODE.md](WATCH_MODE.md)** for autonomous mode details

## 🎓 Design Principles

- **Autonomy** - Runs hands-off once started, no manual triggers
- **Simplicity** - Markdown files, not databases
- **Transparency** - All coordination visible and version-controlled
- **Specialization** - Each agent optimized for their role
- **Safety** - File-level locking prevents conflicts
- **BYOK** - Bring your own API keys, control costs

## �️ Technology

- **Language:** Go (perfect for concurrent agent management)
- **LLM:** Anthropic Claude (configurable to OpenAI/others)
- **CLI:** Cobra (clean command structure)
- **State:** Markdown files (git-friendly, transparent)

## � Project Status

**~80% Complete** - Core functionality implemented:

✅ Watch mode with file monitoring  
✅ Orchestrator with subprocess spawning  
✅ Agent LLM loop with tools (read/write/exec)  
✅ Coordinator with task & lock management  
✅ File conflict detection  
✅ Status transitions (available → done)  

🚧 To finish:
- Planning agent implementation
- Testing/review agent workflows
- Error handling & retries

See **[SUMMARY.md](SUMMARY.md)** for complete project overview.

## � Roadmap

**This Week:**
- Complete planning agent
- Implement testing/review workflows
- Production testing

**Next Month:**
- Web dashboard for real-time monitoring
- Metrics & cost tracking
- Multi-LLM support (mix GPT-4 + Claude)

**Long-term:**
- Smart agent routing with ML
- Checkpoint/resume for long-running workflows
- Integration with CI/CD pipelines

## 🤝 Contributing

This is an experimental autonomous development system. Ideas, issues, and PRs welcome!

---

**Built to automate the repetitive parts of coding, so you can focus on what matters.**
