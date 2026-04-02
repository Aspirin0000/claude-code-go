package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type ZipCommand struct{ *BaseCommand }

func NewZipCommand() *ZipCommand {
	return &ZipCommand{
		BaseCommand: NewBaseCommand(
			"zip",
			"Package and compress files",
			CategoryFiles,
		).WithHelp(`Usage: /zip [options] <archive> [files]

Package and compress files into a ZIP archive.

Common Options:
  -r        Recurse into directories
  -q        Quiet mode
  -9        Maximum compression
  -u        Update files in archive
  -d        Delete files from archive
  -e        Encrypt archive

Examples:
  /zip backup.zip file.txt          Zip a single file
  /zip -r backup.zip dir/           Zip directory recursively
  /zip -9 backup.zip *.txt          Max compression
  /zip -u backup.zip newfile.txt    Update archive`),
	}
}

func (c *ZipCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "zip", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewZipCommand()) }
