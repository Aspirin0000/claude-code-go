// Package utils 提供通用工具函数
// 来源: src/utils/uuid.ts (27行)
// 重构: Go UUID 工具
package utils

import (
	"crypto/rand"
	"encoding/hex"
	"regexp"

	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// uuidRegex UUID 格式正则
// 对应 TS: const uuidRegex = /^[0-9a-f]{8}-...$/i
var uuidRegex = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ValidateUuid 验证 UUID
// 对应 TS: export function validateUuid(maybeUuid: unknown): UUID | null
func ValidateUuid(maybeUuid interface{}) *string {
	// 检查是否为字符串
	str, ok := maybeUuid.(string)
	if !ok {
		return nil
	}

	if uuidRegex.MatchString(str) {
		return &str
	}
	return nil
}

// CreateAgentId 创建代理 ID
// 对应 TS: export function createAgentId(label?: string): AgentId
// Format: a{label-}{16 hex chars}
func CreateAgentId(label *string) types.AgentId {
	// 生成 8 字节 (16 个十六进制字符)
	bytes := make([]byte, 8)
	rand.Read(bytes)
	suffix := hex.EncodeToString(bytes)

	if label != nil && *label != "" {
		return types.AgentId("a" + *label + "-" + suffix)
	}
	return types.AgentId("a" + suffix)
}
