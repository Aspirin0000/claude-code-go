// Package mcp 提供 MCP 客户端实现
// 来源: src/services/mcp/client.ts (3351行)
// 重构: Go MCP Client 完整实现
package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

// ============================================================================
// 错误类型定义
// ============================================================================

// McpAuthError MCP 认证错误
// 对应 TS: export class McpAuthError extends Error
type McpAuthError struct {
	ServerName string
	Message    string
}

func (e *McpAuthError) Error() string {
	return fmt.Sprintf("MCP server '%s' authentication error: %s", e.ServerName, e.Message)
}

// McpSessionExpiredError MCP 会话过期错误
// 对应 TS: class McpSessionExpiredError extends Error
type McpSessionExpiredError struct {
	ServerName string
}

func (e *McpSessionExpiredError) Error() string {
	return fmt.Sprintf("MCP server '%s' session expired", e.ServerName)
}

// McpToolCallError MCP 工具调用错误
// 对应 TS: export class McpToolCallError_I_VERIFIED...
type McpToolCallError struct {
	Message          string
	TelemetryMessage string
	McpMeta          map[string]interface{}
}

func (e *McpToolCallError) Error() string {
	return e.Message
}

// IsMcpSessionExpiredError 检查是否为 MCP 会话过期错误
// 对应 TS: export function isMcpSessionExpiredError(error: Error): boolean
func IsMcpSessionExpiredError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, `"code":-32001`) || strings.Contains(errMsg, `"code": -32001`)
}

// ============================================================================
// 常量定义
// ============================================================================

const (
	// DefaultMcpToolTimeoutMs 默认 MCP 工具调用超时 (~27.8小时)
	DefaultMcpToolTimeoutMs = 100_000_000

	// MaxMcpDescriptionLength MCP 工具描述最大长度
	MaxMcpDescriptionLength = 2048

	// McpAuthCacheTTLMs MCP 认证缓存 TTL (15分钟)
	McpAuthCacheTTLMs = 15 * 60 * 1000

	// MCPRequestTimeoutMs MCP 请求超时时间 (60秒)
	MCPRequestTimeoutMs = 60000

	// MCPStreamableHttpAccept Streamable HTTP 接受的响应类型
	MCPStreamableHttpAccept = "application/json, text/event-stream"
)

// Error codes for MCP
const (
	ErrorCodeParseError      = -32700
	ErrorCodeInvalidRequest  = -32600
	ErrorCodeMethodNotFound  = -32601
	ErrorCodeInvalidParams   = -32602
	ErrorCodeInternalError   = -32603
	ErrorCodeSessionNotFound = -32001
)

// Image MIME types
const (
	ImageMimeTypeJPEG = "image/jpeg"
	ImageMimeTypePNG  = "image/png"
	ImageMimeTypeGIF  = "image/gif"
	ImageMimeTypeWebP = "image/webp"
)

// GetMcpToolTimeoutMs 获取 MCP 工具调用超时时间(毫秒)
func GetMcpToolTimeoutMs() int {
	timeoutStr := os.Getenv("MCP_TOOL_TIMEOUT")
	if timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			return timeout
		}
	}
	return DefaultMcpToolTimeoutMs
}

// GetConnectionTimeoutMs 获取连接超时时间(毫秒)
func GetConnectionTimeoutMs() int {
	timeoutStr := os.Getenv("MCP_TIMEOUT")
	if timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			return timeout
		}
	}
	return 30000 // 默认 30 秒
}

// GetMcpServerConnectionBatchSize 获取 MCP 服务器连接批处理大小
func GetMcpServerConnectionBatchSize() int {
	batchSizeStr := os.Getenv("MCP_SERVER_CONNECTION_BATCH_SIZE")
	if batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil {
			return batchSize
		}
	}
	return 3
}

// GetRemoteMcpServerConnectionBatchSize 获取远程 MCP 服务器连接批处理大小
func GetRemoteMcpServerConnectionBatchSize() int {
	batchSizeStr := os.Getenv("MCP_REMOTE_SERVER_CONNECTION_BATCH_SIZE")
	if batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil {
			return batchSize
		}
	}
	return 20
}

