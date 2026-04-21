package mcp

import (
	"testing"
	"time"
)

// ============================================================================
// ToolExecutor Tests
// ============================================================================

func TestNewToolExecutor(t *testing.T) {
	te := NewToolExecutor()
	if te == nil {
		t.Fatal("expected non-nil tool executor")
	}
}

func TestToolExecutorRegisterAndUnregister(t *testing.T) {
	te := NewToolExecutor()
	client := NewMCPClient("test-server", ScopedMcpServerConfig{})

	te.RegisterServer("test-server", client)

	serverName, exists := te.FindServerForTool("mcp__test-server__tool1")
	if !exists {
		t.Error("expected server to be found")
	}
	if serverName != "test-server" {
		t.Errorf("expected 'test-server', got %q", serverName)
	}

	te.UnregisterServer("test-server")

	_, exists = te.FindServerForTool("mcp__test-server__tool1")
	if exists {
		t.Error("expected server to not be found after unregister")
	}
}

func TestToolExecutorParseToolName(t *testing.T) {
	te := NewToolExecutor()

	tests := []struct {
		toolName       string
		expectedServer string
		expectedTool   string
		expectError    bool
	}{
		{"mcp__server1__tool1", "server1", "tool1", false},
		{"regular-tool", "", "regular-tool", false},
		{"mcp__invalid", "", "", true},
	}

	for _, test := range tests {
		server, tool, err := te.parseToolName(test.toolName)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for %q", test.toolName)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for %q: %v", test.toolName, err)
			continue
		}
		if server != test.expectedServer {
			t.Errorf("expected server %q, got %q for %q", test.expectedServer, server, test.toolName)
		}
		if tool != test.expectedTool {
			t.Errorf("expected tool %q, got %q for %q", test.expectedTool, tool, test.toolName)
		}
	}
}

func TestToolExecutorFindServerForToolInvalid(t *testing.T) {
	te := NewToolExecutor()

	_, exists := te.FindServerForTool("mcp__invalid")
	if exists {
		t.Error("expected false for invalid tool name")
	}
}

// ============================================================================
// MCPManager Tests
// ============================================================================

func TestNewMCPManager(t *testing.T) {
	manager := NewMCPManager()
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestMCPManagerGetStatus(t *testing.T) {
	manager := NewMCPManager()
	statuses := manager.GetStatus()
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(statuses))
	}
}

func TestMCPManagerGetServerStatusNotFound(t *testing.T) {
	manager := NewMCPManager()
	_, exists := manager.GetServerStatus("non-existent")
	if exists {
		t.Error("expected server to not exist")
	}
}

func TestMCPManagerGetConnectedServers(t *testing.T) {
	manager := NewMCPManager()
	connected := manager.GetConnectedServers()
	if len(connected) != 0 {
		t.Errorf("expected 0 connected servers, got %d", len(connected))
	}
}

func TestMCPManagerIsServerConnected(t *testing.T) {
	manager := NewMCPManager()
	if manager.IsServerConnected("non-existent") {
		t.Error("expected server to not be connected")
	}
}

func TestMCPManagerGetState(t *testing.T) {
	manager := NewMCPManager()
	state := manager.GetState()

	if state.Clients == nil {
		t.Error("expected non-nil clients")
	}
	if state.Configs == nil {
		t.Error("expected non-nil configs")
	}
	if state.Tools == nil {
		t.Error("expected non-nil tools")
	}
	if state.Resources == nil {
		t.Error("expected non-nil resources")
	}
	if state.NormalizedNames == nil {
		t.Error("expected non-nil normalized names")
	}
}

func TestMCPManagerGetMCPCache(t *testing.T) {
	manager := NewMCPManager()
	cache := manager.GetMCPCache()
	if cache == nil {
		t.Error("expected non-nil cache")
	}
}

func TestMCPManagerGetConnectionManager(t *testing.T) {
	manager := NewMCPManager()
	cm := manager.GetConnectionManager()
	if cm == nil {
		t.Error("expected non-nil connection manager")
	}
}

func TestMCPManagerGetExecutor(t *testing.T) {
	manager := NewMCPManager()
	executor := manager.GetExecutor()
	if executor == nil {
		t.Error("expected non-nil executor")
	}
}

func TestMCPManagerParseToolName(t *testing.T) {
	manager := NewMCPManager()

	tests := []struct {
		toolName       string
		expectedServer string
		expectedTool   string
		expectError    bool
	}{
		{"mcp__server1__tool1", "server1", "tool1", false},
		{"regular-tool", "", "regular-tool", false},
		{"mcp__invalid", "", "", true},
	}

	for _, test := range tests {
		server, tool, err := manager.parseToolName(test.toolName)
		if test.expectError {
			if err == nil {
				t.Errorf("expected error for %q", test.toolName)
			}
			continue
		}
		if err != nil {
			t.Errorf("unexpected error for %q: %v", test.toolName, err)
			continue
		}
		if server != test.expectedServer {
			t.Errorf("expected server %q, got %q for %q", test.expectedServer, server, test.toolName)
		}
		if tool != test.expectedTool {
			t.Errorf("expected tool %q, got %q for %q", test.expectedTool, tool, test.toolName)
		}
	}
}

func TestMCPManagerRegisterAuthProvider(t *testing.T) {
	manager := NewMCPManager()
	provider := &ClaudeAuthProvider{}

	manager.RegisterAuthProvider("test-server", provider)

	retrieved, exists := manager.GetAuthProvider("test-server")
	if !exists {
		t.Error("expected provider to exist")
	}
	if retrieved != provider {
		t.Error("expected same provider instance")
	}
}

