package commands

import (
	"context"
	"os"
	"os/exec"
)

type WhichCommand struct{ *BaseCommand }

func NewWhichCommand() *WhichCommand {
	return &WhichCommand{
		BaseCommand: NewBaseCommand("which", "Locate a command", CategoryTools).
			WithHelp("Find the path to a command"),
	}
}

func (c *WhichCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return nil
	}
	cmd := exec.Command("which", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewWhichCommand()) }
