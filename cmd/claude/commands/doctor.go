package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DoctorCommand diagnoses system and configuration issues
// Source: src/commands/doctor/
type DoctorCommand struct {
	*BaseCommand
}

// NewDoctorCommand creates the doctor command
func NewDoctorCommand() *DoctorCommand {
	return &DoctorCommand{
		BaseCommand: NewBaseCommand(
			"doctor",
			"诊断系统和配置问题",
			CategoryAdvanced,
		).
			WithHelp(`运行诊断检查，识别常见问题。

检查项目:
- Go 版本
- Git 安装
- 环境变量
- 配置文件
- 网络连接

使用: /doctor`),
	}
}

// Execute runs diagnostics
func (c *DoctorCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println()
	fmt.Println("🔍 运行诊断检查 (Running Diagnostics)")
	fmt.Println("═══════════════════════════════════════")

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{"Go 版本", c.checkGoVersion},
		{"Git 安装", c.checkGit},
		{"环境变量", c.checkEnv},
		{"配置文件", c.checkConfig},
		{"网络连接", c.checkNetwork},
	}

	allPassed := true
	for _, check := range checks {
		passed, msg := check.fn()
		status := "✅"
		if !passed {
			status = "❌"
			allPassed = false
		}
		fmt.Printf("%s %-20s %s\n", status, check.name+":", msg)
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ 所有检查通过！")
	} else {
		fmt.Println("⚠️  发现一些问题，请检查上方输出。")
	}
	fmt.Println()

	return nil
}

func (c *DoctorCommand) checkGoVersion() (bool, string) {
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return false, "未安装"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkGit() (bool, string) {
	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "未安装"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkEnv() (bool, string) {
	required := []string{"HOME", "PATH"}
	missing := []string{}
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}
	if len(missing) > 0 {
		return false, fmt.Sprintf("缺少: %s", strings.Join(missing, ", "))
	}
	return true, "所有必需变量已设置"
}

func (c *DoctorCommand) checkConfig() (bool, string) {
	configDir := getConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, "配置目录不存在（将自动创建）"
	}
	return true, "配置目录存在"
}

func (c *DoctorCommand) checkNetwork() (bool, string) {
	// Simple check: can we resolve a domain?
	cmd := exec.Command("ping", "-c", "1", "-W", "2", "github.com")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "2000", "github.com")
	}
	err := cmd.Run()
	if err != nil {
		return false, "无法连接到网络"
	}
	return true, "网络连接正常"
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude")
}

func init() {
	Register(NewDoctorCommand())
}
