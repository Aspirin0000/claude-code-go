// Package utils 提供通用工具函数
// 来源: src/utils/set.ts (53行)
// 重构: Go Set 工具函数（高性能版本）
package utils

// Difference 计算两个 Set 的差集 (a - b)
// 对应 TS: export function difference<A>(a: Set<A>, b: Set<A>): Set<A>
// Note: this code is hot, so is optimized for speed.
func Difference[T comparable](a, b map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{})
	for item := range a {
		if _, exists := b[item]; !exists {
			result[item] = struct{}{}
		}
	}
	return result
}

// Intersects 检查两个 Set 是否有交集
// 对应 TS: export function intersects<A>(a: Set<A>, b: Set<A>): boolean
// Note: this code is hot, so is optimized for speed.
func Intersects[T comparable](a, b map[T]struct{}) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	// 遍历较小的集合以提高性能
	if len(a) > len(b) {
		a, b = b, a
	}
	for item := range a {
		if _, exists := b[item]; exists {
			return true
		}
	}
	return false
}

// Every 检查 Set a 的所有元素是否都在 Set b 中
// 对应 TS: export function every<A>(a: ReadonlySet<A>, b: ReadonlySet<A>): boolean
// Note: this code is hot, so is optimized for speed.
func Every[T comparable](a, b map[T]struct{}) bool {
	for item := range a {
		if _, exists := b[item]; !exists {
			return false
		}
	}
	return true
}

// Union 计算两个 Set 的并集
// 对应 TS: export function union<A>(a: Set<A>, b: Set<A>): Set<A>
// Note: this code is hot, so is optimized for speed.
func Union[T comparable](a, b map[T]struct{}) map[T]struct{} {
	result := make(map[T]struct{}, len(a)+len(b))
	for item := range a {
		result[item] = struct{}{}
	}
	for item := range b {
		result[item] = struct{}{}
	}
	return result
}

// ToSet 将切片转换为 Set (map[T]struct{})
func ToSet[T comparable](slice []T) map[T]struct{} {
	set := make(map[T]struct{}, len(slice))
	for _, item := range slice {
		set[item] = struct{}{}
	}
	return set
}

// ToSlice 将 Set 转换为切片
func ToSlice[T comparable](set map[T]struct{}) []T {
	slice := make([]T, 0, len(set))
	for item := range set {
		slice = append(slice, item)
	}
	return slice
}
