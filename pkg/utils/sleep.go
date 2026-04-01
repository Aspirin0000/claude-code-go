// Package utils 提供通用工具函数
// 来源: src/utils/sleep.ts (84行)
// 重构: Go 睡眠和超时工具
package utils

import (
	"context"
	"errors"
	"time"
)

// SleepOptions 睡眠选项
type SleepOptions struct {
	ThrowOnAbort bool
	AbortError   func() error
}

// Sleep 可中断的睡眠函数
// 对应 TS: export function sleep(ms: number, signal?: AbortSignal, opts?: {...})
// 使用 Go 的 context 替代 AbortSignal
func Sleep(ctx context.Context, ms int, opts ...SleepOptions) error {
	var opt SleepOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 检查 context 是否已取消
	select {
	case <-ctx.Done():
		if opt.ThrowOnAbort || opt.AbortError != nil {
			if opt.AbortError != nil {
				return opt.AbortError()
			}
			return errors.New("aborted")
		}
		return nil
	default:
	}

	timer := time.NewTimer(time.Duration(ms) * time.Millisecond)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		if opt.ThrowOnAbort || opt.AbortError != nil {
			if opt.AbortError != nil {
				return opt.AbortError()
			}
			return errors.New("aborted")
		}
		return nil
	}
}

// WithTimeout 为 promise 添加超时
// 对应 TS: export function withTimeout<T>(promise: Promise<T>, ms: number, message: string)
func WithTimeout[T any](ctx context.Context, promise func(context.Context) (T, error), ms int, message string) (T, error) {
	var zero T

	ctx, cancel := context.WithTimeout(ctx, time.Duration(ms)*time.Millisecond)
	defer cancel()

	resultCh := make(chan T, 1)
	errCh := make(chan error, 1)

	go func() {
		result, err := promise(ctx)
		if err != nil {
			errCh <- err
		} else {
			resultCh <- result
		}
	}()

	select {
	case result := <-resultCh:
		return result, nil
	case err := <-errCh:
		return zero, err
	case <-ctx.Done():
		return zero, errors.New(message)
	}
}
