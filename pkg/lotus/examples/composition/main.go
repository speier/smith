package main

import (
	"fmt"

	"github.com/speier/smith/pkg/lotus"
)

// Example: Component Composition - The React Way
//
// This demonstrates how to build reusable components with props,
// similar to React's component composition model.

// 1. Define Props Types (like React props)
type CardProps struct {
	Title       string
	Description string
	Color       string
	Width       int
}

type UserProps struct {
	Name   string
	Email  string
	Avatar string
	Online bool
}

type ButtonProps struct {
	Label   string
	Color   string
	OnClick func()
}

// 2. Create Reusable Components (like React functional components)

// Card component - generic card container
func Card(props CardProps) interface{} {
	return lotus.VStack(
		// Header
		lotus.Box("card-header",
			lotus.Text(props.Title),
		).
			Color(props.Color).
			Padding(1),
		// Body
		lotus.Box("card-body",
			lotus.Text(props.Description),
		).
			Padding(1).
			Flex(1),
	).
		Border("1px solid").
		BorderColor(props.Color).
		Width(fmt.Sprintf("%d", props.Width))
}

// UserCard component - composes Card with user-specific content
func UserCard(props UserProps) interface{} {
	status := "● Offline"
	statusColor := "#888"
	if props.Online {
		status = "● Online"
		statusColor = "#0f0"
	}

	return Card(CardProps{
		Title: props.Name,
		Description: fmt.Sprintf("%s\n%s %s",
			props.Email,
			props.Avatar,
			status,
		),
		Color: statusColor,
		Width: 30,
	})
}

// Button component - interactive element
func Button(props ButtonProps) interface{} {
	// Note: Actual onClick handling would need component registration
	return lotus.Box("button",
		lotus.Text(fmt.Sprintf("[ %s ]", props.Label)),
	).
		Color(props.Color).
		Padding(1).
		Border("1px solid")
}

// 3. Compose Components Together (like JSX composition)

type Dashboard struct {
	users []UserProps
}

func NewDashboard() *Dashboard {
	return &Dashboard{
		users: []UserProps{
			{Name: "Alice", Email: "alice@example.com", Avatar: "👩", Online: true},
			{Name: "Bob", Email: "bob@example.com", Avatar: "👨", Online: false},
			{Name: "Charlie", Email: "charlie@example.com", Avatar: "🧑", Online: true},
		},
	}
}

func (d *Dashboard) Render() *lotus.Element {
	return lotus.VStack(
		d.renderHeader(),
		d.renderUserList(),
		d.renderFooter(),
	).Render()
}

func (d *Dashboard) renderHeader() interface{} {
	return Card(CardProps{
		Title:       "User Dashboard",
		Description: "Component Composition Example",
		Color:       "blue",
		Width:       100,
	})
}

func (d *Dashboard) renderUserList() interface{} {
	// Map users to components (like React's .map())
	userCards := make([]interface{}, len(d.users))
	for i, user := range d.users {
		userCards[i] = UserCard(user)
	}

	return lotus.HStack(userCards...).Padding(1)
}

func (d *Dashboard) renderFooter() interface{} {
	return lotus.HStack(
		Button(ButtonProps{Label: "Refresh", Color: "green"}),
		Button(ButtonProps{Label: "Settings", Color: "blue"}),
		Button(ButtonProps{Label: "Exit", Color: "red"}),
	).Padding(1)
}

func main() {
	// Just like React: render your app!
	_ = lotus.Run("dashboard", NewDashboard())
}
