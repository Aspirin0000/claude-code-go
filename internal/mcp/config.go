// Package mcp 提供 MCP (Model Context Protocol) 配置管理
// 来源: src/services/mcp/config.ts (1579行)
// 重构: Go MCP 配置管理（分块实现）
// 进度: 第1-300行 → 基础函数和工具
package mcp

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/settings"
	"github.com/Aspirin0000/claude-code-go/pkg/utils"
)

// ============================================================================
// 基础路径和常量
// ============================================================================

// GetEnterpriseMcpFilePath 获取企业 MCP 配置文件路径
// 对应 TS: export function getEnterpriseMcpFilePath(): string
// 行: 62-64
func GetEnterpriseMcpFilePath() string {
	return filepath.Join(settings.GetManagedFilePath(), "managed-mcp.json")
}

// ============================================================================
// 范围添加工具
// ============================================================================

// AddScopeToServers 为服务器配置添加范围
// 对应 TS: function addScopeToServers(...)
// 行: 69-81
func AddScopeToServers(servers map[string]McpServerConfig, scope ConfigScope) map[string]ScopedMcpServerConfig {
	if servers == nil {
		return make(map[string]ScopedMcpServerConfig)
	}

	scopedServers := make(map[string]ScopedMcpServerConfig)
	for name, config := range servers {
		scopedServers[name] = ScopedMcpServerConfig{
			McpServerConfig: config,
			Scope:           scope,
		}
	}
	return scopedServers
}

// ============================================================================
// 命令和 URL 提取
// ============================================================================

// GetServerCommandArray 从配置中提取命令数组（仅 stdio 服务器）
// 对应 TS: function getServerCommandArray(...)
// 行: 137-144
func GetServerCommandArray(config McpServerConfig) []string {
	// 非 stdio 服务器没有命令
	if config.Type != "" && config.Type != "stdio" {
		return nil
	}

	if config.Command == "" {
		return nil
	}

	cmd := []string{config.Command}
	if len(config.Args) > 0 {
		cmd = append(cmd, config.Args...)
	}
	return cmd
}

// CommandArraysMatch 检查两个命令数组是否完全匹配
// 对应 TS: function commandArraysMatch(...)
// 行: 149-154
func CommandArraysMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, val := range a {
		if val != b[i] {
			return false
		}
	}
	return true
}

// GetServerUrl 从配置中提取 URL（仅远程服务器）
// 对应 TS: function getServerUrl(...)
// 行: 160-162
func GetServerUrl(config McpServerConfig) string {
	return config.URL
}

// ============================================================================
// CCR 代理 URL 处理
// ============================================================================

// CCR_PROXY_PATH_MARKERS CCR 代理路径标记
// 对应 TS: const CCR_PROXY_PATH_MARKERS = [...]
// 行: 171-174
var CCRProxyPathMarkers = []string{
	"/v2/session_ingress/shttp/mcp/",
	"/v2/ccr-sessions/",
}

// UnwrapCcrProxyUrl 如果是 CCR 代理 URL，提取原始供应商 URL
// 对应 TS: export function unwrapCcrProxyUrl(url: string): string
// 行: 182-193
func UnwrapCcrProxyUrl(url string) string {
	isCcrProxy := false
	for _, marker := range CCRProxyPathMarkers {
		if strings.Contains(url, marker) {
			isCcrProxy = true
			break
		}
	}

	if !isCcrProxy {
		return url
	}

	// 解析 URL 并提取 mcp_url 查询参数
	// 注意: URL 解析实现需要进一步开发 - 当前返回原 URL
	return url
}

// ============================================================================
// 服务器签名和去重
// ============================================================================

// GetMcpServerSignature 计算 MCP 服务器配置的签名
// 对应 TS: export function getMcpServerSignature(...)
// 行: 202-212
func GetMcpServerSignature(config McpServerConfig) string {
	cmd := GetServerCommandArray(config)
	if cmd != nil {
		cmdJSON, _ := json.Marshal(cmd)
		return "stdio:" + string(cmdJSON)
	}

	url := GetServerUrl(config)
	if url != "" {
		return "url:" + UnwrapCcrProxyUrl(url)
	}

	return ""
}

