package commands

import (
	"context"
	"os"
	"os/exec"
)

type JqCommand struct{ *BaseCommand }

func NewJqCommand() *JqCommand {
	return &JqCommand{
		BaseCommand: NewBaseCommand("jq", "JSON processor", CategoryTools).
			WithHelp("Command-line JSON processor"),
	}
}

func (c *JqCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("jq", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewJqCommand()) }
