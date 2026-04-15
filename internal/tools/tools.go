// Package tools provides concrete tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	_ = json.Unmarshal(input, &params)

	// Web search implementation - requires a search engine API key
	result := struct {
		Results []string `json:"results"`
		Query   string   `json:"query"`
		Note    string   `json:"note"`
	}{
		Results: []string{},
		Query:   params.Query,
		Note:    "Web search requires a search engine API key. Use the web_fetch tool to retrieve content from a known URL directly.",
	}

	return json.Marshal(result)
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

	// Extended tools
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
