package commands

import (
	"context"
	"os"
	"os/exec"
)

type BcCommand struct{ *BaseCommand }

func NewBcCommand() *BcCommand {
	return &BcCommand{
		BaseCommand: NewBaseCommand("bc", "Basic calculator", CategoryTools).
			WithHelp("Command-line calculator"),
	}
}

func (c *BcCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("bc", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewBcCommand()) }
