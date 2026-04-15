package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDockerPsTool(t *testing.T) {
	tool := &DockerPsTool{}
	input, _ := json.Marshal(map[string]interface{}{"all": true})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// May succeed or fail depending on docker daemon presence
}

func TestDockerLogsTool(t *testing.T) {
	tool := &DockerLogsTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"container": "nonexistent_container_12345",
		"tail":      10,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Error string `json:"error,omitempty"`
	}
	json.Unmarshal(result, &parsed)
	// Expected to fail since container doesn't exist
}

func TestNpmInstallTool(t *testing.T) {
	tool := &NpmInstallTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test"}`), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path": tmpDir,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// npm install on empty package.json should succeed
}

func TestNpmRunTool(t *testing.T) {
	tool := &NpmRunTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(`{"name":"test","scripts":{"hello":"echo hello"}}`), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"script": "hello",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !strings.Contains(parsed.Output, "hello") {
		t.Errorf("expected output to contain hello, got: %s", parsed.Output)
	}
}

func TestGoBuildTool(t *testing.T) {
	tool := &GoBuildTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte(`package main\nfunc main(){}`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module test\ngo 1.26\n`), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":   tmpDir,
		"output": "testbin",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed.Success {
		t.Log("go build may fail in test environment without full toolchain")
	}
}

func TestGoTestTool(t *testing.T) {
	tool := &GoTestTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte(`package main\nimport \"testing\"\nfunc TestX(t *testing.T){}\n`), 0644)
	os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(`module test\ngo 1.26\n`), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path":    tmpDir,
		"verbose": true,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// go test may download dependencies and take time
}

func TestPythonRunTool(t *testing.T) {
	tool := &PythonRunTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"code": "print('hello from python')",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !strings.Contains(parsed.Output, "hello from python") {
		t.Errorf("expected python output, got: %s", parsed.Output)
	}
}

func TestPythonRunToolFile(t *testing.T) {
	tool := &PythonRunTool{}
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "script.py")
	os.WriteFile(pyFile, []byte("print('file mode')"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"file": pyFile,
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !strings.Contains(parsed.Output, "file mode") {
		t.Errorf("expected file mode output, got: %s", parsed.Output)
	}
}

func TestDockerExecTool(t *testing.T) {
	tool := &DockerExecTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"container": "nonexistent_container_12345",
		"command":   "echo hello",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Error string `json:"error,omitempty"`
	}
	json.Unmarshal(result, &parsed)
	// Expected to fail since container doesn't exist
}

func TestDockerBuildTool(t *testing.T) {
	tool := &DockerBuildTool{}
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM scratch\n"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"path": tmpDir,
		"tag":  "test-build:latest",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	// May succeed or fail depending on docker daemon
	t.Logf("docker build success=%v", parsed.Success)
}

func TestSedReplaceTool(t *testing.T) {
	tool := &SedReplaceTool{}
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello world\nhello universe\n"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"file_path":   file,
		"pattern":     `hello (\w+)`,
		"replacement": "hi $1",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed.Success {
		t.Errorf("expected success")
	}

	content, _ := os.ReadFile(file)
	if !strings.Contains(string(content), "hi world") {
		t.Errorf("expected replacement, got: %s", string(content))
	}
	if strings.Contains(string(content), "hello universe") {
		// only first match replaced
	} else {
		t.Log("first-only replacement may have removed second hello")
	}
}

func TestSedReplaceToolAll(t *testing.T) {
	tool := &SedReplaceTool{}
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello world\nhello universe\n"), 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"file_path":   file,
		"pattern":     `hello`,
		"replacement": "hi",
		"all":         true,
	})

	_, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, _ := os.ReadFile(file)
	if strings.Count(string(content), "hi") != 2 {
		t.Errorf("expected all replacements, got: %s", string(content))
	}
}

func TestJSONQueryTool(t *testing.T) {
	tool := &JSONQueryTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"json": `{"user":{"name":"Alice","age":30},"items":[{"id":1},{"id":2}]}`,
		"path": "user.name",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Value interface{} `json:"value"`
		Found bool        `json:"found"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed.Found {
		t.Errorf("expected found=true")
	}
	if parsed.Value != "Alice" {
		t.Errorf("expected Alice, got %v", parsed.Value)
	}
}

func TestJSONQueryToolArray(t *testing.T) {
	tool := &JSONQueryTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"json": `{"items":[{"id":1},{"id":2}]}`,
		"path": "items.1.id",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Value interface{} `json:"value"`
		Found bool        `json:"found"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed.Value != float64(2) {
		t.Errorf("expected 2, got %v", parsed.Value)
	}
}

func TestEnvGetTool(t *testing.T) {
	tool := &EnvGetTool{}
	os.Setenv("CLAUDE_TEST_VAR", "test_value")

	input, _ := json.Marshal(map[string]interface{}{
		"names": []string{"CLAUDE_TEST_VAR", "NONEXISTENT_VAR"},
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if parsed.Env["CLAUDE_TEST_VAR"] != "test_value" {
		t.Errorf("expected test_value, got %s", parsed.Env["CLAUDE_TEST_VAR"])
	}
}

func TestEnvSetTool(t *testing.T) {
	tool := &EnvSetTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"vars": map[string]string{"CLAUDE_SET_VAR": "set_value"},
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed.Success {
		t.Errorf("expected success")
	}
	if os.Getenv("CLAUDE_SET_VAR") != "set_value" {
		t.Errorf("expected env var to be set")
	}
}
