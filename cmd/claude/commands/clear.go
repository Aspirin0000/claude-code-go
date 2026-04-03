package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// ClearCommand 清除终端屏幕
type ClearCommand struct {
	*BaseCommand
}

// NewClearCommand 创建clear命令
func NewClearCommand() *ClearCommand {
	return &ClearCommand{
		BaseCommand: NewBaseCommand(
			"clear",
			"清除终端屏幕",
			CategoryGeneral,
		).WithAliases("cls", "clr").
			WithHelp(`使用: /clear

清除终端屏幕内容。

别名: /cls, /clr`),
	}
}

// Execute 执行清除屏幕操作
func (c *ClearCommand) Execute(ctx context.Context, args []string) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// 如果系统命令失败，使用ANSI转义码作为后备方案
		fmt.Print("\033[2J\033[H")
		return nil
	}

	return nil
}

func init() { Register(NewClearCommand()) }
