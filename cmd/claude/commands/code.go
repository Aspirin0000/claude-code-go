package commands

import (
	"context"
	"os"
	"os/exec"
)

type CodeCommand struct{ *BaseCommand }

func NewCodeCommand() *CodeCommand {
	return &CodeCommand{
		BaseCommand: NewBaseCommand("code", "VS Code editor", CategoryFiles).
			WithAliases("vscode").
			WithHelp("Open file or directory in Visual Studio Code"),
	}
}

func (c *CodeCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("code", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewCodeCommand()) }
