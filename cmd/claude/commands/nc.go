package commands

import (
	"context"
	"os"
	"os/exec"
)

type NcCommand struct{ *BaseCommand }

func NewNcCommand() *NcCommand {
	return &NcCommand{
		BaseCommand: NewBaseCommand("nc", "Netcat networking utility", CategoryTools).
			WithAliases("netcat").
			WithHelp("TCP/UDP networking tool"),
	}
}

func (c *NcCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("nc", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNcCommand()) }
