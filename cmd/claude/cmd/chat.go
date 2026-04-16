// cmd/claude/cmd/chat.go
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/cmd/claude/commands"
	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/config"
	"github.com/Aspirin0000/claude-code-go/internal/hooks"
	"github.com/Aspirin0000/claude-code-go/internal/mcp"
	"github.com/Aspirin0000/claude-code-go/internal/services/analytics"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
	"github.com/Aspirin0000/claude-code-go/internal/types"
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
	jsonFlag    bool
	serveFlag   bool
	portFlag    string
)

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKeyFlag, "api-key", "", "Anthropic API key")
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "AI model to use")
	rootCmd.PersistentFlags().BoolVarP(&verboseFlag, "verbose", "v", false, "Enable verbose output")
	rootCmd.Flags().StringVarP(&promptFlag, "prompt", "p", "", "Initial prompt to send")
	rootCmd.Flags().BoolVar(&jsonFlag, "json", false, "Run in structured JSON mode")
	rootCmd.Flags().BoolVar(&serveFlag, "serve", false, "Run HTTP server mode")
	rootCmd.Flags().StringVar(&portFlag, "port", "8080", "HTTP server port (used with --serve)")
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
	if serveFlag {
		return runServer(portFlag)
	}

	if jsonFlag {
		return runJSONMode()
	}

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

	// Run session start hooks
	if hookMgr := hooks.GetGlobalManager(); hookMgr.HasHooks(types.HookEventSessionStart) {
		if result, err := hookMgr.ExecuteSessionStart(ctx); err == nil && result.Continue {
			if result.InitialUserMessage != "" {
				promptFlag = result.InitialUserMessage
			}
		}
	}

	printWelcome()

	analytics.InitDefaultSink()
	analytics.LogEvent("session_started", analytics.LogEventMetadata{
		"mode": "repl",
	})

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

	// Execute UserPromptSubmit hooks
	hookMgr := hooks.GetGlobalManager()
	if hookMgr.HasHooks(types.HookEventUserPromptSubmit) {
		result, err := hookMgr.ExecuteUserPromptSubmit(ctx, message)
		if err != nil {
			fmt.Printf("Hook error: %v\n", err)
		} else {
			if !result.Continue {
				if result.SuppressOutput {
					return nil
				}
				fmt.Println("🚫 Prompt blocked by hook.")
				return nil
			}
			if result.AdditionalContext != "" {
				message = message + "\n\n[Context: " + result.AdditionalContext + "]"
			}
		}
	}

	state.GlobalState.AddMessage(state.Message{
		Type:    "user",
		Role:    "user",
		Content: message,
	})

	analytics.LogEvent("chat_message_sent", analytics.LogEventMetadata{
		"mode": "repl",
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
		start := time.Now()
		var resp *api.Response
		var err error
		if os.Getenv("CLAUDE_STREAM") == "1" {
			resp, err = streamResponse(ctx, client, apiMessages, toolsList)
		} else {
			resp, err = client.ChatWithBlocks(ctx, apiMessages, toolsList)
		}
		duration := time.Since(start).Milliseconds()
		if err != nil {
			analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
				"success":     false,
				"duration_ms": duration,
				"model":       client.GetModel(),
				"error":       err.Error(),
			})
			return fmt.Errorf("API error: %w", err)
		}
		analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
			"success":     true,
			"duration_ms": duration,
			"model":       client.GetModel(),
			"blocks":      len(resp.Content),
		})

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
				if os.Getenv("CLAUDE_STREAM") != "1" {
					fmt.Printf("\n🤖 %s\n", content)
				}
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
		if err := state.GlobalSessionStorage.AutoSave(state.GlobalState); err == nil {
			analytics.LogEvent("session_autosaved", analytics.LogEventMetadata{
				"mode": "repl",
			})
		}
	}

	return nil
}

