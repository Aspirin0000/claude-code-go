package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type TarCommand struct{ *BaseCommand }

func NewTarCommand() *TarCommand {
	return &TarCommand{
		BaseCommand: NewBaseCommand(
			"tar",
			"Archive files",
			CategoryFiles,
		).WithHelp(`Usage: /tar [options] <archive> [files]

Create and manipulate tar archives.

Common Options:
  -c        Create archive
  -x        Extract from archive
  -t        List contents
  -f file   Archive filename
  -v        Verbose
  -z        Compress with gzip
  -j        Compress with bzip2

Examples:
  /tar -czf backup.tar.gz dir/     Create compressed archive
  /tar -xzf backup.tar.gz          Extract archive
  /tar -tvf backup.tar.gz          List archive contents
  /tar -czvf backup.tar.gz *.txt   Archive all .txt files`),
	}
}

func (c *TarCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "tar", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewTarCommand()) }
