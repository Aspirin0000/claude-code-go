package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type UnzipCommand struct{ *BaseCommand }

func NewUnzipCommand() *UnzipCommand {
	return &UnzipCommand{
		BaseCommand: NewBaseCommand(
			"unzip",
			"Extract files from ZIP archive",
			CategoryFiles,
		).WithHelp(`Usage: /unzip [options] <archive> [files]

Extract files from ZIP archives.

Common Options:
  -l        List contents
  -d dir    Extract to directory
  -o        Overwrite files
  -q        Quiet mode
  -j        Junk paths (flat extraction)

Examples:
  /unzip archive.zip                Extract to current directory
  /unzip -d /tmp archive.zip        Extract to /tmp
  /unzip -l archive.zip             List archive contents
  /unzip archive.zip "*.txt"        Extract only .txt files`),
	}
}

func (c *UnzipCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "unzip", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewUnzipCommand()) }
