package commands

import (
	"context"
	"fmt"
	"os"
)

type HostnameCommand struct{ *BaseCommand }

func NewHostnameCommand() *HostnameCommand {
	return &HostnameCommand{
		BaseCommand: NewBaseCommand("hostname", "Show system hostname", CategoryGeneral).
			WithHelp("Display the system hostname"),
	}
}

func (c *HostnameCommand) Execute(ctx context.Context, args []string) error {
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("cannot get hostname: %w", err)
	}
	fmt.Println(hostname)
	return nil
}

func init() { Register(NewHostnameCommand()) }
