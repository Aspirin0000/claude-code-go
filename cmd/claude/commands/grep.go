package commands

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// GrepResult represents a single grep match
type GrepResult struct {
	Path    string
	LineNum int
	Line    string
	Match   string
}

// GrepCommand search file contents with regex support
type GrepCommand struct {
	*BaseCommand
	defaultLimit int
}

// NewGrepCommand creates a new grep command
func NewGrepCommand() *GrepCommand {
	return &GrepCommand{
		BaseCommand: NewBaseCommand(
			"grep",
			"Search file contents with regex support",
			CategoryTools,
		).WithAliases("grep-files").
			WithHelp(`Usage: /grep [options] <pattern> [path]

Search file contents using regular expressions.

Options:
  -i    Case-insensitive search
  -r    Recursive search (search subdirectories)

Examples:
  /grep "function" ./src           Search for "function" in ./src
  /grep -i "error" ./src       Case-insensitive search in ./src
  /grep -r "config|settings" . Recursive search for configuration terms`),
		defaultLimit: 50,
	}
}

// Execute runs the grep command
func (c *GrepCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		fmt.Println(c.Help())
		return nil
	}

	// Parse flags
	caseInsensitive := false
	recursive := false
	patternStart := 0

	for i, arg := range args {
		if arg == "-i" {
			caseInsensitive = true
			patternStart = i + 1
		} else if arg == "-r" {
			recursive = true
			patternStart = i + 1
		} else if !strings.HasPrefix(arg, "-") {
			if patternStart == 0 {
				patternStart = i
			}
			break
		}
	}

	if patternStart >= len(args) {
		fmt.Println("Error: No pattern specified")
		return nil
	}

	pattern := args[patternStart]
	searchPath := "."

	if patternStart+1 < len(args) {
		searchPath = args[patternStart+1]
	}

	// Compile regex
	flags := ""
	if caseInsensitive {
		flags = "(?i)"
	}

	re, err := regexp.Compile(flags + pattern)
	if err != nil {
		fmt.Printf("Error: Invalid regex pattern: %v\n", err)
		return nil
	}

	// Determine if searchPath is a file pattern or directory
	if strings.Contains(searchPath, "*") || strings.Contains(searchPath, "?") {
		// It's a glob pattern
		return c.searchGlob(ctx, re, searchPath, recursive)
	}

	// Check if it's a file or directory
	info, err := os.Stat(searchPath)
	if err != nil {
		fmt.Printf("Error: Cannot access '%s': %v\n", searchPath, err)
		return nil
	}

	if info.IsDir() {
		return c.searchDirectory(ctx, re, searchPath, recursive)
	}

	// Single file
	return c.searchFile(ctx, re, searchPath)
}

// searchGlob searches files matching a glob pattern
func (c *GrepCommand) searchGlob(ctx context.Context, re *regexp.Regexp, pattern string, recursive bool) error {
	files, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error: Invalid glob pattern: %v\n", err)
		return nil
	}

	if len(files) == 0 {
		fmt.Println("No files match the pattern")
		return nil
	}

	var results []GrepResult
	var count int32
	limit := int32(c.defaultLimit)
	var mu sync.Mutex
	var wg sync.WaitGroup
	resultChan := make(chan GrepResult, 100)

	// Result collector
	go func() {
		for result := range resultChan {
			if atomic.LoadInt32(&count) >= limit {
				continue
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			atomic.AddInt32(&count, 1)
		}
	}()

	// Search files in parallel
	for _, file := range files {
		if atomic.LoadInt32(&count) >= limit {
			break
		}

		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		if info.IsDir() {
			if recursive {
				wg.Add(1)
				go func(dir string) {
					defer wg.Done()
					c.searchDirectoryAsync(ctx, re, dir, resultChan, &count, limit)
				}(file)
			}
		} else {
			wg.Add(1)
			go func(f string) {
				defer wg.Done()
				c.searchFileAsync(ctx, re, f, resultChan, &count, limit)
			}(file)
		}
	}

	wg.Wait()
	close(resultChan)

	// Small delay to ensure all results are collected
	time.Sleep(10 * time.Millisecond)

	c.printResults(results)
	return nil
}

