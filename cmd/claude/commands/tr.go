package commands

import (
	"context"
	"os"
	"os/exec"
)

type TrCommand struct{ *BaseCommand }

func NewTrCommand() *TrCommand {
	return &TrCommand{
		BaseCommand: NewBaseCommand("tr", "Translate or delete characters", CategoryTools).
			WithHelp("Character translation tool"),
	}
}

func (c *TrCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("tr", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTrCommand()) }
