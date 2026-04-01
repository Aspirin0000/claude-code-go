// Package utils 提供通用工具函数
// 来源: src/utils/errors.ts (238行)
// 重构: Go 错误处理工具
package utils

import (
	"errors"
	"fmt"
)

// ClaudeError Claude 错误类型
type ClaudeError struct {
	Msg string
}

func (e *ClaudeError) Error() string {
	return e.Msg
}

// AbortError 中止错误
type AbortError struct {
	Msg string
}

func (e *AbortError) Error() string {
	return e.Msg
}

// ShellError Shell 命令错误
type ShellError struct {
	Stdout      string
	Stderr      string
	Code        int
	Interrupted bool
}

func (e *ShellError) Error() string {
	return fmt.Sprintf("Shell command failed (exit code %d): %s", e.Code, e.Stderr)
}

// GetErrnoCode 提取 errno 代码（如 'ENOENT', 'EACCES'）
// 对应 TS: export function getErrnoCode(e: unknown): string | undefined
// config.ts 使用此函数
func GetErrnoCode(e error) string {
	if e == nil {
		return ""
	}

	// 检查是否为系统错误（有 Code 字段）
	type errnoError interface {
		Code() string
	}

	if err, ok := e.(errnoError); ok {
		return err.Code()
	}

	// 常见错误映射
	if errors.Is(e, errors.New("ENOENT")) {
		return "ENOENT"
	}

	return ""
}

// IsENOENT 检查是否为 ENOENT 错误
func IsENOENT(e error) bool {
	return GetErrnoCode(e) == "ENOENT"
}

// ToError 将未知值转换为 Error
func ToError(e interface{}) error {
	if err, ok := e.(error); ok {
		return err
	}
	return fmt.Errorf("%v", e)
}

// ErrorMessage 提取错误消息
func ErrorMessage(e interface{}) string {
	if err, ok := e.(error); ok {
		return err.Error()
	}
	return fmt.Sprintf("%v", e)
}
