// Package utils 提供通用工具函数
// 来源: src/utils/cwd.ts (32行)
// 重构: Go 工作目录工具（完整实现）
package utils

import (
	"github.com/Aspirin0000/claude-code-go/internal/bootstrap"
)

// Pwd 获取当前工作目录
// 对应 TS: export function pwd(): string
func Pwd() string {
	return bootstrap.GetCwdState()
}

// GetCwd 获取当前工作目录（带降级）
// 对应 TS: export function getCwd(): string
func GetCwd() string {
	cwd := bootstrap.GetCwdState()
	if cwd == "" {
		return bootstrap.GetOriginalCwd()
	}
	return cwd
}
