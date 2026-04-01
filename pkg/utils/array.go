// Package utils 提供通用工具函数
// 来源: src/utils/array.ts (13行)
// 重构: Go 数组工具函数
package utils

// Intersperse 在数组元素之间插入分隔符
// 对应 TS: export function intersperse<A>(as: A[], separator: (index: number) => A): A[]
func Intersperse[T any](as []T, separator func(int) T) []T {
	if len(as) == 0 {
		return []T{}
	}

	result := make([]T, 0, len(as)*2-1)
	for i, a := range as {
		if i > 0 {
			result = append(result, separator(i))
		}
		result = append(result, a)
	}
	return result
}

// Count 计算满足条件的元素数量
// 对应 TS: export function count<T>(arr: readonly T[], pred: (x: T) => unknown): number
func Count[T any](arr []T, pred func(T) bool) int {
	n := 0
	for _, x := range arr {
		if pred(x) {
			n++
		}
	}
	return n
}

// Uniq 去重（保留顺序）
// 对应 TS: export function uniq<T>(xs: Iterable<T>): T[]
func Uniq[T comparable](xs []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(xs))

	for _, x := range xs {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			result = append(result, x)
		}
	}

	return result
}

// UniqWithSlice 去重（切片版本，使用 comparable 约束）
func UniqWithSlice[T comparable](xs []T) []T {
	return Uniq(xs)
}
