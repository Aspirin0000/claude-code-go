package commands

import (
	"context"
	"os"
	"os/exec"
)

type IfconfigCommand struct{ *BaseCommand }

func NewIfconfigCommand() *IfconfigCommand {
	return &IfconfigCommand{
		BaseCommand: NewBaseCommand("ifconfig", "Network interface configuration", CategoryTools).
			WithHelp("Display or configure network interfaces"),
	}
}

func (c *IfconfigCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("ifconfig", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewIfconfigCommand()) }
