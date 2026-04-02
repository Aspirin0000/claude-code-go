package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CdCommand changes the current working directory
// Usage: /cd [directory]
type CdCommand struct {
	*BaseCommand
}

// NewCdCommand creates the /cd command
func NewCdCommand() *CdCommand {
	return &CdCommand{
		BaseCommand: NewBaseCommand(
			"cd",
			"Change current working directory",
			CategoryFiles,
		).WithAliases("chdir").
			WithHelp(`Change the current working directory.

Usage:
  /cd [directory]    Change to directory (default: home directory)

Arguments:
  [directory]    Target directory path (supports ~ for home)

Examples:
  /cd              Change to home directory
  /cd /tmp         Change to /tmp
  /cd ~/projects   Change to ~/projects
  /cd ..           Change to parent directory
  /cd -            Change to previous directory`),
	}
}

// Execute changes the current working directory
func (c *CdCommand) Execute(ctx context.Context, args []string) error {
	if err := c.checkPermissions(); err != nil {
		return err
	}

	var target string

	if len(args) == 0 {
		// Change to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home directory: %w", err)
		}
		target = home
	} else {
		target = args[0]
	}

	// Expand ~ to home directory
	if strings.HasPrefix(target, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot get home directory: %w", err)
		}
		target = home + target[1:]
	}

	// Get absolute path
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return fmt.Errorf("cannot resolve path '%s': %w", target, err)
	}

	// Check if target exists and is a directory
	info, err := os.Stat(absTarget)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot cd to '%s': No such file or directory", target)
		}
		return fmt.Errorf("cannot access '%s': %w", target, err)
	}

	if !info.IsDir() {
		return fmt.Errorf("cannot cd to '%s': Not a directory", target)
	}

	// Change directory
	err = os.Chdir(absTarget)
	if err != nil {
		return fmt.Errorf("cannot change directory to '%s': %w", target, err)
	}

	// Print new directory
	pwd, _ := os.Getwd()
	fmt.Println(pwd)
	return nil
}

// checkPermissions checks if the operation is allowed
func (c *CdCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, _ := IsToolAllowed(level, "file_read")
	if !allowed {
		return fmt.Errorf("directory operations are not allowed in %s permission level", level)
	}
	return nil
}

func init() {
	Register(NewCdCommand())
}
