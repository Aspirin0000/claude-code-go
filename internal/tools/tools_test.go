package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
		// Network-dependent test: allow timeout failures
		t.Logf("web search returned error (possibly network/timeout): %v", err)
		return
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

func TestGitDiffTool(t *testing.T) {
	tool := &GitDiffTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"path":   ".",
		"staged": false,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Diff string `json:"diff"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Path != "." {
		t.Errorf("expected path '.', got %s", parsed.Path)
	}
}

func TestGitLogTool(t *testing.T) {
	tool := &GitLogTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"path":  ".",
		"count": 5,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Log   string `json:"log"`
		Path  string `json:"path"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.Path != "." {
		t.Errorf("expected path '.', got %s", parsed.Path)
	}
	if parsed.Count != 5 {
		t.Errorf("expected count 5, got %d", parsed.Count)
	}
}

func TestGitCommitTool(t *testing.T) {
	tool := &GitCommitTool{}
	tmpDir := t.TempDir()

	// Initialize a git repo
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":    tmpDir,
		"message": "Initial commit",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got error: %s", parsed.Output)
	}
	if parsed.Path != tmpDir {
		t.Errorf("expected path %s, got %s", tmpDir, parsed.Path)
	}
}

func TestGitBranchTool(t *testing.T) {
	tool := &GitBranchTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("x"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"create": "feature-branch",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got: %s", parsed.Output)
	}
}

func TestGitCheckoutTool(t *testing.T) {
	tool := &GitCheckoutTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("x"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"branch": "new-branch",
		"create": true,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got: %s", parsed.Output)
	}
}

func TestGitAddTool(t *testing.T) {
	tool := &GitAddTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	os.WriteFile(filepath.Join(tmpDir, "new.txt"), []byte("new"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":  tmpDir,
		"files": []string{"new.txt"},
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got: %s", parsed.Output)
	}
}

func TestGitResetTool(t *testing.T) {
	tool := &GitResetTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("v1"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path": tmpDir,
		"mode": "mixed",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got: %s", parsed.Output)
	}
}

func TestGitStashTool(t *testing.T) {
	tool := &GitStashTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("stashme"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("changed"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":    tmpDir,
		"action":  "push",
		"message": "test stash",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if !parsed.Success {
		t.Errorf("expected success, got: %s", parsed.Output)
	}

	// Pop the stash
	input2, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"action": "pop",
	})
	result2, err2 := tool.Call(context.Background(), input2)
	if err2 != nil {
		t.Fatalf("unexpected error on pop: %v", err2)
	}
	if err := json.Unmarshal(result2, &parsed); err != nil {
		t.Fatalf("failed to unmarshal pop result: %v", err)
	}
	if !parsed.Success {
		t.Errorf("expected pop success, got: %s", parsed.Output)
	}
}

func TestHttpRequestToolGet(t *testing.T) {
	tool := &HttpRequestTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"url":     "https://httpbin.org/get",
		"method":  "GET",
		"timeout": 10000,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		// Allow network-dependent test to fail gracefully
		t.Logf("http request returned error (possibly network): %v", err)
		return
	}

	var parsed struct {
		StatusCode int    `json:"status_code"`
		Body       string `json:"body"`
		URL        string `json:"url"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", parsed.StatusCode)
	}
	if parsed.URL != "https://httpbin.org/get" {
		t.Errorf("unexpected URL: %s", parsed.URL)
	}
}

func TestHttpRequestToolPost(t *testing.T) {
	tool := &HttpRequestTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"url":     "https://httpbin.org/post",
		"method":  "POST",
		"headers": map[string]string{"Content-Type": "application/json"},
		"body":    `{"hello":"world"}`,
		"timeout": 10000,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Logf("http request returned error (possibly network): %v", err)
		return
	}

	var parsed struct {
		StatusCode int    `json:"status_code"`
		Body       string `json:"body"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if parsed.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", parsed.StatusCode)
	}
	if !strings.Contains(parsed.Body, `"hello"`) {
		t.Errorf("expected echoed body to contain hello, got: %s", parsed.Body)
	}
}
