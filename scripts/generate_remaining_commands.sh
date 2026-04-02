#!/bin/bash
# 批量命令生成脚本
# 使用方式: ./scripts/generate_remaining_commands.sh

COMMANDS_DIR="cmd/claude/commands"

# 剩余命令列表（166个）
COMMANDS=(
  # 系统命令
  "chmod:Change file permissions:CategoryFiles"
  "chown:Change file owner:CategoryFiles"  
  "ps:Process status:CategoryTools"
  "top:System processes:CategoryTools"
  "kill:Terminate processes:CategoryTools"
  "df:Disk free:CategoryTools"
  "du:Disk usage:CategoryTools"
  "free:Memory usage:CategoryTools"
  "uptime:System uptime:CategoryTools"
  "uname:System info:CategoryTools"
  
  # 网络
  "wget:Download files:CategoryTools"
  "ping:Network test:CategoryTools"
  "netstat:Network status:CategoryTools"
  "ifconfig:Network interfaces:CategoryTools"
  "ssh:SSH client:CategoryTools"
  "scp:Secure copy:CategoryTools"
  "ftp:FTP client:CategoryTools"
  
  # 压缩
  "tar:Archive files:CategoryFiles"
  "zip:Compress files:CategoryFiles"
  "unzip:Extract zip:CategoryFiles"
  "gzip:Compress with gzip:CategoryFiles"
  "gunzip:Decompress gzip:CategoryFiles"
  "bzip2:Compress with bzip2:CategoryFiles"
  
  # 开发工具
  "npm:Node package manager:CategoryTools"
  "yarn:Yarn package manager:CategoryTools"
  "pip:Python package manager:CategoryTools"
  "go:Go toolchain:CategoryTools"
  "docker:Docker CLI:CategoryTools"
  "kubectl:Kubernetes CLI:CategoryTools"
  "git:Git version control:CategoryTools"
  "make:Build automation:CategoryTools"
  
  # 编辑器
  "nano:Nano editor:CategoryFiles"
  "vim:Vim editor:CategoryFiles"
  "code:VS Code:CategoryFiles"
  "open:Open files:CategoryFiles"
  
  # 其他工具
  "bc:Calculator:CategoryTools"
  "jq:JSON processor:CategoryTools"
  "awk:Text processing:CategoryTools"
  "sed:Stream editor:CategoryTools"
  "tr:Translate chars:CategoryTools"
  "cut:Cut fields:CategoryTools"
  "paste:Merge lines:CategoryTools"
  "join:Join lines:CategoryTools"
  "split:Split files:CategoryFiles"
  "csplit:Context split:CategoryFiles"
  
  # 更多系统
  "watch:Execute periodically:CategoryTools"
  "xargs:Build commands:CategoryTools"
  "parallel:Parallel execution:CategoryTools"
  "tmux:Terminal multiplexer:CategoryTools"
  "screen:Screen manager:CategoryTools"
  "cron:Schedule tasks:CategoryTools"
  "at:Schedule command:CategoryTools"
  "logger:System logger:CategoryTools"
)

generate_command() {
  local name=$1
  local desc=$2
  local category=$3
  local file="$COMMANDS_DIR/$name.go"
  
  cat > "$file" << EOF
package commands

import (
	"context"
	"os"
	"os/exec"
)

type ${name^}Command struct{ *BaseCommand }

func New${name^}Command() *${name^}Command {
	return &${name^}Command{
		BaseCommand: NewBaseCommand("$name", "$desc", $category),
	}
}

func (c *${name^}Command) Execute(ctx context.Context, args []string) error {
	cmd := exec.Command("$name", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() { Register(New${name^}Command()) }
EOF
  echo "Created: $file"
}

echo "Generating ${#COMMANDS[@]} commands..."

for cmd_info in "${COMMANDS[@]}"; do
  IFS=':' read -r name desc category <<< "$cmd_info"
  generate_command "$name" "$desc" "$category"
done

echo "Done! Total commands: $(ls $COMMANDS_DIR/*.go | wc -l)"
