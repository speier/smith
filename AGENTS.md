# Agent Guidelines - Smith Development

🚨 **CRITICAL: NEVER CREATE MARKDOWN FILES WITHOUT EXPLICIT USER REQUEST** 🚨

🐚 **IMPORTANT: DON'T ASSUME BASH - USE PORTABLE SHELL COMMANDS!** 🐚

📦 **LOTUS FRAMEWORK: THINK IN FLEXBOX - IT'S LIKE HTML/CSS!** 📦

🖥️ **TUI APPS: YOU CAN'T SEE THE OUTPUT!** 🖥️

> **TUI Application Testing:**
> 
> **⚠️ CRITICAL: Terminal UI apps take over the screen - you cannot see their output!**
> 
> When testing Lotus/TUI applications:
> - ❌ **DON'T** try to run TUI apps and read their output
> - ❌ **DON'T** expect to see visual rendering in terminal output
> - ❌ **DON'T** use `go run` on TUI apps expecting to debug visually
> - ❌ **DON'T** add debug logging as a substitute for tests
> - ✅ **DO** write unit tests instead (`go test ./...`)
> - ✅ **DO** ask user to test interactively when visual testing is needed
> - ✅ **DO** focus on code correctness and test coverage
> - ✅ **DO** improve testing infrastructure if something is hard to test
>
> **Why this matters:**
> - TUI apps use raw terminal mode (hide cursor, take full screen)
> - Running them in background doesn't show UI
> - Output appears empty/frozen even when app works
> - Only user can see and interact with the actual UI
> - **We have extensive testing infrastructure - use it!**
>
> **Testing Philosophy:**
> - 🧪 **Tests are your eyes** - you can't see TUI output, but tests can verify behavior
> - 🧪 **Write tests first** - before adding debug logs or trying to run apps
> - 🧪 **Improve test infrastructure** - if something is hard to test, add test helpers
> - 🧪 **Test coverage matters** - aim for >80% on new code
> - 🧪 **Integration tests exist** - use snapshot tests, layout tests, rendering tests
>
> **Example - WRONG:**
> ```bash
> go run examples/chat/main.go  # Can't see anything!
> # Or adding: fmt.Fprintf(os.Stderr, "DEBUG: %v\n", value)
> ```
>
> **Example - CORRECT:**
> ```bash
> go test ./pkg/lotus/primitives/... -v  # Tests show real output
> # Or ask: "Can you test the chat app and let me know if typing works?"
> # Or write: TestInputStatePreservation to verify reconciliation
> ```
>
> **Available Test Infrastructure:**
> - ✅ `primitives/*_test.go` - Unit tests for components
> - ✅ `layout/*_test.go` - Layout engine tests
> - ✅ `snapshot/` - Snapshot testing framework
> - ✅ `testing/` - Test helpers and utilities
> - ✅ Mock implementations for testing
>
> **When something doesn't work:**
> 1. ✅ Write a failing test that reproduces the issue
> 2. ✅ Fix the code to make the test pass
> 3. ✅ Add more tests for edge cases
> 4. ❌ Don't add debug logs and try to run the TUI app

