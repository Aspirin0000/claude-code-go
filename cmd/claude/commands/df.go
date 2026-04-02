package commands

import (
	"context"
	"os"
	"os/exec"
)

type DfCommand struct{ *BaseCommand }

func NewDfCommand() *DfCommand {
	return &DfCommand{
		BaseCommand: NewBaseCommand(
			"df",
			"Report file system disk space usage",
			CategoryTools,
		).WithHelp(`Usage: /df [options] [filesystem]

Display disk space usage for file systems.

Common Options:
  -h       Human-readable sizes (KB, MB, GB)
  -H       Human-readable sizes (1000-based)
  -T       Show filesystem type
  -i       Show inode information

Examples:
  /df -h                  Show all filesystems human-readable
  /df -h .                Show current directory's filesystem
  /df -T /home            Show filesystem type for /home`),
	}
}

func (c *DfCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "df", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewDfCommand()) }