func TestMCPManagerGetAuthProviderNotFound(t *testing.T) {
	manager := NewMCPManager()
	_, exists := manager.GetAuthProvider("non-existent")
	if exists {
		t.Error("expected provider to not exist")
	}
}

// ============================================================================
// Global Manager Tests
// ============================================================================

func TestGetGlobalMCPManager(t *testing.T) {
	manager := GetGlobalMCPManager()
	if manager == nil {
		t.Fatal("expected non-nil global manager")
	}

	// Should return same instance
	manager2 := GetGlobalMCPManager()
	if manager != manager2 {
		t.Error("expected same global manager instance")
	}
}

func TestSetGlobalMCPManager(t *testing.T) {
	original := GetGlobalMCPManager()
	newManager := NewMCPManager()

	SetGlobalMCPManager(newManager)

	current := GetGlobalMCPManager()
	if current != newManager {
		t.Error("expected global manager to be updated")
	}

	// Restore original
	SetGlobalMCPManager(original)
}

// ============================================================================
// ServerStatus Tests
// ============================================================================

func TestServerStatus(t *testing.T) {
	status := ServerStatus{
		Name:        "test-server",
		Type:        MCPServerConnectionTypeConnected,
		Connected:   true,
		LastError:   "",
		ToolCount:   5,
		Config:      ScopedMcpServerConfig{},
		LastChecked: time.Now(),
	}

	if status.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", status.Name)
	}
	if !status.Connected {
		t.Error("expected connected to be true")
	}
	if status.ToolCount != 5 {
		t.Errorf("expected 5 tools, got %d", status.ToolCount)
	}
}

// ============================================================================
// HandleRemoteAuthFailure Tests
// ============================================================================

func TestHandleRemoteAuthFailure(t *testing.T) {
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type: "sse",
			URL:  "http://localhost:3000",
		},
	}

	result := HandleRemoteAuthFailure("test-server", config, "sse")
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", result.Name)
	}
	if result.Type != MCPServerConnectionTypeNeedsAuth {
		t.Errorf("expected 'needs-auth', got %q", result.Type)
	}
}

// ============================================================================
// ExecutionContext Tests
// ============================================================================

func TestNewExecutionContext(t *testing.T) {
	ctx := NewExecutionContext(nil, "server1", "tool1")
	if ctx == nil {
		t.Fatal("expected non-nil execution context")
	}
	if ctx.ServerName != "server1" {
		t.Errorf("expected 'server1', got %q", ctx.ServerName)
	}
	if ctx.ToolName != "tool1" {
		t.Errorf("expected 'tool1', got %q", ctx.ToolName)
	}
	if ctx.Timeout != DefaultExecutionTimeout {
		t.Errorf("expected default timeout, got %v", ctx.Timeout)
	}
}

func TestExecutionContextWithTimeout(t *testing.T) {
	ctx := NewExecutionContext(nil, "server1", "tool1")
	newCtx := ctx.WithTimeout(5000)
	if newCtx.Timeout != 5000 {
		t.Errorf("expected timeout 5000, got %v", newCtx.Timeout)
	}
}

func TestExecutionContextWithProgressCallback(t *testing.T) {
	ctx := NewExecutionContext(nil, "server1", "tool1")
	called := false
	callback := func(progress ExecutionProgress) {
		called = true
	}
	newCtx := ctx.WithProgressCallback(callback)
	if newCtx.ProgressCallback == nil {
		t.Fatal("expected non-nil callback")
	}

	// Test reportProgress
	newCtx.reportProgress(ExecutionStateStarted, "test", 1, 1)
	if !called {
		t.Error("expected callback to be called")
	}
}

// ============================================================================
// ExecutionResult Tests
// ============================================================================

func TestExecutionResultIsSuccess(t *testing.T) {
	result := &ExecutionResult{
		Error:    nil,
		State:    ExecutionStateCompleted,
		Attempts: 1,
	}
	if !result.IsSuccess() {
		t.Error("expected success")
	}

	result2 := &ExecutionResult{
		Error:    &McpToolCallError{Message: "error"},
		State:    ExecutionStateFailed,
		Attempts: 1,
	}
	if result2.IsSuccess() {
		t.Error("expected not success")
	}
}

func TestExecutionResultIsRetried(t *testing.T) {
	result := &ExecutionResult{
		Attempts: 2,
	}
	if !result.IsRetried() {
		t.Error("expected retried")
	}

	result2 := &ExecutionResult{
		Attempts: 1,
	}
	if result2.IsRetried() {
		t.Error("expected not retried")
	}
}

// ============================================================================
// ExecutionState Tests
// ============================================================================

func TestExecutionStateString(t *testing.T) {
	tests := []struct {
		state    ExecutionState
		expected string
	}{
		{ExecutionStateStarted, "started"},
		{ExecutionStateInProgress, "in_progress"},
		{ExecutionStateCompleted, "completed"},
		{ExecutionStateFailed, "failed"},
		{ExecutionState(99), "unknown"},
	}

	for _, test := range tests {
		result := test.state.String()
		if result != test.expected {
			t.Errorf("expected %q, got %q", test.expected, result)
		}
	}
}

// ============================================================================
// ToolExecutorWithRetry Tests
// ============================================================================

func TestNewToolExecutorWithRetry(t *testing.T) {
	cm := NewConnectionManager()
	te := NewToolExecutorWithRetry(cm)
	if te == nil {
		t.Fatal("expected non-nil tool executor with retry")
	}
}

// ============================================================================
// LogMCPDebug Tests
// ============================================================================

func TestLogMCPDebug(t *testing.T) {
	// Just ensure it doesn't panic
	LogMCPDebug("test-server", "test message")
}
