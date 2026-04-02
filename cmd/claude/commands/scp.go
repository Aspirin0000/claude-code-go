package commands

import (
	"context"
	"os"
	"os/exec"
)

type ScpCommand struct{ *BaseCommand }

func NewScpCommand() *ScpCommand {
	return &ScpCommand{
		BaseCommand: NewBaseCommand("scp", "Secure copy", CategoryTools).
			WithHelp("Securely copy files between hosts"),
	}
}

func (c *ScpCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("scp", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewScpCommand()) }
