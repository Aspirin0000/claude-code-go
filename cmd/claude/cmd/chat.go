// cmd/claude/cmd/chat.go
// 来源: src/screens/REPL.tsx + src/query.ts
// 重构: 交互式对话和管道模式实现

package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/cmd/claude/commands"
	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	apiKeyFlag  string
	modelFlag   string
	verboseFlag bool
	promptFlag  string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Anthropic API key")
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "AI model to use")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&promptFlag, "prompt", "p", "", "Initial prompt to send")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "claude",
	Short: "Claude Code - AI-powered coding assistant",
	Long: `Claude Code is an AI-powered coding assistant that helps you:
  - Navigate and understand codebases
  - Write and edit code
  - Execute terminal commands
  - Debug and fix issues
  - Integrate with MCP servers

This is an unofficial Go implementation based on Claude Code v2.1.88.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCLI()
	},
}

// Execute adds all child commands to the root command
func Execute() error {
	return rootCmd.Execute()
}

// runCLI runs the CLI interface (either TUI or simple REPL)
func runCLI() error {
	// Check if we should use TUI or simple REPL
	if os.Getenv("CLAUDE_TUI") == "1" {
		return runTUI()
	}
	return runSimpleREPL()
}

// runTUI runs the Bubble Tea TUI interface
func runTUI() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}

// runSimpleREPL runs a simple command-line REPL
func runSimpleREPL() error {
	ctx := context.Background()

	// Load configuration
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	// Initialize session storage
	state.InitSessionStorage(cfg)

	// Setup API key
	apiKey := apiKeyFlag
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	// Setup model
	model := modelFlag
	if model == "" {
		model = cfg.Model
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	// Initialize components
	toolRegistry := tools.NewDefaultRegistry()
	var client *api.Client
	if apiKey != "" {
		client = api.NewClient(apiKey, model)
	}

	// Print welcome
	printWelcome()

	// Handle initial prompt if provided
	if promptFlag != "" {
		fmt.Printf("\n📝 Initial prompt: %s\n\n", promptFlag)
		if err := processUserMessage(ctx, client, toolRegistry, promptFlag); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	// REPL loop
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\n❯ ")

		input, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("\n👋 Goodbye!")
				return nil
			}
			return fmt.Errorf("read error: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Exit commands
		if input == "exit" || input == "quit" || input == "/exit" {
			fmt.Println("👋 Goodbye!")
			return nil
		}

		// Slash commands
		if strings.HasPrefix(input, "/") {
			if err := handleSlashCommand(ctx, input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}

		// Process message with AI
		if err := processUserMessage(ctx, client, toolRegistry, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║              Claude Code Go - v2.1.88                   ║")
	fmt.Println("║         AI-Powered Coding Assistant (Unofficial)        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Type your message and press Enter to chat with Claude.")
	fmt.Println("Use /help to see available commands, or /exit to quit.")
	fmt.Println()
}

func handleSlashCommand(ctx context.Context, input string) error {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return nil
	}

	cmdName := strings.TrimPrefix(parts[0], "/")
	args := parts[1:]

	cmd, found := commands.GetRegistry().Get(cmdName)
	if !found {
		return fmt.Errorf("unknown command: %s (use /help for available commands)", cmdName)
	}

	return cmd.Execute(ctx, args)
}

func processUserMessage(ctx context.Context, client *api.Client, registry *tools.Registry, message string) error {
	if client == nil {
		fmt.Println("⚠️  AI client not initialized. Please configure API key.")
		return nil
	}

	// Add user message to state
	userMsg := state.Message{
		Type:    "user",
		Role:    "user",
		Content: message,
	}
	state.GlobalState.AddMessage(userMsg)

	// Get conversation history
	messages := state.GlobalState.GetMessages()

	// Convert to API format
	apiMessages := make([]api.Message, 0, len(messages))
	for _, msg := range messages {
		role := msg.Role
		if role == "" {
			role = msg.Type
		}
		apiMessages = append(apiMessages, api.Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Get tools
	toolSchemas := registry.GetToolSchemas()
	toolsList := make([]api.Tool, 0, len(toolSchemas))
	for _, schema := range toolSchemas {
		name, _ := schema["name"].(string)
		desc, _ := schema["description"].(string)
		inputSchema := schema["input_schema"]

		schemaJSON, _ := json.Marshal(inputSchema)
		toolsList = append(toolsList, api.Tool{
			Name:        name,
			Description: desc,
			InputSchema: schemaJSON,
		})
	}

	// Call AI
	fmt.Println("🤔 Thinking...")

	resp, err := client.ChatWithBlocks(ctx, apiMessages, toolsList)
	if err != nil {
		return fmt.Errorf("API error: %w", err)
	}

	// Process response
	var responseContent strings.Builder
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			responseContent.WriteString(block.Text)

		case "tool_use":
			fmt.Printf("\n🔧 Using tool: %s\n", block.Name)

			result, err := registry.Call(ctx, block.Name, block.Input)
			if err != nil {
				fmt.Printf("   Error: %v\n", err)
				responseContent.WriteString(fmt.Sprintf("\n[Tool %s failed: %v]\n", block.Name, err))
			} else if !result.Success {
				fmt.Printf("   Failed: %s\n", result.Error)
				responseContent.WriteString(fmt.Sprintf("\n[Tool %s failed: %s]\n", block.Name, result.Error))
			} else {
				fmt.Printf("   Success\n")
				resultMsg := state.Message{
					Type:    "user",
					Role:    "user",
					Content: fmt.Sprintf("Tool %s result: %s", block.Name, string(result.Data)),
				}
				state.GlobalState.AddMessage(resultMsg)
			}
		}
	}

	// Add AI response to state and display
	if responseContent.Len() > 0 {
		assistantMsg := state.Message{
			Type:    "assistant",
			Role:    "assistant",
			Content: responseContent.String(),
		}
		state.GlobalState.AddMessage(assistantMsg)
		fmt.Printf("\n🤖 %s\n", responseContent.String())
	}

	if state.GlobalSessionStorage != nil {
		_ = state.GlobalSessionStorage.AutoSave(state.GlobalState)
	}

	return nil
}

// 样式定义 (对应原 TS 中的 Chalk/Kleur 样式)
var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Background(lipgloss.Color("#1a1a2e")).Padding(1, 2).Width(50)
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CCFF"))
	systemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	inputStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#2d2d44")).Padding(0, 1)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
)

// Message 消息模型 (对应原 TS 中的 Message 类型)
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// App TUI 应用模型 (对应原 TS 中的 REPL 组件状态)
type App struct {
	config       *config.Config
	apiClient    *api.Client
	toolRegistry *tools.Registry
	messages     []Message
	input        string
	loading      bool
	width        int
	height       int
}

// runInteractive 运行交互式 TUI 模式
// 对应原 TS: src/screens/REPL.tsx - 启动交互式界面
func runInteractive() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("启动 TUI 失败: %w", err)
	}
	return nil
}

// runSingleShot 运行单条命令模式
// 对应原 TS: pipe 模式中的单条提示处理
func runSingleShot(prompt string) error {
	app := NewApp()

	// 构建消息
	messages := []api.Message{{Role: "user", Content: prompt}}

	// 获取工具
	toolsList := buildToolsList(app.toolRegistry)

	// 调用 API
	resp, err := app.apiClient.Chat(context.Background(), messages, toolsList)
	if err != nil {
		return fmt.Errorf("API 调用失败: %w", err)
	}

	fmt.Println(resp.Content)
	return nil
}

// runPipeMode 运行管道模式
// 对应原 TS: -p/--print 模式，从 stdin 读取输入
func runPipeMode(input string) error {
	return runSingleShot(input)
}

// NewApp 创建新的应用实例
// 对应原 TS: REPL 组件的初始化逻辑
func NewApp() *App {
	// 加载配置
	cfg := config.DefaultConfig()

	// 从环境变量读取配置
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		cfg.Model = model
	}
	if provider := os.Getenv("CLAUDE_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	// 创建 API 客户端
	client := api.NewClient(cfg.APIKey, cfg.Model)
	client.SetProvider(cfg.Provider)

	return &App{
		config:       cfg,
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
		messages:     make([]Message, 0),
	}
}

// Init 初始化 TUI
func (a *App) Init() tea.Cmd {
	return tea.EnterAltScreen
}

// Update 处理消息和输入
// 对应原 TS: REPL 组件的事件处理逻辑
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return a, tea.Quit
		case tea.KeyEnter:
			if a.input != "" && !a.loading {
				return a.handleInput()
			}
		case tea.KeyBackspace:
			if len(a.input) > 0 {
				a.input = a.input[:len(a.input)-1]
			}
		case tea.KeyRunes:
			a.input += string(msg.Runes)
		}
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case apiResponseMsg:
		a.loading = false
		if msg.err != nil {
			a.messages = append(a.messages, Message{Role: "system", Content: fmt.Sprintf("错误: %v", msg.err)})
		} else {
			a.messages = append(a.messages, Message{Role: "assistant", Content: msg.content})
		}
		a.input = ""
	}
	return a, nil
}

// View 渲染界面
// 对应原 TS: REPL 组件的渲染逻辑
func (a *App) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 Claude Code Go v0.1.0"))
	b.WriteString("\n\n")

	// 显示消息（限制行数）
	startIdx := 0
	if len(a.messages) > a.height-10 {
		startIdx = len(a.messages) - (a.height - 10)
	}

	for i := startIdx; i < len(a.messages); i++ {
		msg := a.messages[i]
		switch msg.Role {
		case "user":
			b.WriteString(userStyle.Render("你: ") + msg.Content)
		case "assistant":
			b.WriteString(assistantStyle.Render("Claude: ") + msg.Content)
		case "system":
			b.WriteString(systemStyle.Render(msg.Content))
		}
		b.WriteString("\n\n")
	}

	if a.loading {
		b.WriteString(assistantStyle.Render("思考中... ⏳") + "\n\n")
	}

	b.WriteString(inputStyle.Render("> "+a.input+"█") + "\n")
	b.WriteString(helpStyle.Render("Ctrl+C: 退出 | Enter: 发送"))
	return b.String()
}

// handleInput 处理用户输入
func (a *App) handleInput() (tea.Model, tea.Cmd) {
	a.messages = append(a.messages, Message{Role: "user", Content: a.input})
	a.loading = true
	return a, a.processMessage()
}

// apiResponseMsg API 响应消息
type apiResponseMsg struct {
	content string
	err     error
}

// processMessage 处理消息（包括工具调用）
// 对应原 TS: query.ts 中的主查询逻辑
func (a *App) processMessage() tea.Cmd {
	return func() tea.Msg {
		// 转换消息格式
		apiMessages := make([]api.Message, len(a.messages))
		for i, msg := range a.messages {
			apiMessages[i] = api.Message{Role: msg.Role, Content: msg.Content}
		}

		// 获取工具列表
		toolsList := buildToolsList(a.toolRegistry)

		// 调用 API
		resp, err := a.apiClient.ChatWithBlocks(context.Background(), apiMessages, toolsList)
		if err != nil {
			return apiResponseMsg{err: err}
		}

		// 处理响应
		result := processResponse(resp, a.toolRegistry)
		return apiResponseMsg{content: result}
	}
}

// buildToolsList 构建工具列表
func buildToolsList(registry *tools.Registry) []api.Tool {
	toolSchemas := registry.GetToolSchemas()
	toolsList := make([]api.Tool, len(toolSchemas))
	for i, schema := range toolSchemas {
		schemaJSON, _ := json.Marshal(schema["input_schema"])
		toolsList[i] = api.Tool{
			Name:        schema["name"].(string),
			Description: schema["description"].(string),
			InputSchema: schemaJSON,
		}
	}
	return toolsList
}

// processResponse 处理 API 响应（包括工具调用）
// 对应原 TS: query.ts 中的响应处理逻辑
func processResponse(resp *api.Response, registry *tools.Registry) string {
	var result strings.Builder

	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			result.WriteString(block.Text)

		case "tool_use":
			result.WriteString(fmt.Sprintf("\n🔧 执行工具: %s\n", block.Name))

			// 执行工具
			toolResult, err := registry.Call(context.Background(), block.Name, block.Input)
			if err != nil {
				result.WriteString(fmt.Sprintf("❌ 工具执行失败: %v\n", err))
			} else {
				// 格式化输出
				formatToolResult(toolResult.Data, &result)
			}
		}
	}

	return result.String()
}

// formatToolResult 格式化工具执行结果
func formatToolResult(data json.RawMessage, result *strings.Builder) {
	var output map[string]interface{}
	json.Unmarshal(data, &output)

	if stdout, ok := output["stdout"].(string); ok && stdout != "" {
		result.WriteString(fmt.Sprintf("📤 输出:\n%s\n", stdout))
	}
	if content, ok := output["content"].(string); ok && content != "" {
		lines := strings.Split(content, "\n")
		if len(lines) > 10 {
			result.WriteString(fmt.Sprintf("📄 内容 (前10行):\n%s\n... (%d more lines)\n",
				strings.Join(lines[:10], "\n"), len(lines)-10))
		} else {
			result.WriteString(fmt.Sprintf("📄 内容:\n%s\n", content))
		}
	}
	if files, ok := output["files"].([]interface{}); ok {
		result.WriteString(fmt.Sprintf("📁 找到 %d 个文件\n", len(files)))
	}
	result.WriteString("✅ 执行成功\n")
}
