package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// InitCommand initializes Claude Code configuration
type InitCommand struct {
	*BaseCommand
}

// NewInitCommand creates a new init command
func NewInitCommand() *InitCommand {
	return &InitCommand{
		BaseCommand: NewBaseCommand(
			"init",
			"Initialize Claude Code configuration",
			CategoryConfig,
		).WithHelp(`Usage: /init

Initialize Claude Code by creating the configuration directory and files.
This command will:
  - Create ~/.config/claude/ directory
  - Create a sample config.json file
  - Guide you through API key setup

Example:
  /init

Note: You'll need to edit ~/.config/claude/config.json to add your API key.`),
	}
}

// Execute executes the init command
func (c *InitCommand) Execute(ctx context.Context, args []string) error {
	// Use config.GetConfigPath for consistency and testability
	configPath := config.GetConfigPath()
	claudeDir := filepath.Dir(configPath)

	// Create directory
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	fmt.Printf("✅ Created directory: %s\n", claudeDir)

	// Create sample config

	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Config file already exists: %s\n", configPath)
		fmt.Println("   Use /config to view or modify settings")
		return nil
	}

	// Create default config
	cfg := config.DefaultConfig()
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}

	fmt.Printf("✅ Created config file: %s\n", configPath)
	fmt.Println()
	fmt.Println("📝 Next steps:")
	fmt.Println("   1. Edit the config file to add your API key:")
	fmt.Printf("      nano %s\n", configPath)
	fmt.Println()
	fmt.Println("   2. Or set environment variable:")
	fmt.Println("      export ANTHROPIC_API_KEY=your-key-here")
	fmt.Println()
	fmt.Println("   3. Run 'claude' to start chatting!")

	return nil
}

func init() {
	Register(NewInitCommand())
}
