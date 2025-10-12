# Smith Configuration Architecture

## Two-Level Configuration System

Smith uses a two-level configuration system to separate global user settings from project-specific settings.

---

## 1. Global User Configuration

**Location:** `~/.smith/`

### Files:

#### `config.json` - User preferences and provider settings
```json
{
  "provider": "copilot",
  "model": "gpt-4o",
  "providerOpts": {}
}
```

**Purpose:**
- LLM provider selection (copilot, ollama, openai)
- Model selection
- Provider-specific options
- User-level preferences

**Managed by:** `internal/config/config.go`

---

#### `auth.json` - Authentication tokens (0600 permissions)
```json
{
  "provider": "copilot",
  "data": {
    "refresh_token": "gho_...",
    "access_token": "...",
    "expires_at": 1234567890
  }
}
```

**Purpose:**
- Secure storage of authentication tokens
- Provider-specific auth data
- Automatic token refresh

**Managed by:** `internal/config/config.go`

**Commands:**
- `smith auth login` - Authenticate
- `smith auth status` - Check authentication
- `smith auth logout` - Clear tokens

---

## 2. Project/Local Configuration

**Location:** `./.smith/` (in project directory)

### Directory Structure:

```
.smith/
├── backlog/
│   ├── todo/          # Planned, ready to start
│   ├── wip/           # Work in progress (agent working)
│   ├── review/        # Ready for review
│   └── done/          # Completed
├── inbox/             # Agent questions, unplanned ideas
└── agents/
    └── status.json    # Active agents status
```

### Files:

#### Task Files (Markdown with YAML frontmatter)
```markdown
---
id: task-123
title: Add user authentication
created: 2025-10-12T10:00:00Z
assigned: implementation-agent
priority: high
tags: [feature, auth]
---

## Description
Implement user authentication using JWT...

## Acceptance Criteria
- [ ] User can register
- [ ] User can login
- [ ] Tokens expire after 24h
```

**Purpose:**
- Kanban-style task management
- Task metadata and status
- Agent coordination
- File locking and synchronization

**Managed by:** 
- `internal/coordinator/coordinator.go`
- `internal/orchestrator/orchestrator.go`

---

## Separation of Concerns

### Global (~/.smith/)
✅ User identity & authentication  
✅ Provider configuration  
✅ Personal preferences  
✅ Cross-project settings  

### Local (`./.smith/`)
✅ Project-specific tasks  
✅ Agent coordination  
✅ Task backlog & kanban  
✅ Agent messages  
✅ Project-level config (provider, model, auto-level)  
✅ Per-agent configuration  

---

## Benefits

1. **Portability:** Projects can be moved without losing task state
2. **Multi-user:** Different users can work on same project with own auth
3. **Security:** Auth tokens stored in user home, not in project repo
4. **Git-friendly:** `.smith/` can be gitignored or committed as needed
5. **Isolation:** Each project has independent task management

---

## Implementation Status

### ✅ Completed (Global)
- Configuration loading/saving
- Authentication storage (secure 0600 permissions)
- Provider factory
- Auth commands (login/status/logout)

### ✅ Completed (Local)
- Coordinator reads from projectPath
- Task parsing from TODO.md
- Basic file reading/writing helpers
- Directory structure (todo/wip/review/done)
- Local config support (.smith/config.json)
- Per-agent configuration
- Safety system with embedded YAML rules

### 🔄 TODO
- Migrate to `.smith/backlog/` directory structure (todo/wip/review/done)
- Implement YAML frontmatter parsing
- Agent status tracking
- Inbox message system
- File locking mechanism
- Command execution with auto-level safety checks

---

## Usage

### First Time Setup (Global)
```bash
# Authenticate once per machine
smith auth login

# Check authentication
smith auth status
```

### Per Project (Local)
```bash
# Navigate to your project
cd ~/my-project

# Start Smith (creates .smith/ if needed)
smith

# Chat and create tasks
💬 You: Add user authentication with JWT
🤖 Smith: I can help with that! Let me create a plan...
```

The system automatically:
1. Uses global auth from `~/.smith/auth.json`
2. Creates local `.smith/` for task management
3. Keeps them separate and secure

---

**Status:** Architecture is solid! ✨
