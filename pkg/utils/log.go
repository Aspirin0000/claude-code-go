// Package utils 提供通用工具函数
// 来源: src/utils/log.ts
// 重构: Go 日志工具
package utils

import (
	"fmt"
	"os"
)

// LogError 记录错误
// 对应 TS: export function logError(message: string, error?: unknown): void
func LogError(message string, err ...error) {
	if len(err) > 0 && err[0] != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] %s: %v\n", message, err[0])
	} else {
		fmt.Fprintf(os.Stderr, "[ERROR] %s\n", message)
	}
}

// LogInfo 记录信息
func LogInfo(message string) {
	fmt.Fprintf(os.Stdout, "[INFO] %s\n", message)
}
