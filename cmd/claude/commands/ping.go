package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type PingCommand struct{ *BaseCommand }

func NewPingCommand() *PingCommand {
	return &PingCommand{
		BaseCommand: NewBaseCommand(
			"ping",
			"Test network connectivity",
			CategoryTools,
		).WithHelp(`Usage: /ping [options] <host>

Test network connectivity to a host.

Common Options:
  -c count    Stop after count packets
  -i interval Seconds between packets
  -t          Ping until interrupted (Windows)
  -W timeout  Time to wait for response

Examples:
  /ping google.com          Ping google.com
  /ping -c 4 8.8.8.8        Send 4 packets to Google DNS
  /ping -c 10 -i 2 host     Send 10 packets every 2 seconds`),
	}
}

func (c *PingCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "ping", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewPingCommand()) }
