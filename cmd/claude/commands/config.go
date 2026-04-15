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
			"Manage configuration settings",
			CategoryConfig,
		).WithAliases("settings", "cfg").
			WithHelp(`Usage: /config [subcommand] [args...]

Manage Claude Code configuration settings.

Subcommands:
  get <key>       - Get the value of a configuration key
  set <key> <val> - Set the value of a configuration key
  list            - List all available configuration keys
  (no args)       - Show all current configuration values

Supported keys:
  api_key         - API key
  model           - Default AI model
  theme           - UI theme (dark/light)
  verbose         - Enable verbose logging (true/false)
  provider        - API provider
  auto_save       - Enable auto-save (true/false)
  auto_save_dir   - Auto-save directory path

Use dot notation for nested configuration:
  /config get mcp.timeout
  /config set mcp.timeout 30

Aliases: /settings, /cfg`),
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
			fmt.Println("❌ Error: please provide a configuration key")
			fmt.Println("Usage: /config get <key>")
			return nil
		}
		return c.getConfig(args[1])
	case "set":
		if len(args) < 3 {
			fmt.Println("❌ Error: please provide a configuration key and value")
			fmt.Println("Usage: /config set <key> <value>")
			return nil
		}
		return c.setConfig(args[1], strings.Join(args[2:], " "))
	case "list":
		return c.listConfigKeys()
	default:
		fmt.Printf("❌ Error: unknown subcommand '%s'\n", subcommand)
		fmt.Println("\nAvailable subcommands:")
		fmt.Println("  get <key>       - Get a configuration value")
		fmt.Println("  set <key> <val> - Set a configuration value")
		fmt.Println("  list            - List all configuration keys")
		return nil
	}
}

// showAllConfig displays all current configuration
func (c *ConfigCommand) showAllConfig() error {
	configPath := config.GetConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	defaultCfg := config.DefaultConfig()

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║                  Current Configuration                   ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Printf("📁 Config file: %s\n", configPath)
	fmt.Println()

	// Basic settings
	fmt.Println("🔧 Basic settings:")
	fmt.Println("  " + strings.Repeat("─", 50))
	c.printConfigLine("api_key", c.maskAPIKey(cfg.APIKey), c.maskAPIKey(defaultCfg.APIKey), cfg.APIKey != defaultCfg.APIKey)
	c.printConfigLine("model", cfg.Model, defaultCfg.Model, cfg.Model != defaultCfg.Model)
	c.printConfigLine("theme", cfg.Theme, defaultCfg.Theme, cfg.Theme != defaultCfg.Theme)
	c.printConfigLine("verbose", fmt.Sprintf("%v", cfg.Verbose), fmt.Sprintf("%v", defaultCfg.Verbose), cfg.Verbose != defaultCfg.Verbose)
	c.printConfigLine("provider", cfg.Provider, defaultCfg.Provider, cfg.Provider != defaultCfg.Provider)
	c.printConfigLine("auto_save", fmt.Sprintf("%v", cfg.AutoSave), fmt.Sprintf("%v", defaultCfg.AutoSave), cfg.AutoSave != defaultCfg.AutoSave)
	autoSaveDir := cfg.AutoSaveDir
	if autoSaveDir == "" {
		autoSaveDir = "(default)"
	}
	c.printConfigLine("auto_save_dir", autoSaveDir, "", cfg.AutoSaveDir != "")

	// Environment variables
	if len(cfg.Env) > 0 {
		fmt.Println()
		fmt.Println("🌍 Environment variables:")
		fmt.Println("  " + strings.Repeat("─", 50))
		for key, value := range cfg.Env {
			fmt.Printf("  %-20s = %s\n", key, value)
		}
	}

	// Projects
	if len(cfg.Projects) > 0 {
		fmt.Println()
		fmt.Printf("📂 Project configuration (%d projects):\n", len(cfg.Projects))
		for projectPath := range cfg.Projects {
			fmt.Printf("  • %s\n", projectPath)
		}
	}

	fmt.Println()
	fmt.Println("💡 Tip: use /config get <key> to view a specific configuration")
	fmt.Println("      use /config set <key> <value> to update a configuration")
	fmt.Println("      use /config list to see all available configuration keys")
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
		displayValue = "(not set)"
	}

	fmt.Printf("  %s %-18s %-25s (default: %s)\n", indicator, key+":", displayValue, defaultVal)
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
		return fmt.Errorf("failed to load configuration: %w", err)
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
	fmt.Printf("🔑 Configuration key: %s\n", key)
	fmt.Println("  " + strings.Repeat("─", 40))

	// Find config metadata
	meta := c.findConfigMeta(key)
	if meta != nil {
		fmt.Printf("  Description: %s\n", meta.Description)
		fmt.Printf("  Type: %s\n", meta.Type)
	}

	fmt.Printf("  Current value: %v\n", value)
	fmt.Printf("  Default value: %v\n", defaultValue)

	if fmt.Sprintf("%v", value) != fmt.Sprintf("%v", defaultValue) {
		fmt.Println("  Status: ● modified")
	} else {
		fmt.Println("  Status: using default")
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

	return nil, fmt.Errorf("unknown configuration key: %s", key)
}

