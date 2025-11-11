package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
	"github.com/speier/smith/pkg/lotus/components"
)

// KitchenSink demonstrates all Lotus features in one app
type KitchenSink struct {
	// Tabs
	tabs *components.Tabs

	// Components for each tab
	formsDemo    *FormsDemo
	selectDemo   *SelectDemo
	scrollDemo   *ScrollDemo
	progressDemo *ProgressDemo
	modalDemo    *ModalDemo
}

// NewKitchenSink creates the kitchen sink app
func NewKitchenSink() *KitchenSink {
	app := &KitchenSink{
		formsDemo:    NewFormsDemo(),
		selectDemo:   NewSelectDemo(),
		scrollDemo:   NewScrollDemo(),
		progressDemo: NewProgressDemo(),
		modalDemo:    NewModalDemo(),
	}

	// Create tabs
	app.tabs = components.NewTabs().
		WithID("main-tabs").
		WithTabs([]components.Tab{
			{Label: "1:Forms", Content: app.formsDemo},
			{Label: "2:Select", Content: app.selectDemo},
			{Label: "3:Scroll", Content: app.scrollDemo},
			{Label: "4:Progress", Content: app.progressDemo},
			{Label: "5:Modal", Content: app.modalDemo},
		}).
		WithActive(0).
		WithTabBarStyle(components.TabBarStyleLine)

	return app
}

// Render renders the main app
func (app *KitchenSink) Render() *lotus.Element {
	header := lotus.Box(
		lotus.VStack(
			lotus.Text("🪷 Lotus Kitchen Sink").
				WithStyle("color", "#00ff00").
				WithStyle("font-weight", "bold"),
			lotus.Text("Showcasing all components and features").
				WithStyle("color", "#808080"),
			lotus.Text(""),
			lotus.Text("Controls: Tab=next field | ←/→=switch tabs | Ctrl+1-5=direct tab | Space/Enter=select | Esc=close | Ctrl+T=DevTools").
				WithStyle("color", "#606060"),
		),
	).WithBorderStyle(lotus.BorderStyleRounded)

	// Pass Tabs component directly (not pre-rendered) so event routing works
	return lotus.VStack(
		header,
		lotus.Box(app.tabs).WithFlexGrow(1),
	)
}

// --- Forms Demo ---

type FormsDemo struct {
	name      *components.TextInput
	email     *components.TextInput
	subscribe *components.Checkbox
	theme     *components.RadioGroup
	message   *components.TextBox
}

func NewFormsDemo() *FormsDemo {
	demo := &FormsDemo{
		message: components.NewTextBox().WithAutoScroll(true),
	}

	demo.name = components.NewTextInput().
		WithID("name-input").
		WithPlaceholder("Enter your name...").
		WithOnChange(func(value string) {
			demo.message.AppendLine(fmt.Sprintf("Name changed: %s", value))
		})

	demo.email = components.NewTextInput().
		WithID("email-input").
		WithPlaceholder("Enter your email...").
		WithOnChange(func(value string) {
			demo.message.AppendLine(fmt.Sprintf("Email changed: %s", value))
		})

	demo.subscribe = components.NewCheckbox().
		WithID("subscribe-checkbox").
		WithLabel("Subscribe to newsletter").
		WithIcon(components.CheckboxIconSquare).
		WithOnChange(func(checked bool) {
			demo.message.AppendLine(fmt.Sprintf("Subscribe: %v", checked))
		})

	demo.theme = components.NewRadioGroup().
		WithID("theme-radio").
		WithOptions([]components.RadioOption{
			{Label: "Light Theme", Value: "light"},
			{Label: "Dark Theme", Value: "dark"},
			{Label: "Auto Theme", Value: "auto"},
		}).
		WithSelected("dark").
		WithIcon(components.RadioIconDefault).
		WithOnChange(func(value string) {
			demo.message.AppendLine(fmt.Sprintf("Theme changed: %s", value))
		})

	demo.message.AppendLine("🎨 Forms Demo - Try the inputs!")
	demo.message.AppendLine("Changes will be logged here in real-time")
	demo.message.AppendLine("")

	return demo
}

