package commands

import (
	"context"
	"os"
	"os/exec"
)

type DigCommand struct{ *BaseCommand }

func NewDigCommand() *DigCommand {
	return &DigCommand{
		BaseCommand: NewBaseCommand("dig", "DNS lookup", CategoryTools).
			WithHelp("DNS lookup utility"),
	}
}

func (c *DigCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("dig", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewDigCommand()) }