// getNestedConfig handles nested configuration keys
func (c *ConfigCommand) getNestedConfig(cfg *config.Config, key string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid nested key format: %s", key)
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
				fmt.Printf("\n🔑 Configuration key: %s\n", key)
				fmt.Println("  Status: not set")
				fmt.Println()
				return nil
			}
			fmt.Println()
			fmt.Printf("🔑 Configuration key: %s\n", key)
			fmt.Printf("  Value: %s\n", value)
			fmt.Println()
			return nil
		}
		return fmt.Errorf("invalid env key format: %s", key)
	default:
		return fmt.Errorf("unsupported configuration namespace: %s", parts[0])
	}
}

// getMCPConfig retrieves MCP server configuration
func (c *ConfigCommand) getMCPConfig(cfg *config.Config, parts []string) error {
	if len(parts) == 0 {
		fmt.Println()
		fmt.Println("🔌 MCP server configuration:")
		fmt.Println("  " + strings.Repeat("─", 40))

		currentProject := c.getCurrentProject()
		projectCfg := cfg.GetProjectConfig(currentProject)

		if len(projectCfg.MCPServers) == 0 {
			fmt.Println("  (no configured MCP servers)")
		} else {
			for name, serverCfg := range projectCfg.MCPServers {
				fmt.Printf("  • %s:\n", name)
				fmt.Printf("    Command: %s\n", serverCfg.Command)
				if len(serverCfg.Args) > 0 {
					fmt.Printf("    Args: %s\n", strings.Join(serverCfg.Args, " "))
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
		return fmt.Errorf("MCP server '%s' not found", serverName)
	}

	fmt.Println()
	fmt.Printf("🔌 MCP server: %s\n", serverName)
	fmt.Println("  " + strings.Repeat("─", 40))
	fmt.Printf("  Command: %s\n", serverCfg.Command)
	if len(serverCfg.Args) > 0 {
		fmt.Printf("  Args: %v\n", serverCfg.Args)
	}
	if len(serverCfg.Env) > 0 {
		fmt.Println("  Environment:")
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
		return fmt.Errorf("please specify a project path")
	}

	projectPath := parts[0]
	projectCfg := cfg.GetProjectConfig(projectPath)

	fmt.Println()
	fmt.Printf("📂 Project configuration: %s\n", projectPath)
	fmt.Println("  " + strings.Repeat("─", 40))

	if len(projectCfg.AllowedTools) > 0 {
		fmt.Printf("  Allowed tools: %s\n", strings.Join(projectCfg.AllowedTools, ", "))
	} else {
		fmt.Println("  Allowed tools: (all tools)")
	}

	if len(projectCfg.MCPServers) > 0 {
		fmt.Printf("  MCP servers: %d\n", len(projectCfg.MCPServers))
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
		fmt.Printf("❌ Error: unknown configuration key '%s'\n", key)
		fmt.Println("\nAvailable configuration keys:")
		c.listConfigKeys()
		return fmt.Errorf("unknown config key: %s", key)
	}

	// Validate value type
	if meta != nil {
		if err := c.validateValueType(value, meta.Type); err != nil {
			return fmt.Errorf("invalid value type: %w", err)
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
		return fmt.Errorf("failed to load configuration: %w", err)
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
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Println()
	fmt.Println("✅ Configuration updated")
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

	return fmt.Errorf("unknown configuration key: %s", key)
}

// setFieldValue sets a field value based on its type
func (c *ConfigCommand) setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(boolValue)
	case reflect.Int, reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer value: %s", value)
		}
		field.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatValue)
	default:
		return fmt.Errorf("unsupported configuration type: %s", field.Kind())
	}
	return nil
}

// setNestedConfig handles setting nested configuration
func (c *ConfigCommand) setNestedConfig(cfg *config.Config, key, value string) error {
	parts := strings.Split(key, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid nested key format: %s", key)
	}

	switch parts[0] {
	case "env":
		if len(parts) != 2 {
			return fmt.Errorf("invalid env key format: %s", key)
		}
		if cfg.Env == nil {
			cfg.Env = make(map[string]string)
		}
		cfg.Env[parts[1]] = value

		configPath := config.GetConfigPath()
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Println()
		fmt.Println("✅ Environment variable set")
		fmt.Printf("   %s = %s\n", key, value)
		fmt.Println()
		return nil

	default:
		return fmt.Errorf("unsupported configuration namespace or nested configuration '%s' cannot be modified yet", parts[0])
	}
}

// validateAutoSaveDir validates the auto_save_dir path
func (c *ConfigCommand) validateAutoSaveDir(path string) error {
	if path == "" {
		return nil
	}

	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return fmt.Errorf("auto_save_dir must be an absolute path, not a relative path: %s", path)
	}

	// Check if parent directory exists
	parent := filepath.Dir(path)
	info, err := os.Stat(parent)
	if err != nil {
		return fmt.Errorf("unable to access parent directory '%s': %w", parent, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", parent)
	}

	// Try to create the directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Check if we can create it
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("unable to create auto-save directory '%s': %w", path, err)
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
			return fmt.Errorf("'%s' is not a valid boolean (use: true/false, yes/no, 1/0)", value)
		}
	case "int":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("'%s' is not a valid integer", value)
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("'%s' is not a valid number", value)
		}
	}
	return nil
}

