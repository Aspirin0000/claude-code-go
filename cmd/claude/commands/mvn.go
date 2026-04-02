package commands

import (
	"context"
	"os"
	"os/exec"
)

type MvnCommand struct{ *BaseCommand }

func NewMvnCommand() *MvnCommand {
	return &MvnCommand{
		BaseCommand: NewBaseCommand("mvn", "Apache Maven", CategoryTools).
			WithHelp("Java project management and comprehension tool"),
	}
}

func (c *MvnCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("mvn", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewMvnCommand()) }