// streamResponse calls ChatStream, prints text deltas in real-time, and returns the assembled Response.
func streamResponse(ctx context.Context, client *api.Client, messages []api.Message, toolsList []api.Tool) (*api.Response, error) {
	ch, err := client.ChatStream(ctx, messages, toolsList)
	if err != nil {
		return nil, err
	}

	fmt.Print("\n🤖 ")
	resp := api.CollectStreamResponse(ch)
	fmt.Println() // newline after streaming
	return resp, nil
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

// Styles holds all TUI styles for a given theme
type styles struct {
	titleStyle     lipgloss.Style
	userStyle      lipgloss.Style
	assistantStyle lipgloss.Style
	systemStyle    lipgloss.Style
	inputStyle     lipgloss.Style
	helpStyle      lipgloss.Style
	statusBarStyle lipgloss.Style
	dividerStyle   lipgloss.Style
	timestampStyle lipgloss.Style
}

func newStyles(theme string) *styles {
	if theme == "light" {
		return &styles{
			titleStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#5A3FC0")).Background(lipgloss.Color("#F0F0F5")).Padding(1, 2).Width(50),
			userStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#008844")).Bold(true),
			assistantStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#0066AA")),
			systemStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#CC4444")),
			inputStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#222222")).Background(lipgloss.Color("#E8E8EE")).Padding(0, 1),
			helpStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#777777")).Italic(true),
			statusBarStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Background(lipgloss.Color("#F0F0F5")).Padding(0, 1),
			dividerStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")),
			timestampStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Italic(true),
		}
	}
	// default dark theme
	return &styles{
		titleStyle:     lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")).Background(lipgloss.Color("#1a1a2e")).Padding(1, 2).Width(50),
		userStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF88")).Bold(true),
		assistantStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#00CCFF")),
		systemStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")),
		inputStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(lipgloss.Color("#2d2d44")).Padding(0, 1),
		helpStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true),
		statusBarStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")).Background(lipgloss.Color("#1a1a2e")).Padding(0, 1),
		dividerStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")),
		timestampStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true),
	}
}

// App TUI application model
type App struct {
	config       *config.Config
	apiClient    *api.Client
	toolRegistry *tools.Registry
	input        string
	inputHistory []string
	historyIndex int
	scrollOffset int
	loading      bool
	width        int
	height       int
	styles       *styles
	// streaming state
	streamingText string
	streamBlocks  []api.ContentBlock
	streamCh      <-chan api.StreamEvent
}

