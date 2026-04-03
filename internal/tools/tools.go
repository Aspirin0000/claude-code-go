// Package tools 提供具体工具实现
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

// BashTool Bash 命令执行工具
type BashTool struct{}

func (b *BashTool) Name() string        { return "bash" }
func (b *BashTool) Description() string { return "执行 shell 命令" }
func (b *BashTool) IsReadOnly() bool    { return false }
func (b *BashTool) IsDestructive() bool { return true }

func (b *BashTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"command": {"type": "string", "description": "要执行的命令"},
			"timeout": {"type": "number", "description": "超时时间（毫秒）"},
			"description": {"type": "string", "description": "命令描述"}
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

// FileReadTool 文件读取工具
type FileReadTool struct{}

func (f *FileReadTool) Name() string        { return "file_read" }
func (f *FileReadTool) Description() string { return "读取文件内容" }
func (f *FileReadTool) IsReadOnly() bool    { return true }
func (f *FileReadTool) IsDestructive() bool { return false }

func (f *FileReadTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "文件绝对路径"},
			"offset": {"type": "number", "description": "起始行号"},
			"limit": {"type": "number", "description": "读取行数"}
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
		return nil, fmt.Errorf("读取文件失败: %w", err)
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

// FileWriteTool 文件写入工具
type FileWriteTool struct{}

func (f *FileWriteTool) Name() string        { return "file_write" }
func (f *FileWriteTool) Description() string { return "写入或创建新文件" }
func (f *FileWriteTool) IsReadOnly() bool    { return false }
func (f *FileWriteTool) IsDestructive() bool { return true }

func (f *FileWriteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "文件路径"},
			"content": {"type": "string", "description": "文件内容"}
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
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	return json.Marshal(result)
}

// FileEditTool 文件编辑工具 - 使用搜索替换
type FileEditTool struct{}

func (f *FileEditTool) Name() string        { return "file_edit" }
func (f *FileEditTool) Description() string { return "编辑文件内容（搜索并替换）" }
func (f *FileEditTool) IsReadOnly() bool    { return false }
func (f *FileEditTool) IsDestructive() bool { return true }

func (f *FileEditTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"file_path": {"type": "string", "description": "文件路径"},
			"old_string": {"type": "string", "description": "要替换的旧字符串"},
			"new_string": {"type": "string", "description": "新字符串"}
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
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	oldContent := string(content)
	if !strings.Contains(oldContent, params.OldString) {
		return nil, fmt.Errorf("未找到要替换的字符串")
	}

	newContent := strings.Replace(oldContent, params.OldString, params.NewString, 1)

	if err := os.WriteFile(params.FilePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	result := struct {
		Success bool `json:"success"`
	}{
		Success: true,
	}

	return json.Marshal(result)
}

// GrepTool Grep 搜索工具
type GrepTool struct{}

func (g *GrepTool) Name() string        { return "grep" }
func (g *GrepTool) Description() string { return "使用正则表达式搜索文件内容" }
func (g *GrepTool) IsReadOnly() bool    { return true }
func (g *GrepTool) IsDestructive() bool { return false }

func (g *GrepTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "正则表达式模式"},
			"path": {"type": "string", "description": "搜索路径"},
			"glob": {"type": "string", "description": "文件过滤模式"}
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

	// 使用 ripgrep
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
		// ripgrep 返回退出码 1 表示没有找到匹配，这是正常的
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			content = ""
		} else {
			return nil, fmt.Errorf("grep 执行失败: %w", err)
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

// GlobTool Glob 文件匹配工具
type GlobTool struct{}

func (g *GlobTool) Name() string        { return "glob" }
func (g *GlobTool) Description() string { return "按模式查找文件" }
func (g *GlobTool) IsReadOnly() bool    { return true }
func (g *GlobTool) IsDestructive() bool { return false }

func (g *GlobTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {"type": "string", "description": "glob 模式"},
			"path": {"type": "string", "description": "搜索目录"}
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

	// 使用 find 命令实现 glob 功能
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

// TodoWriteTool 待办事项工具
type TodoWriteTool struct{}

func (t *TodoWriteTool) Name() string        { return "todo_write" }
func (t *TodoWriteTool) Description() string { return "创建或更新任务列表" }
func (t *TodoWriteTool) IsReadOnly() bool    { return false }
func (t *TodoWriteTool) IsDestructive() bool { return false }

func (t *TodoWriteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"todos": {
				"type": "array",
				"description": "任务列表",
				"items": {
					"type": "object",
					"properties": {
						"id": {"type": "string", "description": "任务ID"},
						"content": {"type": "string", "description": "任务内容"},
						"status": {"type": "string", "description": "状态: in_progress, done, cancelled"},
						"priority": {"type": "string", "description": "优先级: high, medium, low"}
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

	// 简单返回成功，实际应用可以持久化到文件或数据库
	result := struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}{
		Success: true,
		Count:   len(params.Todos),
	}

	return json.Marshal(result)
}

// WebSearchTool Web 搜索工具
type WebSearchTool struct{}

func (w *WebSearchTool) Name() string        { return "web_search" }
func (w *WebSearchTool) Description() string { return "在网页上搜索信息" }
func (w *WebSearchTool) IsReadOnly() bool    { return true }
func (w *WebSearchTool) IsDestructive() bool { return false }

func (w *WebSearchTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"query": {"type": "string", "description": "搜索查询"}
		},
		"required": ["query"]
	}`)
}

func (w *WebSearchTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	// Web 搜索实现 - 需要配置搜索引擎 API 密钥
	result := struct {
		Results []string `json:"results"`
		Query   string   `json:"query"`
		Note    string   `json:"note"`
	}{
		Results: []string{},
		Query:   "",
		Note:    "Web 搜索功能需要配置搜索引擎 API。请使用 web_fetch 工具直接获取网页内容。",
	}

	return json.Marshal(result)
}

// WebFetchTool Web 获取工具
type WebFetchTool struct{}

func (w *WebFetchTool) Name() string        { return "web_fetch" }
func (w *WebFetchTool) Description() string { return "获取网页内容" }
func (w *WebFetchTool) IsReadOnly() bool    { return true }
func (w *WebFetchTool) IsDestructive() bool { return false }

func (w *WebFetchTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"url": {"type": "string", "description": "网页 URL"}
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

	// 使用 curl 获取网页内容
	cmd := exec.CommandContext(ctx, "curl", "-s", "-L", params.URL)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("获取网页失败: %w", err)
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

// NewDefaultRegistry 创建包含默认工具的注册表
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()

	// 基础工具
	registry.Register(&BashTool{})
	registry.Register(&FileReadTool{})
	registry.Register(&FileWriteTool{})
	registry.Register(&FileEditTool{})
	registry.Register(&GrepTool{})
	registry.Register(&GlobTool{})
	registry.Register(&TodoWriteTool{})
	registry.Register(&WebSearchTool{})
	registry.Register(&WebFetchTool{})

	// 扩展工具
	registry.Register(&NotebookEditTool{})
	registry.Register(&TaskGetTool{})
	registry.Register(&TaskCreateTool{})
	registry.Register(&TaskUpdateTool{})
	registry.Register(&TaskStopTool{})
	registry.Register(&TaskListTool{})
	registry.Register(&AgentTool{})

	return registry
}
