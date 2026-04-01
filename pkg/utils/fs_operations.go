// Package utils 提供通用工具函数
// 来源: src/utils/fsOperations.ts
// 重构: Go 文件系统操作
package utils

import (
	"os"
)

// FsImplementation 文件系统实现接口
type FsImplementation struct {
	MkdirSync      func(path string) error
	AppendFileSync func(path string, data []byte) error
}

// DefaultFsImplementation 默认文件系统实现
var DefaultFsImplementation = &FsImplementation{
	MkdirSync: func(path string) error {
		return os.MkdirAll(path, 0755)
	},
	AppendFileSync: func(path string, data []byte) error {
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.Write(data)
		return err
	},
}

// GetFsImplementation 获取文件系统实现
// 对应 TS: export function getFsImplementation(): FsImplementation
func GetFsImplementation() *FsImplementation {
	return DefaultFsImplementation
}