// DedupPluginMcpServers 去重插件 MCP 服务器
// 对应 TS: export function dedupPluginMcpServers(...)
// 行: 223-266
func DedupPluginMcpServers(
	pluginServers map[string]ScopedMcpServerConfig,
	manualServers map[string]ScopedMcpServerConfig,
) (
	servers map[string]ScopedMcpServerConfig,
	suppressed []struct {
		Name        string
		DuplicateOf string
	},
) {
	// 映射签名 -> 服务器名称
	manualSigs := make(map[string]string)
	for name, config := range manualServers {
		sig := GetMcpServerSignature(config.McpServerConfig)
		if sig != "" {
			if _, exists := manualSigs[sig]; !exists {
				manualSigs[sig] = name
			}
		}
	}

	servers = make(map[string]ScopedMcpServerConfig)
	suppressed = make([]struct {
		Name        string
		DuplicateOf string
	}, 0)
	seenPluginSigs := make(map[string]string)

	for name, config := range pluginServers {
		sig := GetMcpServerSignature(config.McpServerConfig)
		if sig == "" {
			servers[name] = config
			continue
		}

		if manualDup, exists := manualSigs[sig]; exists {
			utils.LogForDebugging(fmt.Sprintf(
				`Suppressing plugin MCP server "%s": duplicates manually-configured "%s"`,
				name, manualDup,
			))
			suppressed = append(suppressed, struct {
				Name        string
				DuplicateOf string
			}{
				Name:        name,
				DuplicateOf: manualDup,
			})
			continue
		}

		if pluginDup, exists := seenPluginSigs[sig]; exists {
			utils.LogForDebugging(fmt.Sprintf(
				`Suppressing plugin MCP server "%s": duplicates earlier plugin server "%s"`,
				name, pluginDup,
			))
			suppressed = append(suppressed, struct {
				Name        string
				DuplicateOf string
			}{
				Name:        name,
				DuplicateOf: pluginDup,
			})
			continue
		}

		seenPluginSigs[sig] = name
		servers[name] = config
	}

	return servers, suppressed
}

// DedupClaudeAiMcpServers 去重 ClaudeAI MCP 服务器
// 对应 TS: export function dedupClaudeAiMcpServers(...)
// 行: 281-310
func DedupClaudeAiMcpServers(
	claudeAiServers map[string]ScopedMcpServerConfig,
	manualServers map[string]ScopedMcpServerConfig,
) (
	servers map[string]ScopedMcpServerConfig,
	suppressed []struct {
		Name        string
		DuplicateOf string
	},
) {
	manualSigs := make(map[string]string)
	for name, config := range manualServers {
		// 跳过禁用的服务器
		if IsMcpServerDisabled(name) {
			continue
		}
		sig := GetMcpServerSignature(config.McpServerConfig)
		if sig != "" {
			if _, exists := manualSigs[sig]; !exists {
				manualSigs[sig] = name
			}
		}
	}

	servers = make(map[string]ScopedMcpServerConfig)
	suppressed = make([]struct {
		Name        string
		DuplicateOf string
	}, 0)

	for name, config := range claudeAiServers {
		sig := GetMcpServerSignature(config.McpServerConfig)
		if sig != "" {
			if manualDup, exists := manualSigs[sig]; exists {
				utils.LogForDebugging(fmt.Sprintf(
					`Suppressing claude.ai connector "%s": duplicates manually-configured "%s"`,
					name, manualDup,
				))
				suppressed = append(suppressed, struct {
					Name        string
					DuplicateOf string
				}{
					Name:        name,
					DuplicateOf: manualDup,
				})
				continue
			}
		}
		servers[name] = config
	}

	return servers, suppressed
}

// ============================================================================
// URL 模式匹配
// ============================================================================

// UrlPatternToRegex 将 URL 模式转换为正则表达式
// 对应 TS: function urlPatternToRegex(pattern: string): RegExp
// 行: 320-326
func UrlPatternToRegex(pattern string) string {
	// 转义正则特殊字符（除了 *）
	escaped := regexp.QuoteMeta(pattern)
	// 将 * 替换为 .*（匹配任意字符）
	regexStr := strings.ReplaceAll(escaped, `\*`, `.*`)
	return "^" + regexStr + "$"
}

// UrlMatchesPattern 检查 URL 是否匹配模式
// 对应 TS: function urlMatchesPattern(url: string, pattern: string): boolean
// 行: 331-334
func UrlMatchesPattern(url, pattern string) bool {
	regex := UrlPatternToRegex(pattern)
	matched, _ := regexp.MatchString(regex, url)
	return matched
}

// ============================================================================
// 策略检查（第301-600行）
// ============================================================================

// GetMcpAllowlistSettings 获取 MCP 允许列表设置
// 对应 TS: function getMcpAllowlistSettings()
// 行: 341-346
func GetMcpAllowlistSettings() *settings.SettingsJson {
	if settings.ShouldAllowManagedMcpServersOnly() {
		return settings.GetSettingsForSource(settings.SettingSourcePolicy)
	}
	return settings.GetInitialSettings()
}

// GetMcpDenylistSettings 获取 MCP 拒绝列表设置
// 对应 TS: function getMcpDenylistSettings()
// 行: 353-355
func GetMcpDenylistSettings() *settings.SettingsJson {
	return settings.GetInitialSettings()
}

