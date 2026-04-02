package commands

import (
	"context"
	"fmt"
	"os"
)

// PwdCommand prints the current working directory
// Usage: /pwd
type PwdCommand struct {
	*BaseCommand
}

// NewPwdCommand creates the /pwd command
func NewPwdCommand() *PwdCommand {
	return &PwdCommand{
		BaseCommand: NewBaseCommand(
			"pwd",
			"Print working directory",
			CategoryFiles,
		).WithHelp(`Print the current working directory.

Usage:
  /pwd    Print current working directory

Examples:
  /pwd

Output:
  /Users/username/projects/myapp`),
	}
}

// Execute prints the current working directory
func (c *PwdCommand) Execute(ctx context.Context, args []string) error {
	if err := c.checkPermissions(); err != nil {
		return err
	}

	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot get current directory: %w", err)
	}

	fmt.Println(pwd)
	return nil
}

// checkPermissions checks if the operation is allowed
func (c *PwdCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, _ := IsToolAllowed(level, "file_read")
	if !allowed {
		return fmt.Errorf("directory operations are not allowed in %s permission level", level)
	}
	return nil
}

func init() {
	Register(NewPwdCommand())
}
