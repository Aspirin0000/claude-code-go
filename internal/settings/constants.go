// Package settings 提供设置管理
// 来源: src/utils/settings/constants.ts
// 重构: Go 设置常量（简化版）
package settings

// VALID_SETTING_SOURCES 有效的设置来源
var VALID_SETTING_SOURCES = []string{"user", "project", "local", "policy"}

// IsSettingSourceEnabled 检查设置来源是否启用（常量版本）
func IsSettingSourceEnabledConst(source string) bool {
	for _, s := range VALID_SETTING_SOURCES {
		if s == source {
			return true
		}
	}
	return false
}
