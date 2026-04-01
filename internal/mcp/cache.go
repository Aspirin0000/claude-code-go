// Package mcp 提供 MCP 客户端缓存实现
// 来源: src/services/mcp/client.ts (lines 1200-1600) - memoizeWithLRU pattern
// 重构: Go MCP Cache 实现
package mcp

import (
	"container/list"
	"encoding/json"
	"os"
	"strings"
	"sync"

	"github.com/Aspirin0000/claude-code-go/internal/api"
)

// MCP_FETCH_CACHE_SIZE 缓存最大大小
const MCP_FETCH_CACHE_SIZE = 20

// MCP_SDK_TOOL_PREFIX_SKIP 需要跳过的 SDK 工具前缀
const MCP_SDK_TOOL_PREFIX_SKIP = "mcp_sdk_"

// ============================================================================
// LRU Cache 实现
// ============================================================================

// LRUCacheEntry 缓存条目
type LRUCacheEntry[V any] struct {
	Key   string
	Value V
}

// LRUCache 泛型 LRU 缓存 (仅支持 string 类型的 key)
type LRUCache[V any] struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
	mu       sync.RWMutex
}

// NewLRUCache 创建新的 LRU 缓存
func NewLRUCache[V any](capacity int) *LRUCache[V] {
	return &LRUCache[V]{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get 获取缓存值
func (c *LRUCache[V]) Get(key string) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if elem, exists := c.items[key]; exists {
		c.order.MoveToFront(elem)
		return elem.Value.(LRUCacheEntry[V]).Value, true
	}

	var zero V
	return zero, false
}

// Set 设置缓存值
func (c *LRUCache[V]) Set(key string, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.order.MoveToFront(elem)
		elem.Value = LRUCacheEntry[V]{Key: key, Value: value}
		return
	}

	if c.order.Len() >= c.capacity {
		back := c.order.Back()
		if back != nil {
			entry := back.Value.(LRUCacheEntry[V])
			delete(c.items, entry.Key)
			c.order.Remove(back)
		}
	}

	elem := c.order.PushFront(LRUCacheEntry[V]{Key: key, Value: value})
	c.items[key] = elem
}

// Delete 删除缓存项
func (c *LRUCache[V]) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		c.order.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

// Clear 清空缓存
func (c *LRUCache[V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*list.Element)
	c.order = list.New()
}

