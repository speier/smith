# Lotus 🪷

React‑style terminal UI components with Flexbox layout in pure Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/speier/smith/pkg/lotus.svg)](https://pkg.go.dev/github.com/speier/smith/pkg/lotus) [![Go Report Card](https://goreportcard.com/badge/github.com/speier/smith)](https://goreportcard.com/report/github.com/speier/smith)

## Why Lotus
- ⚛️ Declarative: describe UI with `Render()` trees (no message boilerplate)
- 📐 Flexbox: CSS‑like layout (gap, padding, align, justify)
- 🪶 Fast: tiny diffs + cached style (sub‑µs updates)
- 🔧 Modular: use vdom, layout, or render independently

## Install
```bash
go get github.com/speier/smith/pkg/lotus
```

## 30‑Second Example
```go
package main
import "github.com/speier/smith/pkg/lotus"

type App struct{}
func (App) Render() *lotus.Element {
  return lotus.VStack(
    lotus.Text("💬 Chat Room"),
    lotus.Text("Messages here..."),
  ).WithGap("1").WithPaddingY("1")
}
func main(){ lotus.Run(App{}) }
```

## Building Blocks
- Layout: `VStack`, `HStack`, `Box`, `Text`
- Components: `TextInput`, `TextBox`, `ProgressBar`, `Tabs`, `Select`, `Checkbox`, `Radio`, `Modal`
- Markup (optional): `lotus.Markup("<box><text>Hello</text></box>")`
- Styles map (inline): `elem.Styles["flex"] = "1"`, percentages and numbers supported

## Flexbox Cheatsheet (Lotus → CSS)
- `VStack` / `HStack` → `display:flex; flex-direction:column|row`
- `.WithGap("1")` → `gap: 1` (terminal cell units)
- `.WithPaddingY("1")` → vertical padding
- `.WithAlignItems(AlignItemsCenter)` → `align-items:center`
- `.WithJustifyContent(...)` → `justify-content:*`
- `.WithFlexGrow(1)` → `flex-grow:1`
- `.WithTextAlign(TextAlignCenter)` on text → `text-align:center`

## Performance Snapshot
- Text change diff: ~245ns
- Complex tree diff (≈10 elems): ~150ns
- Cached style parse: 0 allocs

## Dev Mode
```bash
LOTUS_DEV=true go run main.go
```
DevTools panel (Ctrl+T), hot reload, build errors surfaced inline.

## Minimal API
```go
type App interface { Render() *lotus.Element }
lotus.Run(appOrElementOrMarkup)
```

## More
See `docs/` and `pkg/lotus/examples/` for advanced usage, styling, tests, benchmarks.

## License
MIT (see root `LICENSE`).

Made with 🪷 as part of the Smith project.
