package commands

import (
	"context"
	"os"
	"os/exec"
)

type SedCommand struct{ *BaseCommand }

func NewSedCommand() *SedCommand {
	return &SedCommand{
		BaseCommand: NewBaseCommand("sed", "Stream editor", CategoryTools).
			WithHelp("Text transformation tool"),
	}
}

func (c *SedCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("sed", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewSedCommand()) }
