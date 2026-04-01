// Package utils 提供通用工具函数
// 来源: src/utils/platform.ts
// 重构: Go 平台检测
package utils

import "runtime"

// Platform 平台类型
type Platform string

const (
	PlatformDarwin  Platform = "darwin"
	PlatformLinux   Platform = "linux"
	PlatformWindows Platform = "windows"
)

// GetPlatform 获取当前平台
// 对应 TS: export function getPlatform(): Platform
func GetPlatform() Platform {
	switch runtime.GOOS {
	case "darwin":
		return PlatformDarwin
	case "linux":
		return PlatformLinux
	case "windows":
		return PlatformWindows
	default:
		return PlatformLinux
	}
}
