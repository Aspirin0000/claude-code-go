package mcp

import (
	"encoding/json"
	"testing"
	"time"
)

// ============================================================================
// Error Types Tests
// ============================================================================

func TestMcpAuthError(t *testing.T) {
	err := &McpAuthError{
		ServerName: "test-server",
		Message:    "auth failed",
	}
	expected := "MCP server 'test-server' authentication error: auth failed"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestMcpSessionExpiredError(t *testing.T) {
	err := &McpSessionExpiredError{ServerName: "test-server"}
	expected := "MCP server 'test-server' session expired"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestMcpToolCallError(t *testing.T) {
	err := &McpToolCallError{
		Message:          "tool failed",
		TelemetryMessage: "telemetry",
		McpMeta:          map[string]interface{}{"key": "value"},
	}
	if err.Error() != "tool failed" {
		t.Errorf("expected 'tool failed', got %q", err.Error())
	}
}

func TestIsMcpSessionExpiredError(t *testing.T) {
	if IsMcpSessionExpiredError(nil) {
		t.Error("expected false for nil error")
	}

	err1 := &McpSessionExpiredError{ServerName: "test"}
	if IsMcpSessionExpiredError(err1) {
		t.Error("expected false for McpSessionExpiredError type")
	}

	err2 := &McpToolCallError{Message: `{"code":-32001}`}
	if !IsMcpSessionExpiredError(err2) {
		t.Error("expected true for error with code -32001")
	}

	err3 := &McpToolCallError{Message: `some error with "code": -32001`}
	if !IsMcpSessionExpiredError(err3) {
		t.Error("expected true for error with code -32001 (spaced)")
	}
}

// ============================================================================
// Environment Variable Functions Tests
// ============================================================================

func TestGetMcpToolTimeoutMs(t *testing.T) {
	// Test default
	timeout := GetMcpToolTimeoutMs()
	if timeout != DefaultMcpToolTimeoutMs {
		t.Errorf("expected default %d, got %d", DefaultMcpToolTimeoutMs, timeout)
	}
}

func TestGetConnectionTimeoutMs(t *testing.T) {
	// Test default
	timeout := GetConnectionTimeoutMs()
	if timeout != 30000 {
		t.Errorf("expected default 30000, got %d", timeout)
	}
}

func TestGetMcpServerConnectionBatchSize(t *testing.T) {
	// Test default
	size := GetMcpServerConnectionBatchSize()
	if size != 3 {
		t.Errorf("expected default 3, got %d", size)
	}
}

func TestGetRemoteMcpServerConnectionBatchSize(t *testing.T) {
	// Test default
	size := GetRemoteMcpServerConnectionBatchSize()
	if size != 20 {
		t.Errorf("expected default 20, got %d", size)
	}
}

// ============================================================================
// Server Type Tests
// ============================================================================

func TestIsLocalMcpServer(t *testing.T) {
	tests := []struct {
		config   ScopedMcpServerConfig
		expected bool
	}{
		{ScopedMcpServerConfig{McpServerConfig: McpServerConfig{Type: ""}}, true},
		{ScopedMcpServerConfig{McpServerConfig: McpServerConfig{Type: "stdio"}}, true},
		{ScopedMcpServerConfig{McpServerConfig: McpServerConfig{Type: "sdk"}}, true},
		{ScopedMcpServerConfig{McpServerConfig: McpServerConfig{Type: "sse"}}, false},
		{ScopedMcpServerConfig{McpServerConfig: McpServerConfig{Type: "http"}}, false},
	}

	for _, test := range tests {
		result := IsLocalMcpServer(test.config)
		if result != test.expected {
			t.Errorf("IsLocalMcpServer(%q) = %v, expected %v", test.config.Type, result, test.expected)
		}
	}
}

func TestIsIncludedMcpTool(t *testing.T) {
	// Non-IDE tools should always be included
	if !IsIncludedMcpTool("regular-tool") {
		t.Error("expected regular tool to be included")
	}

	// Allowed IDE tools
	if !IsIncludedMcpTool("mcp__ide__executeCode") {
		t.Error("expected executeCode to be included")
	}
	if !IsIncludedMcpTool("mcp__ide__getDiagnostics") {
		t.Error("expected getDiagnostics to be included")
	}

	// Forbidden IDE tools
	if IsIncludedMcpTool("mcp__ide__forbidden") {
		t.Error("expected forbidden IDE tool to be excluded")
	}
}

func TestIsImageMimeType(t *testing.T) {
	// Supported types
	if !IsImageMimeType(ImageMimeTypeJPEG) {
		t.Error("expected JPEG to be supported")
	}
	if !IsImageMimeType(ImageMimeTypePNG) {
		t.Error("expected PNG to be supported")
	}
	if !IsImageMimeType(ImageMimeTypeGIF) {
		t.Error("expected GIF to be supported")
	}
	if !IsImageMimeType(ImageMimeTypeWebP) {
		t.Error("expected WebP to be supported")
	}

	// Unsupported type
	if IsImageMimeType("image/svg+xml") {
		t.Error("expected SVG to be unsupported")
	}
}

// ============================================================================
// JSON-RPC Tests
// ============================================================================

func TestJSONRPCError(t *testing.T) {
	err := &JSONRPCError{
		Code:    -32600,
		Message: "Invalid Request",
	}
	expected := "JSON-RPC error -32600: Invalid Request"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

// ============================================================================
// Helper Functions Tests
// ============================================================================

func TestGetLoggingSafeMcpBaseUrl(t *testing.T) {
	// Empty URL
	config := ScopedMcpServerConfig{McpServerConfig: McpServerConfig{URL: ""}}
	result := GetLoggingSafeMcpBaseUrl(config)
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}

	// URL without query
	config2 := ScopedMcpServerConfig{McpServerConfig: McpServerConfig{URL: "https://example.com/api"}}
	result2 := GetLoggingSafeMcpBaseUrl(config2)
	if result2 != "https://example.com/api" {
		t.Errorf("expected 'https://example.com/api', got %q", result2)
	}

	// URL with query (should strip query)
	config3 := ScopedMcpServerConfig{McpServerConfig: McpServerConfig{URL: "https://example.com/api?key=secret"}}
	result3 := GetLoggingSafeMcpBaseUrl(config3)
	if result3 != "https://example.com/api" {
		t.Errorf("expected 'https://example.com/api', got %q", result3)
	}
}

func TestMcpBaseUrlAnalytics(t *testing.T) {
	// Empty URL
	config := ScopedMcpServerConfig{McpServerConfig: McpServerConfig{URL: ""}}
	result := McpBaseUrlAnalytics(config)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}

	// Valid URL
	config2 := ScopedMcpServerConfig{McpServerConfig: McpServerConfig{URL: "https://example.com/api"}}
	result2 := McpBaseUrlAnalytics(config2)
	if result2 == nil {
		t.Fatal("expected non-nil analytics")
	}
	if result2["mcpServerBaseUrl"] != "https://example.com/api" {
		t.Errorf("expected 'https://example.com/api', got %q", result2["mcpServerBaseUrl"])
	}
}

func TestGetServerCacheKey(t *testing.T) {
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{Type: "stdio", Command: "node"},
		Scope:           ConfigScopeProject,
	}
	key := GetServerCacheKey("test-server", config)
	if key == "" {
		t.Error("expected non-empty cache key")
	}
	if !contains(key, "test-server") {
		t.Error("expected cache key to contain server name")
	}
}

