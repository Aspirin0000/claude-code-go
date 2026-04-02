package commands

import (
	"context"
	"os"
	"os/exec"
)

type SshCommand struct{ *BaseCommand }

func NewSshCommand() *SshCommand {
	return &SshCommand{
		BaseCommand: NewBaseCommand("ssh", "SSH client", CategoryTools).
			WithHelp("Secure shell remote login"),
	}
}

func (c *SshCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("ssh", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewSshCommand()) }
