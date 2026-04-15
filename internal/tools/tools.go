// Package tools provides concrete tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// BashTool Bash command execution tool
type BashTool struct{}

func (b *BashTool) Name() string        { return "bash" }
func (b *BashTool) Description() string { return "Execute shell commands" }
func (b *BashTool) IsReadOnly() bool    { return false }
func (b *BashTool) IsDestructive() bool { return true }

func (b *BashTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {"type": "string", "description": "Command to execute"},
			"timeout": {"type": "number", "description": "Timeout in milliseconds"},
			"description": {"type": "string", "description": "Command description"}
		},
		"required": ["command"]
	}`)
}

func (b *BashTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Command     string `json:"command"`
		Timeout     int    `json:"timeout"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	timeout := 30 * time.Second
	if params.Timeout > 0 {
		timeout = time.Duration(params.Timeout) * time.Millisecond
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "bash", "-c", params.Command)
	output, err := cmd.CombinedOutput()

	result := struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Code   int    `json:"return_code"`
	}{
		Stdout: string(output),
		Code:   0,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.Code = exitErr.ExitCode()
		} else {
			result.Stderr = err.Error()
		}
	}

	return json.Marshal(result)
}

// FileReadTool File reading tool
type FileReadTool struct{}

func (f *FileReadTool) Name() string        { return "file_read" }
func (f *FileReadTool) Description() string { return "Read file contents" }
func (f *FileReadTool) IsReadOnly() bool    { return true }
func (f *FileReadTool) IsDestructive() bool { return false }

func (f *FileReadTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "Absolute file path"},
			"offset": {"type": "number", "description": "Starting line number"},
			"limit": {"type": "number", "description": "Number of lines to read"}
		},
		"required": ["file_path"]
	}`)
}

func (f *FileReadTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		FilePath string `json:"file_path"`
		Offset   int    `json:"offset"`
		Limit    int    `json:"limit"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	if params.Offset > 0 && params.Offset < len(lines) {
		lines = lines[params.Offset:]
	}

	if params.Limit > 0 && params.Limit < len(lines) {
		lines = lines[:params.Limit]
	}

	result := struct {
		Content    string `json:"content"`
		NumLines   int    `json:"num_lines"`
		TotalLines int    `json:"total_lines"`
		StartLine  int    `json:"start_line"`
	}{
		Content:    strings.Join(lines, "\n"),
		NumLines:   len(lines),
		TotalLines: len(strings.Split(string(content), "\n")),
		StartLine:  params.Offset,
	}

	return json.Marshal(result)
}

// FileWriteTool File writing tool
type FileWriteTool struct{}

func (f *FileWriteTool) Name() string        { return "file_write" }
func (f *FileWriteTool) Description() string { return "Write or create new files" }
func (f *FileWriteTool) IsReadOnly() bool    { return false }
func (f *FileWriteTool) IsDestructive() bool { return true }

func (f *FileWriteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "File path"},
			"content": {"type": "string", "description": "File content"}
		},
		"required": ["file_path", "content"]
	}`)
}

func (f *FileWriteTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if err := os.WriteFile(params.FilePath, []byte(params.Content), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	return json.Marshal(result)
}

// FileEditTool File editing tool - using search and replace
type FileEditTool struct{}

func (f *FileEditTool) Name() string        { return "file_edit" }
func (f *FileEditTool) Description() string { return "Edit file contents (search and replace)" }
func (f *FileEditTool) IsReadOnly() bool    { return false }
func (f *FileEditTool) IsDestructive() bool { return true }

func (f *FileEditTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "File path"},
			"old_string": {"type": "string", "description": "Old string to replace"},
			"new_string": {"type": "string", "description": "New string"}
		},
		"required": ["file_path", "old_string", "new_string"]
	}`)
}