func TestConnectionStats(t *testing.T) {
	stats := ConnectionStats{
		TotalServers: 5,
		StdioCount:   2,
		SseCount:     1,
		HttpCount:    1,
		SseIdeCount:  1,
		WsIdeCount:   0,
	}

	if stats.TotalServers != 5 {
		t.Errorf("expected 5 total servers, got %d", stats.TotalServers)
	}
}

// ============================================================================
// MCPClient Tests
// ============================================================================

func TestNewMCPClient(t *testing.T) {
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{Type: "stdio", Command: "node"},
		Scope:           ConfigScopeProject,
	}
	client := NewMCPClient("test-server", config)

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.name != "test-server" {
		t.Errorf("expected name 'test-server', got %q", client.name)
	}
	if client.connected {
		t.Error("expected client to not be connected initially")
	}
}

func TestMCPClientGetName(t *testing.T) {
	client := NewMCPClient("my-server", ScopedMcpServerConfig{})
	if client.GetName() != "my-server" {
		t.Errorf("expected 'my-server', got %q", client.GetName())
	}
}

func TestMCPClientGetConfig(t *testing.T) {
	config := ScopedMcpServerConfig{
		McpServerConfig: McpServerConfig{Type: "stdio"},
		Scope:           ConfigScopeProject,
	}
	client := NewMCPClient("test", config)
	result := client.GetConfig()
	if result.Scope != ConfigScopeProject {
		t.Errorf("expected scope 'project', got %q", result.Scope)
	}
}

