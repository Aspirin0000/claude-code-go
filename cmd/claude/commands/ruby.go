package commands

import (
	"context"
	"os"
	"os/exec"
)

type RubyCommand struct{ *BaseCommand }

func NewRubyCommand() *RubyCommand {
	return &RubyCommand{
		BaseCommand: NewBaseCommand("ruby", "Ruby interpreter", CategoryTools).
			WithHelp("Run Ruby interpreter"),
	}
}

func (c *RubyCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("ruby", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewRubyCommand()) }
