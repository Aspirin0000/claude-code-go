package commands

import (
	"context"
	"os"
	"os/exec"
)

type JoinCommand struct{ *BaseCommand }

func NewJoinCommand() *JoinCommand {
	return &JoinCommand{
		BaseCommand: NewBaseCommand("join", "Join lines on common field", CategoryTools).
			WithHelp("Join lines of two files on a common field"),
	}
}

func (c *JoinCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("join", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewJoinCommand()) }
