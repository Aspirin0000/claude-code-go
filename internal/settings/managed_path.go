// Package settings 提供设置管理
// 来源: src/utils/settings/managedPath.ts
// 重构: Go 管理路径
package settings

import (
	"os"
	"path/filepath"
)

// GetManagedFilePath 获取管理文件路径
// 对应 TS: export function getManagedFilePath(): string
func GetManagedFilePath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "claude", "managed")
}