func (f *FileEditTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		FilePath  string `json:"file_path"`
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	oldContent := string(content)
	if !strings.Contains(oldContent, params.OldString) {
		return nil, fmt.Errorf("string to replace not found")
	}

	newContent := strings.Replace(oldContent, params.OldString, params.NewString, 1)

	if err := os.WriteFile(params.FilePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	return json.Marshal(result)
}

// GrepTool Grep search tool
type GrepTool struct{}

func (g *GrepTool) Name() string        { return "grep" }
func (g *GrepTool) Description() string { return "Search file contents using regular expressions" }
func (g *GrepTool) IsReadOnly() bool    { return true }
func (g *GrepTool) IsDestructive() bool { return false }

func (g *GrepTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "Regular expression pattern"},
			"path": {"type": "string", "description": "Search path"},
			"glob": {"type": "string", "description": "File filter pattern"}
		},
		"required": ["pattern"]
	}`)
}

func (g *GrepTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
		Glob    string `json:"glob"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	// Use ripgrep
	args := []string{params.Pattern}
	if params.Path != "" {
		args = append(args, params.Path)
	}
	if params.Glob != "" {
		args = append(args, "--glob", params.Glob)
	}

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.CombinedOutput()

	content := string(output)
	if err != nil {
		// ripgrep returns exit code 1 when no matches found, which is normal
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			content = ""
		} else {
			return nil, fmt.Errorf("grep execution failed: %w", err)
		}
	}

	matches := []string{}
	if content != "" {
		lines := strings.Split(strings.TrimSpace(content), "\n")
		for _, line := range lines {
			if line != "" {
				matches = append(matches, line)
			}
		}
	}

	result := struct {
		Content    string   `json:"content"`
		Matches    []string `json:"matches"`
		NumMatches int      `json:"num_matches"`
	}{
		Content:    content,
		Matches:    matches,
		NumMatches: len(matches),
	}

	return json.Marshal(result)
}

// GlobTool Glob file matching tool
type GlobTool struct{}

func (g *GlobTool) Name() string        { return "glob" }
func (g *GlobTool) Description() string { return "Find files by pattern" }
func (g *GlobTool) IsReadOnly() bool    { return true }
func (g *GlobTool) IsDestructive() bool { return false }

func (g *GlobTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "Glob pattern"},
			"path": {"type": "string", "description": "Search directory"}
		},
		"required": ["pattern"]
	}`)
}

func (g *GlobTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Pattern string `json:"pattern"`
		Path    string `json:"path"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	// Use find command for glob functionality
	searchPath := params.Path
	if searchPath == "" {
		searchPath = "."
	}

	cmd := exec.CommandContext(ctx, "find", searchPath, "-name", params.Pattern, "-type", "f")
	output, err := cmd.Output()

	files := []string{}
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		for _, line := range lines {
			if line != "" {
				files = append(files, line)
			}
		}
	}

	result := struct {
		Files     []string `json:"files"`
		NumFiles  int      `json:"num_files"`
		Truncated bool     `json:"truncated"`
	}{
		Files:     files,
		NumFiles:  len(files),
		Truncated: false,
	}

	return json.Marshal(result)
}

// TodoWriteTool Todo list tool
type TodoWriteTool struct{}

func (t *TodoWriteTool) Name() string        { return "todo_write" }
func (t *TodoWriteTool) Description() string { return "Create or update task lists" }
func (t *TodoWriteTool) IsReadOnly() bool    { return false }
func (t *TodoWriteTool) IsDestructive() bool { return false }

func (t *TodoWriteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"todos": {
				"type": "array",
				"description": "Task list",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "string", "description": "Task ID"},
						"content": {"type": "string", "description": "Task content"},
						"status": {"type": "string", "description": "Status: in_progress, done, cancelled"},
						"priority": {"type": "string", "description": "Priority: high, medium, low"}
					}
				}
			}
		},
		"required": ["todos"]
	}`)
}

func (t *TodoWriteTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Todos []struct {
			ID       string `json:"id"`
			Content  string `json:"content"`
			Status   string `json:"status"`
			Priority string `json:"priority"`
		} `json:"todos"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	// Return success; a full app could persist to a file or database
	result := struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}{
		Success: true,
		Count:   len(params.Todos),
	}

	return json.Marshal(result)
}

// DirectoryReadTool Directory listing tool
type DirectoryReadTool struct{}

func (d *DirectoryReadTool) Name() string        { return "dir_read" }
func (d *DirectoryReadTool) Description() string { return "List files and directories" }
func (d *DirectoryReadTool) IsReadOnly() bool    { return true }
func (d *DirectoryReadTool) IsDestructive() bool { return false }

