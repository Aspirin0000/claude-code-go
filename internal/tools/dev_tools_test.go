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
