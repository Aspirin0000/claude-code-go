package commands

import (
	"context"
	"os"
	"os/exec"
)

type AwkCommand struct{ *BaseCommand }

func NewAwkCommand() *AwkCommand {
	return &AwkCommand{
		BaseCommand: NewBaseCommand("awk", "Pattern scanning and processing", CategoryTools).
			WithHelp("Text processing tool"),
	}
}

func (c *AwkCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("awk", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewAwkCommand()) }