func (d *DirectoryReadTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Directory path to list"},
			"recursive": {"type": "boolean", "description": "List recursively"}
		},
		"required": ["path"]
	}`)
}

func (d *DirectoryReadTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path      string `json:"path"`
		Recursive bool   `json:"recursive"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	info, err := os.Stat(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", params.Path)
	}

	type Entry struct {
		Name    string `json:"name"`
		IsDir   bool   `json:"is_dir"`
		Size    int64  `json:"size"`
		ModTime string `json:"mod_time"`
	}

	var entries []Entry
	if params.Recursive {
		err = filepath.Walk(params.Path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if path == params.Path {
				return nil
			}
			rel, _ := filepath.Rel(params.Path, path)
			entries = append(entries, Entry{
				Name:    rel,
				IsDir:   info.IsDir(),
				Size:    info.Size(),
				ModTime: info.ModTime().Format(time.RFC3339),
			})
			return nil
		})
	} else {
		items, err := os.ReadDir(params.Path)
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			info, _ := item.Info()
			size := int64(0)
			modTime := ""
			if info != nil {
				size = info.Size()
				modTime = info.ModTime().Format(time.RFC3339)
			}
			entries = append(entries, Entry{
				Name:    item.Name(),
				IsDir:   item.IsDir(),
				Size:    size,
				ModTime: modTime,
			})
		}
	}
	if err != nil {
		return nil, err
	}

	return json.Marshal(struct {
		Entries []Entry `json:"entries"`
		Path    string  `json:"path"`
		Count   int     `json:"count"`
	}{
		Entries: entries,
		Path:    params.Path,
		Count:   len(entries),
	})
}

// ThinkTool Reasoning tool
type ThinkTool struct{}

func (t *ThinkTool) Name() string { return "think" }
func (t *ThinkTool) Description() string {
	return "Use this tool to think through complex problems step by step before taking action"
}
func (t *ThinkTool) IsReadOnly() bool    { return true }
func (t *ThinkTool) IsDestructive() bool { return false }

func (t *ThinkTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"thought": {"type": "string", "description": "Your step-by-step reasoning"}
		},
		"required": ["thought"]
	}`)
}

func (t *ThinkTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Thought string `json:"thought"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		Status  string `json:"status"`
		Thought string `json:"thought"`
	}{
		Status:  "ok",
		Thought: params.Thought,
	})
}

// WebSearchTool Web search tool
type WebSearchTool struct{}

func (w *WebSearchTool) Name() string        { return "web_search" }
func (w *WebSearchTool) Description() string { return "Search for information on the web" }
func (w *WebSearchTool) IsReadOnly() bool    { return true }
func (w *WebSearchTool) IsDestructive() bool { return false }

