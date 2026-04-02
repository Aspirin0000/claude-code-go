package commands

import (
	"context"
	"os"
	"os/exec"
)

type TreeCommand struct{ *BaseCommand }

func NewTreeCommand() *TreeCommand {
	return &TreeCommand{
		BaseCommand: NewBaseCommand("tree", "Display directory tree", CategoryFiles).
			WithHelp("Show directory structure in tree format"),
	}
}

func (c *TreeCommand) Execute(ctx context.Context, args []string) error {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}

	// Try to use system tree command
	cmd := exec.Command("tree", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTreeCommand()) }
