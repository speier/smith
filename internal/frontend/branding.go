package frontend

import (
	"strings"

	"github.com/speier/smith/internal/version"
)

const (
	// ANSI color codes
	green  = "\033[92m" // Bright green
	blue   = "\033[94m" // Bright blue
	gray   = "\033[90m" // Dark gray
	reset  = "\033[0m"
	bold   = "\033[1m"
	italic = "\033[3m"
)

// GetWelcomeBanner returns the ASCII logo with version and welcome message
func GetWelcomeBanner() string {
	logo := green + bold + `
███████╗███╗   ███╗██╗████████╗██╗  ██╗
██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║
███████╗██╔████╔██║██║   ██║   ███████║
╚════██║██║╚██╔╝██║██║   ██║   ██╔══██║
███████║██║ ╚═╝ ██║██║   ██║   ██║  ██║
╚══════╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝` + reset

	versionText := center(blue+"v"+version.Get()+reset, 46)
	welcome := center(gray+italic+"Mr. Anderson... Welcome back.\nI've been expecting you."+reset, 46)

	return strings.Join([]string{
		"",
		logo,
		"",
		versionText,
		"",
		welcome,
		"",
	}, "\n")
}

// GetWelcomePlain returns a plain text version (no styling) for simple terminals
func GetWelcomePlain() string {
	return `
███████╗███╗   ███╗██╗████████╗██╗  ██╗
██╔════╝████╗ ████║██║╚══██╔══╝██║  ██║
███████╗██╔████╔██║██║   ██║   ███████║
╚════██║██║╚██╔╝██║██║   ██║   ██╔══██║
███████║██║ ╚═╝ ██║██║   ██║   ██║  ██║
╚══════╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝  ╚═╝

v` + version.Get() + `

Mr. Anderson... Welcome back.
I've been expecting you.
`
}

// center centers text within a given width
func center(text string, width int) string {
	lines := strings.Split(text, "\n")
	var centered []string
	for _, line := range lines {
		// Strip ANSI codes to calculate visible length
		visible := stripANSI(line)
		padding := (width - len(visible)) / 2
		if padding < 0 {
			padding = 0
		}
		centered = append(centered, strings.Repeat(" ", padding)+line)
	}
	return strings.Join(centered, "\n")
}

// stripANSI removes ANSI escape codes for length calculation
func stripANSI(s string) string {
	var result []rune
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result = append(result, r)
	}
	return string(result)
}
