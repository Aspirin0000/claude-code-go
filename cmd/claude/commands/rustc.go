package commands

import (
	"context"
	"os"
	"os/exec"
)

type RustcCommand struct{ *BaseCommand }

func NewRustcCommand() *RustcCommand {
	return &RustcCommand{
		BaseCommand: NewBaseCommand("rustc", "Rust compiler", CategoryTools).
			WithHelp("Compile Rust programs"),
	}
}

func (c *RustcCommand) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("rustc", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewRustcCommand()) }
