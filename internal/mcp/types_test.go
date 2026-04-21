package mcp

import (
	"testing"
)

// ============================================================================
// Config Scope Tests
// ============================================================================

func TestConfigScopeConstants(t *testing.T) {
	if ConfigScopeLocal != "local" {
		t.Errorf("expected 'local', got %q", ConfigScopeLocal)
	}
	if ConfigScopeUser != "user" {
		t.Errorf("expected 'user', got %q", ConfigScopeUser)
	}
	if ConfigScopeProject != "project" {
		t.Errorf("expected 'project', got %q", ConfigScopeProject)
	}
	if ConfigScopeGlobal != "global" {
		t.Errorf("expected 'global', got %q", ConfigScopeGlobal)
	}
	if ConfigScopeDynamic != "dynamic" {
		t.Errorf("expected 'dynamic', got %q", ConfigScopeDynamic)
	}
	if ConfigScopeEnterprise != "enterprise" {
		t.Errorf("expected 'enterprise', got %q", ConfigScopeEnterprise)
	}
	if ConfigScopeClaudeAI != "claudeai" {
		t.Errorf("expected 'claudeai', got %q", ConfigScopeClaudeAI)
	}
	if ConfigScopeManaged != "managed" {
		t.Errorf("expected 'managed', got %q", ConfigScopeManaged)
	}
}

// ============================================================================
// Transport Constants Tests
// ============================================================================

func TestTransportConstants(t *testing.T) {
	if TransportStdio != "stdio" {
		t.Errorf("expected 'stdio', got %q", TransportStdio)
	}
	if TransportSSE != "sse" {
		t.Errorf("expected 'sse', got %q", TransportSSE)
	}
	if TransportSSEIDE != "sse-ide" {
		t.Errorf("expected 'sse-ide', got %q", TransportSSEIDE)
	}
	if TransportHTTP != "http" {
		t.Errorf("expected 'http', got %q", TransportHTTP)
	}
	if TransportWebSocket != "ws" {
		t.Errorf("expected 'ws', got %q", TransportWebSocket)
	}
	if TransportSDK != "sdk" {
		t.Errorf("expected 'sdk', got %q", TransportSDK)
	}
}

// ============================================================================
// Server Config Tests
// ============================================================================

func TestMcpServerConfig(t *testing.T) {
	config := McpServerConfig{
		Type:    "stdio",
		Command: "node",
		Args:    []string{"server.js"},
		Env:     map[string]string{"NODE_ENV": "production"},
	}

	if config.Type != "stdio" {
		t.Errorf("expected 'stdio', got %q", config.Type)
	}
	if config.Command != "node" {
		t.Errorf("expected 'node', got %q", config.Command)
	}
}

func TestScopedMcpServerConfig(t *testing.T) {
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{
			Type:    "stdio",
			Command: "node",
		},
		Scope: ConfigScopeProject,
	}

	if config.Scope != ConfigScopeProject {
		t.Errorf("expected 'project', got %q", config.Scope)
	}
}

func TestMcpStdioServerConfig(t *testing.T) {
	config := McpStdioServerConfig{
		Command: "node",
		Args:    []string{"server.js"},
		Env:     map[string]string{"KEY": "value"},
	}

	if config.Command != "node" {
		t.Errorf("expected 'node', got %q", config.Command)
	}
}

