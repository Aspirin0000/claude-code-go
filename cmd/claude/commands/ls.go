package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

const (
	colorReset      = "\033[0m"
	colorBold       = "\033[1m"
	colorDir        = "\033[34;1m"
	colorExecutable = "\033[32;1m"
	colorSymlink    = "\033[36;1m"
	colorHidden     = "\033[90m"
)

// LsCommand lists directory contents
type LsCommand struct {
	*BaseCommand
}

// NewLsCommand creates a new ls command
func NewLsCommand() *LsCommand {
	return &LsCommand{
		BaseCommand: NewBaseCommand(
			"ls",
			"List directory contents",
			CategoryFiles,
		).WithHelp(`Usage: /ls [path] [options]

List directory contents with details including permissions, owner, size, and modification time.

Arguments:
  [path]    Directory path to list (default: current directory)

Options:
  -l        Use long listing format (permissions, owner, size, time)
  -a        Show all files including hidden files (starting with .)
  -la       Combination of -l and -a flags

Examples:
  /ls                    List files in current directory
  /ls /path/to/dir       List files in specified directory
  /ls -l                 List with detailed information
  /ls -la                List all files including hidden with details
  /ls -la /tmp           List all files in /tmp with details

Color Coding:
  Directories    Blue bold
  Executables    Green bold
  Symlinks       Cyan bold
  Hidden files   Gray`),
	}
}

// Execute runs the ls command
func (c *LsCommand) Execute(ctx context.Context, args []string) error {
	if err := c.checkPermissions(); err != nil {
		return err
	}

	path, showAll, longFormat := c.parseArgs(args)

	absPath, err := c.resolvePath(path)
	if err != nil {
		return fmt.Errorf("cannot access '%s': %w", path, err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("cannot access '%s': No such file or directory", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("cannot access '%s': Permission denied", path)
		}
		return fmt.Errorf("cannot access '%s': %w", path, err)
	}

	if !info.IsDir() {
		c.printFileInfo(absPath, info, longFormat)
		return nil
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("cannot open directory '%s': Permission denied", path)
		}
		return fmt.Errorf("error reading directory '%s': %w", path, err)
	}

	var files []fileEntry
	for _, entry := range entries {
		name := entry.Name()

		if !showAll && strings.HasPrefix(name, ".") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, fileEntry{
			name: name,
			path: filepath.Join(absPath, name),
			info: info,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		iIsDir := files[i].info.IsDir()
		jIsDir := files[j].info.IsDir()
		if iIsDir != jIsDir {
			return iIsDir
		}
		return strings.ToLower(files[i].name) < strings.ToLower(files[j].name)
	})

	if longFormat {
		c.printLongFormat(files, absPath)
	} else {
		c.printShortFormat(files)
	}

	return nil
}

type fileEntry struct {
	name string
	path string
	info os.FileInfo
}

func (c *LsCommand) parseArgs(args []string) (path string, showAll, longFormat bool) {
	path = "."

	for _, arg := range args {
		switch arg {
		case "-a":
			showAll = true
		case "-l":
			longFormat = true
		case "-la", "-al":
			showAll = true
			longFormat = true
		default:
			if !strings.HasPrefix(arg, "-") {
				path = arg
			}
		}
	}

	return path, showAll, longFormat
}

