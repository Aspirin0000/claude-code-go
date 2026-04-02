package commands

import (
	"context"
	"os"
	"os/exec"
)

type WatchCommand struct{ *BaseCommand }

func NewWatchCommand() *WatchCommand {
	return &WatchCommand{
		BaseCommand: NewBaseCommand("watch", "Execute command periodically", CategoryTools).
			WithHelp("Execute a command repeatedly, displaying output"),
	}
}

func (c *WatchCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("watch", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewWatchCommand()) }