func TestMCPClientIsConnected(t *testing.T) {
	client := NewMCPClient("test", ScopedMcpServerConfig{})
	if client.IsConnected() {
		t.Error("expected client to not be connected")
	}
}

func TestMCPClientSetOnClose(t *testing.T) {
	client := NewMCPClient("test", ScopedMcpServerConfig{})
	called := false
	client.SetOnClose(func() {
		called = true
	})

	// Simulate close callback
	if client.onClose != nil {
		client.onClose()
	}

	if !called {
		t.Error("expected onClose callback to be called")
	}
}

func TestMCPClientSetOnError(t *testing.T) {
	client := NewMCPClient("test", ScopedMcpServerConfig{})
	var receivedErr error
	client.SetOnError(func(err error) {
		receivedErr = err
	})

	// Simulate error callback
	testErr := &McpAuthError{ServerName: "test", Message: "error"}
	if client.onError != nil {
		client.onError(testErr)
	}

	if receivedErr == nil {
		t.Error("expected onError callback to be called")
	}
}

// ============================================================================
// Auth Cache Tests
// ============================================================================

func TestMcpAuthCache(t *testing.T) {
	cache := &McpAuthCache{
		data:      make(map[string]McpAuthCacheEntry),
		cachePath: "",
	}

	// Test IsCached with non-existent entry
	if cache.IsCached("test-server") {
		t.Error("expected false for non-cached server")
	}

	// Test SetEntry and IsCached
	cache.SetEntry("test-server")
	if !cache.IsCached("test-server") {
		t.Error("expected true after setting entry")
	}

	// Test Clear
	err := cache.Clear()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cache.IsCached("test-server") {
		t.Error("expected false after clear")
	}
}

func TestMcpAuthCacheLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := tmpDir + "/test-cache.json"
	cache := &McpAuthCache{
		data:      make(map[string]McpAuthCacheEntry),
		cachePath: cachePath,
	}

	// Test Load with non-existent file
	err := cache.Load()
	if err != nil {
		t.Errorf("unexpected error loading non-existent cache: %v", err)
	}

	// Set entry and save synchronously
	cache.mu.Lock()
	cache.data["server1"] = McpAuthCacheEntry{
		Timestamp: time.Now().UnixMilli(),
	}
	cache.mu.Unlock()

	// Test Save
	err = cache.Save()
	if err != nil {
		t.Errorf("unexpected error saving cache: %v", err)
	}

	// Create new cache and load
	cache2 := &McpAuthCache{
		data:      make(map[string]McpAuthCacheEntry),
		cachePath: cachePath,
	}
	err = cache2.Load()
	if err != nil {
		t.Errorf("unexpected error loading cache: %v", err)
	}

	if !cache2.IsCached("server1") {
		t.Error("expected server1 to be cached after load")
	}
}

func TestGetGlobalMcpAuthCache(t *testing.T) {
	cache := GetGlobalMcpAuthCache()
	if cache == nil {
		t.Fatal("expected non-nil global auth cache")
	}
}

func TestGetMcpAuthCachePath(t *testing.T) {
	path := GetMcpAuthCachePath()
	if path == "" {
		t.Error("expected non-empty auth cache path")
	}
}

// ============================================================================
// Protocol Types Tests
// ============================================================================

func TestToolInfo(t *testing.T) {
	tool := ToolInfo{
		Name:        "test-tool",
		Description: "A test tool",
		InputSchema: map[string]interface{}{
			"type": "object",
		},
	}

	if tool.Name != "test-tool" {
		t.Errorf("expected 'test-tool', got %q", tool.Name)
	}
}

func TestResourceInfo(t *testing.T) {
	res := ResourceInfo{
		URI:         "file:///test",
		Name:        "test-resource",
		MimeType:    "text/plain",
		Description: "A test resource",
	}

	if res.Name != "test-resource" {
		t.Errorf("expected 'test-resource', got %q", res.Name)
	}
}

