// cmd/claude/cmd/chat.go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Aspirin0000/claude-code-go/cmd/claude/commands"
	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
	"github.com/Aspirin0000/claude-code-go/pkg/utils"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"path/filepath"
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

  This is an unofficial Go implementation based on Claude Code ` + commands.TargetVersion + `.`,
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
	if promptFlag != "" {
		return runSingleShot(promptFlag)
	}

	stdinInfo, err := os.Stdin.Stat()
	if err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0 {
		input, readErr := io.ReadAll(os.Stdin)
		if readErr != nil {
			return fmt.Errorf("failed to read piped input: %w", readErr)
		}
		trimmed := strings.TrimSpace(string(input))
		if trimmed != "" {
			return runPipeMode(trimmed)
		}
	}

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

	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load config: %v\n", err)
		cfg = config.DefaultConfig()
	}

	state.InitSessionStorage(cfg)

	mcpManager := mcp.GetGlobalMCPManager()
	if err := mcpManager.Initialize(cfg); err == nil {
		_ = mcpManager.ConnectAll()
	}

	apiKey := apiKeyFlag
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		apiKey = cfg.APIKey
	}

	model := modelFlag
	if model == "" {
		model = cfg.Model
	}
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}

	toolRegistry := tools.NewDefaultRegistry()
	var client *api.Client
	if apiKey != "" {
		client = api.NewClient(apiKey, model)
	}

	printWelcome()

	historyFile := filepath.Join(utils.GetClaudeConfigHomeDir(), "repl_history")

	// Build slash command completer
	var completers []readline.PrefixCompleterInterface
	for _, name := range commands.GetRegistry().ListNames() {
		completers = append(completers, readline.PcItem("/"+name))
	}
	autoComplete := readline.NewPrefixCompleter(completers...)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "\n❯ ",
		HistoryFile:     historyFile,
		AutoComplete:    autoComplete,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(input) > 0 {
				continue
			}
			fmt.Println("\n👋 Goodbye!")
			return nil
		} else if err == io.EOF {
			fmt.Println("\n👋 Goodbye!")
			return nil
		} else if err != nil {
			return fmt.Errorf("read error: %w", err)
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" || input == "/exit" {
			fmt.Println("👋 Goodbye!")
			return nil
		}

		if strings.HasPrefix(input, "/") {
			if err := handleSlashCommand(ctx, input); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
			continue
		}

		if err := processUserMessage(ctx, client, toolRegistry, input); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func printWelcome() {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Printf("║              Claude Code Go - %-25s ║\n", commands.TargetVersion)
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

	state.GlobalState.AddMessage(state.Message{
		Type:    "user",
		Role:    "user",
		Content: message,
	})

	toolsList := buildToolsList(registry)
	const maxRounds = 10

	for round := 0; round < maxRounds; round++ {
		if round == 0 {
			fmt.Println("🤔 Thinking...")
		} else {
			fmt.Println("🤔 Processing tool results...")
		}

		apiMessages := buildAPIMessagesFromState()
		resp, err := client.ChatWithBlocks(ctx, apiMessages, toolsList)
		if err != nil {
			return fmt.Errorf("API error: %w", err)
		}

		var textParts []string
		var toolUses []api.ContentBlock

		for _, block := range resp.Content {
			switch block.Type {
			case "text":
				textParts = append(textParts, block.Text)
			case "tool_use":
				toolUses = append(toolUses, block)
			}
		}

		if len(toolUses) == 0 {
			content := strings.Join(textParts, "\n")
			if content != "" {
				state.GlobalState.AddMessage(state.Message{
					Type:    "assistant",
					Role:    "assistant",
					Content: content,
				})
				fmt.Printf("\n🤖 %s\n", content)
			}
			break
		}

		// Assistant requested tools - add assistant message with blocks
		assistantBlocks := make([]state.ContentBlock, len(resp.Content))
		for i, b := range resp.Content {
			assistantBlocks[i] = state.ContentBlock{
				Type:      b.Type,
				Text:      b.Text,
				ID:        b.ID,
				Name:      b.Name,
				Input:     b.Input,
				ToolUseID: b.ToolUseID,
			}
		}
		state.GlobalState.AddMessage(state.Message{
			Type:    "assistant",
			Role:    "assistant",
			Content: strings.Join(textParts, "\n"),
			Blocks:  assistantBlocks,
		})

		// Execute tools
		ctx = context.WithValue(ctx, tools.APIClientContextKey, client)
		for _, block := range toolUses {
			fmt.Printf("\n🔧 Using tool: %s\n", block.Name)
			resultText, err := executeTool(ctx, registry, block.Name, block.Input)
			if err != nil {
				fmt.Printf("   Error: %v\n", err)
				resultText = fmt.Sprintf("Error: %v", err)
			} else {
				fmt.Printf("   Success\n")
			}

			state.GlobalState.AddMessage(state.Message{
				Type: "user",
				Role: "user",
				Blocks: []state.ContentBlock{
					{
						Type:      "tool_result",
						ToolUseID: block.ID,
						Content:   resultText,
					},
				},
			})
		}
	}

	if state.GlobalSessionStorage != nil {
		_ = state.GlobalSessionStorage.AutoSave(state.GlobalState)
	}

	return nil
}

func buildAPIMessagesFromState() []api.Message {
	messages := state.GlobalState.GetMessages()
	apiMessages := make([]api.Message, 0, len(messages))
	for _, msg := range messages {
		role := msg.Role
		if role == "" {
			role = msg.Type
		}
		apiMsg := api.Message{
			Role:    role,
			Content: msg.Content,
		}
		if len(msg.Blocks) > 0 {
			apiMsg.Blocks = make([]api.ContentBlock, len(msg.Blocks))
			for i, b := range msg.Blocks {
				apiMsg.Blocks[i] = api.ContentBlock{
					Type:      b.Type,
					Text:      b.Text,
					Content:   b.Content,
					ID:        b.ID,
					Name:      b.Name,
					Input:     b.Input,
					ToolUseID: b.ToolUseID,
				}
			}
		}
		apiMessages = append(apiMessages, apiMsg)
	}
	return apiMessages
}

// Styles
var (
	titleStyle     = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Background(lipgloss.Color("#1a1a2e")).Padding(1, 2).Width(50)
	userStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true)
	assistantStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CCFF"))
	systemStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	inputStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#2d2d44")).Padding(0, 1)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
)

// App TUI application model
type App struct {
	config       *config.Config
	apiClient    *api.Client
	toolRegistry *tools.Registry
	input        string
	inputHistory []string
	historyIndex int
	loading      bool
	width        int
	height       int
}

func runInteractive() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}

func runSingleShot(prompt string) error {
	app := NewApp()
	if app.apiClient == nil {
		return fmt.Errorf("AI client not initialized; please configure API key")
	}
	return processUserMessage(context.Background(), app.apiClient, app.toolRegistry, prompt)
}

func runPipeMode(input string) error {
	return runSingleShot(input)
}

func NewApp() *App {
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		cfg.APIKey = apiKey
	}
	if model := os.Getenv("CLAUDE_MODEL"); model != "" {
		cfg.Model = model
	}
	if provider := os.Getenv("CLAUDE_PROVIDER"); provider != "" {
		cfg.Provider = provider
	}

	state.InitSessionStorage(cfg)

	mcpManager := mcp.GetGlobalMCPManager()
	if err := mcpManager.Initialize(cfg); err == nil {
		_ = mcpManager.ConnectAll()
	}

	var client *api.Client
	if cfg.APIKey != "" {
		client = api.NewClient(cfg.APIKey, cfg.Model)
		client.SetProvider(cfg.Provider)
	}

	return &App{
		config:       cfg,
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.EnterAltScreen
}

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
			a.historyIndex = len(a.inputHistory)
		case tea.KeyUp:
			if a.historyIndex > 0 {
				a.historyIndex--
				a.input = a.inputHistory[a.historyIndex]
			}
		case tea.KeyDown:
			if a.historyIndex < len(a.inputHistory) {
				a.historyIndex++
				if a.historyIndex < len(a.inputHistory) {
					a.input = a.inputHistory[a.historyIndex]
				} else {
					a.input = ""
				}
			}
		}
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case apiResponseMsg:
		a.loading = false
		if msg.err != nil {
			state.GlobalState.AddMessage(state.Message{
				Type:    "system",
				Role:    "system",
				Content: fmt.Sprintf("Error: %v", msg.err),
			})
		}
		a.input = ""
	}
	return a, nil
}

func (a *App) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("🚀 Claude Code Go " + commands.TargetVersion))
	b.WriteString("\n\n")

	messages := state.GlobalState.GetMessages()
	startIdx := 0
	if len(messages) > a.height-10 {
		startIdx = len(messages) - (a.height - 10)
	}

	for i := startIdx; i < len(messages); i++ {
		msg := messages[i]
		switch msg.Role {
		case "user":
			if msg.Content == "" && len(msg.Blocks) > 0 {
				b.WriteString(userStyle.Render("You: ") + "[tool results]")
			} else {
				b.WriteString(userStyle.Render("You: ") + msg.Content)
			}
		case "assistant":
			content := msg.Content
			if len(msg.Blocks) > 0 {
				hasToolUse := false
				for _, block := range msg.Blocks {
					if block.Type == "tool_use" {
						hasToolUse = true
						break
					}
				}
				if hasToolUse {
					if content != "" {
						content = content + "\n[using tools...]"
					} else {
						content = "[using tools...]"
					}
				}
			}
			b.WriteString(assistantStyle.Render("Claude: ") + content)
		case "system":
			b.WriteString(systemStyle.Render(msg.Content))
		}
		b.WriteString("\n\n")
	}

	if a.loading {
		b.WriteString(assistantStyle.Render("Thinking... ⏳") + "\n\n")
	}

	b.WriteString(inputStyle.Render("> "+a.input+"█") + "\n")
	b.WriteString(helpStyle.Render("Ctrl+C: Exit | Enter: Send"))
	return b.String()
}

func (a *App) handleInput() (tea.Model, tea.Cmd) {
	state.GlobalState.AddMessage(state.Message{
		Role:    "user",
		Type:    "user",
		Content: a.input,
	})
	a.inputHistory = append(a.inputHistory, a.input)
	a.historyIndex = len(a.inputHistory)
	a.loading = true
	a.input = ""
	return a, a.processMessage()
}

type apiResponseMsg struct {
	content string
	err     error
}

func (a *App) processMessage() tea.Cmd {
	return func() tea.Msg {
		if a.apiClient == nil {
			return apiResponseMsg{err: fmt.Errorf("AI client not initialized; please configure API key")}
		}

		toolsList := buildToolsList(a.toolRegistry)
		const maxRounds = 10

		for round := 0; round < maxRounds; round++ {
			apiMessages := buildAPIMessagesFromState()
			resp, err := a.apiClient.ChatWithBlocks(context.Background(), apiMessages, toolsList)
			if err != nil {
				return apiResponseMsg{err: err}
			}

			var textParts []string
			var toolUses []api.ContentBlock

			for _, block := range resp.Content {
				switch block.Type {
				case "text":
					textParts = append(textParts, block.Text)
				case "tool_use":
					toolUses = append(toolUses, block)
				}
			}

			if len(toolUses) == 0 {
				content := strings.Join(textParts, "\n")
				if content != "" {
					state.GlobalState.AddMessage(state.Message{
						Type:    "assistant",
						Role:    "assistant",
						Content: content,
					})
				}
				return apiResponseMsg{content: content}
			}

			// Assistant requested tools
			assistantBlocks := make([]state.ContentBlock, len(resp.Content))
			for i, b := range resp.Content {
				assistantBlocks[i] = state.ContentBlock{
					Type:      b.Type,
					Text:      b.Text,
					ID:        b.ID,
					Name:      b.Name,
					Input:     b.Input,
					ToolUseID: b.ToolUseID,
				}
			}
			state.GlobalState.AddMessage(state.Message{
				Type:    "assistant",
				Role:    "assistant",
				Content: strings.Join(textParts, "\n"),
				Blocks:  assistantBlocks,
			})

			// Execute tools
			ctx := context.WithValue(context.Background(), tools.APIClientContextKey, a.apiClient)
			for _, block := range toolUses {
				resultText, err := executeTool(ctx, a.toolRegistry, block.Name, block.Input)
				if err != nil {
					resultText = fmt.Sprintf("Error: %v", err)
				}

				state.GlobalState.AddMessage(state.Message{
					Type: "user",
					Role: "user",
					Blocks: []state.ContentBlock{
						{
							Type:      "tool_result",
							ToolUseID: block.ID,
							Content:   resultText,
						},
					},
				})
			}
		}

		return apiResponseMsg{content: "Maximum tool rounds reached"}
	}
}

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

	mcpTools, err := mcp.GetGlobalMCPManager().GetAllTools()
	if err == nil {
		toolsList = append(toolsList, mcpTools...)
	}

	return toolsList
}

func executeTool(ctx context.Context, registry *tools.Registry, toolName string, input json.RawMessage) (string, error) {
	if strings.HasPrefix(toolName, "mcp__") {
		var arguments map[string]interface{}
		if len(input) > 0 {
			if err := json.Unmarshal(input, &arguments); err != nil {
				return "", fmt.Errorf("failed to parse MCP tool arguments: %w", err)
			}
		}
		result, err := mcp.GetGlobalMCPManager().ExecuteTool(ctx, toolName, arguments)
		if err != nil {
			return "", err
		}
		var parts []string
		for _, block := range result.Content {
			if block.Type == "text" {
				parts = append(parts, block.Text)
			}
		}
		return strings.Join(parts, "\n"), nil
	}

	result, err := registry.Call(ctx, toolName, input)
	if err != nil {
		return "", err
	}
	if !result.Success {
		return "", fmt.Errorf("%s", result.Error)
	}
	return string(result.Data), nil
}
