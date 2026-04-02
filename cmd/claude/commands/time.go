package commands

import (
	"context"
	"os"
	"os/exec"
)

type TimeCommand struct{ *BaseCommand }

func NewTimeCommand() *TimeCommand {
	return &TimeCommand{
		BaseCommand: NewBaseCommand("time", "Time command execution", CategoryTools).
			WithHelp("Measure execution time of a command"),
	}
}

func (c *TimeCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return nil
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func init() { Register(NewTimeCommand()) }
