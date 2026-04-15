// cmd/claude/main.go
// Source: TypeScript CLI entry point
// Refactor: Go entry point using the Cobra framework

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