func runInteractive() error {
	app := NewApp()
	p := tea.NewProgram(app, tea.WithAltScreen(), tea.WithMouseCellMotion())
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

// JSONRequest is the input format for --json mode.
type JSONRequest struct {
	Prompt    string `json:"prompt"`
	System    string `json:"system,omitempty"`
	MaxRounds int    `json:"max_rounds,omitempty"`
}

// JSONToolCall represents a tool invocation in JSON mode output.
type JSONToolCall struct {
	Name   string          `json:"name"`
	Input  json.RawMessage `json:"input"`
	Result string          `json:"result,omitempty"`
	Error  string          `json:"error,omitempty"`
}

// JSONMessage represents a conversation message in JSON mode output.
type JSONMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// JSONResponse is the output format for --json mode.
type JSONResponse struct {
	Success   bool           `json:"success"`
	Response  string         `json:"response,omitempty"`
	Messages  []JSONMessage  `json:"messages,omitempty"`
	ToolCalls []JSONToolCall `json:"tool_calls,omitempty"`
	Error     string         `json:"error,omitempty"`
}

func runJSONMode() error {
	app := NewApp()
	if app.apiClient == nil {
		return outputJSON(os.Stdout, JSONResponse{Success: false, Error: "AI client not initialized; please configure API key"})
	}
	return runJSONModeWithApp(app, os.Stdin, os.Stdout)
}

func runJSONModeWithApp(app *App, stdin io.Reader, stdout io.Writer) error {
	if app.apiClient == nil {
		return outputJSON(stdout, JSONResponse{Success: false, Error: "AI client not initialized; please configure API key"})
	}

	var req JSONRequest
	if stdin != nil {
		if f, ok := stdin.(*os.File); ok {
			stdinInfo, err := f.Stat()
			if err == nil && (stdinInfo.Mode()&os.ModeCharDevice) == 0 {
				data, readErr := io.ReadAll(stdin)
				if readErr != nil {
					return outputJSON(stdout, JSONResponse{Success: false, Error: readErr.Error()})
				}
				if err := json.Unmarshal(data, &req); err != nil {
					return outputJSON(stdout, JSONResponse{Success: false, Error: "invalid JSON input: " + err.Error()})
				}
			} else if promptFlag != "" {
				req.Prompt = promptFlag
			}
		} else {
			data, readErr := io.ReadAll(stdin)
			if readErr != nil {
				return outputJSON(stdout, JSONResponse{Success: false, Error: readErr.Error()})
			}
			if len(data) > 0 {
				if err := json.Unmarshal(data, &req); err != nil {
					return outputJSON(stdout, JSONResponse{Success: false, Error: "invalid JSON input: " + err.Error()})
				}
			} else if promptFlag != "" {
				req.Prompt = promptFlag
			}
		}
	} else if promptFlag != "" {
		req.Prompt = promptFlag
	}

	if req.Prompt == "" {
		return outputJSON(stdout, JSONResponse{Success: false, Error: "missing prompt in JSON input"})
	}
	if req.MaxRounds <= 0 {
		req.MaxRounds = 10
	}

	if req.System != "" {
		state.GlobalState.AddMessage(state.Message{
			Role:    "system",
			Type:    "system",
			Content: req.System,
		})
	}

	state.GlobalState.AddMessage(state.Message{
		Role:    "user",
		Type:    "user",
		Content: req.Prompt,
	})

	analytics.LogEvent("chat_message_sent", analytics.LogEventMetadata{
		"mode": "json",
	})

	toolsList := buildToolsList(app.toolRegistry)
	var toolCalls []JSONToolCall
	var finalResponse string

	ctx := context.WithValue(context.Background(), tools.APIClientContextKey, app.apiClient)

	for round := 0; round < req.MaxRounds; round++ {
		apiMessages := buildAPIMessagesFromState()
		resp, err := app.apiClient.ChatWithBlocks(ctx, apiMessages, toolsList)
		if err != nil {
			return outputJSON(stdout, JSONResponse{Success: false, Error: err.Error(), ToolCalls: toolCalls})
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

		assistantContent := strings.Join(textParts, "\n")
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
			Content: assistantContent,
			Blocks:  assistantBlocks,
		})

		if len(toolUses) == 0 {
			finalResponse = assistantContent
			break
		}

		for _, block := range toolUses {
			resultText, err := executeTool(ctx, app.toolRegistry, block.Name, block.Input)
			call := JSONToolCall{Name: block.Name, Input: block.Input}
			if err != nil {
				call.Error = err.Error()
				resultText = fmt.Sprintf("Error: %v", err)
			} else {
				call.Result = resultText
			}
			toolCalls = append(toolCalls, call)

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

	messages := state.GlobalState.GetMessages()
	jsonMessages := make([]JSONMessage, 0, len(messages))
	for _, msg := range messages {
		if msg.Type == "system" || msg.Role == "system" {
			continue
		}
		if msg.Role == "user" && msg.Content == "" && len(msg.Blocks) > 0 {
			// tool results skip content summary to keep output clean
			continue
		}
		jsonMessages = append(jsonMessages, JSONMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	return outputJSON(stdout, JSONResponse{
		Success:   true,
		Response:  finalResponse,
		Messages:  jsonMessages,
		ToolCalls: toolCalls,
	})
}

func outputJSON(w io.Writer, resp JSONResponse) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(resp)
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

	// Run session start hooks
	if hookMgr := hooks.GetGlobalManager(); hookMgr.HasHooks(types.HookEventSessionStart) {
		if result, err := hookMgr.ExecuteSessionStart(context.Background()); err == nil && result.Continue {
			if result.InitialUserMessage != "" {
				state.GlobalState.AddMessage(state.Message{
					Type:    "user",
					Role:    "user",
					Content: result.InitialUserMessage,
				})
			}
		}
	}

	analytics.InitDefaultSink()
	analytics.LogEvent("session_started", analytics.LogEventMetadata{
		"mode": "tui",
	})

	return &App{
		config:       cfg,
		apiClient:    client,
		toolRegistry: tools.NewDefaultRegistry(),
		styles:       newStyles(cfg.Theme),
	}
}

func (a *App) Init() tea.Cmd {
	return tea.EnterAltScreen
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
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
		case tea.KeyPgUp:
			a.scrollOffset += 3
		case tea.KeyPgDown:
			a.scrollOffset -= 3
			if a.scrollOffset < 0 {
				a.scrollOffset = 0
			}
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			a.scrollOffset += 3
		case tea.MouseButtonWheelDown:
			a.scrollOffset -= 3
			if a.scrollOffset < 0 {
				a.scrollOffset = 0
			}
		}
	case tea.WindowSizeMsg:
		a.width, a.height = msg.Width, msg.Height
	case apiResponseMsg:
		if msg.err != nil {
			a.loading = false
			state.GlobalState.AddMessage(state.Message{
				Type:    "system",
				Role:    "system",
				Content: fmt.Sprintf("Error: %v", msg.err),
			})
			a.input = ""
			a.scrollOffset = 0
			return a, nil
		}
		if msg.resp != nil {
			return a.handleAPIResponse(msg.resp)
		}
		a.loading = false
		a.input = ""
		a.scrollOffset = 0
	case streamEventMsg:
		if msg.ok {
			a.processStreamEvent(msg.event)
			return a, a.pollStream()
		}
		// channel closed or error
		resp, err := a.finishStream()
		if err != nil {
			a.loading = false
			state.GlobalState.AddMessage(state.Message{
				Type:    "system",
				Role:    "system",
				Content: fmt.Sprintf("Error: %v", err),
			})
			a.input = ""
			a.scrollOffset = 0
			return a, nil
		}
		// Continue with normal response handling
		return a.handleAPIResponse(resp)
	case toolRoundCompleteMsg:
		// Trigger next API round after tools complete
		return a, a.processMessage()
	}
	return a, nil
}

func (a *App) View() string {
	var b strings.Builder

	b.WriteString(a.styles.titleStyle.Render("🚀 Claude Code Go " + commands.TargetVersion))
	b.WriteString("\n")

	// Status bar
	modelName := a.config.Model
	if modelName == "" {
		modelName = "default"
	}
	messages := state.GlobalState.GetMessages()
	statusText := fmt.Sprintf(" Model: %s | Messages: %d ", modelName, len(messages))
	if a.scrollOffset > 0 {
		statusText += fmt.Sprintf("| Scroll: %d ", a.scrollOffset)
	}
	b.WriteString(a.styles.statusBarStyle.Render(statusText))
	b.WriteString("\n")
	b.WriteString(a.styles.dividerStyle.Render(strings.Repeat("─", a.width)))
	b.WriteString("\n\n")

	visibleCount := a.height - 12
	startIdx := 0
	if len(messages) > visibleCount {
		startIdx = len(messages) - visibleCount - a.scrollOffset
		if startIdx < 0 {
			startIdx = 0
		}
	}

	for i := startIdx; i < len(messages); i++ {
		msg := messages[i]
		timestampStr := ""
		if !msg.Timestamp.IsZero() {
			timestampStr = a.styles.timestampStyle.Render(msg.Timestamp.Format("15:04")) + " "
		}
		switch msg.Role {
		case "user":
			prefix := timestampStr + a.styles.userStyle.Render("You: ")
			cw := a.width - visibleWidth(prefix)
			if cw < 1 {
				cw = 1
			}
			if msg.Content == "" && len(msg.Blocks) > 0 {
				b.WriteString(prefix + "[tool results]")
			} else {
				b.WriteString(prefix + wrapText(msg.Content, cw))
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
			prefix := timestampStr + a.styles.assistantStyle.Render("Claude: ")
			cw := a.width - visibleWidth(prefix)
			if cw < 1 {
				cw = 1
			}
			b.WriteString(prefix + wrapText(content, cw))
		case "system":
			prefix := timestampStr
			cw := a.width - visibleWidth(prefix)
			if cw < 1 {
				cw = 1
			}
			b.WriteString(prefix + a.styles.systemStyle.Render(wrapText(msg.Content, cw)))
		}
		b.WriteString("\n\n")
	}

	if a.loading && a.streamingText == "" {
		b.WriteString(a.styles.assistantStyle.Render("Thinking... ⏳") + "\n\n")
	}

	// Show streaming text preview if active
	if a.loading && a.streamingText != "" {
		prefix := a.styles.assistantStyle.Render("Claude: ")
		cw := a.width - visibleWidth(prefix)
		if cw < 1 {
			cw = 1
		}
		b.WriteString(prefix + wrapText(a.streamingText, cw) + "▌")
		b.WriteString("\n\n")
	}

	b.WriteString(a.styles.dividerStyle.Render(strings.Repeat("─", a.width)))
	b.WriteString("\n")
	b.WriteString(a.styles.inputStyle.Render("> "+a.input+"█") + "\n")
	b.WriteString(a.styles.helpStyle.Render("Ctrl+C / Esc: Exit | Enter: Send | ↑↓: History | PgUp/PgDn: Scroll"))
	return b.String()
}

// wrapText wraps text to fit within the given width.
// It preserves existing newlines and breaks long words.
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}
	var result strings.Builder
	lines := strings.Split(text, "\n")
	for lineIdx, line := range lines {
		if lineIdx > 0 {
			result.WriteString("\n")
		}
		if len(line) == 0 {
			continue
		}
		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for len(currentLine) > width {
			result.WriteString(currentLine[:width])
			result.WriteString("\n")
			currentLine = currentLine[width:]
		}

		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) <= width {
				currentLine += " " + word
			} else {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
				for len(currentLine) > width {
					result.WriteString(currentLine[:width])
					result.WriteString("\n")
					currentLine = currentLine[width:]
				}
			}
		}
		if currentLine != "" {
			result.WriteString(currentLine)
		}
	}
	return result.String()
}

