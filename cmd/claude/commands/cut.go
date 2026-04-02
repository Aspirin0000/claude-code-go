package commands

import (
	"context"
	"os"
	"os/exec"
)

type CutCommand struct{ *BaseCommand }

func NewCutCommand() *CutCommand {
	return &CutCommand{
		BaseCommand: NewBaseCommand("cut", "Remove sections from lines", CategoryTools).
			WithHelp("Extract columns from text"),
	}
}

func (c *CutCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("cut", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewCutCommand()) }
