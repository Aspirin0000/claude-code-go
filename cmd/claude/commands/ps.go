package commands

import (
	"context"
	"os"
	"os/exec"
)

type PsCommand struct{ *BaseCommand }

func NewPsCommand() *PsCommand {
	return &PsCommand{
		BaseCommand: NewBaseCommand(
			"ps",
			"Report a snapshot of current processes",
			CategoryTools,
		).WithHelp(`Usage: /ps [options]

Display information about active processes.

Common Options:
  aux      Show all processes for all users
  -ef      Full format listing
  -e       All processes
  -f       Full format
  -u user  Show processes for user

Examples:
  /ps aux                Show all processes
  /ps -ef | grep nginx   Find nginx processes
  /ps -u root            Show root's processes`),
	}
}

func (c *PsCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "ps", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewPsCommand()) }