func (d *FormsDemo) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(lotus.Text("📝 Forms & Inputs")).
			WithBorderStyle(lotus.BorderStyleRounded).
			WithStyle("color", "#00ff00"),

		lotus.Box(
			lotus.VStack(
				lotus.Text("Name:"),
				d.name,
				lotus.Text(""),
				lotus.Text("Email:"),
				d.email,
				lotus.Text(""),
				d.subscribe,
				lotus.Text(""),
				lotus.Text("Theme Preference:"),
				d.theme,
			),
		).WithBorderStyle(lotus.BorderStyleRounded),

		lotus.Box(d.message).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),
	)
}

// IsNode marks FormsDemo as a vdom.Node
func (d *FormsDemo) IsNode() {}

// --- Select Demo ---

type SelectDemo struct {
	country *components.Select
	size    *components.Select
	message *components.TextBox
}

func NewSelectDemo() *SelectDemo {
	demo := &SelectDemo{
		message: components.NewTextBox().WithAutoScroll(true),
	}

	demo.country = components.NewSelect().
		WithID("country-select").
		WithOptions([]components.SelectOption{
			{Label: "🇺🇸 United States", Value: "us"},
			{Label: "🇬🇧 United Kingdom", Value: "gb"},
			{Label: "🇨🇦 Canada", Value: "ca"},
			{Label: "🇩🇪 Germany", Value: "de"},
			{Label: "🇫🇷 France", Value: "fr"},
			{Label: "🇯🇵 Japan", Value: "jp"},
			{Label: "🇦🇺 Australia", Value: "au"},
		}).
		WithPlaceholder("Select a country...").
		WithOnChange(func(index int, value string) {
			demo.message.AppendLine(fmt.Sprintf("Country: %s (index: %d)", value, index))
		})

	demo.size = components.NewSelect().
		WithID("size-select").
		WithStringOptions([]string{"Small", "Medium", "Large", "Extra Large"}).
		WithSelected(1).
		WithOnChange(func(index int, value string) {
			demo.message.AppendLine(fmt.Sprintf("Size: %s", value))
		})

	demo.message.AppendLine("📋 Select Demo - Choose from dropdowns!")
	demo.message.AppendLine("Press Space/Enter to open, Arrow keys to navigate")
	demo.message.AppendLine("")

	return demo
}

func (d *SelectDemo) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(lotus.Text("📋 Select/Dropdown")).
			WithBorderStyle(lotus.BorderStyleRounded).
			WithStyle("color", "#00ff00"),

		lotus.Box(
			lotus.VStack(
				lotus.Text("Country:"),
				d.country,
				lotus.Text(""),
				lotus.Text("Size:"),
				d.size,
			),
		).WithBorderStyle(lotus.BorderStyleRounded),

		lotus.Box(d.message).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),
	)
}

// IsNode marks SelectDemo as a vdom.Node
func (d *SelectDemo) IsNode() {}

// --- Scroll Demo ---

type ScrollDemo struct {
	scrollView *components.ScrollView
}

func NewScrollDemo() *ScrollDemo {
	// Generate lots of content to scroll
	lines := make([]any, 50)
	for i := 0; i < 50; i++ {
		line := fmt.Sprintf("Line %d: This is a very long line that demonstrates scrolling in both directions. Lorem ipsum dolor sit amet.", i+1)
		lines[i] = lotus.Text(line)
	}
	content := lotus.VStack(lines...)

	demo := &ScrollDemo{
		scrollView: components.NewScrollView().
			WithID("scroll-view").
			WithContent(content).
			WithSize(80, 20).
			WithOnScroll(func(x, y int) {
				// Scroll callback
			}),
	}

	return demo
}

func (d *ScrollDemo) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(lotus.Text("📜 ScrollView")).
			WithBorderStyle(lotus.BorderStyleRounded).
			WithStyle("color", "#00ff00"),

		lotus.Box(lotus.Text("Use Arrow keys to scroll | Page Up/Down | Home/End")).
			WithStyle("color", "#808080"),

		lotus.Box(d.scrollView).
			WithFlexGrow(1).
			WithBorderStyle(lotus.BorderStyleRounded),
	)
}

// IsNode marks ScrollDemo as a vdom.Node
func (d *ScrollDemo) IsNode() {}

// --- Progress Demo ---

type ProgressDemo struct {
	progress1 *components.ProgressBar
	progress2 *components.ProgressBar
	progress3 *components.ProgressBar
}

