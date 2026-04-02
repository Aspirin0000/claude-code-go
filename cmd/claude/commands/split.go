package commands

import (
	"context"
	"os"
	"os/exec"
)

type SplitCommand struct{ *BaseCommand }

func NewSplitCommand() *SplitCommand {
	return &SplitCommand{
		BaseCommand: NewBaseCommand("split", "Split file into pieces", CategoryFiles).
			WithHelp("Split a file into pieces"),
	}
}

func (c *SplitCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("split", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewSplitCommand()) }
