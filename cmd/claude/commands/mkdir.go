// Package commands provides file operation commands
package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// MkdirCommand creates directories
// Usage: /mkdir <directory>
type MkdirCommand struct {
	*BaseCommand
}

// NewMkdirCommand creates the /mkdir command
func NewMkdirCommand() *MkdirCommand {
	return &MkdirCommand{
		BaseCommand: NewBaseCommand(
			"mkdir",
			"Create directories",
			CategoryFiles,
		).WithAliases("md", "mkd").
			WithHelp(`Create directories.

Usage:
  /mkdir <directory>         - Create a directory
  /mkdir -p <path/to/dir>  - Create parent directories as needed

Examples:
  /mkdir new_folder
  /mkdir -p projects/go/app

Options:
  -p, --parents    Create parent directories as needed`),
	}
}

// Execute creates directories
func (c *MkdirCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /mkdir [options] <directory>")
	}

	parents := false
	var dirs []string

	for _, arg := range args {
		switch arg {
		case "-p", "--parents":
			parents = true
		default:
			if !strings.HasPrefix(arg, "-") {
				dirs = append(dirs, arg)
			}
		}
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directory specified")
	}

	for _, dir := range dirs {
		var err error
		if parents {
			err = os.MkdirAll(dir, 0755)
		} else {
			err = os.Mkdir(dir, 0755)
		}

		if err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		fmt.Printf("✓ Created directory: %s\n", dir)
	}

	return nil
}

func init() {
	Register(NewMkdirCommand())
}
