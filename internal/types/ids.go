// Package types 提供核心类型定义
// 来源: src/types/ids.ts (44行)
// 重构: Go 类型系统
package types

import (
	"regexp"
)

// ============================================================================
// ID 类型定义
// 使用自定义类型模拟 TypeScript branded types
// ============================================================================

// SessionId 唯一标识一个 Claude Code 会话
// 对应 TS: export type SessionId = string & { readonly __brand: 'SessionId' }
type SessionId string

// AgentId 唯一标识会话内的子代理
// 对应 TS: export type AgentId = string & { readonly __brand: 'AgentId' }
type AgentId string

// AGENT_ID_PATTERN 匹配 AgentId 格式: 'a' + 可选的 '<label>-' + 16位十六进制字符
// 对应 TS: const AGENT_ID_PATTERN = /^a(?:.+-)?[0-9a-f]{16}$/
var AGENT_ID_PATTERN = regexp.MustCompile(`^a(?:.+-)?[0-9a-f]{16}$`)

// NewSessionId 将原始字符串转换为 SessionId
// 对应 TS: export function asSessionId(id: string): SessionId
func NewSessionId(id string) SessionId {
	return SessionId(id)
}

// NewAgentId 将原始字符串转换为 AgentId（不进行验证）
// 对应 TS: export function asAgentId(id: string): AgentId
func NewAgentId(id string) AgentId {
	return AgentId(id)
}

// ParseAgentId 验证并转换字符串为 AgentId
// 匹配格式: 'a' + 可选的 '<label>-' + 16位十六进制字符
// 如果不匹配返回 nil（例如队友名称、团队寻址）
// 对应 TS: export function toAgentId(s: string): AgentId | null
func ParseAgentId(s string) *AgentId {
	if AGENT_ID_PATTERN.MatchString(s) {
		id := AgentId(s)
		return &id
	}
	return nil
}

// String 返回 SessionId 的字符串表示
func (s SessionId) String() string {
	return string(s)
}

// String 返回 AgentId 的字符串表示
func (a AgentId) String() string {
	return string(a)
}

// IsValid 验证 AgentId 格式是否正确
func (a AgentId) IsValid() bool {
	return AGENT_ID_PATTERN.MatchString(string(a))
}
