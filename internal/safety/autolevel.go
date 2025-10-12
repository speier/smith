package safety

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed rules.yaml
var rulesYAML []byte

// AutoLevel constants
const (
	AutoLevelLow    = "low"
	AutoLevelMedium = "medium"
	AutoLevelHigh   = "high"
)

// Rules represents the complete safety ruleset
type Rules struct {
	Version          string                 `yaml:"version"`
	Levels           map[string]Level       `yaml:"levels"`
	Blocked          BlockedRules           `yaml:"blocked"`
	SessionAllowlist SessionAllowlistConfig `yaml:"sessionAllowlist"`
}

// Level represents a single auto-level configuration
type Level struct {
	Description   string   `yaml:"description"`
	AllowPatterns []string `yaml:"allowPatterns"`
	AllowTools    []string `yaml:"allowTools,omitempty"`
}

// BlockedRules contains patterns that are always blocked
type BlockedRules struct {
	Patterns []string `yaml:"patterns"`
	Features []string `yaml:"features"`
}

// SessionAllowlistConfig controls runtime allowlist behavior
type SessionAllowlistConfig struct {
	Enabled bool `yaml:"enabled"`
}

// CheckResult represents the result of a command safety check
type CheckResult struct {
	Allowed bool
	Reason  string
	Level   string
}

var LoadedRules *Rules
var sessionAllowlist = make([]string, 0)

// init loads the embedded rules at startup
func init() {
	LoadedRules = &Rules{}
	if err := yaml.Unmarshal(rulesYAML, LoadedRules); err != nil {
		panic("failed to load safety rules: " + err.Error())
	}
}

// EnsureRulesInConfigDir copies the bundled rules.yaml to the config directory
// if it doesn't already exist, allowing users to customize it
func EnsureRulesInConfigDir(configDir string) error {
	rulesPath := filepath.Join(configDir, "rules.yaml")

	// Check if rules file already exists
	if _, err := os.Stat(rulesPath); err == nil {
		// File exists, load from it instead of embedded
		data, err := os.ReadFile(rulesPath)
		if err != nil {
			return fmt.Errorf("reading custom rules: %w", err)
		}
		customRules := &Rules{}
		if err := yaml.Unmarshal(data, customRules); err != nil {
			return fmt.Errorf("parsing custom rules: %w", err)
		}
		LoadedRules = customRules
		return nil
	}

	// File doesn't exist, copy bundled rules
	if err := os.WriteFile(rulesPath, rulesYAML, 0600); err != nil {
		return fmt.Errorf("writing rules file: %w", err)
	}

	return nil
}

// IsCommandAllowed checks if a command is allowed at the given auto-level
func IsCommandAllowed(cmd string, level string) CheckResult {
	cmd = strings.TrimSpace(cmd)

	// Check blocked patterns first (always dangerous)
	for _, pattern := range LoadedRules.Blocked.Patterns {
		matched, err := regexp.MatchString(pattern, cmd)
		if err == nil && matched {
			return CheckResult{
				Allowed: false,
				Reason:  "blocked: dangerous command pattern detected",
				Level:   level,
			}
		}
	}

	// Check for command substitution
	if hasCommandSubstitution(cmd) {
		return CheckResult{
			Allowed: false,
			Reason:  "blocked: command substitution detected $(..`) or backticks",
			Level:   level,
		}
	}

	// Check for pipe to shell
	if hasPipeToShell(cmd) {
		return CheckResult{
			Allowed: false,
			Reason:  "blocked: pipe to shell detected (| sh or | bash)",
			Level:   level,
		}
	}

	// Check session allowlist
	if isInSessionAllowlist(cmd) {
		return CheckResult{
			Allowed: true,
			Reason:  "session allowlist",
			Level:   level,
		}
	}

	// Check level-specific rules
	switch level {
	case AutoLevelLow:
		if isMatchingPatterns(cmd, LoadedRules.Levels["low"].AllowPatterns) {
			return CheckResult{Allowed: true, Reason: "low-risk command", Level: level}
		}
		return CheckResult{Allowed: false, Reason: "not in low-risk allowlist", Level: level}

	case AutoLevelMedium:
		// Medium includes low patterns
		if isMatchingPatterns(cmd, LoadedRules.Levels["low"].AllowPatterns) {
			return CheckResult{Allowed: true, Reason: "low-risk command", Level: level}
		}
		if isMatchingPatterns(cmd, LoadedRules.Levels["medium"].AllowPatterns) {
			return CheckResult{Allowed: true, Reason: "medium-risk command", Level: level}
		}
		return CheckResult{Allowed: false, Reason: "not in low/medium allowlist", Level: level}

	case AutoLevelHigh:
		// High allows everything except blocked
		return CheckResult{Allowed: true, Reason: "high-risk allowed", Level: level}

	default:
		return CheckResult{Allowed: false, Reason: "unknown auto-level", Level: level}
	}
}

// IsToolAllowed checks if a tool is allowed at the given auto-level
func IsToolAllowed(tool string, level string) bool {
	switch level {
	case AutoLevelLow:
		return isInList(tool, LoadedRules.Levels["low"].AllowTools)
	case AutoLevelMedium, AutoLevelHigh:
		// Medium and high allow all tools
		return true
	default:
		return false
	}
}

// AddToSessionAllowlist adds a command to the runtime allowlist
func AddToSessionAllowlist(cmd string) {
	if !LoadedRules.SessionAllowlist.Enabled {
		return
	}
	cmd = strings.TrimSpace(cmd)
	if !isInSessionAllowlist(cmd) {
		sessionAllowlist = append(sessionAllowlist, cmd)
	}
}

// ClearSessionAllowlist clears the runtime allowlist
func ClearSessionAllowlist() {
	sessionAllowlist = make([]string, 0)
}

// GetSessionAllowlist returns the current session allowlist
func GetSessionAllowlist() []string {
	return append([]string{}, sessionAllowlist...)
}

// GetVersion returns the rules version
func GetVersion() string {
	return LoadedRules.Version
}

// GetLevelDescription returns the description for a level
func GetLevelDescription(level string) string {
	if lvl, ok := LoadedRules.Levels[level]; ok {
		return lvl.Description
	}
	return ""
}

// Helper functions

func isMatchingPatterns(cmd string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, cmd)
		if err == nil && matched {
			return true
		}
	}
	return false
}

func isInList(item string, list []string) bool {
	for _, allowed := range list {
		if item == allowed {
			return true
		}
	}
	return false
}

func isInSessionAllowlist(cmd string) bool {
	for _, allowed := range sessionAllowlist {
		if cmd == allowed {
			return true
		}
	}
	return false
}

func hasCommandSubstitution(cmd string) bool {
	// Check for $(...)
	if strings.Contains(cmd, "$(") {
		return true
	}
	// Check for backticks
	if strings.Contains(cmd, "`") {
		return true
	}
	return false
}

func hasPipeToShell(cmd string) bool {
	pipedCommands := []string{"| sh", "| bash", "|sh", "|bash"}
	for _, piped := range pipedCommands {
		if strings.Contains(cmd, piped) {
			return true
		}
	}
	return false
}

// FormatCheckResult returns a human-readable message for a check result
func FormatCheckResult(result CheckResult) string {
	if result.Allowed {
		return fmt.Sprintf("✅ Allowed (%s): %s", result.Level, result.Reason)
	}
	return fmt.Sprintf("❌ Blocked: %s", result.Reason)
}
