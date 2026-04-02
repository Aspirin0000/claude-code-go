package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type PipCommand struct{ *BaseCommand }

func NewPipCommand() *PipCommand {
	return &PipCommand{
		BaseCommand: NewBaseCommand(
			"pip",
			"Python Package Installer",
			CategoryTools,
		).WithHelp(`Usage: /pip <command> [args]

Python package installer.

Common Commands:
  install      Install packages
  uninstall    Uninstall packages
  list         List installed packages
  show         Show package details
  freeze       Output installed packages
  search       Search packages (deprecated)

Examples:
  /pip install requests
  /pip install -r requirements.txt
  /pip list
  /pip freeze > requirements.txt
  /pip uninstall package`),
	}
}

func (c *PipCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "pip", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewPipCommand()) }
