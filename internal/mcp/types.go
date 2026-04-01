// Package mcp 提供 MCP (Model Context Protocol) 服务类型
// 来源: src/services/mcp/types.ts (258行)
// 重构: Go MCP 类型系统
package mcp

// ============================================================================
// 配置范围
// ============================================================================

// ConfigScope 配置范围
// 对应 TS: export type ConfigScope = ...
type ConfigScope string

const (
	ConfigScopeLocal      ConfigScope = "local"
	ConfigScopeUser       ConfigScope = "user"
	ConfigScopeProject    ConfigScope = "project"
	ConfigScopeGlobal     ConfigScope = "global"
	ConfigScopeDynamic    ConfigScope = "dynamic"
	ConfigScopeEnterprise ConfigScope = "enterprise"
	ConfigScopeClaudeAI   ConfigScope = "claudeai"
	ConfigScopeManaged    ConfigScope = "managed"
)

// ============================================================================
// 传输类型
// ============================================================================

// Transport 传输类型
type Transport string

const (
	TransportStdio     Transport = "stdio"
	TransportSSE       Transport = "sse"
	TransportSSEIDE    Transport = "sse-ide"
	TransportHTTP      Transport = "http"
	TransportWebSocket Transport = "ws"
	TransportSDK       Transport = "sdk"
)

// ============================================================================
// 服务器配置
// ============================================================================

