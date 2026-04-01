// Package utils 提供通用工具函数
// 来源: src/utils/envUtils.ts / src/services/analytics/index.ts
// 重构: Go 环境工具和分析日志
package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetClaudeConfigHomeDir 获取 Claude 配置主目录
// 对应 TS: export function getClaudeConfigHomeDir(): string
func GetClaudeConfigHomeDir() string {
	// 首先检查环境变量
	if envDir := os.Getenv("CLAUDE_CONFIG_HOME"); envDir != "" {
		return envDir
	}

	// 使用用户配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		// 降级到用户主目录
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "claude")
	}

	return filepath.Join(configDir, "claude")
}

// LogEvent 记录分析事件
// 对应 TS: export function logEvent(...)
// 当前为无操作实现 - 分析功能需要配置分析后端
func LogEvent(eventName string, properties map[string]interface{}) {
	// 无操作实现 - 分析功能需要配置分析后端
	// 实际实现需要发送到分析后端
	// 仅在调试模式下输出
	if IsDebugMode() {
		fmt.Printf("[Analytics] %s: %v\n", eventName, properties)
	}
}

// DebugLog 记录调试日志
func DebugLog(message string) {
	if IsDebugMode() {
		fmt.Printf("[Debug] %s\n", message)
	}
}
