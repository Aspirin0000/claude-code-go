package commands

import (
	"context"
	"os"
	"os/exec"
)

type VimCommand struct{ *BaseCommand }

func NewVimCommand() *VimCommand {
	return &VimCommand{
		BaseCommand: NewBaseCommand("vim", "Vim text editor", CategoryFiles).
			WithAliases("vi").
			WithHelp("Open file in vim editor"),
	}
}

func (c *VimCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("vim", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewVimCommand()) }
