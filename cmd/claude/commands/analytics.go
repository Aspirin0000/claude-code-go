package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AnalyticsCommand manages analytics settings
type AnalyticsCommand struct {
	*BaseCommand
}

// NewAnalyticsCommand creates the /analytics command
func NewAnalyticsCommand() *AnalyticsCommand {
	return &AnalyticsCommand{
		BaseCommand: NewBaseCommand(
			"analytics",
			"View or manage analytics settings",
			CategoryConfig,
		).WithAliases("stats").
			WithHelp(`Usage: /analytics [status|enable|disable]

View or manage analytics and telemetry settings.

Subcommands:
  status   Show current analytics status
  enable   Enable analytics collection
  disable  Disable analytics collection

Examples:
  /analytics        Show analytics status
  /analytics status Same as above
  /analytics enable Enable analytics

Aliases: /stats`),
	}
}

// Execute runs the analytics command
func (c *AnalyticsCommand) Execute(ctx context.Context, args []string) error {
	action := "status"
	if len(args) > 0 {
		action = args[0]
	}

	switch action {
	case "status":
		return c.showStatus()
	case "enable":
		return c.setEnabled(true)
	case "disable":
		return c.setEnabled(false)
	default:
		return fmt.Errorf("unknown action: %s; use status, enable, or disable", action)
	}
}

func (c *AnalyticsCommand) showStatus() error {
	configDir, _ := os.UserConfigDir()
	analyticsDir := filepath.Join(configDir, "claude", "analytics")

	enabled := true
	if _, err := os.Stat(analyticsDir); os.IsNotExist(err) {
		// Check if disabled file exists
		disableFile := filepath.Join(configDir, "claude", ".analytics-disabled")
		if _, err := os.Stat(disableFile); err == nil {
			enabled = false
		}
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                   Analytics Status                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	if enabled {
		fmt.Println("📊 Analytics: ENABLED")
		fmt.Println("   Events are being collected to improve the application.")
		fmt.Printf("   Data directory: %s\n", analyticsDir)
	} else {
		fmt.Println("📊 Analytics: DISABLED")
		fmt.Println("   No events are being collected.")
	}

	fmt.Println()
	fmt.Println("Use '/analytics enable' or '/analytics disable' to change.")
	fmt.Println()
	return nil
}

func (c *AnalyticsCommand) setEnabled(enabled bool) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	disableFile := filepath.Join(configDir, "claude", ".analytics-disabled")

	if enabled {
		// Remove disable file if exists
		if err := os.Remove(disableFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to enable analytics: %w", err)
		}
		fmt.Println("✅ Analytics enabled.")
	} else {
		// Create disable file
		if err := os.MkdirAll(filepath.Dir(disableFile), 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}
		if err := os.WriteFile(disableFile, []byte("disabled at "+time.Now().Format(time.RFC3339)), 0644); err != nil {
			return fmt.Errorf("failed to disable analytics: %w", err)
		}
		fmt.Println("✅ Analytics disabled.")
	}
	fmt.Println()
	return nil
}

func init() {
	Register(NewAnalyticsCommand())
}