func TestMcpSSEServerConfig(t *testing.T) {
	config := McpSSEServerConfig{
		Type: "sse",
		URL:  "http://localhost:3000",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	if config.Type != "sse" {
		t.Errorf("expected 'sse', got %q", config.Type)
	}
	if config.URL != "http://localhost:3000" {
		t.Errorf("expected 'http://localhost:3000', got %q", config.URL)
	}
}

func TestMcpHTTPServerConfig(t *testing.T) {
	config := McpHTTPServerConfig{
		Type: "http",
		URL:  "http://localhost:3000",
	}

	if config.Type != "http" {
		t.Errorf("expected 'http', got %q", config.Type)
	}
}

func TestMcpWebSocketServerConfig(t *testing.T) {
	config := McpWebSocketServerConfig{
		Type: "ws",
		URL:  "ws://localhost:3000",
	}

	if config.Type != "ws" {
		t.Errorf("expected 'ws', got %q", config.Type)
	}
}

func TestMcpSdkServerConfig(t *testing.T) {
	config := McpSdkServerConfig{
		Type: "sdk",
		Name: "test-sdk",
	}

	if config.Type != "sdk" {
		t.Errorf("expected 'sdk', got %q", config.Type)
	}
	if config.Name != "test-sdk" {
		t.Errorf("expected 'test-sdk', got %q", config.Name)
	}
}

func TestMcpClaudeAIProxyServerConfig(t *testing.T) {
	config := McpClaudeAIProxyServerConfig{
		Type: "claudeai-proxy",
		URL:  "https://claude.ai/proxy",
		ID:   "proxy-123",
	}

	if config.Type != "claudeai-proxy" {
		t.Errorf("expected 'claudeai-proxy', got %q", config.Type)
	}
}

func TestMcpOAuthConfig(t *testing.T) {
	clientID := "test-client"
	callbackPort := 8080
	config := McpOAuthConfig{
		ClientID:     &clientID,
		CallbackPort: &callbackPort,
	}

	if config.ClientID == nil || *config.ClientID != "test-client" {
		t.Error("expected client ID to be set")
	}
	if config.CallbackPort == nil || *config.CallbackPort != 8080 {
		t.Error("expected callback port to be 8080")
	}
}

// ============================================================================
// Server Connection Type Tests
// ============================================================================

func TestMCPServerConnectionTypeConstants(t *testing.T) {
	if MCPServerConnectionTypeConnected != "connected" {
		t.Errorf("expected 'connected', got %q", MCPServerConnectionTypeConnected)
	}
	if MCPServerConnectionTypeFailed != "failed" {
		t.Errorf("expected 'failed', got %q", MCPServerConnectionTypeFailed)
	}
	if MCPServerConnectionTypeNeedsAuth != "needs-auth" {
		t.Errorf("expected 'needs-auth', got %q", MCPServerConnectionTypeNeedsAuth)
	}
	if MCPServerConnectionTypePending != "pending" {
		t.Errorf("expected 'pending', got %q", MCPServerConnectionTypePending)
	}
	if MCPServerConnectionTypeDisabled != "disabled" {
		t.Errorf("expected 'disabled', got %q", MCPServerConnectionTypeDisabled)
	}
}

// ============================================================================
// MCPCliState Tests
// ============================================================================

func TestMCPCliState(t *testing.T) {
	state := MCPCliState{
		Clients:         make([]SerializedClient, 0),
		Configs:         make(map[string]ScopedMcpServerConfig),
		Tools:           make([]SerializedTool, 0),
		Resources:       make(map[string][]ServerResource),
		NormalizedNames: make(map[string]string),
	}

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

func TestSerializedClient(t *testing.T) {
	client := SerializedClient{
		Name:         "test-client",
		Type:         MCPServerConnectionTypeConnected,
		Capabilities: map[string]interface{}{},
	}

	if client.Name != "test-client" {
		t.Errorf("expected 'test-client', got %q", client.Name)
	}
}

func TestSerializedTool(t *testing.T) {
	isMcp := true
	tool := SerializedTool{
		Name:            "test-tool",
		Description:     "A test tool",
		InputJSONSchema: map[string]interface{}{"type": "object"},
		IsMcp:           &isMcp,
	}

	if tool.Name != "test-tool" {
		t.Errorf("expected 'test-tool', got %q", tool.Name)
	}
	if tool.IsMcp == nil || !*tool.IsMcp {
		t.Error("expected IsMcp to be true")
	}
}

// ============================================================================
// ServerResource Tests
// ============================================================================

func TestServerResource(t *testing.T) {
	res := ServerResource{
		Server:      "test-server",
		Name:        "test-resource",
		Description: strPtr("A test resource"),
		MIMEType:    strPtr("text/plain"),
	}

	if res.Server != "test-server" {
		t.Errorf("expected 'test-server', got %q", res.Server)
	}
	if res.Name != "test-resource" {
		t.Errorf("expected 'test-resource', got %q", res.Name)
	}
}

// ============================================================================
// McpJsonConfig Tests
// ============================================================================

func TestMcpJsonConfig(t *testing.T) {
	config := McpJsonConfig{
		McpServers: map[string]McpServerConfig{
			"server1": {Type: "stdio", Command: "node"},
		},
	}

	if len(config.McpServers) != 1 {
		t.Errorf("expected 1 server, got %d", len(config.McpServers))
	}
}


