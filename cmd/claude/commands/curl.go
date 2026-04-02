package commands

import (
	"context"
	"os"
	"os/exec"
)

type CurlCommand struct{ *BaseCommand }

func NewCurlCommand() *CurlCommand {
	return &CurlCommand{
		BaseCommand: NewBaseCommand("curl", "Transfer data from/to server", CategoryTools).
			WithHelp("Make HTTP requests"),
	}
}

func (c *CurlCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("curl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewCurlCommand()) }
