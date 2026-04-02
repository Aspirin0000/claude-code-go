package commands

import (
	"context"
	"os"
	"os/exec"
)

type NslookupCommand struct{ *BaseCommand }

func NewNslookupCommand() *NslookupCommand {
	return &NslookupCommand{
		BaseCommand: NewBaseCommand("nslookup", "Query DNS servers", CategoryTools).
			WithHelp("Query Internet name servers"),
	}
}

func (c *NslookupCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("nslookup", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNslookupCommand()) }
