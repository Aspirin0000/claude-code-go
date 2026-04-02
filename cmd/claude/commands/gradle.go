package commands

import (
	"context"
	"os"
	"os/exec"
)

type GradleCommand struct{ *BaseCommand }

func NewGradleCommand() *GradleCommand {
	return &GradleCommand{
		BaseCommand: NewBaseCommand("gradle", "Gradle build tool", CategoryTools).
			WithHelp("Build automation system"),
	}
}

func (c *GradleCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("gradle", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewGradleCommand()) }
