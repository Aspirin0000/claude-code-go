package commands

import (
	"context"
	"os"
	"os/exec"
)

type ScreenCommand struct{ *BaseCommand }

func NewScreenCommand() *ScreenCommand {
	return &ScreenCommand{
		BaseCommand: NewBaseCommand("screen", "Screen manager", CategoryTools).
			WithHelp("Terminal multiplexer with session management"),
	}
}

func (c *ScreenCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("screen", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewScreenCommand()) }
