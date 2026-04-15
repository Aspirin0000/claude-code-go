package commands

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// ThemeCommand manages the TUI color theme
type ThemeCommand struct {
	*BaseCommand
}

// NewThemeCommand creates the /theme command
func NewThemeCommand() *ThemeCommand {
	return &ThemeCommand{
		BaseCommand: NewBaseCommand(
			"theme",
			"Switch between light and dark TUI themes",
			CategoryConfig,
		).WithAliases("themes").
			WithHelp(`Usage: /theme [light|dark]

Switch the TUI color theme between light and dark modes.
Without an argument, shows the current theme.

Available themes:
  dark  - Dark color palette (default)
  light - Light color palette

Aliases: /themes`),
	}
}

// Execute runs the theme command
func (c *ThemeCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showCurrentTheme()
	}

	theme := strings.ToLower(args[0])
	if theme != "light" && theme != "dark" {
		return fmt.Errorf("unknown theme %q; available themes: light, dark", theme)
	}

	return c.switchTheme(theme)
}

func (c *ThemeCommand) showCurrentTheme() error {
	current := c.getCurrentTheme()
	fmt.Printf("Current theme: %s\n", current)
	fmt.Println("Use /theme <light|dark> to switch themes")
	return nil
}

func (c *ThemeCommand) switchTheme(theme string) error {
	current := c.getCurrentTheme()
	if current == theme {
		fmt.Printf("Already using %s theme.\n", theme)
		return nil
	}

	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	cfg.Theme = theme
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Switched to %s theme. Restart TUI to apply changes.\n", theme)
	return nil
}

func (c *ThemeCommand) getCurrentTheme() string {
	if envTheme := os.Getenv("CLAUDE_THEME"); envTheme != "" {
		return strings.ToLower(envTheme)
	}

	configPath := config.GetConfigPath()
	if cfg, err := config.Load(configPath); err == nil && cfg.Theme != "" {
		return cfg.Theme
	}

	return config.DefaultConfig().Theme
}

func init() {
	Register(NewThemeCommand())
}
