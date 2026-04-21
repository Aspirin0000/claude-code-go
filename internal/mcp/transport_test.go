package mcp

import (
	"testing"
	"time"
)

// ============================================================================
// TransportBase Tests
// ============================================================================

func TestTransportBaseSetOnMessage(t *testing.T) {
	tb := &TransportBase{}
	called := false
	tb.SetOnMessage(func(msg JSONRPCMessage) {
		called = true
	})

	tb.triggerOnMessage(JSONRPCMessage{})
	if !called {
		t.Error("expected onMessage to be called")
	}
}

func TestTransportBaseSetOnClose(t *testing.T) {
	tb := &TransportBase{}
	called := false
	tb.SetOnClose(func() {
		called = true
	})

	tb.triggerOnClose()
	if !called {
		t.Error("expected onClose to be called")
	}
}

func TestTransportBaseSetOnError(t *testing.T) {
	tb := &TransportBase{}
	var receivedErr error
	tb.SetOnError(func(err error) {
		receivedErr = err
	})

	testErr := &McpAuthError{ServerName: "test", Message: "error"}
	tb.triggerOnError(testErr)
	if receivedErr == nil {
		t.Error("expected onError to be called")
	}
}

func TestTransportBaseIsClosed(t *testing.T) {
	tb := &TransportBase{}
	if tb.IsClosed() {
		t.Error("expected not closed initially")
	}

	tb.MarkClosed()
	if !tb.IsClosed() {
		t.Error("expected closed after MarkClosed")
	}
}

// ============================================================================
// HTTPTransport Tests
// ============================================================================

