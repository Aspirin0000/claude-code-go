package commands

import (
	"context"
	"os"
	"os/exec"
)

type XargsCommand struct{ *BaseCommand }

func NewXargsCommand() *XargsCommand {
	return &XargsCommand{
		BaseCommand: NewBaseCommand("xargs", "Build and execute commands", CategoryTools).
			WithHelp("Build and execute command lines from stdin"),
	}
}

func (c *XargsCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("xargs", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewXargsCommand()) }
