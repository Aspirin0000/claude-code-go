package commands

import (
	"context"
	"os"
	"os/exec"
)

type CmakeCommand struct{ *BaseCommand }

func NewCmakeCommand() *CmakeCommand {
	return &CmakeCommand{
		BaseCommand: NewBaseCommand("cmake", "Cross-platform build system", CategoryTools).
			WithHelp("Build system generator"),
	}
}

func (c *CmakeCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("cmake", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewCmakeCommand()) }
