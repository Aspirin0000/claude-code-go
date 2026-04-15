package tools

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

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

func TestGitRemoteTool(t *testing.T) {
	tool := &GitRemoteTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"action": "list",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	// No remotes yet, output should be empty or contain a message
}

func TestGitMergeTool(t *testing.T) {
	tool := &GitMergeTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("main"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()
	exec.Command("git", "-C", tmpDir, "checkout", "-b", "feature").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("feature"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "feature commit").Run()
	exec.Command("git", "-C", tmpDir, "checkout", "main").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"branch": "feature",
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

func TestGitShowTool(t *testing.T) {
	tool := &GitShowTool{}
	tmpDir := t.TempDir()
	exec.Command("git", "init", tmpDir).Run()
	exec.Command("git", "-C", tmpDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", tmpDir, "config", "user.name", "Test").Run()
	os.WriteFile(filepath.Join(tmpDir, "file.txt"), []byte("content"), 0644)
	exec.Command("git", "-C", tmpDir, "add", "file.txt").Run()
	exec.Command("git", "-C", tmpDir, "commit", "-m", "init").Run()

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"commit": "HEAD",
		"stat":   true,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if parsed.Output == "" {
		t.Error("expected non-empty output")
	}
}
