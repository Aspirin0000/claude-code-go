package commands

import (
	"context"
	"os"
	"os/exec"
)

type IpCommand struct{ *BaseCommand }

func NewIpCommand() *IpCommand {
	return &IpCommand{
		BaseCommand: NewBaseCommand("ip", "Network configuration", CategoryTools).
			WithHelp("Show/manipulate routing, devices, policy routing and tunnels"),
	}
}

func (c *IpCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("ip", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewIpCommand()) }
