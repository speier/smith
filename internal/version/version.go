package version

import (
	"fmt"
	"runtime/debug"
)

// Version is the current version of Smith
// Can be overridden at build time with:
// go build -ldflags "-X github.com/speier/smith/internal/version.Version=1.0.0"
var Version = "dev"

// Get returns the current version with commit hash in format: "version (commit)"
func Get() string {
	version := Version
	var revision string

	// Try to get VCS info from build metadata
	if info, ok := debug.ReadBuildInfo(); ok {
		var modified string
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					revision = setting.Value[:7]
				}
			case "vcs.modified":
				if setting.Value == "true" {
					modified = "-dirty"
				}
			}
		}
		if revision != "" {
			return fmt.Sprintf("%s (%s%s)", version, revision, modified)
		}
	}

	return version
}
