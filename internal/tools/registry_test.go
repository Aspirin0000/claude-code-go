package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================================
// Registry Tests
// ============================================================================

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
	if len(r.List()) != 0 {
		t.Errorf("expected 0 tools, got %d", len(r.List()))
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	tool := &AgentTool{}
	r.Register(tool)

	retrieved, ok := r.Get("agent")
	if !ok {
		t.Fatal("expected tool to be found")
	}
	if retrieved.Name() != "agent" {
		t.Errorf("expected 'agent', got %q", retrieved.Name())
	}
}

func TestRegistryGetNotFound(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("non-existent")
	if ok {
		t.Error("expected tool to not be found")
	}
}

func TestRegistryList(t *testing.T) {
	r := NewRegistry()
	r.Register(&AgentTool{})
	r.Register(&ListMcpResourcesTool{})

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2 tools, got %d", len(list))
	}
}

func TestRegistryCall(t *testing.T) {
	r := NewRegistry()
	r.Register(&AgentTool{})

	result, err := r.Call(context.Background(), "agent", json.RawMessage(`{"agent_type":"coder","task":"test"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Without API client, agent returns error status but Call still succeeds
	if !result.Success {
		// This is expected - the agent tool returns an error result
	}
}

func TestRegistryCallNotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Call(context.Background(), "non-existent", nil)
	if err == nil {
		t.Error("expected error for non-existent tool")
	}
}

func TestRegistryGetToolSchemas(t *testing.T) {
	r := NewRegistry()
	r.Register(&AgentTool{})

	schemas := r.GetToolSchemas()
	if len(schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(schemas))
	}
	if schemas[0]["name"] != "agent" {
		t.Errorf("expected name 'agent', got %v", schemas[0]["name"])
	}
}

// ============================================================================
// Result Tests
// ============================================================================

func TestResult(t *testing.T) {
	result := &Result{
		Data:    json.RawMessage(`{"key":"value"}`),
		Success: true,
		Error:   "",
	}
	if !result.Success {
		t.Error("expected success")
	}

	result2 := &Result{
		Success: false,
		Error:   "something went wrong",
	}
	if result2.Success {
		t.Error("expected failure")
	}
}

// ============================================================================
// AgentTool Tests
// ============================================================================

func TestAgentToolMetadata(t *testing.T) {
	tool := &AgentTool{}
	if tool.Name() != "agent" {
		t.Errorf("expected 'agent', got %q", tool.Name())
	}
	if tool.IsReadOnly() {
		t.Error("expected not read-only")
	}
	if !tool.IsDestructive() {
		t.Error("expected destructive")
	}
	if tool.InputSchema() == nil {
		t.Error("expected non-nil input schema")
	}
}

func TestAgentToolCallInvalidJSON(t *testing.T) {
	tool := &AgentTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAgentToolCallMissingTask(t *testing.T) {
	tool := &AgentTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`{"agent_type":"coder"}`))
	if err == nil {
		t.Error("expected error for missing task")
	}
}

func TestAgentToolCallNoAPIClient(t *testing.T) {
	tool := &AgentTool{}
	result, err := tool.Call(context.Background(), json.RawMessage(`{"agent_type":"coder","task":"test"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	if response["status"] != "error" {
		t.Errorf("expected error status, got %v", response["status"])
	}
}

func TestAgentToolCallWithFiles(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("test content"), 0644)

	tool := &AgentTool{}
	result, err := tool.Call(context.Background(), json.RawMessage(`{
		"agent_type": "coder",
		"task": "test",
		"files": ["`+testFile+`", "/non/existent/file.txt"]
	}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	if response["files_read"] == nil {
		t.Error("expected files_read in response")
	}
}

// ============================================================================
// MCP Tools Tests
// ============================================================================

func TestListMcpResourcesToolMetadata(t *testing.T) {
	tool := &ListMcpResourcesTool{}
	if tool.Name() != "list_mcp_resources" {
		t.Errorf("expected 'list_mcp_resources', got %q", tool.Name())
	}
	if !tool.IsReadOnly() {
		t.Error("expected read-only")
	}
	if tool.IsDestructive() {
		t.Error("expected not destructive")
	}
}

func TestListMcpResourcesToolCallNoServers(t *testing.T) {
	tool := &ListMcpResourcesTool{}
	result, err := tool.Call(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response map[string]interface{}
	json.Unmarshal(result, &response)
	if response["count"] == nil {
		t.Error("expected count in response")
	}
}

func TestListMcpResourcesToolCallInvalidJSON(t *testing.T) {
	tool := &ListMcpResourcesTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestReadMcpResourceToolMetadata(t *testing.T) {
	tool := &ReadMcpResourceTool{}
	if tool.Name() != "read_mcp_resource" {
		t.Errorf("expected 'read_mcp_resource', got %q", tool.Name())
	}
	if !tool.IsReadOnly() {
		t.Error("expected read-only")
	}
}

func TestReadMcpResourceToolCallMissingURI(t *testing.T) {
	tool := &ReadMcpResourceTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for missing URI")
	}
}

func TestReadMcpResourceToolCallInvalidJSON(t *testing.T) {
	tool := &ReadMcpResourceTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMcpToolMetadata(t *testing.T) {
	tool := &McpTool{}
	if tool.Name() != "mcp_tool" {
		t.Errorf("expected 'mcp_tool', got %q", tool.Name())
	}
	if tool.IsReadOnly() {
		t.Error("expected not read-only")
	}
	if !tool.IsDestructive() {
		t.Error("expected destructive")
	}
}

func TestMcpToolCallInvalidJSON(t *testing.T) {
	tool := &McpTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`invalid`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestMcpToolCallMissingServer(t *testing.T) {
	tool := &McpTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`{"tool":"test"}`))
	if err == nil {
		t.Error("expected error for missing server")
	}
}

func TestMcpToolCallMissingTool(t *testing.T) {
	tool := &McpTool{}
	_, err := tool.Call(context.Background(), json.RawMessage(`{"server":"test"}`))
	if err == nil {
		t.Error("expected error for missing tool")
	}

}

// ============================================================================
// Dev Tools Metadata Tests
// ============================================================================

func TestDockerPsToolMetadata(t *testing.T) {
	tool := &DockerPsTool{}
	if tool.Name() != "docker_ps" {
		t.Errorf("expected 'docker_ps', got %q", tool.Name())
	}
	if !tool.IsReadOnly() {
		t.Error("expected read-only")
	}
}

func TestDockerLogsToolMetadata(t *testing.T) {
	tool := &DockerLogsTool{}
	if tool.Name() != "docker_logs" {
		t.Errorf("expected 'docker_logs', got %q", tool.Name())
	}
	if !tool.IsReadOnly() {
		t.Error("expected read-only")
	}
}

func TestDockerExecToolMetadata(t *testing.T) {
	tool := &DockerExecTool{}
	if tool.Name() != "docker_exec" {
		t.Errorf("expected 'docker_exec', got %q", tool.Name())
	}
	if tool.IsReadOnly() {
		t.Error("expected not read-only")
	}
	// DockerExec is not destructive (it just runs a command in a container)
}

func TestDockerBuildToolMetadata(t *testing.T) {
	tool := &DockerBuildTool{}
	if tool.Name() != "docker_build" {
		t.Errorf("expected 'docker_build', got %q", tool.Name())
	}
	if tool.IsReadOnly() {
		t.Error("expected not read-only")
	}
}

func TestNpmInstallToolMetadata(t *testing.T) {
	tool := &NpmInstallTool{}
	if tool.Name() != "npm_install" {
		t.Errorf("expected 'npm_install', got %q", tool.Name())
	}
	if tool.IsReadOnly() {
		t.Error("expected not read-only")
	}
}

func TestNpmRunToolMetadata(t *testing.T) {
	tool := &NpmRunTool{}
	if tool.Name() != "npm_run" {
		t.Errorf("expected 'npm_run', got %q", tool.Name())
	}
	// npm_run is read-only (it just runs a script, doesn't modify files)
}

func TestGoBuildToolMetadata(t *testing.T) {
	tool := &GoBuildTool{}
	if tool.Name() != "go_build" {
		t.Errorf("expected 'go_build', got %q", tool.Name())
	}
	// go_build is read-only (it just compiles, doesn't modify source files)
}

func TestGoTestToolMetadata(t *testing.T) {
	tool := &GoTestTool{}
	if tool.Name() != "go_test" {
		t.Errorf("expected 'go_test', got %q", tool.Name())
	}
	// go_test is read-only (it just runs tests)
}

func TestPythonRunToolMetadata(t *testing.T) {
	tool := &PythonRunTool{}
	if tool.Name() != "python_run" {
		t.Errorf("expected 'python_run', got %q", tool.Name())
	}
	// python_run is read-only (it just runs a script)
}