> **Lotus UI Framework Philosophy:**
> 
> **🎯 Core Principle: "How would I do this in HTML/CSS Flexbox?"**
> 
> Lotus is designed to mirror standard CSS Flexbox behavior. When building UIs:
> 1. ✅ **First**: Think about how you'd solve it in HTML/CSS
> 2. ✅ **Then**: Apply the same approach in Lotus
> 3. ✅ **If missing**: Add the CSS feature to Lotus (don't work around it!)
> 
> **CSS Flexbox Equivalents:**
> - ✅ `VStack` = `<div style="display: flex; flex-direction: column">`
> - ✅ `HStack` = `<div style="display: flex; flex-direction: row">`
> - ✅ `WithGap("1")` = `gap: 1rem` (space BETWEEN children)
> - ✅ `WithPaddingY("1")` = `padding: 1rem 0` (space before/after ALL children)
> - ✅ `WithMarginY("1")` = `margin: 1rem 0` (outer spacing on container)
> - ✅ `WithAlignItems(AlignItemsCenter)` = `align-items: center` (cross-axis)
> - ✅ `WithJustifyContent(...)` = `justify-content: ...` (main-axis)
> - ✅ `WithFlexGrow(1)` = `flex-grow: 1` (fill available space)
> - ✅ `.WithTextAlign(TextAlignCenter)` = `text-align: center` (for text)
>
> **❌ ANTI-PATTERNS - Don't Do These:**
> - ❌ Using empty `Text("")` elements for spacing (use `gap`, `padding`, `margin` instead)
> - ❌ Manual centering calculations (use `align-items`, `justify-content`, `text-align`)
> - ❌ Custom layout logic (leverage flexbox engine)
> - ❌ Working around missing features (add them to Lotus instead!)
>
> **✅ CORRECT PATTERNS:**
> ```go
> // GOOD: CSS flexbox approach
> lotus.VStack(
>     lotus.Text("Logo"),
>     lotus.Text("Content"),
> ).
>     WithAlignItems(lotus.AlignItemsCenter).  // Center horizontally
>     WithGap("1").                             // 1 line between items
>     WithPaddingY("1")                         // 1 line top/bottom padding
>
> // BAD: Empty elements for spacing
> lotus.VStack(
>     lotus.Text(""),      // ❌ Don't do this
>     lotus.Text("Logo"),
>     lotus.Text(""),      // ❌ Don't do this
>     lotus.Text("Content"),
>     lotus.Text(""),      // ❌ Don't do this
> )
> ```
>
> **Important Details:**
> - ✅ `align-items` positions the CHILD BOX in cross-axis
> - ✅ `text-align` centers TEXT LINES within the box
> - ✅ Multi-line text needs BOTH: `WithAlignItems()` on parent AND `WithTextAlign()` on text
> - ✅ `gap` adds space BETWEEN children (not before first or after last)
> - ✅ `padding` adds space INSIDE the container (before first and after last child)
> - ✅ Default: `VStack`/`Box`/`HStack` all have `align-items: stretch` (CSS default)
>
> **When Missing a Feature:**
> 1. ✅ Check if CSS has it (gap, padding, margin, align-items, etc.)
> 2. ✅ Add it to Lotus: ComputedStyle → resolver → layout engine → Element API
> 3. ✅ Write tests to verify it works like CSS
> 4. ❌ Don't create workarounds with empty elements or manual calculations
>
> **Testing:**
> - 🧪 Write tests that mimic HTML/CSS behavior
> - 🧪 Tests catch subtle bugs (e.g., rune vs byte counting in Unicode)
> - 🧪 Integration tests with real component structures reveal layout issues

> **Shell Commands:**
> - ❌ NO bash-specific syntax (heredocs, `[[`, `source`, `&&`, etc.)
> - ✅ ONLY POSIX-compliant commands that work in fish, bash, zsh, sh
> - ✅ Check user's shell from context before writing commands
> - ✅ When in doubt: use `printf`, `test`, separate lines, or Go code
> - See Rule #5 below for details

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
- ❌ **NO README.md files inside packages** (internal/*, pkg/*)

**If documentation is TRULY mandatory:**
- ✅ Put it in `docs/` folder, not in code packages
- ✅ Use descriptive names: `docs/session-architecture.md` not `internal/session/README.md`
- ✅ But still: **ASK FIRST** before creating any .md file

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
**NEVER use shell-specific features** - contributors use different shells (bash, zsh, fish, sh, etc.)

**⚠️ CRITICAL: Don't assume bash is available!** 
- Check user's shell from context (terminal output, environment info)
- Use POSIX-compliant commands that work in ALL shells
- Test commands work in sh/bash/zsh/fish before suggesting them
- When in doubt, ask or use Go code instead

**Shell-Specific Features to AVOID:**
- ❌ Heredocs (`<< EOF`) - fish doesn't support them
- ❌ Bash arrays (`arr=(1 2 3)`) - not portable
- ❌ Process substitution (`<(command)`) - bash/zsh only
- ❌ Bash-specific syntax (`[[`, `source`, `&&` in same line, etc.)
- ❌ Command substitution in command position - fish requires different syntax

**Use PORTABLE alternatives:**
- ✅ `printf "line1\nline2\n" > file` instead of heredocs
- ✅ `echo "content" > file` for simple content
- ✅ Temporary files instead of process substitution
- ✅ POSIX-compliant syntax (`[`, `.` instead of `source`)
- ✅ Separate commands with `;` or multiple lines
- ✅ Use `test -d` or `test -f` instead of `[[ ]]`

**Fish Shell Specifics:**
- ❌ `$(command)` in command position - fish doesn't allow this
- ✅ Use `set var (command)` and then `$var` for command substitution
- ❌ `command1 && command2` - works but prefer separate lines
- ✅ `if test -d dir; then mv dir dest; end` for conditionals

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

### Rule #6: All Tests Must Pass 100%
**NO EXCEPTIONS** - Every test run must show 100% passing tests.

**Test Requirements:**
- ✅ Run `go test ./...` after making code changes
- ✅ ALL tests must pass - no failures, no skips (unless pre-existing)
- ✅ Fix any test failures before reporting completion
- ✅ Aim for >80% code coverage on new code
- ✅ Use table-driven tests where appropriate

**Token Conservation:**
- ⚠️ Don't run tests after EVERY small change during iteration
- ✅ Run tests when: logic changes, before reporting completion, or user asks
- ✅ Skip tests for: simple comment changes, test updates, style-only changes
- ✅ Use judgment - balance confidence with token efficiency

**When Tests Fail:**
1. ❌ DO NOT report "done" if tests are failing
2. ✅ Fix the failing tests immediately
3. ✅ Re-run tests until 100% pass
4. ✅ Only then report completion

**Example - CORRECT:**
```
Running: go test ./...
✅ All tests pass! (100%)
```

**Example - WRONG:**
```
❌ "Most tests pass, just one small failure"
❌ "Tests pass except for that edge case"
❌ "Done! (Tests are failing but the feature works)"
```

### Rule #7: Zero Linter Errors - ALWAYS
**NO EXCEPTIONS** - Code must pass `golangci-lint run ./...` with ZERO errors.

**Why This Matters:**
- Prevents wasting tokens writing code that needs fixing
- Catches bugs early (unchecked errors, unused code, etc.)
- Maintains consistent code quality
- Saves time - fix issues as you write, not after

**Linting Requirements:**
- ✅ Check for errors: `_ =` for intentionally ignored errors
- ✅ No unused variables, functions, imports, or types
- ✅ No empty if branches (use comments if intentional)
- ✅ Handle all error returns (or explicitly ignore with `_`)
- ✅ **Test cleanup is important**: Use `defer func() { _ = resource.Close() }()` in tests
- ✅ **Production code must be perfect**: Zero tolerance for unchecked errors

**When Writing Code:**
1. ✅ Always check error returns: `if err != nil` or `_ = funcCall()`
2. ✅ Remove unused code immediately
3. ✅ Run `golangci-lint run ./...` before reporting completion
4. ✅ Fix ALL issues before saying "done"
5. ✅ In tests: Wrap cleanup in anonymous functions to satisfy linter

**Example - CORRECT:**
```go
// Intentionally ignore non-critical error
_ = cfg.Save()

// Handle error properly
if err := db.Update(); err != nil {
    return fmt.Errorf("update failed: %w", err)
}
```

**Example - WRONG:**
```go
cfg.Save()  // ❌ Unchecked error return

func unused() {}  // ❌ Unused function

import "fmt"  // ❌ Unused import (if not used)
```

**Before Reporting Done:**
```bash
golangci-lint run ./...
# Must show: (empty output = success)
✅ No issues found!
```

### Rule #8: NEVER Delete Files During Refactoring
**CRITICAL** - Files can't be recovered if not in git!

**When refactoring/consolidating/moving files:**
- ❌ DO NOT delete source files until replacement is fully working
- ❌ DO NOT `rm file.go` and then create `new_file.go`
- ✅ DO create the new file FIRST, verify it works, THEN delete old
- ✅ DO move to archive/temp folder if unsure: `mv old.go /tmp/old.go.bak`

**Safe Refactoring Pattern:**
```bash
# WRONG - Delete first, create later (file lost if creation fails!)
rm old_test.go
cat > new_test.go << EOF
...
EOF

# CORRECT - Create first, verify, then delete
cat > new_test.go << EOF
...
EOF
go test ./...  # Verify it works!
rm old_test.go  # Only now safe to delete

# BEST - Move to backup first
mv old_test.go /tmp/old_test.go.bak
cat > new_test.go << EOF
...
EOF
go test ./...  # Verify it works!
# If success: rm /tmp/old_test.go.bak
# If failure: mv /tmp/old_test.go.bak old_test.go
```

**When consolidating multiple files:**
```bash
# WRONG
cat file1.go file2.go > combined.go && rm file1.go file2.go

# CORRECT
cat file1.go file2.go > combined.go  # Create first
go test ./...  # Verify works
rm file1.go file2.go  # Only delete after success

# BEST
mv file1.go /tmp/file1.go.bak
mv file2.go /tmp/file2.go.bak
cat /tmp/file1.go.bak /tmp/file2.go.bak > combined.go
go test ./...  # Verify
# Only delete backups after confirming success
```

**Why This Matters:**
- Files not in git = permanently lost if deleted prematurely
- Refactoring often fails halfway through
- Always have an escape route

### Rule #9: Lotus Import Guidelines
**PREFERRED** - Use public API imports for application code

**Import Pattern:**
```go
import (
    "github.com/speier/smith/pkg/lotus"
    "github.com/speier/smith/pkg/lotusui"
)
```

**Guidelines:**
- ✅ **DON'T** import `pkg/lotus/primitives` in application code
- ✅ **DON'T** import `pkg/lotus/vdom` in application code
- ✅ **DON'T** import `pkg/lotus/runtime` in application code
- ✅ **DO** use `lotus.CreateInput()` instead of `primitives.CreateInput()`
- ✅ **DO** use `lotus.Command` instead of `primitives.Command`
- ✅ **DO** use `lotus.VStack` instead of `vdom.VStack`
- ✅ **DO** import lotusui directly (package name matches import path now)

**Exposing New APIs:**
- When adding features that apps need, expose them in `pkg/lotus/lotus.go`
- Only expose what's needed - don't over-expose internal APIs
- Add incrementally as use cases arise
- Document with clear comments showing usage examples

**Example - WRONG:**
```go
import (
    "github.com/speier/smith/pkg/lotus"
    "github.com/speier/smith/pkg/lotus/primitives"  // ❌ Don't import internal packages
)

func MyApp() {
    input := primitives.CreateInput("...", handler)  // ❌
}
```

**Example - CORRECT:**
```go
import (
    "github.com/speier/smith/pkg/lotus"
    "github.com/speier/smith/pkg/lotusui"           // ✅ Clean import
)

func MyApp() {
    input := lotus.CreateInput("...", handler)      // ✅ Use public API
    modal := lotusui.NewModal()                     // ✅ Direct usage
}
```

---

## 📋 Development Workflow

### When Starting Work
1. Read AGENTS.md (this file)
2. Understand what the user actually asked for
3. Only create/modify files necessary for that request
4. Don't add "extras" or "improvements" unless asked

### When User Asks to Build/Test
1. **PREFER `go test ./...` over building** - Run tests to validate changes
2. **ALL TESTS MUST PASS 100%** - No failing tests are acceptable
3. **During development: Use `go run .` instead of building** - User often runs directly without building
4. Only build when specifically asked or for production/release
5. Use `go build` directly in terminal (no wrapper scripts)
6. Run the binary directly for manual testing when needed
7. Report results
8. **DO NOT** create wrapper scripts unless explicitly requested
9. **DO NOT** automatically build after making code changes - just run tests or `go run .`

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
3. **ALL TESTS MUST PASS 100%** - If tests fail, fix them before reporting
4. Report what was done (and test results if tests were run)
5. Stop and wait for next instruction
6. **DO NOT** automatically build unless explicitly asked
7. **DO NOT** create documentation about the change unless asked
8. **DO NOT** run tests for trivial changes (prompt text, styling, documentation-only changes)

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

## 🕶️ Agent Responsibilities

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

- 🏛️ **Planning Agent** (The Architect) - Breaks down features into tasks
- 🔑 **Implementation Agent** (The Keymaker) - Writes code for tasks
- 🦑 **Testing Agent** (Sentinels) - Creates test suites
- 🔮 **Review Agent** (The Oracle) - Reviews code quality

These will be defined in the application code, not in AGENTS.md.

---

**Version:** 1.0  
**Last Updated:** 2025-10-11  
**These rules take precedence over everything else.**

