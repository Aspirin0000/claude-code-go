package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// EnvCommand manages environment variables
type EnvCommand struct {
	*BaseCommand
}

// NewEnvCommand creates the /env command
func NewEnvCommand() *EnvCommand {
	return &EnvCommand{
		BaseCommand: NewBaseCommand(
			"env",
			"Show or set environment variables",
			CategoryConfig,
		).WithAliases("environment").
			WithHelp(`Usage: /env [name] [value]

Show or manage environment variables.

Without arguments, shows all environment variables.
With one argument, shows the value of that variable.
With two arguments, sets the variable to the value.

Examples:
  /env                    Show all environment variables
  /env PATH               Show the PATH variable
  /env MY_VAR hello       Set MY_VAR to "hello"

Note: Setting variables affects only the current session.

Aliases: /environment`),
	}
}

// Execute runs the env command
func (c *EnvCommand) Execute(ctx context.Context, args []string) error {
	switch len(args) {
	case 0:
		return c.showAllEnv()
	case 1:
		return c.showEnv(args[0])
	default:
		return c.setEnv(args[0], strings.Join(args[1:], " "))
	}
}

func (c *EnvCommand) showAllEnv() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Environment Variables                       ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Get common/important variables first
	important := []string{"PATH", "HOME", "USER", "SHELL", "PWD", "EDITOR", "LANG"}
	shown := make(map[string]bool)

	fmt.Println("Common variables:")
	for _, key := range important {
		if value := os.Getenv(key); value != "" {
			fmt.Printf("  %s=%s\n", key, truncate(value, 60))
			shown[key] = true
		}
	}

	fmt.Println("\nAll variables:")
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if !shown[key] {
				fmt.Printf("  %s=%s\n", key, truncate(parts[1], 60))
			}
		}
	}
	fmt.Println()
	return nil
}

func (c *EnvCommand) showEnv(name string) error {
	value := os.Getenv(name)
	if value == "" {
		fmt.Printf("%s is not set or is empty\n", name)
		return nil
	}
	fmt.Printf("%s=%s\n", name, value)
	return nil
}

func (c *EnvCommand) setEnv(name, value string) error {
	if err := os.Setenv(name, value); err != nil {
		return fmt.Errorf("failed to set %s: %w", name, err)
	}
	fmt.Printf("✅ Set %s=%s\n", name, value)
	return nil
}

// getShell returns the current shell
func getShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "/bin/sh"
	}
	return shell
}

func init() {
	Register(NewEnvCommand())
}
