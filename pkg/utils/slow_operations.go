// Package utils 提供通用工具函数
// 来源: src/utils/slowOperations.ts
// 重构: Go 慢速操作工具
package utils

import "encoding/json"

// JSONStringify 慢速 JSON 序列化（保持 API 兼容）
// 对应 TS: export function jsonStringify(...)
func JSONStringifySlow(obj interface{}, space ...string) string {
	indent := ""
	if len(space) > 0 {
		indent = space[0]
	}

	var data []byte
	var err error

	if indent != "" {
		data, err = json.MarshalIndent(obj, "", indent)
	} else {
		data, err = json.Marshal(obj)
	}

	if err != nil {
		return "{}"
	}

	return string(data)
}
