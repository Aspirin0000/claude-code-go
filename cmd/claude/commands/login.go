package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// LoginCommand authenticates the user with an Anthropic API key
// Source: src/commands/login/
type LoginCommand struct {
	*BaseCommand
}

// NewLoginCommand creates the /login command
func NewLoginCommand() *LoginCommand {
	return &LoginCommand{
		BaseCommand: NewBaseCommand(
			"login",
			"Sign in with your Anthropic API key",
			CategoryConfig,
		).WithHelp(`Usage: /login [api-key]

Authenticate with your Anthropic API key.
If no API key is provided, you will be prompted to enter one securely.

Examples:
  /login                      Prompt for API key interactively
  /login sk-ant-api03-...     Provide API key directly

Note: Your API key will be saved to the configuration file.`),
	}
}

// Execute runs the login command
func (c *LoginCommand) Execute(ctx context.Context, args []string) error {
	var apiKey string

	if len(args) > 0 {
		apiKey = strings.TrimSpace(args[0])
	} else {
		fmt.Print("Enter your Anthropic API key: ")
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(input)
	}

	if apiKey == "" {
		return fmt.Errorf("API key cannot be empty")
	}

	// Load current config
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update API key
	cfg.APIKey = apiKey

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ API key saved successfully.")
	fmt.Printf("   Key prefix: %s...\n", apiKey[:min(8, len(apiKey))])
	fmt.Println("   You can now start chatting with Claude.")

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// LogoutCommand signs out the user by removing the API key
// Source: src/commands/logout/
type LogoutCommand struct {
	*BaseCommand
}

// NewLogoutCommand creates the /logout command
func NewLogoutCommand() *LogoutCommand {
	return &LogoutCommand{
		BaseCommand: NewBaseCommand(
			"logout",
			"Sign out and remove the saved API key",
			CategoryConfig,
		).WithHelp(`Usage: /logout

Remove the saved Anthropic API key from the configuration file.

Examples:
  /logout

Note: This does not invalidate your API key on Anthropic's servers.
It only removes it from the local configuration file.`),
	}
}

// Execute runs the logout command
func (c *LogoutCommand) Execute(ctx context.Context, args []string) error {
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if cfg.APIKey == "" {
		fmt.Println("ℹ️  No API key is currently saved.")
		return nil
	}

	// Clear API key
	cfg.APIKey = ""

	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println("✅ Signed out successfully.")
	fmt.Println("   Your API key has been removed from the local configuration.")

	return nil
}

func init() {
	Register(NewLoginCommand())
	Register(NewLogoutCommand())
}