func (w *WebSearchTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "Search query"}
		},
		"required": ["query"]
	}`)
}

func (w *WebSearchTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	searchURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(params.Query)
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type SearchResult struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Snippet string `json:"snippet"`
	}

	var results []SearchResult

	// Parse DuckDuckGo HTML results
	re := regexp.MustCompile(`<a[^>]*class="result__a"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	snippetRe := regexp.MustCompile(`<a[^>]*class="result__snippet"[^>]*>(.*?)</a>`)
	snippetMatches := snippetRe.FindAllStringSubmatch(string(body), -1)

	for i, m := range matches {
		if i >= 10 {
			break
		}
		title := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[2], "")
		snippet := ""
		if i < len(snippetMatches) {
			snippet = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(snippetMatches[i][1], "")
		}
		results = append(results, SearchResult{
			Title:   strings.TrimSpace(title),
			URL:     strings.TrimSpace(m[1]),
			Snippet: strings.TrimSpace(snippet),
		})
	}

	if len(results) == 0 {
		// Fallback: try lite.duckduckgo.com
		liteURL := "https://lite.duckduckgo.com/lite/?q=" + url.QueryEscape(params.Query)
		req2, _ := http.NewRequestWithContext(ctx, "GET", liteURL, nil)
		req2.Header.Set("User-Agent", "Mozilla/5.0")
		resp2, err2 := client.Do(req2)
		if err2 == nil {
			defer resp2.Body.Close()
			body2, _ := io.ReadAll(resp2.Body)
			liteRe := regexp.MustCompile(`<a[^>]*class="result-link"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
			liteMatches := liteRe.FindAllStringSubmatch(string(body2), -1)
			for i, m := range liteMatches {
				if i >= 10 {
					break
				}
				title := regexp.MustCompile(`<[^>]+>`).ReplaceAllString(m[2], "")
				results = append(results, SearchResult{
					Title: strings.TrimSpace(title),
					URL:   strings.TrimSpace(m[1]),
				})
			}
		}
	}

	return json.Marshal(struct {
		Results []SearchResult `json:"results"`
		Query   string         `json:"query"`
		Count   int            `json:"count"`
	}{
		Results: results,
		Query:   params.Query,
		Count:   len(results),
	})
}

// WebFetchTool Web fetch tool
type WebFetchTool struct{}

func (w *WebFetchTool) Name() string        { return "web_fetch" }
func (w *WebFetchTool) Description() string { return "Fetch web page content" }
func (w *WebFetchTool) IsReadOnly() bool    { return true }
func (w *WebFetchTool) IsDestructive() bool { return false }

func (w *WebFetchTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"url": {"type": "string", "description": "Web page URL"}
		},
		"required": ["url"]
	}`)
}

func (w *WebFetchTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		URL string `json:"url"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	// Use curl to fetch web page content
	cmd := exec.CommandContext(ctx, "curl", "-s", "-L", params.URL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch web page: %w", err)
	}

	result := struct {
		Content string `json:"content"`
		URL     string `json:"url"`
	}{
		Content: string(output),
		URL:     params.URL,
	}

	return json.Marshal(result)
}

// FileDeleteTool File deletion tool
type FileDeleteTool struct{}

func (f *FileDeleteTool) Name() string        { return "file_delete" }
func (f *FileDeleteTool) Description() string { return "Delete a file or empty directory" }
func (f *FileDeleteTool) IsReadOnly() bool    { return false }
func (f *FileDeleteTool) IsDestructive() bool { return true }

func (f *FileDeleteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Path to delete"}
		},
		"required": ["path"]
	}`)
}

func (f *FileDeleteTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if err := os.Remove(params.Path); err != nil {
		return nil, fmt.Errorf("failed to delete: %w", err)
	}

	return json.Marshal(struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
	}{
		Success: true,
		Path:    params.Path,
	})
}

// DirWriteTool Directory creation tool
type DirWriteTool struct{}

func (d *DirWriteTool) Name() string        { return "dir_write" }
func (d *DirWriteTool) Description() string { return "Create a directory (including parents)" }
func (d *DirWriteTool) IsReadOnly() bool    { return false }
func (d *DirWriteTool) IsDestructive() bool { return false }

func (d *DirWriteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Directory path to create"}
		},
		"required": ["path"]
	}`)
}

func (d *DirWriteTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(params.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return json.Marshal(struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
	}{
		Success: true,
		Path:    params.Path,
	})
}

// FileMoveTool File move/rename tool
type FileMoveTool struct{}

func (f *FileMoveTool) Name() string        { return "file_move" }
func (f *FileMoveTool) Description() string { return "Move or rename a file or directory" }
func (f *FileMoveTool) IsReadOnly() bool    { return false }
func (f *FileMoveTool) IsDestructive() bool { return false }

func (f *FileMoveTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"source": {"type": "string", "description": "Source path"},
			"destination": {"type": "string", "description": "Destination path"}
		},
		"required": ["source", "destination"]
	}`)
}

func (f *FileMoveTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	if err := os.Rename(params.Source, params.Destination); err != nil {
		return nil, fmt.Errorf("failed to move: %w", err)
	}

	return json.Marshal(struct {
		Success     bool   `json:"success"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}{
		Success:     true,
		Source:      params.Source,
		Destination: params.Destination,
	})
}

// GitStatusTool Git status tool
type GitStatusTool struct{}

func (g *GitStatusTool) Name() string        { return "git_status" }
func (g *GitStatusTool) Description() string { return "Check git repository status" }
func (g *GitStatusTool) IsReadOnly() bool    { return true }
func (g *GitStatusTool) IsDestructive() bool { return false }

func (g *GitStatusTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"}
		}
	}`)
}

func (g *GitStatusTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain", "-b")
	output, err := cmd.CombinedOutput()

	result := struct {
		Status string `json:"status"`
		Branch string `json:"branch"`
		Path   string `json:"path"`
		Error  string `json:"error,omitempty"`
	}{
		Status: string(output),
		Path:   path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	// Parse branch from first line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			result.Branch = strings.TrimPrefix(line, "## ")
			break
		}
	}

	return json.Marshal(result)
}

