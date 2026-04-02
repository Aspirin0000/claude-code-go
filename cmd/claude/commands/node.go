package commands

import (
	"context"
	"os"
	"os/exec"
)

type NodeCommand struct{ *BaseCommand }

func NewNodeCommand() *NodeCommand {
	return &NodeCommand{
		BaseCommand: NewBaseCommand("node", "Node.js runtime", CategoryTools).
			WithHelp("Run JavaScript with Node.js"),
	}
}

func (c *NodeCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("node", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNodeCommand()) }
