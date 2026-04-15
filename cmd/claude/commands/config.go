package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// ConfigCommand manages configuration settings
type ConfigCommand struct {
	*BaseCommand
}

// ConfigMeta contains metadata about configuration keys
type ConfigMeta struct {
	Key         string
	Type        string
	Default     interface{}
	Description string
	Nested      bool
}

// ConfigDefinitions defines all available configuration keys
var ConfigDefinitions = []ConfigMeta{
	{Key: "api_key", Type: "string", Default: "", Description: "Anthropic API key"},
	{Key: "model", Type: "string", Default: "claude-sonnet-4-20250514", Description: "Default AI model"},
	{Key: "theme", Type: "string", Default: "dark", Description: "UI theme (dark/light)"},
	{Key: "verbose", Type: "bool", Default: false, Description: "Enable verbose logging"},
	{Key: "provider", Type: "string", Default: "anthropic", Description: "API provider"},
	{Key: "auto_save", Type: "bool", Default: true, Description: "Enable automatic session saving"},
	{Key: "auto_save_dir", Type: "string", Default: "", Description: "Custom auto-save directory"},
}

// NewConfigCommand creates a new config command
func NewConfigCommand() *ConfigCommand {
	return &ConfigCommand{
		BaseCommand: NewBaseCommand(
			"config",
			"管理配置设置",
			CategoryConfig,
		).WithAliases("settings", "cfg").
			WithHelp(`使用: /config [subcommand] [args...]

管理Claude Code的配置设置。

子命令:
  get <key>       - 获取指定配置项的值
  set <key> <val> - 设置配置项的值
  list            - 列出所有可用的配置键
  (无参数)        - 显示所有当前配置

支持的配置键:
  api_key         - API密钥
  model           - 默认AI模型
  theme           - 界面主题 (dark/light)
  verbose         - 启用详细日志输出 (true/false)
  provider        - API提供商
  auto_save       - 启用自动保存 (true/false)
  auto_save_dir   - 自动保存目录路径

使用点符号访问嵌套配置:
  /config get mcp.timeout
  /config set mcp.timeout 30

别名: /settings, /cfg`),
	}
}

// Execute runs the config command
func (c *ConfigCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.showAllConfig()
	}

	subcommand := strings.ToLower(strings.TrimSpace(args[0]))

	switch subcommand {
	case "get":
		if len(args) < 2 {
			fmt.Println("❌ 错误: 请提供配置键名")
			fmt.Println("用法: /config get <key>")
			return nil
		}
		return c.getConfig(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("❌ 错误: 请提供配置键名和值")
			fmt.Println("用法: /config set <key> <value>")
			return nil
		}
		return c.setConfig(args[1], strings.Join(args[2:], " "))
	case "list":
		return c.listConfigKeys()
	default:
		fmt.Printf("❌ 错误: 未知子命令 '%s'\n", subcommand)
		fmt.Println("\n可用子命令:")
		fmt.Println("  get <key>       - 获取配置值")
		fmt.Println("  set <key> <val> - 设置配置值")
		fmt.Println("  list            - 列出所有配置键")
		return nil
	}
}