// IsMcpServerDenied 检查服务器是否被策略拒绝
// 对应 TS: function isMcpServerDenied(...)
// 行: 364-408
func IsMcpServerDenied(serverName string, config *McpServerConfig) bool {
	s := GetMcpDenylistSettings()
	if s == nil || len(s.DeniedMcpServers) == 0 {
		return false // No restrictions
	}

	// Check name-based denial
	for _, entry := range s.DeniedMcpServers {
		if settings.IsMcpServerNameEntry(entry) && entry.ServerName != nil && *entry.ServerName == serverName {
			return true
		}
	}

	// Check command-based denial (stdio servers only) and URL-based denial (remote servers only)
	if config != nil {
		serverCommand := GetServerCommandArray(*config)
		if serverCommand != nil {
			for _, entry := range s.DeniedMcpServers {
				if settings.IsMcpServerCommandEntry(entry) && CommandArraysMatch(entry.ServerCommand, serverCommand) {
					return true
				}
			}
		}

		serverUrl := GetServerUrl(*config)
		if serverUrl != "" {
			for _, entry := range s.DeniedMcpServers {
				if settings.IsMcpServerUrlEntry(entry) && entry.ServerUrl != nil && UrlMatchesPattern(serverUrl, *entry.ServerUrl) {
					return true
				}
			}
		}
	}

	return false
}

// IsMcpServerAllowedByPolicy 检查服务器是否被策略允许
// 对应 TS: function isMcpServerAllowedByPolicy(...)
// 行: 417-508
func IsMcpServerAllowedByPolicy(serverName string, config *McpServerConfig) bool {
	// Denylist takes absolute precedence
	if IsMcpServerDenied(serverName, config) {
		return false
	}

	s := GetMcpAllowlistSettings()
	if s == nil || len(s.AllowedMcpServers) == 0 {
		return true // No allowlist restrictions (undefined or empty means allow all)
	}

	// Check if allowlist contains any command-based or URL-based entries
	hasCommandEntries := false
	hasUrlEntries := false
	for _, entry := range s.AllowedMcpServers {
		if settings.IsMcpServerCommandEntry(entry) {
			hasCommandEntries = true
		}
		if settings.IsMcpServerUrlEntry(entry) {
			hasUrlEntries = true
		}
	}

	if config != nil {
		serverCommand := GetServerCommandArray(*config)
		serverUrl := GetServerUrl(*config)

		if serverCommand != nil {
			// This is a stdio server
			if hasCommandEntries {
				// If ANY serverCommand entries exist, stdio servers MUST match one of them
				for _, entry := range s.AllowedMcpServers {
					if settings.IsMcpServerCommandEntry(entry) && CommandArraysMatch(entry.ServerCommand, serverCommand) {
						return true
					}
				}
				return false // Stdio server doesn't match any command entry
			} else {
				// No command entries, check name-based allowance
				for _, entry := range s.AllowedMcpServers {
					if settings.IsMcpServerNameEntry(entry) && entry.ServerName != nil && *entry.ServerName == serverName {
						return true
					}
				}
				return false
			}
		} else if serverUrl != "" {
			// This is a remote server (sse, http, ws, etc.)
			if hasUrlEntries {
				// If ANY serverUrl entries exist, remote servers MUST match one of them
				for _, entry := range s.AllowedMcpServers {
					if settings.IsMcpServerUrlEntry(entry) && entry.ServerUrl != nil && UrlMatchesPattern(serverUrl, *entry.ServerUrl) {
						return true
					}
				}
				return false // Remote server doesn't match any URL entry
			} else {
				// No URL entries, check name-based allowance
				for _, entry := range s.AllowedMcpServers {
					if settings.IsMcpServerNameEntry(entry) && entry.ServerName != nil && *entry.ServerName == serverName {
						return true
					}
				}
				return false
			}
		} else {
			// Unknown server type - check name-based allowance only
			for _, entry := range s.AllowedMcpServers {
				if settings.IsMcpServerNameEntry(entry) && entry.ServerName != nil && *entry.ServerName == serverName {
					return true
				}
			}
			return false
		}
	}

	// No config provided - check name-based allowance only
	for _, entry := range s.AllowedMcpServers {
		if settings.IsMcpServerNameEntry(entry) && entry.ServerName != nil && *entry.ServerName == serverName {
			return true
		}
	}
	return false
}

// FilterMcpServersByPolicy 按策略过滤服务器
// SDK-type 服务器豁免 - 它们是 SDK 管理的传输，不是 CLI 管理的连接
// 对应 TS: export function filterMcpServersByPolicy<T>(...)
// 行: 536-551
func FilterMcpServersByPolicy(configs map[string]McpServerConfig) (map[string]McpServerConfig, []string) {
	allowed := make(map[string]McpServerConfig)
	blocked := []string{}

	for name, config := range configs {
		// SDK-type servers are exempt from policy checks
		if config.Type == "sdk" || IsMcpServerAllowedByPolicy(name, &config) {
			allowed[name] = config
		} else {
			blocked = append(blocked, name)
		}
	}

	return allowed, blocked
}
