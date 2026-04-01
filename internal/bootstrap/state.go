// Package bootstrap 提供启动时状态管理
// 来源: src/bootstrap/state.ts
// 重构: Go 启动状态
package bootstrap

import (
	"fmt"
	"os"
	"time"
)

var (
	sessionID    string
	sessionStart time.Time
	cwd          string
	originalCwd  string
)

// InitState 初始化状态
func InitState() {
	sessionID = generateSessionID()
	sessionStart = time.Now()
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

// GetSessionStartTime 获取会话开始时间
// 对应 TS: export function getSessionStartTime(): Date
func GetSessionStartTime() time.Time {
	if sessionStart.IsZero() {
		InitState()
	}
	return sessionStart
}

// SetCwd 设置当前工作目录
func SetCwd(newCwd string) {
	cwd = newCwd
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	// 使用进程 ID 和时间戳作为会话标识
	return fmt.Sprintf("session-%d-%d", os.Getpid(), time.Now().Unix())
}
