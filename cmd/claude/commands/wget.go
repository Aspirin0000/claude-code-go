package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type WgetCommand struct{ *BaseCommand }

func NewWgetCommand() *WgetCommand {
	return &WgetCommand{
		BaseCommand: NewBaseCommand(
			"wget",
			"Network downloader",
			CategoryTools,
		).WithHelp(`Usage: /wget [options] <url>

Download files from the web.

Common Options:
  -O file     Save to specific filename
  -P dir      Save to directory
  -c          Continue partial download
  -q          Quiet mode
  --limit-rate=RATE   Limit download speed

Examples:
  /wget https://example.com/file.zip
  /wget -O output.txt https://example.com/data
  /wget -c https://example.com/large-file.iso
  /wget -P /tmp https://example.com/file.txt`),
	}
}

func (c *WgetCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "wget", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewWgetCommand()) }
