package commands

import (
	"context"
	"os"
	"os/exec"
)

type PasteCommand struct{ *BaseCommand }

func NewPasteCommand() *PasteCommand {
	return &PasteCommand{
		BaseCommand: NewBaseCommand("paste", "Merge lines of files", CategoryTools).
			WithHelp("Merge corresponding lines of files"),
	}
}

func (c *PasteCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("paste", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewPasteCommand()) }
