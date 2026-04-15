package commands

import (
	"context"
	"fmt"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// ReloadCommand reloads the configuration file
type ReloadCommand struct {
	*BaseCommand
}

// NewReloadCommand creates the reload command
func NewReloadCommand() *ReloadCommand {
	return &ReloadCommand{
		BaseCommand: NewBaseCommand(
			"reload",
			"Reload configuration from disk",
			CategoryConfig,
		).WithHelp(`Usage: /reload

Reload the configuration file from disk.
This is useful if you have manually edited the config file.

Examples:
  /reload        Reload configuration`),
	}
}

// Execute runs the reload command
func (c *ReloadCommand) Execute(ctx context.Context, args []string) error {
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to reload configuration: %w", err)
	}

	fmt.Println("✓ Configuration reloaded successfully.")
	fmt.Printf("  Model:    %s\n", cfg.Model)
	fmt.Printf("  Provider: %s\n", cfg.Provider)
	if cfg.APIKey != "" {
		fmt.Println("  API Key:  configured")
	} else {
		fmt.Println("  API Key:  not set")
	}

	return nil
}

func init() {
	Register(NewReloadCommand())
}
