package frontend

import (
	"strings"

	"github.com/speier/smith/internal/version"
	"github.com/speier/smith/pkg/lotus"
)

var (
	// Matrix-style colors using terminal color 10 (bright green)
	greenStyle   = lotus.NewStyle().Foreground("10").Bold(true)
	versionStyle = lotus.NewStyle().Foreground("10")
	welcomeStyle = lotus.NewStyle().Foreground("10")
)

// GetWelcomeBanner returns the full welcome banner with ASCII logo and version
// Text is pre-colored using terminal color 10 (bright green)
func GetWelcomeBanner() string {
	logoLines := GetLogoLines()

	// Build banner parts with color applied
	var parts []string
	parts = append(parts, "") // Empty line at top

	// Add logo lines with green color and bold
	for _, line := range logoLines {
		parts = append(parts, greenStyle.Render(line))
	}

	parts = append(parts, "")                                     // Empty line
	parts = append(parts, versionStyle.Render("v"+version.Get())) // Version
	parts = append(parts, "")                                     // Empty line
	parts = append(parts, welcomeStyle.Render("Mr. Anderson... Welcome back."))
	parts = append(parts, welcomeStyle.Render("I've been expecting you."))
	parts = append(parts, "") // Empty line at bottom

	return strings.Join(parts, "\n")
}

// GetGoodbyeBanner returns a simple goodbye message (no logo)
func GetGoodbyeBanner() string {
	// Matrix green, bold (matching original)
	return greenStyle.Render("👋 Goodbye, Mr. Anderson...")
}

// GetLogoLines returns the ASCII logo lines for use in UI components
func GetLogoLines() []string {
	return []string{
		"███████╗███╗   ███╗██╗████████╗██╗  ██╗",
		"██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║",
		"███████╗██╔████╔██║██║   ██║   ███████║",
		"╚════██║██║╚██╔╝██║██║   ██║   ██╔══██║",
		"███████║██║ ╚═╝ ██║██║   ██║   ██║  ██║",
		"╚══════╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝",
	}
}