// McpStdioServerConfig stdio 服务器配置
type McpStdioServerConfig struct {
	Type    *string           `json:"type,omitempty"` // "stdio"
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// McpSSEServerConfig SSE 服务器配置
type McpSSEServerConfig struct {
	Type          string            `json:"type"` // "sse"
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	HeadersHelper *string           `json:"headersHelper,omitempty"`
	OAuth         *McpOAuthConfig   `json:"oauth,omitempty"`
}

// McpSSEIDEServerConfig SSE IDE 服务器配置
type McpSSEIDEServerConfig struct {
	Type                string `json:"type"` // "sse-ide"
	URL                 string `json:"url"`
	IdeName             string `json:"ideName"`
	IdeRunningInWindows *bool  `json:"ideRunningInWindows,omitempty"`
}

// McpWebSocketIDEServerConfig WebSocket IDE 服务器配置
type McpWebSocketIDEServerConfig struct {
	Type                string  `json:"type"` // "ws-ide"
	URL                 string  `json:"url"`
	IdeName             string  `json:"ideName"`
	AuthToken           *string `json:"authToken,omitempty"`
	IdeRunningInWindows *bool   `json:"ideRunningInWindows,omitempty"`
}

// McpHTTPServerConfig HTTP 服务器配置
type McpHTTPServerConfig struct {
	Type          string            `json:"type"` // "http"
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	HeadersHelper *string           `json:"headersHelper,omitempty"`
	OAuth         *McpOAuthConfig   `json:"oauth,omitempty"`
}

// McpWebSocketServerConfig WebSocket 服务器配置
type McpWebSocketServerConfig struct {
	Type          string            `json:"type"` // "ws"
	URL           string            `json:"url"`
	Headers       map[string]string `json:"headers,omitempty"`
	HeadersHelper *string           `json:"headersHelper,omitempty"`
}

// McpSdkServerConfig SDK 服务器配置
type McpSdkServerConfig struct {
	Type string `json:"type"` // "sdk"
	Name string `json:"name"`
}

// McpClaudeAIProxyServerConfig ClaudeAI 代理服务器配置
type McpClaudeAIProxyServerConfig struct {
	Type string `json:"type"` // "claudeai-proxy"
	URL  string `json:"url"`
	ID   string `json:"id"`
}

// McpOAuthConfig OAuth 配置
type McpOAuthConfig struct {
	ClientID              *string `json:"clientId,omitempty"`
	CallbackPort          *int    `json:"callbackPort,omitempty"`
	AuthServerMetadataURL *string `json:"authServerMetadataUrl,omitempty"`
	XAA                   *bool   `json:"xaa,omitempty"`
}

// McpServerConfig MCP 服务器配置（联合类型）
type McpServerConfig struct {
	// 通用字段
	Type string `json:"type,omitempty"`

	// stdio 字段
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// sse/http/ws 字段
	URL           string            `json:"url,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	HeadersHelper *string           `json:"headersHelper,omitempty"`
	OAuth         *McpOAuthConfig   `json:"oauth,omitempty"`

	// sse-ide/ws-ide 字段
	IdeName             string  `json:"ideName,omitempty"`
	IdeRunningInWindows *bool   `json:"ideRunningInWindows,omitempty"`
	AuthToken           *string `json:"authToken,omitempty"`

	// sdk 字段
	Name string `json:"name,omitempty"`

	// claudeai-proxy 字段
	ID string `json:"id,omitempty"`
}

// ScopedMcpServerConfig 带范围的 MCP 服务器配置
type ScopedMcpServerConfig struct {
	McpServerConfig
	Scope        ConfigScope `json:"scope"`
	PluginSource *string     `json:"pluginSource,omitempty"`
}

// McpJsonConfig MCP JSON 配置
type McpJsonConfig struct {
	McpServers map[string]McpServerConfig `json:"mcpServers"`
}

// ============================================================================
// 服务器连接状态
// ============================================================================

// MCPServerConnectionType 连接类型
type MCPServerConnectionType string

const (
	MCPServerConnectionTypeConnected MCPServerConnectionType = "connected"
	MCPServerConnectionTypeFailed    MCPServerConnectionType = "failed"
	MCPServerConnectionTypeNeedsAuth MCPServerConnectionType = "needs-auth"
	MCPServerConnectionTypePending   MCPServerConnectionType = "pending"
	MCPServerConnectionTypeDisabled  MCPServerConnectionType = "disabled"
)

// ConnectedMCPServer 已连接的 MCP 服务器
type ConnectedMCPServer struct {
	Name         string                  `json:"name"`
	Type         MCPServerConnectionType `json:"type"`   // "connected"
	Client       interface{}             `json:"client"` // MCP Client
	Capabilities interface{}             `json:"capabilities"`
	ServerInfo   *ServerInfo             `json:"serverInfo,omitempty"`
	Instructions *string                 `json:"instructions,omitempty"`
	Config       ScopedMcpServerConfig   `json:"config"`
}

// ServerInfo 服务器信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// FailedMCPServer 失败的 MCP 服务器
type FailedMCPServer struct {
	Name   string                  `json:"name"`
	Type   MCPServerConnectionType `json:"type"` // "failed"
	Config ScopedMcpServerConfig   `json:"config"`
	Error  *string                 `json:"error,omitempty"`
}

// NeedsAuthMCPServer 需要认证的 MCP 服务器
type NeedsAuthMCPServer struct {
	Name   string                  `json:"name"`
	Type   MCPServerConnectionType `json:"type"` // "needs-auth"
	Config ScopedMcpServerConfig   `json:"config"`
}

// PendingMCPServer 待连接的 MCP 服务器
type PendingMCPServer struct {
	Name                 string                  `json:"name"`
	Type                 MCPServerConnectionType `json:"type"` // "pending"
	Config               ScopedMcpServerConfig   `json:"config"`
	ReconnectAttempt     *int                    `json:"reconnectAttempt,omitempty"`
	MaxReconnectAttempts *int                    `json:"maxReconnectAttempts,omitempty"`
}

// DisabledMCPServer 禁用的 MCP 服务器
type DisabledMCPServer struct {
	Name   string                  `json:"name"`
	Type   MCPServerConnectionType `json:"type"` // "disabled"
	Config ScopedMcpServerConfig   `json:"config"`
}

// MCPServerConnection MCP 服务器连接状态
type MCPServerConnection interface{}

// ============================================================================
// 资源类型
// ============================================================================

// ServerResource 服务器资源
type ServerResource struct {
	Server string `json:"server"`
	// 继承 Resource 字段
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	MIMEType    *string `json:"mimeType,omitempty"`
}

// ============================================================================
// CLI 状态类型
// ============================================================================

// SerializedTool 序列化工具
type SerializedTool struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	InputJSONSchema  map[string]interface{} `json:"inputJSONSchema,omitempty"`
	IsMcp            *bool                  `json:"isMcp,omitempty"`
	OriginalToolName *string                `json:"originalToolName,omitempty"`
}

// SerializedClient 序列化客户端
type SerializedClient struct {
	Name         string                  `json:"name"`
	Type         MCPServerConnectionType `json:"type"`
	Capabilities interface{}             `json:"capabilities,omitempty"`
}

// MCPCliState MCP CLI 状态
type MCPCliState struct {
	Clients         []SerializedClient               `json:"clients"`
	Configs         map[string]ScopedMcpServerConfig `json:"configs"`
	Tools           []SerializedTool                 `json:"tools"`
	Resources       map[string][]ServerResource      `json:"resources"`
	NormalizedNames map[string]string                `json:"normalizedNames,omitempty"`
}

// ============================================================================
// MCP 管理器接口
// ============================================================================

// MCPManager MCP 管理器接口
type MCPManager interface {
	ConnectServer(name string, config ScopedMcpServerConfig) (MCPServerConnection, error)
	DisconnectServer(name string) error
	GetServer(name string) (MCPServerConnection, bool)
	ListServers() []MCPServerConnection
	GetTools() []SerializedTool
	GetResources() map[string][]ServerResource
	GetState() MCPCliState
}
