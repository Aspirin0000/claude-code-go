// Package types 提供核心类型定义
// 来源: src/types/utils.ts (3行)
// 重构: Go 泛型工具类型
package types

// DeepImmutable 递归将类型变为不可变
// 对应 TS: export type DeepImmutable<T> = T
// 注意: Go 中通过值传递和不可变设计模式实现，无需泛型包装
type DeepImmutable[T any] struct {
	value T
}

// NewDeepImmutable 创建深不可变包装
func NewDeepImmutable[T any](v T) DeepImmutable[T] {
	return DeepImmutable[T]{value: v}
}

// Value 获取内部值（返回副本）
func (d DeepImmutable[T]) Value() T {
	return d.value
}

// Permutations 生成类型的所有排列
// 对应 TS: export type Permutations<T> = T[]
// Go 实现: 使用切片即可
type Permutations[T any] []T
