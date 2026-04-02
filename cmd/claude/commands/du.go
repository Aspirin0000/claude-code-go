package commands

import (
	"context"
	"os"
	"os/exec"
)

type DuCommand struct{ *BaseCommand }

func NewDuCommand() *DuCommand {
	return &DuCommand{
		BaseCommand: NewBaseCommand(
			"du",
			"Estimate file space usage",
			CategoryTools,
		).WithHelp(`Usage: /du [options] [path]

Estimate file space usage.

Common Options:
  -h       Human-readable sizes
  -s       Summarize (total only)
  -a       Show all files, not just directories
  -c       Show grand total
  --max-depth=N   Limit directory depth

Examples:
  /du -sh *               Show sizes of all items in current dir
  /du -sh /var/log        Show total size of /var/log
  /du -ah --max-depth=1   Show all files, 1 level deep`),
	}
}

func (c *DuCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "du", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewDuCommand()) }
