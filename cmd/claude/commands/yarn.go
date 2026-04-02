package commands

import (
	"context"
	"os"
	"os/exec"
)

type YarnCommand struct{ *BaseCommand }

func NewYarnCommand() *YarnCommand {
	return &YarnCommand{
		BaseCommand: NewBaseCommand("yarn", "Yarn package manager", CategoryTools).
			WithHelp("Fast, reliable, and secure dependency management"),
	}
}

func (c *YarnCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("yarn", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewYarnCommand()) }
