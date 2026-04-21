package mcp

import (
	"testing"
)

// ============================================================================
// ConnectionManager Tests
// ============================================================================

func TestNewConnectionManager(t *testing.T) {
	cm := NewConnectionManager()
	if cm == nil {
		t.Fatal("expected non-nil connection manager")
	}
}

func TestConnectionManagerGetClientNotFound(t *testing.T) {
	cm := NewConnectionManager()
	client, exists := cm.GetClient("non-existent")
	if exists {
		t.Error("expected client to not exist")
	}
	if client != nil {
		t.Error("expected nil client")
	}
}

func TestConnectionManagerGetStatus(t *testing.T) {
	cm := NewConnectionManager()
	status := cm.GetStatus("non-existent")
	if status != StatusDisconnected {
		t.Errorf("expected StatusDisconnected, got %v", status)
	}
}

func TestConnectionManagerListConnected(t *testing.T) {
	cm := NewConnectionManager()
	connected := cm.ListConnected()
	if len(connected) != 0 {
		t.Errorf("expected 0 connected servers, got %d", len(connected))
	}
}

func TestConnectionManagerDisconnectNotFound(t *testing.T) {
	cm := NewConnectionManager()
	err := cm.DisconnectServer("non-existent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConnectionStatusString(t *testing.T) {
	tests := []struct {
		status   ConnectionStatus
		expected string
	}{
		{StatusDisconnected, "disconnected"},
		{StatusConnecting, "connecting"},
		{StatusConnected, "connected"},
		{StatusFailed, "failed"},
		{StatusNeedsAuth, "needs-auth"},
		{ConnectionStatus(99), "unknown"},
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("expected %q, got %q", test.expected, result)
		}
	}
}

func TestConnectionManagerCreateTransportStdio(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type:    "stdio",
			Command: "echo",
			Args:    []string{"hello"},
		},
	}

	transport, err := cm.createTransport(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	// Check it's a StdioTransport
	stdioTransport, ok := transport.(*StdioTransport)
	if !ok {
		t.Fatal("expected StdioTransport")
	}
	if stdioTransport.command != "echo" {
		t.Errorf("expected command 'echo', got %q", stdioTransport.command)
	}
}

func TestConnectionManagerCreateTransportStdioEmptyCommand(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type:    "stdio",
			Command: "",
		},
	}

	_, err := cm.createTransport(config)
	if err == nil {
		t.Error("expected error for empty command")
	}
}

func TestConnectionManagerCreateTransportSSE(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type: "sse",
			URL:  "http://localhost:3000",
		},
	}

	transport, err := cm.createTransport(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	_, ok := transport.(*SSETransport)
	if !ok {
		t.Fatal("expected SSETransport")
	}
}

func TestConnectionManagerCreateTransportHTTP(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type: "http",
			URL:  "http://localhost:3000",
		},
	}

	transport, err := cm.createTransport(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	_, ok := transport.(*HTTPTransport)
	if !ok {
		t.Fatal("expected HTTPTransport")
	}
}

func TestConnectionManagerCreateTransportDefault(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type:    "",
			Command: "echo",
		},
	}

	transport, err := cm.createTransport(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}

	_, ok := transport.(*StdioTransport)
	if !ok {
		t.Fatal("expected StdioTransport for default type")
	}
}

func TestConnectionManagerCreateTransportUnsupported(t *testing.T) {
	cm := NewConnectionManager()
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type: "unsupported",
		},
	}

	_, err := cm.createTransport(config)
	if err == nil {
		t.Error("expected error for unsupported transport type")
	}
}

// ============================================================================
// BatchConnectResult Tests
// ============================================================================

func TestBatchConnectResult(t *testing.T) {
	result := BatchConnectResult{
		Name:       "test-server",
		Connection: nil,
		Error:      nil,
	}

	if result.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", result.Name)
	}
}

// ============================================================================
// ConnectedMCPServer Tests
// ============================================================================

func TestConnectedMCPServer(t *testing.T) {
	server := &ConnectedMCPServer{
		Name:         "test-server",
		Type:         MCPServerConnectionTypeConnected,
		Client:       nil,
		Capabilities: map[string]interface{}{},
		ServerInfo:   &ServerInfo{Name: "test", Version: "1.0"},
		Instructions: strPtr("test instructions"),
		Config:       ScopedMcpServerConfig{},
	}

	if server.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", server.Name)
	}
	if server.Type != MCPServerConnectionTypeConnected {
		t.Errorf("expected 'connected', got %q", server.Type)
	}
}

func TestFailedMCPServer(t *testing.T) {
	errMsg := "connection failed"
	server := &FailedMCPServer{
		Name:   "test-server",
		Type:   MCPServerConnectionTypeFailed,
		Config: ScopedMcpServerConfig{},
		Error:  &errMsg,
	}

	if server.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", server.Name)
	}
	if server.Error == nil {
		t.Error("expected non-nil error")
	}
}

func TestNeedsAuthMCPServer(t *testing.T) {
	server := &NeedsAuthMCPServer{
		Name:   "test-server",
		Type:   MCPServerConnectionTypeNeedsAuth,
		Config: ScopedMcpServerConfig{},
	}

	if server.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", server.Name)
	}
	if server.Type != MCPServerConnectionTypeNeedsAuth {
		t.Errorf("expected 'needs-auth', got %q", server.Type)
	}
}

func TestPendingMCPServer(t *testing.T) {
	server := &PendingMCPServer{
		Name:   "test-server",
		Type:   MCPServerConnectionTypePending,
		Config: ScopedMcpServerConfig{},
	}

	if server.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", server.Name)
	}
	if server.Type != MCPServerConnectionTypePending {
		t.Errorf("expected 'pending', got %q", server.Type)
	}
}

func TestDisabledMCPServer(t *testing.T) {
	server := &DisabledMCPServer{
		Name:   "test-server",
		Type:   MCPServerConnectionTypeDisabled,
		Config: ScopedMcpServerConfig{},
	}

	if server.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", server.Name)
	}
	if server.Type != MCPServerConnectionTypeDisabled {
		t.Errorf("expected 'disabled', got %q", server.Type)
	}
}

// ============================================================================
// ServerInfo Tests
// ============================================================================

func TestServerInfo(t *testing.T) {
	info := ServerInfo{
		Name:    "test-server",
		Version: "1.0.0",
	}

	if info.Name != "test-server" {
		t.Errorf("expected 'test-server', got %q", info.Name)
	}
	if info.Version != "1.0.0" {
		t.Errorf("expected '1.0.0', got %q", info.Version)
	}
}

// ============================================================================
// Helper function
// ============================================================================

func strPtr(s string) *string {
	return &s
}