// visibleWidth returns the visible width of a string rendered by lipgloss.
// For plain text this is just the length; for styled strings we approximate
// by stripping ANSI sequences. Since lipgloss.Render is called on short
// labels, we approximate by measuring the input to Render().
func visibleWidth(s string) int {
	// Simple ANSI strip: look for \x1b[...m sequences
	clean := ansiStripRegex(s)
	return len(clean)
}

func ansiStripRegex(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\x1b' {
			// skip until 'm'
			for j := i + 1; j < len(s); j++ {
				if s[j] == 'm' {
					i = j
					break
				}
			}
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func (a *App) processStreamEvent(event api.StreamEvent) {
	switch event.Type {
	case "content_block_start":
		if event.ContentBlock != nil {
			a.streamBlocks = append(a.streamBlocks, *event.ContentBlock)
		}
	case "content_block_delta":
		if event.Index >= 0 && event.Index < len(a.streamBlocks) {
			block := &a.streamBlocks[event.Index]
			switch event.Delta.Type {
			case "text_delta":
				block.Text += event.Delta.Text
				a.streamingText += event.Delta.Text
			case "input_json_delta":
				// accumulate for tool_use later
			}
		}
	}
}

func (a *App) pollStream() tea.Cmd {
	return func() tea.Msg {
		if a.streamCh == nil {
			return streamEventMsg{ok: false}
		}
		event, ok := <-a.streamCh
		return streamEventMsg{event: event, ok: ok}
	}
}

func (a *App) finishStream() (*api.Response, error) {
	resp := &api.Response{Role: "assistant", Content: a.streamBlocks}
	a.streamingText = ""
	a.streamBlocks = nil
	a.streamCh = nil
	return resp, nil
}

// handleAPIResponse processes a complete API response in TUI mode.
func (a *App) handleAPIResponse(resp *api.Response) (tea.Model, tea.Cmd) {
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

	if len(toolUses) == 0 {
		a.loading = false
		a.input = ""
		a.scrollOffset = 0
		return a, nil
	}

	// Execute tools asynchronously
	return a, a.executeTools(toolUses)
}

func (a *App) executeTools(toolUses []api.ContentBlock) tea.Cmd {
	return func() tea.Msg {
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
		// After tool execution, trigger another API round
		return toolRoundCompleteMsg{}
	}
}

type toolRoundCompleteMsg struct{}

func (a *App) handleInput() (tea.Model, tea.Cmd) {
	input := a.input

	// Execute UserPromptSubmit hooks
	hookMgr := hooks.GetGlobalManager()
	if hookMgr.HasHooks(types.HookEventUserPromptSubmit) {
		result, err := hookMgr.ExecuteUserPromptSubmit(context.Background(), input)
		if err != nil {
			state.GlobalState.AddMessage(state.Message{
				Role:    "system",
				Type:    "system",
				Content: fmt.Sprintf("Hook error: %v", err),
			})
			return a, nil
		}
		if !result.Continue {
			if !result.SuppressOutput {
				state.GlobalState.AddMessage(state.Message{
					Role:    "system",
					Type:    "system",
					Content: "🚫 Prompt blocked by hook.",
				})
			}
			a.input = ""
			return a, nil
		}
		if result.AdditionalContext != "" {
			input = input + "\n\n[Context: " + result.AdditionalContext + "]"
		}
	}

	state.GlobalState.AddMessage(state.Message{
		Role:    "user",
		Type:    "user",
		Content: input,
	})
	analytics.LogEvent("chat_message_sent", analytics.LogEventMetadata{
		"mode": "tui",
	})
	a.inputHistory = append(a.inputHistory, a.input)
	a.historyIndex = len(a.inputHistory)
	a.loading = true
	a.input = ""
	return a, a.processMessage()
}

type apiResponseMsg struct {
	content string
	resp    *api.Response
	err     error
}

// streamEventMsg carries a single SSE event from ChatStream.
type streamEventMsg struct {
	event api.StreamEvent
	ok    bool
}

func (a *App) processMessage() tea.Cmd {
	return func() tea.Msg {
		if a.apiClient == nil {
			return apiResponseMsg{err: fmt.Errorf("AI client not initialized; please configure API key")}
		}

		toolsList := buildToolsList(a.toolRegistry)
		apiMessages := buildAPIMessagesFromState()
		start := time.Now()

		if os.Getenv("CLAUDE_STREAM") == "1" {
			ch, err := a.apiClient.ChatStream(context.Background(), apiMessages, toolsList)
			if err != nil {
				analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
					"success":     false,
					"duration_ms": time.Since(start).Milliseconds(),
					"model":       a.apiClient.GetModel(),
					"error":       err.Error(),
				})
				return apiResponseMsg{err: err}
			}
			analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
				"success":     true,
				"duration_ms": time.Since(start).Milliseconds(),
				"model":       a.apiClient.GetModel(),
			})
			a.streamCh = ch
			a.streamingText = ""
			a.streamBlocks = nil
			return a.pollStream()
		}

		resp, err := a.apiClient.ChatWithBlocks(context.Background(), apiMessages, toolsList)
		duration := time.Since(start).Milliseconds()
		if err != nil {
			analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
				"success":     false,
				"duration_ms": duration,
				"model":       a.apiClient.GetModel(),
				"error":       err.Error(),
			})
			return apiResponseMsg{err: err}
		}
		analytics.LogEvent("api_request_completed", analytics.LogEventMetadata{
			"success":     true,
			"duration_ms": duration,
			"model":       a.apiClient.GetModel(),
			"blocks":      len(resp.Content),
		})
		return apiResponseMsg{resp: resp}
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
	// Execute PreToolUse hooks
	hookMgr := hooks.GetGlobalManager()
	var toolInput map[string]interface{}
	if len(input) > 0 {
		_ = json.Unmarshal(input, &toolInput)
	}
	if hookMgr.HasHooks(types.HookEventPreToolUse) {
		result, err := hookMgr.ExecutePreToolUse(ctx, toolName, toolInput)
		if err != nil {
			return "", fmt.Errorf("hook error: %w", err)
		}
		if !result.Continue || result.Decision == types.PermissionBehaviorDeny {
			analytics.LogEvent("tool_executed", analytics.LogEventMetadata{
				"tool_name": toolName,
				"success":   false,
				"blocked":   true,
				"reason":    result.DecisionReason,
			})
			return "", fmt.Errorf("tool use blocked by hook: %s", result.DecisionReason)
		}
		if len(result.UpdatedInput) > 0 {
			updatedInput, _ := json.Marshal(result.UpdatedInput)
			input = updatedInput
			toolInput = result.UpdatedInput
		}
	}

	start := time.Now()
	var resultText string
	var execErr error

	if strings.HasPrefix(toolName, "mcp__") {
		var arguments map[string]interface{}
		if len(input) > 0 {
			if err := json.Unmarshal(input, &arguments); err != nil {
				execErr = fmt.Errorf("failed to parse MCP tool arguments: %w", err)
			}
		}
		if execErr == nil {
			result, err := mcp.GetGlobalMCPManager().ExecuteTool(ctx, toolName, arguments)
			if err != nil {
				execErr = err
			} else {
				var parts []string
				for _, block := range result.Content {
					if block.Type == "text" {
						parts = append(parts, block.Text)
					}
				}
				resultText = strings.Join(parts, "\n")
			}
		}
	} else {
		result, err := registry.Call(ctx, toolName, input)
		if err != nil {
			execErr = err
		} else if !result.Success {
			execErr = fmt.Errorf("%s", result.Error)
		} else {
			resultText = string(result.Data)
		}
	}

	analytics.LogEvent("tool_executed", analytics.LogEventMetadata{
		"tool_name":   toolName,
		"success":     execErr == nil,
		"duration_ms": time.Since(start).Milliseconds(),
	})

	if execErr != nil {
		return "", execErr
	}
	return resultText, nil
}
