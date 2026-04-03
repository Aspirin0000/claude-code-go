package commands

import (
	"context"
	"fmt"
	"os"
)

// ExitCommand exits the application
type ExitCommand struct {
	*BaseCommand
}

// NewExitCommand creates a new exit command
func NewExitCommand() *ExitCommand {
	return &ExitCommand{
		BaseCommand: NewBaseCommand(
			"exit",
			"Exit the application",
			CategoryGeneral,
		).WithAliases("quit", "q", "bye").WithHelp(`Exit the application.

Usage: /exit
   or: /quit
   or: /q
   or: /bye

Aliases:
  /quit, /q, /bye - Same as /exit

Exits the application gracefully. Any unsaved session state will be preserved.`),
	}
}

// Execute runs the exit command
func (c *ExitCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println("Goodbye!")
	os.Exit(0)
	return nil // Never reached
}

func init() { Register(NewExitCommand()) }
