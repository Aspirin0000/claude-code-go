package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	readColorReset  = "\033[0m"
	readColorGray   = "\033[90m"
	readColorCyan   = "\033[36m"
	readColorYellow = "\033[33m"
	readColorGreen  = "\033[32m"
	readColorRed    = "\033[31m"
	readMaxLines    = 100
	readMaxFileSize = 10 * 1024 * 1024 // 10MB
)

// ReadCommand reads and displays file contents
type ReadCommand struct {
	*BaseCommand
}

// NewReadCommand creates a new read command
func NewReadCommand() *ReadCommand {
	return &ReadCommand{
		BaseCommand: NewBaseCommand(
			"read",
			"Read and display file contents",
			CategoryFiles,
		).WithAliases("cat", "view").
			WithHelp(`Usage: /read <file> [options]

Read and display file contents with optional line numbers and syntax highlighting.

Arguments:
  <file>    File path to read

Options:
  -n        Show line numbers
  -l        Limit output to first 100 lines (default)
  --all     Show entire file (no line limit)

Examples:
  /read file.txt           Read file.txt
  /read file.txt -n        Read with line numbers
  /read main.go -n         Read Go file with syntax highlighting
  /read large.log --all    Read entire log file

Features:
  - Syntax highlighting for code files
  - Line numbers option
  - Large file protection (shows first 100 lines by default)
  - Binary file detection

Aliases: /cat, /view`),
	}
}

// Execute runs the read command
func (c *ReadCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println(c.Help())
		return nil
	}

	if err := c.checkPermissions(); err != nil {
		return err
	}

	filePath, showLines, noLimit := c.parseArgs(args)

	absPath, err := c.resolvePath(filePath)
	if err != nil {
		return err
	}

	if err := c.validateFile(absPath); err != nil {
		return err
	}

	return c.readFile(absPath, showLines, noLimit)
}

func (c *ReadCommand) parseArgs(args []string) (filePath string, showLines, noLimit bool) {
	for _, arg := range args {
		switch arg {
		case "-n":
			showLines = true
		case "--all":
			noLimit = true
		default:
			if !strings.HasPrefix(arg, "-") && filePath == "" {
				filePath = arg
			}
		}
	}
	return filePath, showLines, noLimit
}

func (c *ReadCommand) resolvePath(path string) (string, error) {
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

	return absPath, nil
}

func (c *ReadCommand) validateFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied: %s", path)
		}
		return fmt.Errorf("cannot access file: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a file", path)
	}

	if info.Size() > readMaxFileSize {
		return fmt.Errorf("file too large (%.1f MB > %.1f MB limit)",
			float64(info.Size())/(1024*1024),
			float64(readMaxFileSize)/(1024*1024))
	}

	return nil
}

func (c *ReadCommand) checkPermissions() error {
	level := GetCurrentPermissionLevel()
	allowed, _ := IsToolAllowed(level, "file_read")
	if !allowed {
		return fmt.Errorf("file read operations are not allowed in %s permission level", level)
	}
	return nil
}

func (c *ReadCommand) readFile(path string, showLines, noLimit bool) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("cannot open file: %w", err)
	}
	defer file.Close()

	// Check if file is binary
	if c.isBinaryFile(file) {
		fmt.Printf("%s⚠ Warning:%s This appears to be a binary file. Displaying may produce garbled output.\n\n",
			readColorYellow, readColorReset)
		file.Seek(0, 0)
	}

	ext := strings.ToLower(filepath.Ext(path))
	isCodeFile := c.isCodeFile(ext)

	fmt.Printf("%s%sFile:%s %s%s\n", readColorGray, strings.Repeat("-", 10), readColorReset, path, readColorGray)
	fmt.Println(strings.Repeat("-", 50) + readColorReset)

	scanner := bufio.NewScanner(file)
	lineNum := 0
	limit := readMaxLines
	if noLimit {
		limit = -1
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if limit > 0 && lineNum > limit {
			fmt.Printf("\n%s... (%d more lines, use --all to show entire file)%s\n",
				readColorGray, c.countRemainingLines(file, lineNum), readColorReset)
			break
		}

		if showLines {
			fmt.Printf("%s%4d%s  ", readColorGray, lineNum, readColorReset)
		}

		if isCodeFile {
			line = c.highlightSyntax(line, ext)
		}

		fmt.Println(line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	fmt.Printf("\n%s%s%s\n", readColorGray, strings.Repeat("-", 50), readColorReset)
	fmt.Printf("%s%d lines read%s\n", readColorGray, lineNum, readColorReset)

	return nil
}

func (c *ReadCommand) isBinaryFile(file *os.File) bool {
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true
		}
	}
	return false
}

