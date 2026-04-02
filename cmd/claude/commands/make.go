package commands

import (
	"context"
	"os"
	"os/exec"
)

type MakeCommand struct{ *BaseCommand }

func NewMakeCommand() *MakeCommand {
	return &MakeCommand{
		BaseCommand: NewBaseCommand("make", "GNU Make utility", CategoryTools).
			WithHelp("Build automation tool"),
	}
}

func (c *MakeCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("make", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewMakeCommand()) }
