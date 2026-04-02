package commands

import (
	"context"
	"os"
	"os/exec"
)

type TracerouteCommand struct{ *BaseCommand }

func NewTracerouteCommand() *TracerouteCommand {
	return &TracerouteCommand{
		BaseCommand: NewBaseCommand("traceroute", "Trace route to host", CategoryTools).
			WithAliases("tracepath").
			WithHelp("Display route and measure transit delays"),
	}
}

func (c *TracerouteCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("traceroute", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTracerouteCommand()) }
