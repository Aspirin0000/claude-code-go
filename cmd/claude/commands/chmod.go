package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type ChmodCommand struct{ *BaseCommand }

func NewChmodCommand() *ChmodCommand {
	return &ChmodCommand{
		BaseCommand: NewBaseCommand(
			"chmod",
			"Change file permissions",
			CategoryFiles,
		).WithHelp(`Usage: /chmod <permissions> <file>

Change file mode bits (permissions).

Arguments:
  permissions   Permission mode (e.g., 755, u+x, go-w)
  file          File or directory path

Examples:
  /chmod 755 script.sh       Set permissions to rwxr-xr-x
  /chmod u+x file            Add execute permission for user
  /chmod go-w file.txt       Remove write permission for group and others
  /chmod -R 644 *.txt        Recursively change permissions`),
	}
}

func (c *ChmodCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		fmt.Println(c.Help())
		return nil
	}

	cmd := exec.CommandContext(ctx, "chmod", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(NewChmodCommand()) }