// GitDiffTool Git diff tool
type GitDiffTool struct{}

func (g *GitDiffTool) Name() string        { return "git_diff" }
func (g *GitDiffTool) Description() string { return "Show git diff for the repository" }
func (g *GitDiffTool) IsReadOnly() bool    { return true }
func (g *GitDiffTool) IsDestructive() bool { return false }

func (g *GitDiffTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"staged": {"type": "boolean", "description": "Show staged changes"}
		}
	}`)
}

func (g *GitDiffTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Staged bool   `json:"staged"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "diff"}
	if params.Staged {
		args = append(args, "--staged")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Diff  string `json:"diff"`
		Path  string `json:"path"`
		Error string `json:"error,omitempty"`
	}{
		Diff: string(output),
		Path: path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitLogTool Git log tool
type GitLogTool struct{}

func (g *GitLogTool) Name() string        { return "git_log" }
func (g *GitLogTool) Description() string { return "Show recent git commit history" }
func (g *GitLogTool) IsReadOnly() bool    { return true }
func (g *GitLogTool) IsDestructive() bool { return false }

func (g *GitLogTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"count": {"type": "number", "description": "Number of commits to show (default: 10)"}
		}
	}`)
}

func (g *GitLogTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path  string `json:"path"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}
	count := params.Count
	if count <= 0 {
		count = 10
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "log", "--oneline", "-n", fmt.Sprintf("%d", count))
	output, err := cmd.CombinedOutput()

	result := struct {
		Log   string `json:"log"`
		Path  string `json:"path"`
		Count int    `json:"count"`
		Error string `json:"error,omitempty"`
	}{
		Log:   string(output),
		Path:  path,
		Count: count,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitCommitTool Git commit tool
type GitCommitTool struct{}

func (g *GitCommitTool) Name() string        { return "git_commit" }
func (g *GitCommitTool) Description() string { return "Create a git commit with a message" }
func (g *GitCommitTool) IsReadOnly() bool    { return false }
func (g *GitCommitTool) IsDestructive() bool { return false }

func (g *GitCommitTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"message": {"type": "string", "description": "Commit message"},
			"all": {"type": "boolean", "description": "Stage all modified/deleted files before commit"}
		},
		"required": ["message"]
	}`)
}

func (g *GitCommitTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path    string `json:"path"`
		Message string `json:"message"`
		All     bool   `json:"all"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "commit", "-m", params.Message}
	if params.All {
		args = append(args, "-a")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// NewDefaultRegistry creates a registry with the default tools
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()

	// Base tools
	registry.Register(&BashTool{})
	registry.Register(&FileReadTool{})
	registry.Register(&FileWriteTool{})
	registry.Register(&FileEditTool{})
	registry.Register(&GrepTool{})
	registry.Register(&GlobTool{})
	registry.Register(&TodoWriteTool{})
	registry.Register(&WebSearchTool{})
	registry.Register(&WebFetchTool{})
	registry.Register(&ThinkTool{})
	registry.Register(&FileDeleteTool{})
	registry.Register(&DirWriteTool{})
	registry.Register(&FileMoveTool{})
	registry.Register(&GitStatusTool{})
	registry.Register(&GitDiffTool{})
	registry.Register(&GitLogTool{})
	registry.Register(&GitCommitTool{})

	// Extended tools
	registry.Register(&DirectoryReadTool{})
	registry.Register(&NotebookEditTool{})
	registry.Register(&TaskGetTool{})
	registry.Register(&TaskCreateTool{})
	registry.Register(&TaskUpdateTool{})
	registry.Register(&TaskStopTool{})
	registry.Register(&TaskListTool{})
	registry.Register(&AgentTool{})

	// MCP tools
	registry.Register(&ListMcpResourcesTool{})
	registry.Register(&ReadMcpResourceTool{})
	registry.Register(&McpTool{})

	return registry
}