// showAllConfig displays all current configuration
func (c *ConfigCommand) showAllConfig() error {
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	defaultCfg := config.DefaultConfig()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              当前配置 (Current Configuration)             ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("📁 配置文件: %s\n", configPath)
	fmt.Println()

	// Basic settings
	fmt.Println("🔧 基本设置:")
	fmt.Println("  " + strings.Repeat("─", 50))
	c.printConfigLine("api_key", c.maskAPIKey(cfg.APIKey), c.maskAPIKey(defaultCfg.APIKey), cfg.APIKey != defaultCfg.APIKey)
	c.printConfigLine("model", cfg.Model, defaultCfg.Model, cfg.Model != defaultCfg.Model)
	c.printConfigLine("theme", cfg.Theme, defaultCfg.Theme, cfg.Theme != defaultCfg.Theme)
	c.printConfigLine("verbose", fmt.Sprintf("%v", cfg.Verbose), fmt.Sprintf("%v", defaultCfg.Verbose), cfg.Verbose != defaultCfg.Verbose)
	c.printConfigLine("provider", cfg.Provider, defaultCfg.Provider, cfg.Provider != defaultCfg.Provider)
	c.printConfigLine("auto_save", fmt.Sprintf("%v", cfg.AutoSave), fmt.Sprintf("%v", defaultCfg.AutoSave), cfg.AutoSave != defaultCfg.AutoSave)
	autoSaveDir := cfg.AutoSaveDir
	if autoSaveDir == "" {
		autoSaveDir = "(默认)"
	}
	c.printConfigLine("auto_save_dir", autoSaveDir, "", cfg.AutoSaveDir != "")

	// Environment variables
	if len(cfg.Env) > 0 {
		fmt.Println()
		fmt.Println("🌍 环境变量:")
		fmt.Println("  " + strings.Repeat("─", 50))
		for key, value := range cfg.Env {
			fmt.Printf("  %-20s = %s\n", key, value)
		}
	}

	// Projects
	if len(cfg.Projects) > 0 {
		fmt.Println()
		fmt.Printf("📂 项目配置 (%d个项目):\n", len(cfg.Projects))
		for projectPath := range cfg.Projects {
			fmt.Printf("  • %s\n", projectPath)
		}
	}

	fmt.Println()
	fmt.Println("💡 提示: 使用 /config get <key> 查看特定配置")
	fmt.Println("        使用 /config set <key> <value> 修改配置")
	fmt.Println("        使用 /config list 查看所有可用配置键")
	fmt.Println()

	return nil
}

// printConfigLine prints a configuration line with modification indicator
func (c *ConfigCommand) printConfigLine(key, value, defaultVal string, modified bool) {
	indicator := " "
	if modified {
		indicator = "●"
	}

	displayValue := value
	if value == "" {
		displayValue = "(未设置)"
	}

	fmt.Printf("  %s %-18s %-25s (默认: %s)\n", indicator, key+":", displayValue, defaultVal)
}

// maskAPIKey masks the API key for display
func (c *ConfigCommand) maskAPIKey(key string) string {
	if key == "" {
		return ""
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// getConfig retrieves a specific configuration value
func (c *ConfigCommand) getConfig(key string) error {
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	defaultCfg := config.DefaultConfig()

	// Handle nested keys (e.g., "mcp.timeout")
	if strings.Contains(key, ".") {
		return c.getNestedConfig(cfg, key)
	}

	// Get basic config value
	value, err := c.getBasicConfigValue(cfg, key)
	if err != nil {
		return err
	}

	defaultValue, _ := c.getBasicConfigValue(defaultCfg, key)

	fmt.Println()
	fmt.Printf("🔑 配置键: %s\n", key)
	fmt.Println("  " + strings.Repeat("─", 40))

	// Find config metadata
	meta := c.findConfigMeta(key)
	if meta != nil {
		fmt.Printf("  描述: %s\n", meta.Description)
		fmt.Printf("  类型: %s\n", meta.Type)
	}

	fmt.Printf("  当前值: %v\n", value)
	fmt.Printf("  默认值: %v\n", defaultValue)

	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", defaultValue) {
		fmt.Println("  状态: ● 已修改")
	} else {
		fmt.Println("  状态: 使用默认值")
	}

	fmt.Println()
	return nil
}

// getBasicConfigValue gets a basic (non-nested) config value using reflection
func (c *ConfigCommand) getBasicConfigValue(cfg *config.Config, key string) (interface{}, error) {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		jsonName := strings.Split(jsonTag, ",")[0]

		if jsonName == key {
			return v.Field(i).Interface(), nil
		}
	}

	return nil, fmt.Errorf("未知配置键: %s", key)
}

// getNestedConfig handles nested configuration keys
func (c *ConfigCommand) getNestedConfig(cfg *config.Config, key string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("无效的嵌套键格式: %s", key)
	}

	switch parts[0] {
	case "mcp":
		return c.getMCPConfig(cfg, parts[1:])
	case "project":
		return c.getProjectConfig(cfg, parts[1:])
	case "env":
		if len(parts) == 2 {
			value, exists := cfg.Env[parts[1]]
			if !exists {
				fmt.Printf("\n🔑 配置键: %s\n", key)
				fmt.Println("  状态: 未设置")
				fmt.Println()
				return nil
			}
			fmt.Println()
			fmt.Printf("🔑 配置键: %s\n", key)
			fmt.Printf("  值: %s\n", value)
			fmt.Println()
			return nil
		}
		return fmt.Errorf("无效的 env 键格式: %s", key)
	default:
		return fmt.Errorf("不支持的配置命名空间: %s", parts[0])
	}
}

