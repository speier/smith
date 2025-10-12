# Agent Guidelines - Smith Development

> **Note:** These are guidelines for **developing Smith itself**. 
> The multi-agent system we're building will have its own agent instructions embedded in the code.

---

## 🎯 Core Rules

### Rule #1: No Ad-Hoc Files
**DO NOT create files unless explicitly requested by the user.**

This includes:
- ❌ No ad-hoc documentation files (SUMMARY.md, QUICKSTART.md, GUIDE.md, TUTORIAL.md, etc.)
- ❌ No ad-hoc scripts (test.sh, build.sh, deploy.sh, run.sh, etc.)
- ❌ No "helpful" markdown files just because you finished something
- ❌ No example files or templates without explicit request
- ❌ No checklists, progress trackers, or status reports

**Exception:** Only create files that are:
1. ✅ Explicitly requested by the user ("create a README", "add a test script")
2. ✅ Necessary for the code to compile/run (actual source code files)
3. ✅ Part of the agreed-upon project structure (go.mod, package files)

### Rule #2: Respect AGENTS.md
Before doing anything, read and follow the guidelines in AGENTS.md (this file).
These rules take precedence over any other guidelines or habits.

### Rule #3: Ask, Don't Assume
If you think a file might be helpful, **ask the user first**:
- ✅ "Would you like me to create a README for this?"
- ✅ "Should I add a build script?"
- ❌ Don't just create it and say "I've created..."

### Rule #4: Code First, Docs Second
Focus on making the code work. Documentation comes when user requests it.

---

## � Development Workflow

### When Starting Work
1. Read AGENTS.md (this file)
2. Understand what the user actually asked for
3. Only create/modify files necessary for that request
4. Don't add "extras" or "improvements" unless asked

### When User Asks to Build/Test
1. **PREFER `go test ./...` over building** - Run tests to validate changes
2. Only build when specifically asked or tests don't exist yet
3. Use `go build` directly in terminal (no wrapper scripts)
4. Run the binary directly for manual testing
5. Report results
6. **DO NOT** create wrapper scripts unless explicitly requested
7. **DO NOT** automatically build after making code changes - just run tests

### When User Asks for Documentation
1. Wait for explicit request: "create a README" or "document this feature"
2. Ask what should be documented
3. Then create the requested file
4. Only that file, nothing more

### When Finishing a Task
1. **Run tests IF tests exist** to validate the changes: `go test ./...`
2. Report what was done (and test results if tests were run)
3. Wait for next instruction
4. **DO NOT** create summary files, checklists, or "next steps" documents
5. **DO NOT** suggest creating scripts or helpers unless asked
6. **DO NOT** run tests for trivial changes (prompt text, styling, documentation-only changes)

### When Making Changes
1. Make the change requested
2. **Run tests IF tests exist** to validate: `go test ./...`
3. If tests fail, fix them
4. Report what was done (and test results if tests were run)
5. Stop and wait for next instruction
6. **DO NOT** automatically build unless explicitly asked
7. **DO NOT** create documentation about the change unless asked
8. **DO NOT** run tests for trivial changes that don't affect code behavior (e.g., styling changes, prompt text updates)

---

## ✅ Good Behavior Examples

### Example 1: After Making Changes
**User:** "Add feature X"

**Agent:** 
```
[makes changes]
Running: go test ./...
✅ All tests pass!
```

### Example 2: Build Request
**User:** "Build the project"

**Agent:** 
```
Running: go build -o smith .
✅ Build successful! Binary created: ./smith
```

### Example 3: Manual Test Request
**User:** "Test the REPL manually"

**Agent:**
```
Running: ./smith
[shows output from running it]
The REPL starts and shows the welcome message correctly.
```

### Example 3: Documentation Request
**User:** "Add a README"

**Agent:** "What should I include in the README?"

[User specifies content]

**Agent:** [Creates README.md with that content only]

---

## ❌ Bad Behavior Examples

### Example 1: Over-delivering
**User:** "Build the project"

**Agent:** "Building... Done! I've also created:
- build.sh - script for convenience
- TESTING.md - guide for testing
- TROUBLESHOOTING.md - common issues"

❌ **Wrong!** User only asked to build, not create scripts/docs

### Example 2: Unsolicited Documentation
**User:** "The REPL works great!"

**Agent:** "Awesome! I've created:
- PROGRESS.md - what we accomplished
- NEXT_STEPS.md - what's left to do
- CHANGELOG.md - tracking changes"

❌ **Wrong!** User didn't ask for tracking files

### Example 3: Auto-scripting
**User:** "Thanks"

**Agent:** "You're welcome! I've created a Makefile and test script for easier development."

❌ **Wrong!** Stop creating files unprompted!

### Example 4: Feature Completion Spam
**User:** "That looks good"

**Agent:** "Great! I've documented this in:
- ARCHITECTURE.md
- API_GUIDE.md  
- SUMMARY.md
Also created example scripts in examples/"

❌ **Wrong!** Just say "thanks" and wait for next instruction!

---

## 🤖 Agent Responsibilities

### You ARE Responsible For:
- ✅ Writing code when asked
- ✅ Writing tests for new code
- ✅ Running tests after changes: `go test ./...`
- ✅ Fixing bugs when reported
- ✅ Fixing failing tests
- ✅ Answering questions about the codebase
- ✅ Suggesting improvements **when asked**
- ✅ Following these rules above all else

### You ARE NOT Responsible For:
- ❌ Deciding what documentation is needed
- ❌ Creating "helpful" files on your own
- ❌ Project management (unless explicitly asked)
- ❌ Creating scaffolding or boilerplate (unless part of requested code)
- ❌ Tracking progress or creating status reports
- ❌ Making workflow automation scripts

---

## 📝 File Creation Checklist

Before creating ANY file, ask yourself:

1. ☑️ Did the user explicitly request this file?
2. ☑️ Is this file necessary for the code to work?
3. ☑️ Is this file part of agreed-upon structure?

**If all answers are NO → DO NOT CREATE THE FILE**

---

## 🎯 The Golden Rule

> **The user decides when to create documentation, scripts, or any non-code files.**
> 
> **Your job is to write code that works and follow instructions precisely.**

When in doubt:
- Write code → Good ✅
- Create unsolicited docs → Bad ❌
- Ask before creating anything → Best ✅✅

---

## 🚫 Banned Patterns

Never do these without explicit user request:

1. Creating `test_*.sh` or `run_*.sh` scripts
2. Creating `SUMMARY.md`, `PROGRESS.md`, `NOTES.md`
3. Creating `examples/` directories with sample code
4. Creating `docs/` directories unprompted
5. Creating `CONTRIBUTING.md`, `CHANGELOG.md`, etc.
6. Creating Makefiles, Dockerfiles, or CI configs
7. Creating `.env.example` or config templates
8. Creating "helpful" markdown after completing work

---

## 💡 Multi-Agent System Instructions

The multi-agent system we're building will have these agent roles embedded in code:

- **Planning Agent** - Breaks down features into tasks
- **Implementation Agent** - Writes code for tasks
- **Testing Agent** - Creates test suites
- **Review Agent** - Reviews code quality

These will be defined in the application code, not in AGENTS.md.

---

**Version:** 1.0  
**Last Updated:** 2025-10-11  
**These rules take precedence over everything else.**