// listConfigKeys displays all available configuration keys
func (c *ConfigCommand) listConfigKeys() error {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Available Configuration Keys               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("🔧 Basic configuration:")
	fmt.Println("  " + strings.Repeat("─", 50))
	for _, meta := range ConfigDefinitions {
		displayDefault := fmt.Sprintf("%v", meta.Default)
		if displayDefault == "" {
			displayDefault = "(empty)"
		}
		fmt.Printf("  %-18s %-12s default: %-15s %s\n",
			meta.Key,
			"["+meta.Type+"]",
			displayDefault,
			meta.Description,
		)
	}

	fmt.Println()
	fmt.Println("🔌 Nested configuration (use dot notation):")
	fmt.Println("  " + strings.Repeat("─", 50))
	fmt.Println("  env.<name>         [string]    Environment variable")
	fmt.Println("  mcp.<server>       [object]    MCP server configuration")
	fmt.Println("  project.<path>     [object]    Project-specific configuration")

	fmt.Println()
	fmt.Println("💡 Examples:")
	fmt.Println("  /config get model           - View the current model")
	fmt.Println("  /config set model opus      - Switch models")
	fmt.Println("  /config set verbose true    - Enable verbose mode")
	fmt.Println("  /config set env.API_KEY xxx - Set an environment variable")
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
