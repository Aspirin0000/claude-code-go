package commands

import (
	"context"
	"os"
	"os/exec"
)

type TelnetCommand struct{ *BaseCommand }

func NewTelnetCommand() *TelnetCommand {
	return &TelnetCommand{
		BaseCommand: NewBaseCommand("telnet", "Telnet client", CategoryTools).
			WithHelp("User interface to the TELNET protocol"),
	}
}

func (c *TelnetCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("telnet", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTelnetCommand()) }