// getMCPConfig retrieves MCP server configuration
func (c *ConfigCommand) getMCPConfig(cfg *config.Config, parts []string) error {
	if len(parts) == 0 {
		fmt.Println()
		fmt.Println("🔌 MCP 服务器配置:")
		fmt.Println("  " + strings.Repeat("─", 40))

		currentProject := c.getCurrentProject()
		projectCfg := cfg.GetProjectConfig(currentProject)

		if len(projectCfg.MCPServers) == 0 {
			fmt.Println("  (无配置的MCP服务器)")
		} else {
			for name, serverCfg := range projectCfg.MCPServers {
				fmt.Printf("  • %s:\n", name)
				fmt.Printf("    命令: %s\n", serverCfg.Command)
				if len(serverCfg.Args) > 0 {
					fmt.Printf("    参数: %s\n", strings.Join(serverCfg.Args, " "))
				}
			}
		}
		fmt.Println()
		return nil
	}

	// Get specific MCP server config
	serverName := parts[0]
	currentProject := c.getCurrentProject()
	projectCfg := cfg.GetProjectConfig(currentProject)

	serverCfg, exists := projectCfg.MCPServers[serverName]
	if !exists {
		return fmt.Errorf("MCP服务器 '%s' 未找到", serverName)
	}

	fmt.Println()
	fmt.Printf("🔌 MCP 服务器: %s\n", serverName)
	fmt.Println("  " + strings.Repeat("─", 40))
	fmt.Printf("  命令: %s\n", serverCfg.Command)
	if len(serverCfg.Args) > 0 {
		fmt.Printf("  参数: %v\n", serverCfg.Args)
	}
	if len(serverCfg.Env) > 0 {
		fmt.Println("  环境变量:")
		for k, v := range serverCfg.Env {
			fmt.Printf("    %s = %s\n", k, v)
		}
	}
	fmt.Println()

	return nil
}

// getProjectConfig retrieves project configuration
func (c *ConfigCommand) getProjectConfig(cfg *config.Config, parts []string) error {
	if len(parts) < 1 {
		return fmt.Errorf("请指定项目路径")
	}

	projectPath := parts[0]
	projectCfg := cfg.GetProjectConfig(projectPath)

	fmt.Println()
	fmt.Printf("📂 项目配置: %s\n", projectPath)
	fmt.Println("  " + strings.Repeat("─", 40))

	if len(projectCfg.AllowedTools) > 0 {
		fmt.Printf("  允许的工具: %s\n", strings.Join(projectCfg.AllowedTools, ", "))
	} else {
		fmt.Println("  允许的工具: (所有工具)")
	}

	if len(projectCfg.MCPServers) > 0 {
		fmt.Printf("  MCP服务器: %d个\n", len(projectCfg.MCPServers))
		for name := range projectCfg.MCPServers {
			fmt.Printf("    • %s\n", name)
		}
	}

	fmt.Println()
	return nil
}

// setConfig sets a configuration value
func (c *ConfigCommand) setConfig(key, value string) error {
	// Validate key exists in definitions
	meta := c.findConfigMeta(key)
	if meta == nil && !strings.Contains(key, ".") {
		fmt.Printf("❌ 错误: 未知配置键 '%s'\n", key)
		fmt.Println("\n可用配置键:")
		c.listConfigKeys()
		return fmt.Errorf("unknown config key: %s", key)
	}

	// Validate value type
	if meta != nil {
		if err := c.validateValueType(value, meta.Type); err != nil {
			return fmt.Errorf("值类型错误: %w", err)
		}
	}

	// Validate auto_save_dir path
	if key == "auto_save_dir" && value != "" {
		if err := c.validateAutoSaveDir(value); err != nil {
			return err
		}
	}

	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	// Handle nested keys
	if strings.Contains(key, ".") {
		return c.setNestedConfig(cfg, key, value)
	}

	// Set basic config value
	if err := c.setBasicConfigValue(cfg, key, value); err != nil {
		return err
	}

	// Save config
	if err := cfg.Save(configPath); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ 配置已更新")
	fmt.Printf("   %s = %s\n", key, value)
	fmt.Println()

	return nil
}

