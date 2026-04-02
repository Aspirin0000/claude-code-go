package commands

import (
	"context"
	"os"
	"os/exec"
)

type NetstatCommand struct{ *BaseCommand }

func NewNetstatCommand() *NetstatCommand {
	return &NetstatCommand{
		BaseCommand: NewBaseCommand("netstat", "Network statistics", CategoryTools).
			WithHelp("Display network connections and statistics"),
	}
}

func (c *NetstatCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("netstat", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNetstatCommand()) }
