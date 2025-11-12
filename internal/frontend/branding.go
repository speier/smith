package frontend

import (
	"strings"

	"github.com/speier/smith/internal/version"
)

// ANSI color codes for Matrix-style green (terminal color 10 = bright green)
const (
	green     = "\x1b[38;5;10m"   // Bright green foreground
	greenBold = "\x1b[38;5;10;1m" // Bright green + bold
	reset     = "\x1b[0m"         // Reset all formatting
)

// GetWelcomeText returns the version and greeting text (multi-line, pre-colored)
func GetWelcomeText() string {
	var text strings.Builder
	text.WriteString(green + "v" + version.Get() + reset)
	text.WriteString("\n\n")
	text.WriteString(green + "Mr. Anderson... Welcome back." + reset)
	text.WriteString("\n")
	text.WriteString(green + "I've been expecting you." + reset)
	return text.String()
}

// GetLogoLines returns the ASCII logo (multi-line, pre-colored)
func GetLogoLines() string {
	lines := []string{
		"███████╗███╗   ███╗██╗████████╗██╗  ██╗",
		"██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║",
		"███████╗██╔████╔██║██║   ██║   ███████║",
		"╚════██║██║╚██╔╝██║██║   ██║   ██╔══██║",
		"███████║██║ ╚═╝ ██║██║   ██║   ██║  ██║",
		"╚══════╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝",
	}

	colored := make([]string, len(lines))
	for i, line := range lines {
		colored[i] = greenBold + line + reset
	}
	return strings.Join(colored, "\n")
}

// GetGoodbyeBanner returns a simple goodbye message
func GetGoodbyeBanner() string {
	// Matrix green, bold
	return greenBold + "👋 Goodbye, Mr. Anderson..." + reset
}
