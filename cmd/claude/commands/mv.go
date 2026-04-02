package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MvCommand moves files or directories
// Usage: /mv <source> <destination>
type MvCommand struct {
	*BaseCommand
}

// NewMvCommand creates the /mv command
func NewMvCommand() *MvCommand {
	return &MvCommand{
		BaseCommand: NewBaseCommand(
			"mv",
			"Move files or directories",
			CategoryFiles,
		).WithAliases("move").
			WithHelp(`Move files or directories.

Usage:
  /mv <source> <destination>    Move file or directory

Examples:
  /mv file.txt newname.txt
  /mv mydir/ otherdir/
  /mv file.txt /path/to/dest/`),
	}
}

// Execute moves files or directories
func (c *MvCommand) Execute(ctx context.Context, args []string) error {
	if err := c.checkPermissions(); err != nil {
		return err
	}

	if len(args) < 2 {
		return fmt.Errorf("usage: /mv <source> <destination>")
	}

	src := args[0]
	dst := args[1]

	// Expand ~ in paths
	src = expandPath(src)
	dst = expandPath(dst)

	// Get absolute paths
	absSrc, err := filepath.Abs(src)
	if err != nil {
		return fmt.Errorf("cannot resolve source path '%s': %w", src, err)
	}

	absDst, err := filepath.Abs(dst)
	if err != nil {
		return fmt.Errorf("cannot resolve destination path '%s': %w", dst, err)
	}

	// Check if source exists
	srcInfo, err := os.Stat(absSrc)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot stat '%s': No such file or directory", src)
		}
		return fmt.Errorf("cannot access '%s': %w", src, err)
	}

	// If destination exists and is a directory, move source into it
	dstInfo, err := os.Stat(absDst)
	if err == nil && dstInfo.IsDir() {
		absDst = filepath.Join(absDst, filepath.Base(absSrc))
	}

	// Perform the move
	err = os.Rename(absSrc, absDst)
	if err != nil {
		// Handle cross-device move by copying then deleting
		if strings.Contains(err.Error(), "cross-device") || strings.Contains(err.Error(), "invalid link") {
			err = c.moveCrossDevice(absSrc, absDst, srcInfo.IsDir())
			if err != nil {
				return fmt.Errorf("move failed: %w", err)
			}
		} else {
			return fmt.Errorf("move failed: %w", err)
		}
	}

	fmt.Printf("✓ Moved: %s → %s\n", src, dst)
	return nil
}

// moveCrossDevice handles moving files across different filesystems
func (c *MvCommand) moveCrossDevice(src, dst string, isDir bool) error {
	if isDir {
		// Copy directory then remove source
		err := c.copyDir(src, dst)
		if err != nil {
			return err
		}
		return os.RemoveAll(src)
	}
	// Copy file then remove source
	err := c.copyFile(src, dst)
	if err != nil {
		return err
	}
	return os.Remove(src)
}

// copyFile copies a single file
func (c *MvCommand) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	// Create destination directory if needed
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(dst, data, 0644)
}

// copyDir copies a directory recursively
func (c *MvCommand) copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return c.copyFile(path, dstPath)
	})
}

// checkPermissions checks if the operation is allowed
func (c *MvCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, _ := IsToolAllowed(level, "file_write")
	if !allowed {
		return fmt.Errorf("file write operations are not allowed in %s permission level", level)
	}
	return nil
}

// expandPath expands ~ to home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			path = home + path[1:]
		}
	}
	return path
}

func init() {
	Register(NewMvCommand())
}
