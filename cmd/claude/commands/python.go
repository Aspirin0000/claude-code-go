package commands

import (
	"context"
	"os"
	"os/exec"
)

type PythonCommand struct{ *BaseCommand }

func NewPythonCommand() *PythonCommand {
	return &PythonCommand{
		BaseCommand: NewBaseCommand("python", "Python interpreter", CategoryTools).
			WithAliases("python3", "py").
			WithHelp("Run Python interpreter"),
	}
}

func (c *PythonCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("python3", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewPythonCommand()) }
