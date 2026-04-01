// Package utils 提供通用工具函数
// 来源: src/utils/json.ts
// 重构: Go JSON 工具（简化版）
package utils

import (
	"encoding/json"
	"fmt"
)

// SafeParseJSON 安全解析 JSON
// 对应 TS: export function safeParseJSON<T>(json: string): T | null
func SafeParseJSON(data string) (interface{}, error) {
	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SafeParseJSONString 安全解析 JSON 字符串
func SafeParseJSONString(data string) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}
	return result, nil
}

// JSONStringify JSON 序列化
// 对应 TS: export function jsonStringify(obj: unknown, space?: string | number): string
func JSONStringify(obj interface{}, indent ...bool) string {
	var data []byte
	var err error

	if len(indent) > 0 && indent[0] {
		data, err = json.MarshalIndent(obj, "", "  ")
	} else {
		data, err = json.Marshal(obj)
	}

	if err != nil {
		return fmt.Sprintf("%v", obj)
	}

	return string(data)
}