func NewProgressDemo() *ProgressDemo {
	progress1 := components.NewProgressBar(40)
	progress1.SetValue(0.25)

	progress2 := components.NewProgressBar(40)
	progress2.SetValue(0.65)

	progress3 := components.NewProgressBar(40)
	progress3.SetValue(1.0)

	demo := &ProgressDemo{
		progress1: progress1,
		progress2: progress2,
		progress3: progress3,
	}

	return demo
}

func (d *ProgressDemo) Render() *lotus.Element {
	return lotus.VStack(
		lotus.Box(lotus.Text("📊 Progress Bars")).
			WithBorderStyle(lotus.BorderStyleRounded).
			WithStyle("color", "#00ff00"),

		lotus.Box(
			lotus.VStack(
				lotus.Text("Upload - 25%"),
				d.progress1,
				lotus.Text(""),
				lotus.Text("Download - 65%"),
				d.progress2,
				lotus.Text(""),
				lotus.Text("Complete - 100%"),
				d.progress3,
				lotus.Text(""),
				lotus.Text("Progress bars show completion status").
					WithStyle("color", "#808080"),
				lotus.Text("Useful for uploads, downloads, loading states").
					WithStyle("color", "#808080"),
			),
		).WithBorderStyle(lotus.BorderStyleRounded).WithFlexGrow(1),
	)
}

// IsNode marks ProgressDemo as a vdom.Node
func (d *ProgressDemo) IsNode() {}

// --- Modal Demo ---

type ModalDemo struct {
	modal   *components.Modal
	message *components.TextBox
}

func NewModalDemo() *ModalDemo {
	demo := &ModalDemo{
		message: components.NewTextBox().WithAutoScroll(true),
	}

	demo.modal = components.NewModal().
		WithID("demo-modal").
		WithTitle("⚠️ Confirm Action").
		WithContent(
			lotus.VStack(
				lotus.Text("Are you sure you want to proceed?"),
				lotus.Text("This action cannot be undone.").
					WithStyle("color", "#808080"),
			),
		).
		WithButtons([]components.ModalButton{
			{
				Label:   "Cancel",
				Variant: "secondary",
				OnClick: func() {
					demo.message.AppendLine("❌ Action cancelled")
					demo.modal.Close()
				},
			},
			{
				Label:   "Confirm",
				Variant: "danger",
				OnClick: func() {
					demo.message.AppendLine("✅ Action confirmed!")
					demo.modal.Close()
				},
			},
		}).
		WithCloseOnEscape(true)

	demo.message.AppendLine("💬 Modal Demo")
	demo.message.AppendLine("Click 'Show Modal' button to see a dialog")
	demo.message.AppendLine("")

	return demo
}

func (d *ModalDemo) Render() *lotus.Element {
	// Button to show modal
	showButton := lotus.Box(lotus.Text(" Show Modal ")).
		WithBorderStyle(lotus.BorderStyleRounded).
		WithStyle("color", "#ffffff").
		WithStyle("background-color", "#007acc")

	var content *lotus.Element
	if d.modal.Open {
		// Show modal overlay
		content = lotus.VStack(
			lotus.Box(lotus.Text("💬 Modal Demo")).
				WithBorderStyle(lotus.BorderStyleRounded).
				WithStyle("color", "#00ff00"),
			showButton,
			lotus.Box(d.message).
				WithFlexGrow(1).
				WithBorderStyle(lotus.BorderStyleRounded),
			d.modal, // Modal overlay
		)
	} else {
		content = lotus.VStack(
			lotus.Box(lotus.Text("💬 Modal Demo")).
				WithBorderStyle(lotus.BorderStyleRounded).
				WithStyle("color", "#00ff00"),
			showButton,
			lotus.Box(d.message).
				WithFlexGrow(1).
				WithBorderStyle(lotus.BorderStyleRounded),
		)
	}

	return content
}

// IsNode marks ModalDemo as a vdom.Node
func (d *ModalDemo) IsNode() {}

func main() {
	// Create kitchen sink app
	app := NewKitchenSink()

	// Run with DevTools & HMR enabled via LOTUS_DEV=true
	if err := lotus.Run(app); err != nil {
		panic(err)
	}
}
