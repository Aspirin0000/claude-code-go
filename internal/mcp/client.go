// Package mcp 提供 MCP 客户端实现
// 来源: src/services/mcp/client.ts (3351行)
// 批次: C-1/8 - 基础类型、错误定义、常量、缓存 (1-400行等效)
package mcp

import (
	"encoding/json"
	"errors"
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
// 对应 TS: export class McpToolCallError_I_VERIFIED_THIS_IS_NOT_CODE_OR_FILEPATHS
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
// 检查 HTTP 404 状态码和 JSON-RPC -32001 错误码
func IsMcpSessionExpiredError(err error) bool {
	if err == nil {
		return false
	}

	// 检查错误消息中是否包含 -32001 错误码
	// MCP 服务器返回: {"error":{"code":-32001,"message":"Session not found"}}
	errMsg := err.Error()
	return strings.Contains(errMsg, `"code":-32001`) || strings.Contains(errMsg, `"code": -32001`)
}

// ============================================================================
// 常量定义
// ============================================================================

const (
	// DefaultMcpToolTimeoutMs 默认 MCP 工具调用超时 (~27.8小时)
	// 对应 TS: const DEFAULT_MCP_TOOL_TIMEOUT_MS = 100_000_000
	DefaultMcpToolTimeoutMs = 100_000_000

	// MaxMcpDescriptionLength MCP 工具描述和服务器指令的最大长度
	// 对应 TS: const MAX_MCP_DESCRIPTION_LENGTH = 2048
	MaxMcpDescriptionLength = 2048

	// McpAuthCacheTTLMs MCP 认证缓存 TTL (15分钟)
	// 对应 TS: const MCP_AUTH_CACHE_TTL_MS = 15 * 60 * 1000
	McpAuthCacheTTLMs = 15 * 60 * 1000
)

// GetMcpToolTimeoutMs 获取 MCP 工具调用超时时间(毫秒)
// 对应 TS: function getMcpToolTimeoutMs(): number
func GetMcpToolTimeoutMs() int {
	timeoutStr := os.Getenv("MCP_TOOL_TIMEOUT")
	if timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			return timeout
		}
	}
	return DefaultMcpToolTimeoutMs
}

// ============================================================================
// MCP 认证缓存
// ============================================================================

// McpAuthCacheEntry 认证缓存条目
// 对应 TS: type McpAuthCacheData = Record<string, { timestamp: number }>
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
		// 尝试加载已有缓存
		_ = globalAuthCache.Load()
	})
	return globalAuthCache
}

// GetMcpAuthCachePath 获取认证缓存文件路径
// 对应 TS: function getMcpAuthCachePath(): string
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
			return nil // 缓存文件不存在，使用空缓存
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

	// 确保目录存在
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
// 对应 TS: async function isMcpAuthCached(serverId: string): Promise<boolean>
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
// 对应 TS: function setMcpAuthCacheEntry(serverId: string): void
func (c *McpAuthCache) SetEntry(serverId string) {
	c.mu.Lock()
	c.data[serverId] = McpAuthCacheEntry{
		Timestamp: time.Now().UnixMilli(),
	}
	c.mu.Unlock()

	// 异步保存到文件
	go func() {
		_ = c.Save()
	}()
}

// Clear 清除缓存
// 对应 TS: export function clearMcpAuthCache(): void
func (c *McpAuthCache) Clear() error {
	c.mu.Lock()
	c.data = make(map[string]McpAuthCacheEntry)
	c.mu.Unlock()

	// 删除缓存文件
	if err := os.Remove(c.cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("删除缓存文件失败: %w", err)
	}
	return nil
}

// ============================================================================
// 分析辅助函数
// ============================================================================

// McpBaseUrlAnalytics 获取 MCP 服务器基础 URL 分析数据
// 对应 TS: function mcpBaseUrlAnalytics(serverRef: ScopedMcpServerConfig)
func McpBaseUrlAnalytics(serverRef ScopedMcpServerConfig) map[string]string {
	url := GetLoggingSafeMcpBaseUrl(serverRef)
	if url == "" {
		return nil
	}
	return map[string]string{
		"mcpServerBaseUrl": url,
	}
}

// HandleRemoteAuthFailure 处理远程认证失败
// 对应 TS: function handleRemoteAuthFailure(name, serverRef, transportType)
func HandleRemoteAuthFailure(name string, serverRef ScopedMcpServerConfig, transportType string) *NeedsAuthMCPServer {
	// 记录分析事件
	analytics := McpBaseUrlAnalytics(serverRef)
	if analytics != nil {
		utils.LogEvent("tengu_mcp_server_needs_auth", map[string]interface{}{
			"transportType":    transportType,
			"mcpServerBaseUrl": analytics["mcpServerBaseUrl"],
		})
	}

	// 获取标签
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

	// 设置认证缓存
	GetGlobalMcpAuthCache().SetEntry(name)

	return &NeedsAuthMCPServer{
		Name:   name,
		Type:   MCPServerConnectionTypeNeedsAuth,
		Config: serverRef,
	}
}

// Error codes for MCP
const (
	ErrorCodeParseError      = -32700
	ErrorCodeInvalidRequest  = -32600
	ErrorCodeMethodNotFound  = -32601
	ErrorCodeInvalidParams   = -32602
	ErrorCodeInternalError   = -32603
	ErrorCodeSessionNotFound = -32001
)

// MCPError represents an MCP error response
type MCPError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *MCPError) Error() string {
	return fmt.Sprintf("MCP error %d: %s", e.Code, e.Message)
}

// IsMCPError checks if an error is an MCPError
func IsMCPError(err error) (*MCPError, bool) {
	var mcpErr *MCPError
	if errors.As(err, &mcpErr) {
		return mcpErr, true
	}
	return nil, false
}

// NewMCPError creates a new MCPError
func NewMCPError(code int, message string, data ...interface{}) *MCPError {
	err := &MCPError{
		Code:    code,
		Message: message,
	}
	if len(data) > 0 {
		err.Data = data[0]
	}
	return err
}

// ============================================================================
// 类型定义 (来自 types.ts 的补充)
// ============================================================================

// GetLoggingSafeMcpBaseUrl 获取安全日志的 MCP 基础 URL
// 对应 TS: import { getLoggingSafeMcpBaseUrl } from './utils.js'
// 注意: 这是一个存根实现，完整的实现将在 C-2/8 中提供
func GetLoggingSafeMcpBaseUrl(serverRef ScopedMcpServerConfig) string {
	// 移除查询参数，只返回基础 URL
	url := GetServerUrl(serverRef.McpServerConfig)
	if url == "" {
		return ""
	}
	// 简单实现：如果有 ? 则截断
	if idx := strings.Index(url, "?"); idx != -1 {
		return url[:idx]
	}
	return url
}

// LogMCPDebug 记录 MCP 调试日志
// 对应 TS: import { logMCPDebug } from '../../utils/log.js'
func LogMCPDebug(serverName string, message string) {
	// 使用标准日志记录 MCP 调试信息
	fmt.Printf("[MCP:%s] %s\n", serverName, message)
}