// IsLocalMcpServer 检查是否为本地 MCP 服务器
func IsLocalMcpServer(config ScopedMcpServerConfig) bool {
	return config.Type == "" || config.Type == "stdio" || config.Type == "sdk"
}

// AllowedIDETools IDE 服务器允许的工具列表
var AllowedIDETools = []string{"mcp__ide__executeCode", "mcp__ide__getDiagnostics"}

// IsIncludedMcpTool 检查 MCP 工具是否应该被包含
func IsIncludedMcpTool(toolName string) bool {
	if !strings.HasPrefix(toolName, "mcp__ide__") {
		return true
	}
	for _, allowed := range AllowedIDETools {
		if toolName == allowed {
			return true
		}
	}
	return false
}

// IsImageMimeType 检查是否为支持的图片 MIME 类型
func IsImageMimeType(mimeType string) bool {
	switch mimeType {
	case ImageMimeTypeJPEG, ImageMimeTypePNG, ImageMimeTypeGIF, ImageMimeTypeWebP:
		return true
	default:
		return false
	}
}

// ============================================================================
// JSON-RPC 类型定义
// ============================================================================

// JSONRPCMessage JSON-RPC 消息
// 对应 TS: JSONRPCMessage
type JSONRPCMessage struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *JSONRPCError   `json:"error,omitempty"`
}

// JSONRPCError JSON-RPC 错误
// 对应 TS: McpError
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *JSONRPCError) Error() string {
	return fmt.Sprintf("JSON-RPC error %d: %s", e.Code, e.Message)
}

// ============================================================================
// MCP 协议类型
// ============================================================================

// ToolInfo 工具信息
// 对应 TS: Tool
type ToolInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
}

// ResourceInfo 资源信息
// 对应 TS: ResourceLink
type ResourceInfo struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	MimeType    string `json:"mimeType,omitempty"`
	Description string `json:"description,omitempty"`
}

// PromptMessage 提示消息
// 对应 TS: PromptMessage
type PromptMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// PromptDetail 提示详情
// 对应 TS: Prompt
type PromptDetail struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages,omitempty"`
}

// ContentBlock 内容块
// 对应 TS: ContentBlockParam
type ContentBlock struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	Image    *ImageSource `json:"image,omitempty"`
	MIMEType string       `json:"mimeType,omitempty"`
	Data     string       `json:"data,omitempty"`
}

// ImageSource 图片源
// 对应 TS: Base64ImageSource
type ImageSource struct {
	Type     string `json:"type"`
	Data     string `json:"data"`
	MimeType string `json:"mime_type"`
}

// CallToolRequest 调用工具请求
// 对应 TS: CallToolRequestSchema
type CallToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult 调用工具结果
// 对应 TS: CallToolResultSchema
type CallToolResult struct {
	Content []ContentBlock         `json:"content"`
	IsError bool                   `json:"isError,omitempty"`
	Meta    map[string]interface{} `json:"_meta,omitempty"`
}

// ListToolsResult 工具列表结果
// 对应 TS: ListToolsResultSchema
type ListToolsResult struct {
	Tools []ToolInfo `json:"tools"`
}

// ListResourcesResult 资源列表结果
// 对应 TS: ListResourcesResultSchema
type ListResourcesResult struct {
	Resources []ResourceInfo `json:"resources"`
}

// ReadResourceResult 资源读取结果
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent 资源内容
type ResourceContent struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"`
}

// ListPromptsResult 提示列表结果
// 对应 TS: ListPromptsResultSchema
type ListPromptsResult struct {
	Prompts []PromptDetail `json:"prompts"`
}

// ListRootsRequest 列出根请求
// 对应 TS: ListRootsRequestSchema
type ListRootsRequest struct {
	ID interface{} `json:"id"`
}

