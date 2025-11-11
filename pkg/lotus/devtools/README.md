# Lotus DevTools & Hot Module Reload (HMR)

## Features

- **DevTools Panel**: In-app debug console for logging
- **Hot Module Reload**: Auto-rebuild and restart on file changes
- **State Preservation**: Maintains app state across reloads (coming soon)

## Usage

### Enable DevTools & HMR

Set the `LOTUS_DEV=true` environment variable:

```bash
LOTUS_DEV=true go run main.go
```

### DevTools Panel

- **Toggle**: Press `Ctrl+T` to show/hide the DevTools panel
- **Layout**: DevTools appears at the bottom (30% of screen)
- **Logging**: Use `devtools.Log()` for debug messages

### Hot Module Reload

When `LOTUS_DEV=true`:
- Watches all `.go` files in your project directory (recursive)
- Debounces changes (300ms) to avoid rebuild spam
- Auto-rebuilds on file save
- **Shows build errors in DevTools** - See compile errors line by line
- **DevTools updates in real-time** - Logs appear immediately
- Restarts process with state preservation (WIP)

**What gets watched:**
- All `.go` files recursively
- Excludes: `.git`, `node_modules`, `vendor`, `.idea`, `.vscode`, and hidden directories

**Build error handling:**
- Capture compile errors from `go build`
- Display first 10 lines of errors in DevTools
- Doesn't restart on build failures

## Example

```go
package main

import (
    "github.com/speier/smith/pkg/lotus"
    "github.com/speier/smith/pkg/lotus/components"
)

type MyApp struct {
    input *components.TextInput
}

func NewMyApp() *MyApp {
    return &MyApp{
        input: components.NewTextInput().
            WithPlaceholder("Type here..."),
    }
}

func (app *MyApp) Render() *lotus.Element {
    return lotus.VStack(
        lotus.Box("My App"),
        lotus.Box(app.input),
    )
}

func main() {
    // DevTools + HMR auto-enabled when LOTUS_DEV=true
    if err := lotus.Run(NewMyApp()); err != nil {
        panic(err)
    }
}
```

**Important:** Run from the package directory (where `main.go` is):

```bash
cd myapp
LOTUS_DEV=true go run main.go
```

Now edit `main.go` and save - the app will auto-reload!

**Advanced:** Override build target with `LOTUS_BUILD_TARGET` env var if needed.

## Keyboard Shortcuts

- `Ctrl+C` or `Ctrl+D` - Exit application
- `Ctrl+T` - Toggle DevTools panel

## Implementation Details

### File Watching
- Uses `fsnotify` for file system events
- 300ms debounce to batch rapid changes
- Filters to `.go` files only

### Rebuild Process
1. File change detected → debounced
2. Save current app state to `/tmp/lotus-state-{pid}.json`
3. Run `go build -o /tmp/lotus-hmr-app`
4. Start new process with `LOTUS_STATE_PATH` env var
5. New process loads state and resumes
6. Old process exits

### State Preservation (TODO)
Currently saves placeholder state. Full state serialization coming soon for:
- TextInput values and cursor position
- TextBox content and scroll position
- Custom component state

## Limitations

- State preservation is basic (work in progress)
- Only watches current working directory and subdirectories
- Build errors shown in DevTools but don't prevent retry
- Terminal must support ANSI escape codes
