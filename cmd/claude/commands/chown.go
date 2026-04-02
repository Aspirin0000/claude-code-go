package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type ChownCommand struct{ *BaseCommand }

func NewChownCommand() *ChownCommand {
	return &ChownCommand{
		BaseCommand: NewBaseCommand(
			"chown",
			"Change file owner and group",
			CategoryFiles,
		).WithHelp(`Usage: /chown <owner>[:group] <file>

Change file owner and group.

Arguments:
  owner[:group]  New owner and optional group
  file           File or directory path

Examples:
  /chown user file.txt         Change owner to user
  /chown user:group file.txt   Change owner and group
  /chown :group file.txt       Change group only
  /chown -R user:group /dir    Recursively change ownership`),
	}
}

func (c *ChownCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "chown", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewChownCommand()) }