// setBasicConfigValue sets a basic config value using reflection
func (c *ConfigCommand) setBasicConfigValue(cfg *config.Config, key, value string) error {
	v := reflect.ValueOf(cfg).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		jsonTag := field.Tag.Get("json")
		jsonName := strings.Split(jsonTag, ",")[0]

		if jsonName == key {
			fieldValue := v.Field(i)
			return c.setFieldValue(fieldValue, value)
		}
	}

	return fmt.Errorf("未知配置键: %s", key)
}

// setFieldValue sets a field value based on its type
func (c *ConfigCommand) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("无效的布尔值: %s", value)
		}
		field.SetBool(boolValue)
	case reflect.Int, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("无效的整数值: %s", value)
		}
		field.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("无效的浮点值: %s", value)
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("不支持的配置类型: %s", field.Kind())
	}
	return nil
}

// setNestedConfig handles setting nested configuration
func (c *ConfigCommand) setNestedConfig(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("无效的嵌套键格式: %s", key)
	}

	switch parts[0] {
	case "env":
		if len(parts) != 2 {
			return fmt.Errorf("无效的 env 键格式: %s", key)
		}
		if cfg.Env == nil {
			cfg.Env = make(map[string]string)
		}
		cfg.Env[parts[1]] = value

		configPath := config.GetConfigPath()
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("保存配置失败: %w", err)
		}

		fmt.Println()
		fmt.Println("✅ 环境变量已设置")
		fmt.Printf("   %s = %s\n", key, value)
		fmt.Println()
		return nil

	default:
		return fmt.Errorf("不支持的配置命名空间或嵌套配置 '%s' 暂不支持修改", parts[0])
	}
}

// validateAutoSaveDir validates the auto_save_dir path
func (c *ConfigCommand) validateAutoSaveDir(path string) error {
	if path == "" {
		return nil
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("auto_save_dir 必须是绝对路径，不能是相对路径: %s", path)
	}

	// Check if parent directory exists
	parent := filepath.Dir(path)
	info, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("无法访问父目录 '%s': %w", parent, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' 不是目录", parent)
	}

	// Try to create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check if we can create it
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("无法创建自动保存目录 '%s': %w", path, err)
		}
		// Clean up the test directory if it was created empty
		if dir, _ := os.ReadDir(path); len(dir) == 0 {
			os.Remove(path)
		}
	}

	return nil
}
func (c *ConfigCommand) validateValueType(value, expectedType string) error {
	switch expectedType {
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("'%s' 不是有效的布尔值 (请使用: true/false, yes/no, 1/0)", value)
		}
	case "int":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("'%s' 不是有效的整数值", value)
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("'%s' 不是有效的数值", value)
		}
	}
	return nil
}

// listConfigKeys displays all available configuration keys
func (c *ConfigCommand) listConfigKeys() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           可用配置键 (Available Configuration Keys)        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🔧 基本配置:")
	fmt.Println("  " + strings.Repeat("─", 50))
	for _, meta := range ConfigDefinitions {
		displayDefault := fmt.Sprintf("%v", meta.Default)
		if displayDefault == "" {
			displayDefault = "(空)"
		}
		fmt.Printf("  %-18s %-12s 默认: %-15s %s\n",
			meta.Key,
			"["+meta.Type+"]",
			displayDefault,
			meta.Description,
		)
	}

	fmt.Println()
	fmt.Println("🔌 嵌套配置 (使用点符号访问):")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Println("  env.<name>         [string]    环境变量")
	fmt.Println("  mcp.<server>       [object]    MCP服务器配置")
	fmt.Println("  project.<path>     [object]    项目特定配置")

	fmt.Println()
	fmt.Println("💡 用法示例:")
	fmt.Println("  /config get model           - 查看当前模型")
	fmt.Println("  /config set model opus      - 切换模型")
	fmt.Println("  /config set verbose true    - 启用详细模式")
	fmt.Println("  /config set env.API_KEY xxx - 设置环境变量")
	fmt.Println()

	return nil
}

// findConfigMeta finds configuration metadata by key
func (c *ConfigCommand) findConfigMeta(key string) *ConfigMeta {
	for _, meta := range ConfigDefinitions {
		if meta.Key == key {
			return &meta
		}
	}
	return nil
}

// getCurrentProject returns the current project path
func (c *ConfigCommand) getCurrentProject() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return cwd
}

func init() {
	// Register the config command
	Register(NewConfigCommand())
}
