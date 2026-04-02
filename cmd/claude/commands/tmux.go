package commands

import (
	"context"
	"os"
	"os/exec"
)

type TmuxCommand struct{ *BaseCommand }

func NewTmuxCommand() *TmuxCommand {
	return &TmuxCommand{
		BaseCommand: NewBaseCommand("tmux", "Terminal multiplexer", CategoryTools).
			WithHelp("Terminal session manager"),
	}
}

func (c *TmuxCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("tmux", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTmuxCommand()) }