func TestNewHTTPTransport(t *testing.T) {
	transport, err := NewHTTPTransport("http://localhost:3000", map[string]string{
		"Authorization": "Bearer token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestNewHTTPTransportInvalidURL(t *testing.T) {
	_, err := NewHTTPTransport("://invalid-url", nil)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestNewHTTPTransportWithAuthToken(t *testing.T) {
	transport, err := NewHTTPTransport("http://localhost:3000", nil, "my-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport.authToken != "my-token" {
		t.Errorf("expected auth token 'my-token', got %q", transport.authToken)
	}
}

func TestHTTPTransportClose(t *testing.T) {
	transport, _ := NewHTTPTransport("http://localhost:3000", nil)
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !transport.IsClosed() {
		t.Error("expected transport to be closed")
	}
}

func TestHTTPTransportCloseAlreadyClosed(t *testing.T) {
	transport, _ := NewHTTPTransport("http://localhost:3000", nil)
	transport.Close()
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestHTTPTransportSendClosed(t *testing.T) {
	transport, _ := NewHTTPTransport("http://localhost:3000", nil)
	transport.Close()
	err := transport.Send(JSONRPCMessage{})
	if err == nil {
		t.Error("expected error when sending on closed transport")
	}
}

// ============================================================================
// StdioTransport Tests
// ============================================================================

func TestNewStdioTransport(t *testing.T) {
	transport := NewStdioTransport("echo", []string{"hello"}, map[string]string{"KEY": "value"})
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.command != "echo" {
		t.Errorf("expected command 'echo', got %q", transport.command)
	}
	if len(transport.args) != 1 || transport.args[0] != "hello" {
		t.Errorf("expected args ['hello'], got %v", transport.args)
	}
}

func TestStdioTransportGetPID(t *testing.T) {
	transport := NewStdioTransport("echo", []string{"hello"}, nil)
	if transport.GetPID() != 0 {
		t.Errorf("expected PID 0 before connect, got %d", transport.GetPID())
	}
}

func TestStdioTransportCloseNotConnected(t *testing.T) {
	transport := NewStdioTransport("echo", []string{"hello"}, nil)
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStdioTransportSendClosed(t *testing.T) {
	transport := NewStdioTransport("echo", []string{"hello"}, nil)
	transport.Close()
	err := transport.Send(JSONRPCMessage{})
	if err == nil {
		t.Error("expected error when sending on closed transport")
	}
}

// ============================================================================
// SSETransport Tests
// ============================================================================

func TestNewSSETransport(t *testing.T) {
	transport := NewSSETransport("http://localhost:3000", map[string]string{
		"Authorization": "Bearer token",
	})
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.serverURL != "http://localhost:3000" {
		t.Errorf("expected URL 'http://localhost:3000', got %q", transport.serverURL)
	}
}

func TestSSETransportClose(t *testing.T) {
	transport := NewSSETransport("http://localhost:3000", nil)
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !transport.IsClosed() {
		t.Error("expected transport to be closed")
	}
}

func TestSSETransportCloseAlreadyClosed(t *testing.T) {
	transport := NewSSETransport("http://localhost:3000", nil)
	transport.Close()
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestSSETransportSendClosed(t *testing.T) {
	transport := NewSSETransport("http://localhost:3000", nil)
	transport.Close()
	err := transport.Send(JSONRPCMessage{})
	if err == nil {
		t.Error("expected error when sending on closed transport")
	}
}

// ============================================================================
// TransportType Tests
// ============================================================================

func TestTransportTypeConstants(t *testing.T) {
	if TransportTypeStdio != "stdio" {
		t.Errorf("expected 'stdio', got %q", TransportTypeStdio)
	}
	if TransportTypeSSE != "sse" {
		t.Errorf("expected 'sse', got %q", TransportTypeSSE)
	}
	if TransportTypeHTTP != "http" {
		t.Errorf("expected 'http', got %q", TransportTypeHTTP)
	}
	if TransportTypeWebSocket != "ws" {
		t.Errorf("expected 'ws', got %q", TransportTypeWebSocket)
	}
	if TransportTypeClaudeAIProxy != "claudeai-proxy" {
		t.Errorf("expected 'claudeai-proxy', got %q", TransportTypeClaudeAIProxy)
	}
}

// ============================================================================
// CreateTransport Tests
// ============================================================================

func TestCreateTransportStdio(t *testing.T) {
	_, err := CreateTransport(TransportTypeStdio, "", nil)
	if err == nil {
		t.Error("expected error for stdio transport")
	}
}

func TestCreateTransportSSE(t *testing.T) {
	transport, err := CreateTransport(TransportTypeSSE, "http://localhost:3000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestCreateTransportHTTP(t *testing.T) {
	transport, err := CreateTransport(TransportTypeHTTP, "http://localhost:3000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestCreateTransportClaudeAIProxy(t *testing.T) {
	transport, err := CreateTransport(TransportTypeClaudeAIProxy, "http://localhost:3000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestCreateTransportWebSocket(t *testing.T) {
	transport, err := CreateTransport(TransportTypeWebSocket, "ws://localhost:3000", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
}

func TestCreateTransportUnsupported(t *testing.T) {
	_, err := CreateTransport("unsupported", "", nil)
	if err == nil {
		t.Error("expected error for unsupported transport type")
	}
}

// ============================================================================
// CreateStdioTransportFromConfig Tests
// ============================================================================

func TestCreateStdioTransportFromConfig(t *testing.T) {
	config := McpStdioServerConfig{
		Command: "node",
		Args:    []string{"server.js"},
		Env:     map[string]string{"NODE_ENV": "production"},
	}

	transport := CreateStdioTransportFromConfig(config)
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.command != "node" {
		t.Errorf("expected command 'node', got %q", transport.command)
	}
}

func TestCreateStdioTransportFromConfigWithShellPrefix(t *testing.T) {
	// This test would need to set environment variable, skip for now
	// as it would affect other tests
	t.Skip("Skipping test that requires environment variable modification")
}

// ============================================================================
// WebSocketTransport Tests
// ============================================================================

func TestNewWebSocketTransport(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", map[string]string{
		"Authorization": "Bearer token",
	})
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.serverURL != "ws://localhost:3000" {
		t.Errorf("expected URL 'ws://localhost:3000', got %q", transport.serverURL)
	}
	if transport.maxReconnectAttempts != 3 {
		t.Errorf("expected 3 reconnect attempts, got %d", transport.maxReconnectAttempts)
	}
}

func TestNewWebSocketTransportWithConfig(t *testing.T) {
	config := WebSocketConfig{
		ServerURL:            "wss://example.com",
		Headers:              map[string]string{"X-Custom": "value"},
		MaxReconnectAttempts: 5,
		ReconnectDelay:       1000,
		PingInterval:         15000,
		PongTimeout:          5000,
	}

	transport := NewWebSocketTransportWithConfig(config)
	if transport == nil {
		t.Fatal("expected non-nil transport")
	}
	if transport.maxReconnectAttempts != 5 {
		t.Errorf("expected 5 reconnect attempts, got %d", transport.maxReconnectAttempts)
	}
}

func TestNewWebSocketTransportWithConfigDefaults(t *testing.T) {
	config := WebSocketConfig{
		ServerURL: "ws://localhost:3000",
	}

	transport := NewWebSocketTransportWithConfig(config)
	if transport.maxReconnectAttempts != 3 {
		t.Errorf("expected default 3 reconnect attempts, got %d", transport.maxReconnectAttempts)
	}
	if transport.reconnectDelay != 2*time.Second {
		t.Errorf("expected default 2s reconnect delay, got %v", transport.reconnectDelay)
	}
}

func TestWebSocketTransportIsConnected(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", nil)
	if transport.IsConnected() {
		t.Error("expected not connected initially")
	}
}

func TestWebSocketTransportGetReconnectCount(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", nil)
	if transport.GetReconnectCount() != 0 {
		t.Errorf("expected 0 reconnect count initially, got %d", transport.GetReconnectCount())
	}
}

func TestWebSocketTransportClose(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", nil)
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !transport.IsClosed() {
		t.Error("expected transport to be closed")
	}
}

func TestWebSocketTransportCloseAlreadyClosed(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", nil)
	transport.Close()
	err := transport.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestWebSocketTransportSendClosed(t *testing.T) {
	transport := NewWebSocketTransport("ws://localhost:3000", nil)
	transport.Close()
	err := transport.Send(JSONRPCMessage{})
	if err == nil {
		t.Error("expected error when sending on closed transport")
	}
}
