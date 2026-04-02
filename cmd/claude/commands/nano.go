package commands

import (
	"context"
	"os"
	"os/exec"
)

type NanoCommand struct{ *BaseCommand }

func NewNanoCommand() *NanoCommand {
	return &NanoCommand{
		BaseCommand: NewBaseCommand("nano", "Nano text editor", CategoryFiles).
			WithAliases("pico").
			WithHelp("Open file in nano editor"),
	}
}

func (c *NanoCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("nano", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNanoCommand()) }