// Len 返回缓存项数量
func (c *LRUCache[V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// ============================================================================
// MCP Cache 实现
// ============================================================================

// MCPCache MCP 数据缓存
type MCPCache struct {
	toolsCache     *LRUCache[[]ToolInfo]
	resourcesCache *LRUCache[[]ResourceInfo]
	promptsCache   *LRUCache[[]PromptDetail]
}

// NewMCPCache 创建新的 MCP 缓存
func NewMCPCache() *MCPCache {
	return &MCPCache{
		toolsCache:     NewLRUCache[[]ToolInfo](MCP_FETCH_CACHE_SIZE),
		resourcesCache: NewLRUCache[[]ResourceInfo](MCP_FETCH_CACHE_SIZE),
		promptsCache:   NewLRUCache[[]PromptDetail](MCP_FETCH_CACHE_SIZE),
	}
}

// 全局 MCP 缓存实例
var (
	globalMCPCache     *MCPCache
	globalMCPCacheOnce sync.Once
)

// GetGlobalMCPCache 获取全局 MCP 缓存
func GetGlobalMCPCache() *MCPCache {
	globalMCPCacheOnce.Do(func() {
		globalMCPCache = NewMCPCache()
	})
	return globalMCPCache
}

// ============================================================================
// Fetch Tools
// ============================================================================

// FetchToolsForClient 获取客户端的工具列表（带缓存）
// 对应 TS: fetchToolsForClient with memoizeWithLRU
func FetchToolsForClient(client *MCPClient) ([]api.Tool, error) {
	cache := GetGlobalMCPCache()
	clientName := client.GetName()

	// 检查缓存
	if cachedTools, found := cache.toolsCache.Get(clientName); found {
		return convertToolsToAPITools(cachedTools, clientName), nil
	}

	// 从服务器获取
	tools, err := client.GetTools()
	if err != nil {
		return nil, err
	}

	// 应用权限过滤
	filteredTools := filterToolsByPermission(tools, clientName)

	// 存储到缓存
	cache.toolsCache.Set(clientName, filteredTools)

	// 转换为 API Tool 格式
	return convertToolsToAPITools(filteredTools, clientName), nil
}

// filterToolsByPermission 根据权限过滤工具
func filterToolsByPermission(tools []ToolInfo, clientName string) []ToolInfo {
	var filtered []ToolInfo
	for _, tool := range tools {
		// 跳过 IDE 服务器中不允许的工具
		if !IsIncludedMcpTool(tool.Name) {
			continue
		}

		// 检查是否需要跳过 SDK 前缀
		if shouldSkipSDKPrefix(clientName) && strings.HasPrefix(tool.Name, MCP_SDK_TOOL_PREFIX_SKIP) {
			continue
		}

		filtered = append(filtered, tool)
	}
	return filtered
}

// shouldSkipSDKPrefix 检查是否应该跳过 SDK 前缀
func shouldSkipSDKPrefix(clientName string) bool {
	// 根据环境变量判断是否跳过 SDK 前缀
	return os.Getenv("MCP_SKIP_SDK_PREFIX") == "true"
}

// convertToolsToAPITools 将 ToolInfo 转换为 api.Tool
func convertToolsToAPITools(tools []ToolInfo, clientName string) []api.Tool {
	var result []api.Tool
	for _, tool := range tools {
		inputSchema, _ := json.Marshal(tool.InputSchema)
		result = append(result, api.Tool{
			Name:        formatToolName(tool.Name, clientName),
			Description: tool.Description,
			InputSchema: inputSchema,
		})
	}
	return result
}

// formatToolName 格式化工具名称（添加服务器前缀）
func formatToolName(toolName, clientName string) string {
	// 如果工具名已经包含服务器前缀，直接返回
	if strings.HasPrefix(toolName, "mcp__") {
		return toolName
	}
	return "mcp__" + clientName + "__" + toolName
}

// ============================================================================
// Fetch Resources
// ============================================================================

// FetchResourcesForClient 获取客户端的资源列表（带缓存）
// 对应 TS: fetchResourcesForClient with memoizeWithLRU
func FetchResourcesForClient(client *MCPClient) ([]ResourceInfo, error) {
	cache := GetGlobalMCPCache()
	clientName := client.GetName()

	// 检查缓存
	if cachedResources, found := cache.resourcesCache.Get(clientName); found {
		return addServerNameToResources(cachedResources, clientName), nil
	}

	// 从服务器获取
	resources, err := client.GetResources()
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	cache.resourcesCache.Set(clientName, resources)

	// 添加服务器名称
	return addServerNameToResources(resources, clientName), nil
}

// addServerNameToResources 为每个资源添加服务器名称
func addServerNameToResources(resources []ResourceInfo, serverName string) []ResourceInfo {
	result := make([]ResourceInfo, len(resources))
	for i, res := range resources {
		res.Name = serverName + "/" + res.Name
		result[i] = res
	}
	return result
}

// ============================================================================
// Fetch Prompts
// ============================================================================

// FetchPromptsForClient 获取客户端的提示列表（带缓存）
// 对应 TS: fetchPromptsForClient with memoizeWithLRU
func FetchPromptsForClient(client *MCPClient) ([]PromptCommand, error) {
	cache := GetGlobalMCPCache()
	clientName := client.GetName()

	// 检查缓存
	if cachedPrompts, found := cache.promptsCache.Get(clientName); found {
		return convertPromptsToCommands(cachedPrompts, clientName), nil
	}

	// 从服务器获取
	prompts, err := client.GetPrompts()
	if err != nil {
		return nil, err
	}

	// 存储到缓存
	cache.promptsCache.Set(clientName, prompts)

	// 转换为 Command 格式
	return convertPromptsToCommands(prompts, clientName), nil
}

// PromptCommand 提示命令
type PromptCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Source      string `json:"source"`
	ServerName  string `json:"serverName"`
}

// convertPromptsToCommands 将 PromptDetail 转换为 Command 格式
func convertPromptsToCommands(prompts []PromptDetail, clientName string) []PromptCommand {
	var commands []PromptCommand
	for _, prompt := range prompts {
		commands = append(commands, PromptCommand{
			Name:        formatPromptName(prompt.Name, clientName),
			Description: prompt.Description,
			Source:      "mcp",
			ServerName:  clientName,
		})
	}
	return commands
}

// formatPromptName 格式化提示名称
func formatPromptName(promptName, clientName string) string {
	if strings.HasPrefix(promptName, "mcp__") {
		return promptName
	}
	return "mcp__" + clientName + "__" + promptName
}

// ============================================================================
// Clear Cache
// ============================================================================

// ClearClientCache 清除指定客户端的所有缓存
// 对应 TS: clearClientCache
func ClearClientCache(clientName string) {
	cache := GetGlobalMCPCache()
	cache.toolsCache.Delete(clientName)
	cache.resourcesCache.Delete(clientName)
	cache.promptsCache.Delete(clientName)
}

// ClearAllCache 清除所有缓存
func ClearAllCache() {
	cache := GetGlobalMCPCache()
	cache.toolsCache.Clear()
	cache.resourcesCache.Clear()
	cache.promptsCache.Clear()
}