func (c *LsCommand) resolvePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("cannot get home directory: %w", err)
		}
		path = home + path[1:]
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if err := c.validatePath(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

func (c *LsCommand) validatePath(path string) error {
	sensitivePaths := []string{
		"/proc",
		"/sys",
		"/dev",
	}

	for _, sensitive := range sensitivePaths {
		if strings.HasPrefix(path, sensitive) {
			return fmt.Errorf("access to system directory '%s' is not allowed", sensitive)
		}
	}

	return nil
}

func (c *LsCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, _ := IsToolAllowed(level, "file_read")
	if !allowed {
		return fmt.Errorf("file read operations are not allowed in %s permission level", level)
	}
	return nil
}

func (c *LsCommand) printLongFormat(files []fileEntry, dir string) {
	if len(files) == 0 {
		fmt.Println("total 0")
		return
	}

	maxSizeLen := 0
	maxOwnerLen := 0
	for _, f := range files {
		sizeStr := c.formatSize(f.info.Size())
		if len(sizeStr) > maxSizeLen {
			maxSizeLen = len(sizeStr)
		}

		owner := c.getOwner(f.info)
		if len(owner) > maxOwnerLen {
			maxOwnerLen = len(owner)
		}
	}

	fmt.Printf("%stotal %d%s\n", colorBold, len(files), colorReset)

	for _, f := range files {
		mode := c.formatMode(f.info.Mode())
		owner := c.getOwner(f.info)
		size := c.formatSize(f.info.Size())
		time := c.formatTime(f.info.ModTime())
		name := c.colorizeName(f.name, f.info)

		fmt.Printf("%s %-*s %-*s %*s %s %s\n",
			mode,
			maxOwnerLen, owner,
			maxOwnerLen, owner,
			maxSizeLen, size,
			time,
			name,
		)
	}
}

func (c *LsCommand) printShortFormat(files []fileEntry) {
	if len(files) == 0 {
		return
	}

	termWidth := 80
	maxLen := 0
	for _, f := range files {
		if len(f.name) > maxLen {
			maxLen = len(f.name)
		}
	}

	colWidth := maxLen + 2
	cols := termWidth / colWidth
	if cols < 1 {
		cols = 1
	}

	for i, f := range files {
		name := c.colorizeName(f.name, f.info)
		if i > 0 && i%cols == 0 {
			fmt.Println()
		}
		fmt.Printf("%-*s", colWidth, name+colorReset)
	}
	fmt.Println()
}

func (c *LsCommand) printFileInfo(path string, info os.FileInfo, longFormat bool) {
	if longFormat {
		mode := c.formatMode(info.Mode())
		owner := c.getOwner(info)
		size := c.formatSize(info.Size())
		time := c.formatTime(info.ModTime())
		name := c.colorizeName(info.Name(), info)

		fmt.Printf("%s %s %s %s %s\n", mode, owner, size, time, name)
	} else {
		name := c.colorizeName(info.Name(), info)
		fmt.Println(name + colorReset)
	}
}

func (c *LsCommand) formatMode(mode os.FileMode) string {
	var buf strings.Builder

	switch {
	case mode.IsDir():
		buf.WriteString("d")
	case mode&os.ModeSymlink != 0:
		buf.WriteString("l")
	default:
		buf.WriteString("-")
	}

	perm := uint32(mode.Perm())
	perms := []struct {
		bit uint32
		ch  byte
	}{
		{0400, 'r'}, {0200, 'w'}, {0100, 'x'},
		{0040, 'r'}, {0020, 'w'}, {0010, 'x'},
		{0004, 'r'}, {0002, 'w'}, {0001, 'x'},
	}

	for _, p := range perms {
		if perm&p.bit != 0 {
			buf.WriteByte(p.ch)
		} else {
			buf.WriteByte('-')
		}
	}

	return buf.String()
}

func (c *LsCommand) getOwner(info os.FileInfo) string {
	if stat, ok := info.Sys().(*syscall.Stat_t); ok {
		return fmt.Sprintf("%d", stat.Uid)
	}
	return "0"
}

func (c *LsCommand) formatSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.1fG", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.1fM", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.1fK", float64(size)/KB)
	default:
		return fmt.Sprintf("%d", size)
	}
}

func (c *LsCommand) formatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < 6*30*24*time.Hour {
		return t.Format("Jan 02 15:04")
	}
	return t.Format("Jan 02  2006")
}

func (c *LsCommand) colorizeName(name string, info os.FileInfo) string {
	mode := info.Mode()

	if strings.HasPrefix(name, ".") {
		return colorHidden + name
	}

	if mode&os.ModeSymlink != 0 {
		return colorSymlink + name + colorReset + " ->"
	}

	if mode.IsDir() {
		return colorDir + name
	}

	if mode&0111 != 0 {
		return colorExecutable + name
	}

	return name
}

func init() {
	Register(NewLsCommand())
}
