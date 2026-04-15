package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDirectoryReadTool(t *testing.T) {
	tool := &DirectoryReadTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hello"), 0644)
	os.Mkdir(filepath.Join(tmpDir, "sub"), 0755)

	input, _ := json.Marshal(map[string]interface{}{
		"path":      tmpDir,
		"recursive": false,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Entries []struct {
			Name  string `json:"name"`
			IsDir bool   `json:"is_dir"`
		} `json:"entries"`
		Count int `json:"count"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Count != 2 {
		t.Fatalf("expected 2 entries, got %d", parsed.Count)
	}

	names := make(map[string]bool)
	for _, e := range parsed.Entries {
		names[e.Name] = e.IsDir
	}
	if _, ok := names["a.txt"]; !ok {
		t.Errorf("expected a.txt in entries")
	}
	if isDir, ok := names["sub"]; !ok || !isDir {
		t.Errorf("expected sub directory in entries")
	}
}

func TestDirectoryReadToolRecursive(t *testing.T) {
	tool := &DirectoryReadTool{}
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub")
	os.Mkdir(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "b.txt"), []byte("world"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":      tmpDir,
		"recursive": true,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Entries []struct {
			Name string `json:"name"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	found := false
	for _, e := range parsed.Entries {
		if e.Name == filepath.Join("sub", "b.txt") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected sub/b.txt in recursive entries")
	}
}

func TestThinkTool(t *testing.T) {
	tool := &ThinkTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"thought": "Step 1: analyze the problem",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Status  string `json:"status"`
		Thought string `json:"thought"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Status != "ok" {
		t.Errorf("expected status ok, got %s", parsed.Status)
	}
	if parsed.Thought != "Step 1: analyze the problem" {
		t.Errorf("unexpected thought: %s", parsed.Thought)
	}
}

func TestWebSearchTool(t *testing.T) {
	tool := &WebSearchTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"query": "golang",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Results []struct {
			Title string `json:"title"`
			URL   string `json:"url"`
		} `json:"results"`
		Query string `json:"query"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Query != "golang" {
		t.Errorf("expected query golang, got %s", parsed.Query)
	}
	if parsed.Count == 0 {
		t.Log("warning: no search results returned (may be due to network or rate limiting)")
	}
}

func TestWebFetchTool(t *testing.T) {
	tool := &WebFetchTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"url": "https://example.com",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Content string `json:"content"`
		URL     string `json:"url"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.URL != "https://example.com" {
		t.Errorf("expected url https://example.com, got %s", parsed.URL)
	}
	if parsed.Content == "" {
		t.Log("warning: empty content from example.com")
	}
}

func TestFileDeleteTool(t *testing.T) {
	tool := &FileDeleteTool{}
	tmpFile := filepath.Join(t.TempDir(), "to_delete.txt")
	os.WriteFile(tmpFile, []byte("delete me"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path": tmpFile,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success")
	}
	if _, statErr := os.Stat(tmpFile); !os.IsNotExist(statErr) {
		t.Errorf("expected file to be deleted")
	}
}

func TestDirWriteTool(t *testing.T) {
	tool := &DirWriteTool{}
	tmpDir := filepath.Join(t.TempDir(), "nested", "dir")

	input, _ := json.Marshal(map[string]interface{}{
		"path": tmpDir,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success")
	}
	if info, statErr := os.Stat(tmpDir); statErr != nil || !info.IsDir() {
		t.Errorf("expected directory to be created")
	}
}

func TestFileMoveTool(t *testing.T) {
	tool := &FileMoveTool{}
	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, "source.txt")
	dst := filepath.Join(tmpDir, "dest.txt")
	os.WriteFile(src, []byte("content"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"source":      src,
		"destination": dst,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success     bool   `json:"success"`
		Source      string `json:"source"`
		Destination string `json:"destination"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success")
	}
	if _, statErr := os.Stat(src); !os.IsNotExist(statErr) {
		t.Errorf("expected source file to be moved")
	}
	if _, statErr := os.Stat(dst); os.IsNotExist(statErr) {
		t.Errorf("expected destination file to exist")
	}
}

func TestGitStatusTool(t *testing.T) {
	tool := &GitStatusTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"path": ".",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Status string `json:"status"`
		Branch string `json:"branch"`
		Path   string `json:"path"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Path != "." {
		t.Errorf("expected path '.', got %s", parsed.Path)
	}
	// This test runs inside a git repo, so branch should be present
	if parsed.Branch == "" {
		t.Log("warning: branch not detected (may not be a git repo in test environment)")
	}
}
