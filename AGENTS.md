# Agent Guidelines - Smith Development

🚨 **CRITICAL: NEVER CREATE MARKDOWN FILES WITHOUT EXPLICIT USER REQUEST** 🚨

> **When you finish something:**
> - ❌ DO NOT create SUMMARY.md, PROGRESS.md, STATUS.md, MIGRATION_PHASE2.md, etc.
> - ✅ DO just say: "Done! Tests pass." and STOP
>
> **ONLY create .md files when user explicitly says:**
> - "Document this migration"
> - "Create a guide for X"
> - "Write a README explaining Y"
>
> **Default behavior: When in doubt, CREATE NOTHING.**

> **Note:** These are guidelines for **developing Smith itself**. 
> The multi-agent system we're building will have its own agent instructions embedded in the code.

---

## 🎯 Core Rules

### Rule #1: No Ad-Hoc Files - EVER
**NEVER EVER create .md files after completing work.**

**When you finish a task:**
- ❌ DO NOT create MIGRATION_PHASE2.md, PROGRESS.md, SUMMARY.md, STATUS.md
- ❌ DO NOT create documentation "to help track what we built"
- ❌ DO NOT create files "to summarize our progress"
- ✅ DO just say "Done! Tests pass." and **STOP**

**This includes:**
- ❌ No ad-hoc documentation files (SUMMARY.md, QUICKSTART.md, GUIDE.md, TUTORIAL.md, etc.)
- ❌ No ad-hoc scripts (test.sh, build.sh, deploy.sh, run.sh, etc.)
- ❌ No "helpful" markdown files just because you finished something
- ❌ No example files or templates without explicit request
- ❌ No checklists, progress trackers, or status reports
- ❌ No documentation about what you just built

**THE ONLY EXCEPTION - User explicitly says:**
1. ✅ "Create a README" / "Document this feature"
2. ✅ "Add a test script" / "Make a build script"
3. ✅ Actual source code files necessary for the code to compile/run
4. ✅ Files that are part of agreed-upon project structure (go.mod, package files)

### Rule #2: Respect AGENTS.md
Before doing anything, read and follow the guidelines in AGENTS.md (this file).
These rules take precedence over any other guidelines or habits.

### Rule #3: Ask First, NEVER Assume
If you think a file might be helpful, **ask the user first**:
- ✅ "Would you like me to document this migration?"
- ✅ "Should I create a guide for X?"
- ❌ NEVER just create it and say "I've created..."
- ❌ NEVER create it and then delete it (shows you violated the rule)

**Remember:** The user will tell you when they want documentation.

### Rule #4: Code First, Docs ONLY When Requested
Focus on making the code work. Documentation comes **ONLY when user explicitly requests it**.

- ✅ Write code, run tests, report results
- ❌ Write code, run tests, create PROGRESS.md
- ✅ "Done! All tests pass."
- ❌ "Done! I've created MIGRATION_STATUS.md to track progress"

### Rule #5: Shell Commands Must Be Portable
**NEVER use shell-specific features** - team members use different shells (bash, zsh, fish, etc.)

**Shell-Specific Features to AVOID:**
- ❌ Heredocs (`<< EOF`) - fish doesn't support them
- ❌ Bash arrays (`arr=(1 2 3)`) - not portable
- ❌ Process substitution (`<(command)`) - bash/zsh only
- ❌ Bash-specific syntax (`[[`, `source`, etc.)

**Use PORTABLE alternatives:**
- ✅ `printf "line1\nline2\n" > file` instead of heredocs
- ✅ `echo "content" > file` for simple content
- ✅ Temporary files instead of process substitution
- ✅ POSIX-compliant syntax (`[`, `.` instead of `source`)

**Example - WRONG:**
```bash
cat > file.txt << EOF
This won't work in fish
EOF
```

**Example - CORRECT:**
```bash
printf "This works everywhere\n" > file.txt
```

**Or use Go directly:**
```bash
go run script.go  # Better: write a small Go program
```

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
3. **STOP. DO NOT CREATE ANY FILES.**
4. Wait for next instruction
5. **DO NOT** create summary files, checklists, or "next steps" documents
6. **DO NOT** suggest creating scripts or helpers unless asked
7. **DO NOT** run tests for trivial changes (prompt text, styling, documentation-only changes)

**Correct completion response:**
```
✅ All tests pass!
```

**WRONG completion responses:**
```
❌ "Done! I've created SUMMARY.md to track what we built"
❌ "Tests pass! See MIGRATION_PHASE2.md for details"
❌ "Complete! Created STATUS.md for reference"
```

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

### Example 4: Task Completion (CORRECT ✅)
**User:** "That looks good"

**Agent:** "Great! All tests passing. Ready for next step."

[Agent stops here - creates NOTHING]

### Example 5: Task Completion (WRONG ❌)  
**User:** "Perfect!"

**Agent:** "Awesome! I've created MIGRATION_COMPLETE.md to document what we built."

❌ **Wrong!** User didn't ask for documentation

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

### Example 3: Feature Completion Spam
**User:** "Thanks"

**Agent:** "You're welcome! I've created MIGRATION_PHASE2.md to summarize our progress."

❌ **Wrong!** Stop creating files unprompted!

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

**If user says "what's next?" or "that's good" → Answer verbally, create NO files.**

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
9. Creating `MIGRATION_*.md`, `STATUS.md`, `QUICKSTART.md`
10. Creating documentation "to track what we built"

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

