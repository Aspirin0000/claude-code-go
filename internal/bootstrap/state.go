// Package bootstrap 提供启动时状态管理
// 来源: src/bootstrap/state.ts
// 重构: Go 启动状态（简化版）
package bootstrap

import (
	"os"
)

var (
	sessionID   string
	cwd         string
	originalCwd string
)

// InitState 初始化状态
func InitState() {
	sessionID = generateSessionID()
	cwd, _ = os.Getwd()
	originalCwd = cwd
}

// GetSessionId 获取会话 ID
// 对应 TS: export function getSessionId(): string
func GetSessionId() string {
	if sessionID == "" {
		InitState()
	}
	return sessionID
}

// GetCwdState 获取当前工作目录状态
// 对应 TS: export function getCwdState(): string
func GetCwdState() string {
	if cwd == "" {
		cwd, _ = os.Getwd()
	}
	return cwd
}

// GetOriginalCwd 获取原始工作目录
// 对应 TS: export function getOriginalCwd(): string
func GetOriginalCwd() string {
	if originalCwd == "" {
		originalCwd, _ = os.Getwd()
	}
	return originalCwd
}

// SetCwd 设置当前工作目录
func SetCwd(newCwd string) {
	cwd = newCwd
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	// 简化实现：使用时间戳
	return "session-" + string(rune(os.Getpid()))
}
