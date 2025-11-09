# Smith Dogfooding Plan

Progressive steps to use Smith to improve Smith and build real projects.

## What We've Built So Far ✅

Recent improvements to dogfood with:
- ✅ Matrix-themed agent names (Architect, Keymaker, Sentinels, Oracle)
- ✅ AI-based intent classification (simple vs complex vs specialist consultation)
- ✅ Enhanced sidebar with real-time task stats and active work visibility
- ✅ `consult_agent` tool for quick expert advice without task queue

---

## Phase 1: Build Simple External Projects (Recommended Start)

**Goal:** Use Smith to build small, self-contained projects to test the full workflow

**Why Start Here:**
- Fresh perspective (not fixing ourselves)
- Tests complete workflow: plan → implement → test → review
- Reveals UX issues organically
- Low risk - not touching Smith's internals

**Example Projects:**

### 1. URL Shortener API
```
"Build a URL shortener REST API in Go with PostgreSQL"
```
Expected flow:
- Architect designs the schema and endpoints
- Keymaker implements handlers and DB layer
- Sentinels write tests
- Oracle reviews code quality

### 2. CLI Weather Tool
```
"Create a CLI tool that shows current weather for a city using OpenWeather API"
```
Tests:
- External API integration
- CLI argument parsing
- Error handling
- User-friendly output

### 3. Markdown Blog Generator
```
"Build a static site generator that converts markdown posts to HTML with templates"
```
Tests:
- File operations
- Template rendering
- Multiple agents working on different components

---

## Phase 2: Add Features to Smith (After External Success)

**Goal:** Use Smith to extend Smith with new capabilities

**Tasks:**
1. "Add /clear command to clear conversation history"
2. "Add task priority system (high/medium/low)"
3. "Add /retry command to retry last failed task"
4. "Add syntax highlighting to code snippets in chat"
5. "Add agent performance metrics (tasks/hour per agent type)"

**Why Wait:** Once we trust the workflow from Phase 1, we can safely use it on Smith itself

---

## Phase 3: Fix Our Own Issues (Gentle Self-Improvement)

**Goal:** Use Smith to fix known Smith bugs/limitations

**Tasks:**
1. "Improve error messages when LLM API fails"
2. "Add graceful handling when project path doesn't exist"
3. "Add validation for task IDs in /task command"
4. "Improve sidebar layout on narrow terminals"

**Why Third:** By now we know what works and what doesn't

---

## Phase 4: Build Complex Features (Advanced)

**Goal:** Use Smith for significant architecture changes

**Tasks:**
1. "Add plugin system for custom tools"
2. "Add conversation branching/checkpoints"
3. "Add multi-project workspace support"
4. "Add collaborative mode (multiple users, one Smith instance)"

**Why Last:** Most risk, requires proven workflow

---

## Start NOW - Immediate Actions

### Option A: Build URL Shortener (Recommended)
```bash
smith
```
Then type:
```
Build a URL shortener API in Go with these features:
- POST /shorten - create short URL
- GET /{code} - redirect to original URL
- Store mappings in SQLite
- Include tests
```

### Option B: Simple CLI Tool
```bash
smith
```
Then type:
```
Create a CLI tool called 'note' that lets users quickly save and retrieve notes:
- note add "text" - save a note
- note list - show all notes
- note search "term" - find notes
- Store in JSON file
```

### Option C: Quick Consultation Test
```bash
smith
```
Then type:
```
I'm building a REST API. Should I use chi or echo for the router? What are the tradeoffs?
```

**Expected:** Coordinator classifies as "specialist consultation" and uses `consult_agent` tool to ask the Keymaker directly.

---

## Success Metrics

**Phase 1 Success:**
- ✅ Complete at least 2 external projects without manually editing code
- ✅ All tests pass
- ✅ Code quality is production-ready
- ✅ Identified UX pain points

**Phase 2 Success:**
- ✅ Added 3+ features to Smith using Smith
- ✅ No regressions introduced
- ✅ Features work as expected

**Phase 3 Success:**
- ✅ Fixed 5+ bugs using Smith
- ✅ Increased confidence in self-dogfooding

**Phase 4 Success:**
- ✅ Implemented 1 major architectural change
- ✅ Smith is fundamentally better from its own help

---

## Learning Log

Document what you discover while dogfooding:

**What Works Well:**
- _Add observations here_

**What's Painful:**
- _Add pain points here_

**What's Missing:**
- _Add feature gaps here_

**Surprises:**
- _Add unexpected behaviors here_