func TestPromptDetail(t *testing.T) {
	prompt := PromptDetail{
		Name:        "test-prompt",
		Description: "A test prompt",
		Messages: []PromptMessage{
			{Role: "user", Content: "Hello"},
		},
	}

	if prompt.Name != "test-prompt" {
		t.Errorf("expected 'test-prompt', got %q", prompt.Name)
	}
	if len(prompt.Messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(prompt.Messages))
	}
}

func TestCallToolRequest(t *testing.T) {
	req := CallToolRequest{
		Name: "test-tool",
		Arguments: map[string]interface{}{
			"arg1": "value1",
		},
	}

	if req.Name != "test-tool" {
		t.Errorf("expected 'test-tool', got %q", req.Name)
	}
}

func TestCallToolResult(t *testing.T) {
	result := CallToolResult{
		Content: []ContentBlock{
			{Type: "text", Text: "result"},
		},
		IsError: false,
	}

	if len(result.Content) != 1 {
		t.Errorf("expected 1 content block, got %d", len(result.Content))
	}
}

func TestListToolsResult(t *testing.T) {
	result := ListToolsResult{
		Tools: []ToolInfo{
			{Name: "tool1"},
			{Name: "tool2"},
		},
	}

	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}
}

func TestListResourcesResult(t *testing.T) {
	result := ListResourcesResult{
		Resources: []ResourceInfo{
			{Name: "res1"},
		},
	}

	if len(result.Resources) != 1 {
		t.Errorf("expected 1 resource, got %d", len(result.Resources))
	}
}

func TestReadResourceResult(t *testing.T) {
	result := ReadResourceResult{
		Contents: []ResourceContent{
			{URI: "file:///test", Text: "content"},
		},
	}

	if len(result.Contents) != 1 {
		t.Errorf("expected 1 content, got %d", len(result.Contents))
	}
}

func TestListPromptsResult(t *testing.T) {
	result := ListPromptsResult{
		Prompts: []PromptDetail{
			{Name: "prompt1"},
		},
	}

	if len(result.Prompts) != 1 {
		t.Errorf("expected 1 prompt, got %d", len(result.Prompts))
	}
}

// ============================================================================
// JSON-RPC Message Tests
// ============================================================================

func TestJSONRPCMessage(t *testing.T) {
	msg := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
	}

	if msg.JSONRPC != "2.0" {
		t.Errorf("expected JSON-RPC 2.0, got %q", msg.JSONRPC)
	}
	if msg.Method != "initialize" {
		t.Errorf("expected method 'initialize', got %q", msg.Method)
	}
}

// ============================================================================
// ContentBlock Tests
// ============================================================================

func TestContentBlock(t *testing.T) {
	block := ContentBlock{
		Type: "text",
		Text: "Hello, world!",
	}

	if block.Type != "text" {
		t.Errorf("expected type 'text', got %q", block.Type)
	}
	if block.Text != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", block.Text)
	}
}

func TestImageSource(t *testing.T) {
	img := ImageSource{
		Type:     "base64",
		Data:     "dGVzdA==",
		MimeType: "image/png",
	}

	if img.Type != "base64" {
		t.Errorf("expected type 'base64', got %q", img.Type)
	}
}

// ============================================================================
// Elicit Types Tests
// ============================================================================

func TestElicitRequest(t *testing.T) {
	req := ElicitRequest{
		URL: "https://example.com",
		Params: ElicitURLParams{
			Key:    "test-key",
			Prompt: "test prompt",
		},
	}

	if req.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", req.URL)
	}
}

func TestElicitResult(t *testing.T) {
	result := ElicitResult{
		Value:   "test-value",
		Success: true,
	}

	if result.Value != "test-value" {
		t.Errorf("expected value 'test-value', got %q", result.Value)
	}
	if !result.Success {
		t.Error("expected success to be true")
	}
}

// ============================================================================
// ResourceContent Tests
// ============================================================================

func TestResourceContent(t *testing.T) {
	content := ResourceContent{
		URI:      "file:///test.txt",
		MIMEType: "text/plain",
		Text:     "Hello",
	}

	if content.URI != "file:///test.txt" {
		t.Errorf("expected URI 'file:///test.txt', got %q", content.URI)
	}
}
