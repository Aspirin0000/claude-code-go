// cmd/claude/main.go
// 来源: TypeScript CLI 入口
// 重构: 使用 Cobra 框架的 Go 入口

package main

import (
	"os"

	"github.com/Aspirin0000/claude-code-go/cmd/claude/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
