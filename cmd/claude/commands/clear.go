package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ClearCommand clears the terminal screen
type ClearCommand struct {
	*BaseCommand
}

// NewClearCommand creates the clear command
func NewClearCommand() *ClearCommand {
	return &ClearCommand{
		BaseCommand: NewBaseCommand(
			"clear",
			"Clear the terminal screen",
			CategoryGeneral,
		).WithAliases("cls", "clr").
			WithHelp(`Usage: /clear

Clear the terminal screen.

Aliases: /cls, /clr`),
	}
}

// Execute clears the screen
func (c *ClearCommand) Execute(ctx context.Context, args []string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Fallback to ANSI escape codes if system command fails
		fmt.Print("\033[2J\033[H")
		return nil
	}

	return nil
}

func init() { Register(NewClearCommand()) }
