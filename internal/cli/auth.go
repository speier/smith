package cli

import (
	"fmt"
	"time"

	"github.com/speier/smith/internal/config"
	"github.com/speier/smith/internal/llm"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with GitHub Copilot",
	Long: `Authenticate with GitHub Copilot using device flow.

This will:
1. Display a device code and verification URL
2. Wait for you to authorize in your browser
3. Save the authentication token securely

Example:
  smith auth login
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("🔐 GitHub Copilot Authentication")
		fmt.Println()

		provider := llm.NewCopilotProvider()

		// Start device flow
		fmt.Println("Starting device flow authorization...")
		deviceCode, err := provider.Authorize()
		if err != nil {
			return fmt.Errorf("failed to start authorization: %w", err)
		}

		// Display instructions
		fmt.Println()
		fmt.Println("┌──────────────────────────────────────────────────┐")
		fmt.Printf("│  Please visit: %s  │\n", deviceCode.VerificationURI)
		fmt.Printf("│  Enter code:   %-31s│\n", deviceCode.UserCode)
		fmt.Println("└──────────────────────────────────────────────────┘")
		fmt.Println()
		fmt.Println("⏳ Waiting for authorization...")

		// Poll for token
		interval := time.Duration(deviceCode.Interval) * time.Second
		timeout := time.Duration(deviceCode.ExpiresIn) * time.Second
		deadline := time.Now().Add(timeout)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			if time.Now().After(deadline) {
				return fmt.Errorf("authorization timed out")
			}

			<-ticker.C

			token, err := provider.PollForToken(deviceCode.DeviceCode)
			if err != nil {
				return fmt.Errorf("polling for token: %w", err)
			}

			if token == "pending" {
				continue // Still waiting
			}

			// Success! Save the token
			if err := provider.SetAuth(token); err != nil {
				return fmt.Errorf("saving authentication: %w", err)
			}

			fmt.Println()
			fmt.Println("✅ Successfully authenticated with GitHub Copilot!")
			fmt.Println()

			// Update config to use copilot
			cfg, _ := config.Load()
			cfg.Provider = "copilot"
			cfg.Model = "gpt-4o"
			if err := cfg.Save(); err != nil {
				fmt.Printf("⚠️  Warning: failed to update config: %v\n", err)
			} else {
				fmt.Println("📝 Config updated to use GitHub Copilot")
			}

			return nil
		}
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		fmt.Printf("Current provider: %s\n", cfg.Provider)
		fmt.Printf("Current model: %s\n", cfg.Model)
		fmt.Println()

		auth, err := config.LoadAuth(cfg.Provider)
		if err != nil {
			return fmt.Errorf("loading auth: %w", err)
		}

		if auth == nil {
			fmt.Println("Status: ❌ Not authenticated")
			fmt.Println()
			fmt.Println("Run 'smith auth login' to authenticate")
			return nil
		}

		fmt.Printf("Status: ✅ Authenticated as %s\n", auth.Provider)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear authentication",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.ClearAuth(); err != nil {
			return fmt.Errorf("clearing auth: %w", err)
		}
		fmt.Println("✅ Logged out successfully")
		return nil
	},
}

// Note: Moved to REPL /auth command
// Auth functionality available via /auth inside interactive mode
