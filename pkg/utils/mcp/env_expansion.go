// Package mcp 提供 MCP 相关工具函数
// 来源: src/services/mcp/envExpansion.ts
// 重构: Go 环境变量展开
package mcp

import (
	"os"
	"regexp"
)

// ExpandEnvVarsInString 展开字符串中的环境变量
// 对应 TS: export function expandEnvVarsInString(input: string): string
// 支持 $VAR 和 ${VAR} 格式
func ExpandEnvVarsInString(input string) string {
	// 使用 Go 的 os.ExpandEnv，它支持 $VAR 和 ${VAR}
	return os.ExpandEnv(input)
}

// ExpandEnvVarsInConfig 展开配置中的所有环境变量
func ExpandEnvVarsInConfig(config map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range config {
		switch val := v.(type) {
		case string:
			result[k] = ExpandEnvVarsInString(val)
		case map[string]interface{}:
			result[k] = ExpandEnvVarsInConfig(val)
		default:
			result[k] = v
		}
	}
	return result
}

// envVarRegex 环境变量正则表达式
var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
