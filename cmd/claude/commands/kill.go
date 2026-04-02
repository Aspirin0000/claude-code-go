package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type KillCommand struct{ *BaseCommand }

func NewKillCommand() *KillCommand {
	return &KillCommand{
		BaseCommand: NewBaseCommand(
			"kill",
			"Terminate a process",
			CategoryTools,
		).WithHelp(`Usage: /kill [signal] <pid>

Send a signal to a process.

Arguments:
  signal   Signal to send (-9, -TERM, -HUP, etc.)
  pid      Process ID

Signals:
  -15, -TERM   Graceful termination (default)
  -9, -KILL    Force kill
  -1, -HUP     Hang up (restart)

Examples:
  /kill 1234               Send TERM to PID 1234
  /kill -9 1234            Force kill PID 1234
  /kill -HUP $(cat pidfile) Restart process`),
	}
}

func (c *KillCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "kill", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewKillCommand()) }