func (c *ReadCommand) isCodeFile(ext string) bool {
	codeExts := []string{
		".go", ".js", ".ts", ".py", ".java", ".c", ".cpp", ".h", ".hpp",
		".rs", ".rb", ".php", ".swift", ".kt", ".scala", ".r", ".m",
		".cs", ".fs", ".fsx", ".vb", ".pl", ".pm", ".t", ".sh",
		".bash", ".zsh", ".fish", ".ps1", ".psm1",
		".html", ".htm", ".xml", ".json", ".yaml", ".yml", ".toml",
		".css", ".scss", ".sass", ".less", ".sql",
		".md", ".markdown", ".txt", ".log",
	}

	for _, codeExt := range codeExts {
		if ext == codeExt {
			return true
		}
	}
	return false
}

func (c *ReadCommand) highlightSyntax(line, ext string) string {
	switch ext {
	case ".go", ".js", ".ts", ".java", ".c", ".cpp", ".cs":
		return c.highlightCLike(line)
	case ".py", ".rb":
		return c.highlightPythonLike(line)
	case ".html", ".htm", ".xml":
		return c.highlightMarkup(line)
	case ".json":
		return c.highlightJSON(line)
	case ".md", ".markdown":
		return c.highlightMarkdown(line)
	default:
		return line
	}
}

func (c *ReadCommand) highlightCLike(line string) string {
	keywords := []string{
		"package", "import", "func", "type", "struct", "interface", "map", "chan",
		"var", "const", "if", "else", "for", "range", "switch", "case", "default",
		"return", "break", "continue", "goto", "defer", "go", "select",
		"class", "public", "private", "protected", "static", "void", "int", "string",
		"function", "const", "let", "var", "async", "await", "export", "import",
	}

	result := line
	for _, kw := range keywords {
		result = strings.ReplaceAll(result, kw, readColorCyan+kw+readColorReset)
	}
	return result
}

func (c *ReadCommand) highlightPythonLike(line string) string {
	keywords := []string{
		"def", "class", "import", "from", "as", "if", "elif", "else", "for", "while",
		"try", "except", "finally", "with", "return", "yield", "lambda", "pass",
		"break", "continue", "raise", "assert", "del", "global", "nonlocal",
	}

	result := line
	for _, kw := range keywords {
		result = strings.ReplaceAll(result, kw, readColorCyan+kw+readColorReset)
	}
	return result
}

func (c *ReadCommand) highlightMarkup(line string) string {
	// Simple tag highlighting
	if strings.HasPrefix(strings.TrimSpace(line), "<") {
		return readColorRed + line + readColorReset
	}
	return line
}

func (c *ReadCommand) highlightJSON(line string) string {
	// Highlight keys and values
	if strings.Contains(line, `":`) {
		parts := strings.SplitN(line, `":`, 2)
		if len(parts) == 2 {
			return readColorCyan + parts[0] + `":` + readColorGreen + parts[1] + readColorReset
		}
	}
	return line
}

func (c *ReadCommand) highlightMarkdown(line string) string {
	trimmed := strings.TrimSpace(line)

	// Headers
	if strings.HasPrefix(trimmed, "#") {
		return readColorCyan + readColorBold + line + readColorReset
	}

	// Bold/italic
	if strings.Contains(line, "**") || strings.Contains(line, "*") {
		return readColorYellow + line + readColorReset
	}

	// Code blocks
	if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "`") {
		return readColorGreen + line + readColorReset
	}

	return line
}

func (c *ReadCommand) countRemainingLines(file *os.File, currentLine int) int {
	// Simple estimation - count newlines in remaining file
	buf := make([]byte, 32*1024)
	count := 0
	for {
		n, err := file.Read(buf)
		if n == 0 || err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if buf[i] == '\n' {
				count++
			}
		}
	}
	return count
}

var readColorBold = "\033[1m"

func init() {
	Register(NewReadCommand())
}
