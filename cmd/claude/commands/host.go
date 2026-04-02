package commands

import (
	"context"
	"os"
	"os/exec"
)

type HostCommand struct{ *BaseCommand }

func NewHostCommand() *HostCommand {
	return &HostCommand{
		BaseCommand: NewBaseCommand("host", "DNS lookup utility", CategoryTools).
			WithHelp("Simple DNS lookup"),
	}
}

func (c *HostCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("host", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewHostCommand()) }
