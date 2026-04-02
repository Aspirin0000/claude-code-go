// Package commands provides file operation commands
package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CpCommand copies files or directories
// Usage: /cp <source> <destination>
type CpCommand struct {
	*BaseCommand
}

// NewCpCommand creates the /cp command
func NewCpCommand() *CpCommand {
	return &CpCommand{
		BaseCommand: NewBaseCommand(
			"cp",
			"Copy files or directories",
			CategoryFiles,
		).WithAliases("copy").
			WithHelp(`Copy files or directories.

Usage:
  /cp <source> <dest>       - Copy file
  /cp -r <source> <dest>   - Copy directory recursively

Examples:
  /cp file.txt backup.txt
  /cp -r myapp/ backup/myapp/

Options:
  -r, --recursive    Copy directories recursively`),
	}
}

// Execute copies files or directories
func (c *CpCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("usage: /cp [options] <source> <destination>")
	}

	recursive := false
	var paths []string

	for _, arg := range args {
		switch arg {
		case "-r", "-R", "--recursive":
			recursive = true
		default:
			if !strings.HasPrefix(arg, "-") {
				paths = append(paths, arg)
			}
		}
	}

	if len(paths) < 2 {
		return fmt.Errorf("source and destination required")
	}

	src := paths[0]
	dst := paths[1]

	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source not found: %s", src)
	}

	if srcInfo.IsDir() && !recursive {
		return fmt.Errorf("source is a directory (use -r for recursive copy)")
	}

	// Copy
	if srcInfo.IsDir() {
		err = c.copyDir(src, dst)
	} else {
		err = c.copyFile(src, dst)
	}

	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	fmt.Printf("✓ Copied: %s → %s\n", src, dst)
	return nil
}

// copyFile copies a single file
func (c *CpCommand) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination directory if needed
	dir := filepath.Dir(dst)
	os.MkdirAll(dir, 0755)

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// copyDir copies a directory recursively
func (c *CpCommand) copyDir(src, dst string) error {
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

func init() {
	Register(NewCpCommand())
}