// searchDirectory searches all files in a directory
func (c *GrepCommand) searchDirectory(ctx context.Context, re *regexp.Regexp, dir string, recursive bool) error {
	var results []GrepResult
	var count int32
	limit := int32(c.defaultLimit)
	var mu sync.Mutex
	resultChan := make(chan GrepResult, 100)
	var wg sync.WaitGroup

	// Result collector
	done := make(chan struct{})
	go func() {
		for result := range resultChan {
			if atomic.LoadInt32(&count) >= limit {
				continue
			}
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
			atomic.AddInt32(&count, 1)
		}
		close(done)
	}()

	// Walk directory
	walkFunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if atomic.LoadInt32(&count) >= limit {
			return filepath.SkipDir
		}

		if info.IsDir() {
			if !recursive && path != dir {
				return filepath.SkipDir
			}
			// Skip hidden directories and common non-source directories
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") ||
				base == "node_modules" ||
				base == "vendor" ||
				base == "dist" ||
				base == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary and common non-text files
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".dll" || ext == ".so" || ext == ".dylib" ||
			ext == ".bin" || ext == ".o" || ext == ".a" ||
			ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" ||
			ext == ".mp4" || ext == ".mp3" || ext == ".zip" || ext == ".tar" ||
			ext == ".gz" || ext == ".7z" || ext == ".rar" {
			return nil
		}

		wg.Add(1)
		go func(file string) {
			defer wg.Done()
			c.searchFileAsync(ctx, re, file, resultChan, &count, limit)
		}(path)

		return nil
	}

	err := filepath.Walk(dir, walkFunc)
	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	wg.Wait()
	close(resultChan)
	<-done

	c.printResults(results)
	return nil
}

// searchDirectoryAsync searches directory asynchronously
func (c *GrepCommand) searchDirectoryAsync(ctx context.Context, re *regexp.Regexp, dir string, resultChan chan<- GrepResult, count *int32, limit int32) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if atomic.LoadInt32(count) >= limit {
			return filepath.SkipDir
		}

		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") ||
				base == "node_modules" ||
				base == "vendor" ||
				base == "dist" ||
				base == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".exe" || ext == ".dll" || ext == ".so" || ext == ".dylib" ||
			ext == ".bin" || ext == ".o" || ext == ".a" ||
			ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".gif" ||
			ext == ".mp4" || ext == ".mp3" || ext == ".zip" || ext == ".tar" ||
			ext == ".gz" || ext == ".7z" || ext == ".rar" {
			return nil
		}

		c.searchFileAsync(ctx, re, path, resultChan, count, limit)
		return nil
	})
}

// searchFile searches a single file
func (c *GrepCommand) searchFile(ctx context.Context, re *regexp.Regexp, filePath string) error {
	results := c.searchFileInternal(ctx, re, filePath)
	c.printResults(results)
	return nil
}

// searchFileAsync searches a file asynchronously
func (c *GrepCommand) searchFileAsync(ctx context.Context, re *regexp.Regexp, filePath string, resultChan chan<- GrepResult, count *int32, limit int32) {
	if atomic.LoadInt32(count) >= limit {
		return
	}

	results := c.searchFileInternal(ctx, re, filePath)
	for _, result := range results {
		if atomic.LoadInt32(count) >= limit {
			return
		}
		select {
		case resultChan <- result:
			atomic.AddInt32(count, 1)
		case <-ctx.Done():
			return
		}
	}
}

// searchFileInternal searches a file and returns results
func (c *GrepCommand) searchFileInternal(ctx context.Context, re *regexp.Regexp, filePath string) []GrepResult {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var results []GrepResult
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return results
		default:
		}

		lineNum++
		line := scanner.Text()

		if loc := re.FindStringIndex(line); loc != nil {
			results = append(results, GrepResult{
				Path:    filePath,
				LineNum: lineNum,
				Line:    line,
				Match:   line[loc[0]:loc[1]],
			})
		}
	}

	return results
}

// printResults prints grep results with colored output
func (c *GrepCommand) printResults(results []GrepResult) {
	if len(results) == 0 {
		fmt.Println("No matches found")
		return
	}

	// Color codes
	const (
		colorReset   = "\033[0m"
		colorFile    = "\033[36m" // Cyan
		colorLineNum = "\033[33m" // Yellow
		colorMatch   = "\033[31m" // Red
	)

	for _, result := range results {
		// Colorize the match in the line
		line := result.Line
		match := result.Match

		// Escape ANSI sequences in the line itself to prevent display issues
		line = strings.ReplaceAll(line, "\x1b", "")

		// Highlight match
		highlightedLine := strings.Replace(line, match, colorMatch+match+colorReset, 1)

		fmt.Printf("%s%s%s:%s%d%s: %s\n",
			colorFile, result.Path, colorReset,
			colorLineNum, result.LineNum, colorReset,
			highlightedLine,
		)
	}

	if len(results) >= c.defaultLimit {
		fmt.Printf("\n(Showing first %d results - refine your search to see more)\n", c.defaultLimit)
	} else {
		fmt.Printf("\nFound %d match(es)\n", len(results))
	}
}

func init() {
	Register(NewGrepCommand())
}
