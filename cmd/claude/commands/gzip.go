package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type GzipCommand struct{ *BaseCommand }

func NewGzipCommand() *GzipCommand {
	return &GzipCommand{
		BaseCommand: NewBaseCommand(
			"gzip",
			"Compress or expand files",
			CategoryFiles,
		).WithHelp(`Usage: /gzip [options] [files]

Compress or decompress files using gzip.

Common Options:
  -d        Decompress (like gunzip)
  -k        Keep input files
  -f        Force overwrite
  -r        Recurse into directories
  -t        Test integrity
  -9        Best compression
  -1        Fastest compression

Examples:
  /gzip file.txt              Compress file.txt to file.txt.gz
  /gzip -d file.txt.gz        Decompress file.txt.gz
  /gzip -k file.txt           Compress but keep original
  /gzip -r dir/               Recursively compress files`),
	}
}

func (c *GzipCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "gzip", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewGzipCommand()) }
