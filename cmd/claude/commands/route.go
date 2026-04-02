package commands

import (
	"context"
	"os"
	"os/exec"
)

type RouteCommand struct{ *BaseCommand }

func NewRouteCommand() *RouteCommand {
	return &RouteCommand{
		BaseCommand: NewBaseCommand("route", "Show/manipulate routing table", CategoryTools).
			WithHelp("Display or alter the IP routing table"),
	}
}

func (c *RouteCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("route", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewRouteCommand()) }
