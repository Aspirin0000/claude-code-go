// Package commands provides file operation commands
package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RmCommand removes files or directories
// Usage: /rm <path>
type RmCommand struct {
	*BaseCommand
}

// NewRmCommand creates the /rm command
func NewRmCommand() *RmCommand {
	return &RmCommand{
		BaseCommand: NewBaseCommand(
			"rm",
			"Remove files or directories",
			CategoryFiles,
		).WithAliases("remove", "del", "delete").
			WithHelp(`Remove files or directories.

Usage:
  /rm <file>              - Remove a file
  /rm -r <directory>     - Remove directory recursively
  /rm -f <file>          - Force remove without confirmation

Examples:
  /rm old.txt
  /rm -r temp/
  /rm -rf build/

Safety:
  - Asks for confirmation by default
  - Use -f to skip confirmation
  - Use -r for directories
  - Cannot remove root directory /`),
	}
}

// Execute removes files or directories
func (c *RmCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: /rm [options] <path>")
	}

	recursive := false
	force := false
	var paths []string

	for _, arg := range args {
		switch arg {
		case "-r", "-R", "--recursive":
			recursive = true
		case "-f", "--force":
			force = true
		case "-rf", "-fr":
			recursive = true
			force = true
		default:
			if !strings.HasPrefix(arg, "-") {
				paths = append(paths, arg)
			}
		}
	}

	if len(paths) == 0 {
		return fmt.Errorf("no path specified")
	}

	for _, path := range paths {
		// Security: prevent removing root
		absPath, _ := filepath.Abs(path)
		if absPath == "/" {
			return fmt.Errorf("cannot remove root directory /")
		}

		// Check if exists
		info, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				if !force {
					fmt.Printf("✗ File not found: %s\n", path)
				}
				continue
			}
			return fmt.Errorf("failed to stat %s: %w", path, err)
		}

		// Confirm if directory without -r
		if info.IsDir() && !recursive {
			return fmt.Errorf("%s is a directory (use -r to remove recursively)", path)
		}

		// Confirm unless forced
		if !force {
			action := "Remove"
			if info.IsDir() {
				action = "Remove directory"
			}
			fmt.Printf("%s %s? [y/N]: ", action, path)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled")
				continue
			}
		}

		// Remove
		if info.IsDir() {
			err = os.RemoveAll(path)
		} else {
			err = os.Remove(path)
		}

		if err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}

		fmt.Printf("✓ Removed: %s\n", path)
	}

	return nil
}

func init() {
	Register(NewRmCommand())
}
