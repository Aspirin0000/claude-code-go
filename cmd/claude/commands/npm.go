package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type NpmCommand struct{ *BaseCommand }

func NewNpmCommand() *NpmCommand {
	return &NpmCommand{
		BaseCommand: NewBaseCommand(
			"npm",
			"Node Package Manager",
			CategoryTools,
		).WithHelp(`Usage: /npm <command> [args]

Node.js package manager.

Common Commands:
  install      Install packages
  ci           Clean install
  run <script> Run package script
  test         Run tests
  build        Build project
  start        Start application
  publish      Publish package
  update       Update packages

Examples:
  /npm install
  /npm install express
  /npm run build
  /npm test
  /npm update`),
	}
}

func (c *NpmCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewNpmCommand()) }
