// Package main 提供代码简化检测
// 此脚本自动扫描所有生成的 Go 文件，检测简化行为
package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// 禁止词列表
var forbiddenWords = []string{
	"TODO",
	"FIXME",
	"stub",
	"placeholder",
	"simplified",
	"简化",
	"占位",
}

func main() {
	fmt.Println("🔍 扫描简化行为...")

	violations := 0

	// 遍历所有 Go 文件
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过 vendor 和 .git 目录
		if strings.Contains(path, "vendor/") || strings.Contains(path, ".git/") {
			return nil
		}

		// 只检查 .go 文件
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// 跳过此脚本自身
		if strings.Contains(path, "verify_no_simplification.go") {
			return nil
		}

		// 检查文件
		fileViolations := checkFile(path)
		violations += fileViolations

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ 扫描失败: %v\n", err)
		os.Exit(1)
	}

	if violations > 0 {
		fmt.Printf("\n❌ 发现 %d 处简化违规\n", violations)
		fmt.Println("请删除简化代码，补充完整实现")
		os.Exit(1)
	}

	fmt.Println("\n✅ 零简化政策验证通过")
}

// checkFile 检查单个文件
func checkFile(path string) int {
	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开文件 %s: %v\n", path, err)
		return 0
	}
	defer file.Close()

	violations := 0
	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// 检查禁止词
		for _, word := range forbiddenWords {
			if strings.Contains(line, word) {
				fmt.Printf("❌ %s:%d - 发现禁止词: %s\n", path, lineNum, word)
				fmt.Printf("   %s\n", strings.TrimSpace(line))
				violations++
			}
		}

		// 检查空函数体（简化特征）
		// 注意：检查行尾是否以 "{}" 结束，避免误判 interface{}
		trimmed := strings.TrimSpace(line)
		if strings.Contains(line, "func ") && strings.HasSuffix(trimmed, "{}") {
			fmt.Printf("⚠️  %s:%d - 发现空函数体\n", path, lineNum)
			fmt.Printf("   %s\n", trimmed)
			violations++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "读取文件 %s 失败: %v\n", path, err)
	}

	return violations
}
