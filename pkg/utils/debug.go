// Package utils 提供通用工具函数
// 来源: src/utils/debug.ts (268行)
// 重构: Go 调试日志工具（简化版 - 仅核心功能）
package utils

import (
	"fmt"
	"os"
	"time"
)

// DebugLogLevel 日志级别
// 对应 TS: export type DebugLogLevel = 'verbose' | 'debug' | 'info' | 'warn' | 'error'
type DebugLogLevel string

const (
	DebugLogLevelVerbose DebugLogLevel = "verbose"
	DebugLogLevelDebug   DebugLogLevel = "debug"
	DebugLogLevelInfo    DebugLogLevel = "info"
	DebugLogLevelWarn    DebugLogLevel = "warn"
	DebugLogLevelError   DebugLogLevel = "error"
)

var levelOrder = map[DebugLogLevel]int{
	DebugLogLevelVerbose: 0,
	DebugLogLevelDebug:   1,
	DebugLogLevelInfo:    2,
	DebugLogLevelWarn:    3,
	DebugLogLevelError:   4,
}

var (
	runtimeDebugEnabled = false
	debugFilePath       string
)

// IsDebugMode 检查是否为调试模式
// 对应 TS: export const isDebugMode = memoize(...)
func IsDebugMode() bool {
	if runtimeDebugEnabled {
		return true
	}

	// 检查环境变量
	if os.Getenv("DEBUG") != "" || os.Getenv("DEBUG_SDK") != "" {
		return true
	}

	// 检查命令行参数
	for _, arg := range os.Args {
		if arg == "--debug" || arg == "-d" || arg == "--debug-to-stderr" {
			return true
		}
		if arg == "--debug-file" {
			return true
		}
	}

	return false
}

// GetDebugFilePath 获取调试文件路径
func GetDebugFilePath() string {
	if debugFilePath != "" {
		return debugFilePath
	}

	// 从命令行参数解析
	for i, arg := range os.Args {
		if arg == "--debug-file" && i+1 < len(os.Args) {
			debugFilePath = os.Args[i+1]
			return debugFilePath
		}
		if len(arg) > 13 && arg[:13] == "--debug-file=" {
			debugFilePath = arg[13:]
			return debugFilePath
		}
	}

	return ""
}

// GetMinDebugLogLevel 获取最小日志级别
func GetMinDebugLogLevel() DebugLogLevel {
	raw := os.Getenv("CLAUDE_CODE_DEBUG_LOG_LEVEL")
	if raw != "" {
		switch raw {
		case "verbose":
			return DebugLogLevelVerbose
		case "debug":
			return DebugLogLevelDebug
		case "info":
			return DebugLogLevelInfo
		case "warn":
			return DebugLogLevelWarn
		case "error":
			return DebugLogLevelError
		}
	}
	return DebugLogLevelDebug
}

// LogForDebugging 记录调试日志
// 对应 TS: export function logForDebugging(message: string, { level }: { level: DebugLogLevel })
// config.ts 使用此函数
func LogForDebugging(message string, level ...DebugLogLevel) {
	lvl := DebugLogLevelDebug
	if len(level) > 0 {
		lvl = level[0]
	}

	// 检查日志级别
	if levelOrder[lvl] < levelOrder[GetMinDebugLogLevel()] {
		return
	}

	if !IsDebugMode() {
		return
	}

	timestamp := time.Now().Format(time.RFC3339)
	output := fmt.Sprintf("%s [%s] %s\n", timestamp, lvl, message)

	// 输出到标准错误
	fmt.Fprint(os.Stderr, output)
}

// EnableDebugLogging 启用调试日志
func EnableDebugLogging() bool {
	wasActive := IsDebugMode()
	runtimeDebugEnabled = true
	return wasActive
}