// ElicitRequest 引出请求
// 对应 TS: ElicitRequestSchema
type ElicitRequest struct {
	URL    string          `json:"url"`
	Params ElicitURLParams `json:"params,omitempty"`
}

// ElicitURLParams 引出 URL 参数
// 对应 TS: ElicitRequestURLParams
type ElicitURLParams struct {
	Key    string `json:"key,omitempty"`
	Prompt string `json:"prompt,omitempty"`
}

// ElicitResult 引出结果
// 对应 TS: ElicitResult
type ElicitResult struct {
	Value   string `json:"value"`
	Success bool   `json:"success"`
}

// ============================================================================
// MCP 认证缓存
// ============================================================================

// McpAuthCacheEntry 认证缓存条目
type McpAuthCacheEntry struct {
	Timestamp int64 `json:"timestamp"`
}

// McpAuthCache MCP 认证缓存
type McpAuthCache struct {
	data      map[string]McpAuthCacheEntry
	mu        sync.RWMutex
	cachePath string
}

// 全局认证缓存实例
var (
	globalAuthCache     *McpAuthCache
	globalAuthCacheOnce sync.Once
)

// GetGlobalMcpAuthCache 获取全局 MCP 认证缓存
func GetGlobalMcpAuthCache() *McpAuthCache {
	globalAuthCacheOnce.Do(func() {
		globalAuthCache = &McpAuthCache{
			data:      make(map[string]McpAuthCacheEntry),
			cachePath: GetMcpAuthCachePath(),
		}
		_ = globalAuthCache.Load()
	})
	return globalAuthCache
}

// GetMcpAuthCachePath 获取认证缓存文件路径
func GetMcpAuthCachePath() string {
	configDir := utils.GetClaudeConfigHomeDir()
	return filepath.Join(configDir, "mcp-needs-auth-cache.json")
}

// Load 从文件加载缓存
func (c *McpAuthCache) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取认证缓存失败: %w", err)
	}

	var cacheData map[string]McpAuthCacheEntry
	if err := json.Unmarshal(data, &cacheData); err != nil {
		return fmt.Errorf("解析认证缓存失败: %w", err)
	}

	c.data = cacheData
	return nil
}

// Save 保存缓存到文件
func (c *McpAuthCache) Save() error {
	c.mu.RLock()
	data := c.data
	c.mu.RUnlock()

	dir := filepath.Dir(c.cachePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建缓存目录失败: %w", err)
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	if err := os.WriteFile(c.cachePath, jsonData, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// IsCached 检查服务器认证是否在缓存中且未过期
func (c *McpAuthCache) IsCached(serverId string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.data[serverId]
	if !exists {
		return false
	}

	return time.Now().UnixMilli()-entry.Timestamp < McpAuthCacheTTLMs
}

// SetEntry 设置缓存条目
func (c *McpAuthCache) SetEntry(serverId string) {
	c.mu.Lock()
	c.data[serverId] = McpAuthCacheEntry{
		Timestamp: time.Now().UnixMilli(),
	}
	c.mu.Unlock()

	go func() {
		_ = c.Save()
	}()
}

// Clear 清除缓存
func (c *McpAuthCache) Clear() error {
	c.mu.Lock()
	c.data = make(map[string]McpAuthCacheEntry)
	c.mu.Unlock()

	if err := os.Remove(c.cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除缓存文件失败: %w", err)
	}
	return nil
}

// ============================================================================
// 工具函数
// ============================================================================

// GetLoggingSafeMcpBaseUrl 获取安全日志的 MCP 基础 URL
func GetLoggingSafeMcpBaseUrl(serverRef ScopedMcpServerConfig) string {
	url := GetServerUrl(serverRef.McpServerConfig)
	if url == "" {
		return ""
	}
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
}

// McpBaseUrlAnalytics 获取 MCP 服务器基础 URL 分析数据
func McpBaseUrlAnalytics(serverRef ScopedMcpServerConfig) map[string]string {
	url := GetLoggingSafeMcpBaseUrl(serverRef)
	if url == "" {
		return nil
	}
	return map[string]string{
		"mcpServerBaseUrl": url,
	}
}

// LogMCPDebug 记录 MCP 调试日志
func LogMCPDebug(serverName string, message string) {
	fmt.Printf("[MCP:%s] %s\n", serverName, message)
}

// GetServerCacheKey 生成服务器连接的缓存键
func GetServerCacheKey(name string, serverRef ScopedMcpServerConfig) string {
	configJSON, _ := json.Marshal(serverRef)
	return fmt.Sprintf("%s-%s", name, string(configJSON))
}

// HandleRemoteAuthFailure 处理远程认证失败
func HandleRemoteAuthFailure(name string, serverRef ScopedMcpServerConfig, transportType string) *NeedsAuthMCPServer {
	analytics := McpBaseUrlAnalytics(serverRef)
	if analytics != nil {
		utils.LogEvent("tengu_mcp_server_needs_auth", map[string]interface{}{
			"transportType":    transportType,
			"mcpServerBaseUrl": analytics["mcpServerBaseUrl"],
		})
	}

	labels := map[string]string{
		"sse":            "SSE",
		"http":           "HTTP",
		"claudeai-proxy": "claude.ai proxy",
	}
	label := labels[transportType]
	if label == "" {
		label = transportType
	}

	LogMCPDebug(name, fmt.Sprintf("Authentication required for %s server", label))
	GetGlobalMcpAuthCache().SetEntry(name)

	return &NeedsAuthMCPServer{
		Name:   name,
		Type:   MCPServerConnectionTypeNeedsAuth,
		Config: serverRef,
	}
}

// ConnectionStats 连接统计信息
type ConnectionStats struct {
	TotalServers int
	StdioCount   int
	SseCount     int
	HttpCount    int
	SseIdeCount  int
	WsIdeCount   int
}

// ============================================================================
// MCP Client 结构体和方法
// ============================================================================

// ClientTransport MCP 传输接口
type ClientTransport interface {
	Connect() error
	Close() error
	Send(message JSONRPCMessage) error
	SetOnMessage(handler func(message JSONRPCMessage))
	SetOnClose(handler func())
	SetOnError(handler func(error))
}

// MCPClient MCP 客户端结构体
type MCPClient struct {
	name         string
	config       ScopedMcpServerConfig
	transport    ClientTransport
	capabilities map[string]interface{}
	serverInfo   ServerInfo
	instructions string
	tools        []ToolInfo
	resources    []ResourceInfo
	prompts      []PromptDetail
	connected    bool
	mu           sync.RWMutex
	onClose      func()
	onError      func(error)
}

// NewMCPClient 创建新的 MCP 客户端
func NewMCPClient(name string, config ScopedMcpServerConfig) *MCPClient {
	return &MCPClient{
		name:         name,
		config:       config,
		capabilities: make(map[string]interface{}),
		tools:        make([]ToolInfo, 0),
		resources:    make([]ResourceInfo, 0),
		prompts:      make([]PromptDetail, 0),
		connected:    false,
	}
}

// Connect 连接到 MCP 服务器
func (c *MCPClient) Connect(transport ClientTransport) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("client already connected")
	}

	c.transport = transport

	transport.SetOnMessage(func(message JSONRPCMessage) {
		LogMCPDebug(c.name, fmt.Sprintf("Received: %s", message.Method))
	})

	transport.SetOnClose(func() {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		if c.onClose != nil {
			c.onClose()
		}
	})

	transport.SetOnError(func(err error) {
		if c.onError != nil {
			c.onError(err)
		}
	})

	if err := transport.Connect(); err != nil {
		return fmt.Errorf("transport connect failed: %w", err)
	}

	if err := c.initialize(); err != nil {
		transport.Close()
		return fmt.Errorf("initialization failed: %w", err)
	}

	c.connected = true
	return nil
}

// initialize 执行 MCP 初始化握手
func (c *MCPClient) initialize() error {
	initRequest := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"claude-code","version":"1.0.0"}}`),
	}

	response, err := c.SendRequest(initRequest)
	if err != nil {
		return err
	}

	if response.Error != nil {
		return response.Error
	}

	var initResult struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		Capabilities    map[string]interface{} `json:"capabilities"`
		ServerInfo      ServerInfo             `json:"serverInfo"`
		Instructions    string                 `json:"instructions,omitempty"`
	}

	if err := json.Unmarshal(response.Result, &initResult); err != nil {
		return fmt.Errorf("failed to parse initialize response: %w", err)
	}

	c.capabilities = initResult.Capabilities
	c.serverInfo = initResult.ServerInfo
	c.instructions = initResult.Instructions

	initializedNotification := JSONRPCMessage{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
	}

	return c.transport.Send(initializedNotification)
}

// SendRequest 发送请求并等待响应
func (c *MCPClient) SendRequest(request JSONRPCMessage) (*JSONRPCMessage, error) {
	if !c.connected {
		return nil, fmt.Errorf("client not connected")
	}

	responseChan := make(chan *JSONRPCMessage, 1)

	c.transport.SetOnMessage(func(msg JSONRPCMessage) {
		if msg.ID != nil && fmt.Sprintf("%v", msg.ID) == fmt.Sprintf("%v", request.ID) {
			responseChan <- &msg
		}
	})

	if err := c.transport.Send(request); err != nil {
		return nil, err
	}

	select {
	case response := <-responseChan:
		return response, nil
	case <-time.After(MCPRequestTimeoutMs * time.Millisecond):
		return nil, fmt.Errorf("request timeout")
	}
}

// Close 关闭客户端连接
func (c *MCPClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.connected = false
	return c.transport.Close()
}

// IsConnected 检查客户端是否已连接
func (c *MCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetName 获取客户端名称
func (c *MCPClient) GetName() string {
	return c.name
}

// GetConfig 获取客户端配置
func (c *MCPClient) GetConfig() ScopedMcpServerConfig {
	return c.config
}

// GetTools 获取服务器工具列表
func (c *MCPClient) GetTools() ([]ToolInfo, error) {
	c.mu.RLock()
	if len(c.tools) > 0 {
		defer c.mu.RUnlock()
		return c.tools, nil
	}
	c.mu.RUnlock()

	request := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	response, err := c.SendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	var result ListToolsResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tools list: %w", err)
	}

	c.mu.Lock()
	c.tools = result.Tools
	c.mu.Unlock()

	return result.Tools, nil
}

// CallTool 调用工具
func (c *MCPClient) CallTool(name string, arguments map[string]interface{}) (*CallToolResult, error) {
	params := CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}
	paramsData, _ := json.Marshal(params)

	request := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  paramsData,
	}

	response, err := c.SendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	var result CallToolResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse tool call result: %w", err)
	}

	return &result, nil
}

// GetResources 获取资源列表
func (c *MCPClient) GetResources() ([]ResourceInfo, error) {
	request := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "resources/list",
	}

	response, err := c.SendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	var result ListResourcesResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse resources list: %w", err)
	}

	return result.Resources, nil
}

// ReadResource reads a specific resource by URI
func (c *MCPClient) ReadResource(uri string) (*ReadResourceResult, error) {
	params := map[string]string{"uri": uri}
	paramsData, _ := json.Marshal(params)

	request := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "resources/read",
		Params:  paramsData,
	}

	response, err := c.SendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	var result ReadResourceResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse resource read result: %w", err)
	}

	return &result, nil
}

// GetPrompts 获取提示列表
func (c *MCPClient) GetPrompts() ([]PromptDetail, error) {
	request := JSONRPCMessage{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "prompts/list",
	}

	response, err := c.SendRequest(request)
	if err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	var result ListPromptsResult
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse prompts list: %w", err)
	}

	return result.Prompts, nil
}

// SetOnClose 设置关闭回调
func (c *MCPClient) SetOnClose(handler func()) {
	c.onClose = handler
}

// SetOnError 设置错误回调
func (c *MCPClient) SetOnError(handler func(error)) {
	c.onError = handler
}
