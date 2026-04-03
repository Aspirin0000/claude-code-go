package commands

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
)

// LoadCommand 加载会话命令
type LoadCommand struct {
	*BaseCommand
}

// ConversationData 会话数据结构
type ConversationData struct {
	SessionID      string                 `json:"sessionId,omitempty"`
	ConversationID string                 `json:"conversationId,omitempty"`
	Messages       []ConversationMessage  `json:"messages"`
	CreatedAt      time.Time              `json:"createdAt,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConversationMessage 会话消息
type ConversationMessage struct {
	UUID      string    `json:"uuid"`
	Type      string    `json:"type"`
	Role      string    `json:"role,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}

// NewLoadCommand 创建load命令
func NewLoadCommand() *LoadCommand {
	return &LoadCommand{
		BaseCommand: NewBaseCommand(
			"load",
			"从文件加载会话",
			CategorySession,
		).WithAliases("import", "restore-file").
			WithHelp(`使用: /load <filename>

从文件加载会话历史。
支持格式: JSON (.json) 和 Markdown (.md)

参数:
  filename - 要加载的文件路径

别名: /import, /restore-file

示例:
  /load my-session.json
  /load backup.md`),
	}
}

// Execute 执行加载操作
func (c *LoadCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("请提供文件路径: /load <filename>")
	}

	filename := args[0]

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filename)
	}

	// 根据文件扩展名决定加载方式
	ext := strings.ToLower(filepath.Ext(filename))
	var data *ConversationData
	var err error

	switch ext {
	case ".json":
		data, err = c.loadJSON(filename)
	case ".md", ".markdown":
		data, err = c.loadMarkdown(filename)
	default:
		// 尝试自动检测格式
		data, err = c.autoDetectAndLoad(filename)
	}

	if err != nil {
		return fmt.Errorf("加载文件失败: %w", err)
	}

	// 验证数据
	if err := c.validateData(data); err != nil {
		return fmt.Errorf("数据验证失败: %w", err)
	}

	// 恢复会话状态
	if err := c.restoreSession(data); err != nil {
		return fmt.Errorf("恢复会话失败: %w", err)
	}

	fmt.Printf("✓ 会话已从 %s 加载成功\n", filename)
	fmt.Printf("  - 会话ID: %s\n", data.SessionID)
	fmt.Printf("  - 消息数量: %d\n", len(data.Messages))

	return nil
}

// loadJSON 从JSON文件加载
func (c *LoadCommand) loadJSON(filename string) (*ConversationData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ConversationData
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("JSON解析失败: %w", err)
	}

	return &data, nil
}

// loadMarkdown 从Markdown文件加载
func (c *LoadCommand) loadMarkdown(filename string) (*ConversationData, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data := &ConversationData{
		Messages: make([]ConversationMessage, 0),
		Metadata: make(map[string]interface{}),
	}

	scanner := bufio.NewScanner(file)
	var currentMsg *ConversationMessage
	var contentLines []string

	for scanner.Scan() {
		line := scanner.Text()

		// 检测角色标题 (## User, ## Assistant, ## System)
		if strings.HasPrefix(line, "## ") {
			// 保存之前的消息
			if currentMsg != nil && len(contentLines) > 0 {
				currentMsg.Content = strings.Join(contentLines, "\n")
				data.Messages = append(data.Messages, *currentMsg)
			}

			// 创建新消息
			role := strings.TrimPrefix(line, "## ")
			role = strings.TrimSpace(role)
			role = strings.ToLower(role)

			currentMsg = &ConversationMessage{
				UUID:      generateUUID(),
				Type:      role,
				Role:      role,
				Timestamp: time.Now(),
			}
			contentLines = []string{}
		} else if currentMsg != nil {
			// 累积内容
			contentLines = append(contentLines, line)
		} else {
			// 文件头部可能是元数据
			if strings.HasPrefix(line, "# ") {
				data.SessionID = strings.TrimPrefix(line, "# ")
				data.SessionID = strings.TrimSpace(data.SessionID)
			}
		}
	}

	// 保存最后一条消息
	if currentMsg != nil && len(contentLines) > 0 {
		currentMsg.Content = strings.Join(contentLines, "\n")
		data.Messages = append(data.Messages, *currentMsg)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return data, nil
}

// autoDetectAndLoad 自动检测并加载文件
func (c *LoadCommand) autoDetectAndLoad(filename string) (*ConversationData, error) {
	// 读取文件内容进行格式检测
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	// 尝试JSON解析
	var jsonData ConversationData
	if err := json.Unmarshal(content, &jsonData); err == nil && len(jsonData.Messages) > 0 {
		return &jsonData, nil
	}

	// 尝试Markdown解析
	return c.loadMarkdown(filename)
}

// validateData 验证加载的数据
func (c *LoadCommand) validateData(data *ConversationData) error {
	if data == nil {
		return fmt.Errorf("数据为空")
	}

	if len(data.Messages) == 0 {
		return fmt.Errorf("消息列表为空")
	}

	// 验证每条消息
	for i, msg := range data.Messages {
		if msg.Type == "" {
			return fmt.Errorf("消息 %d 缺少类型", i)
		}
		if msg.Content == "" {
			return fmt.Errorf("消息 %d 缺少内容", i)
		}
	}

	return nil
}

// restoreSession 恢复会话状态
func (c *LoadCommand) restoreSession(data *ConversationData) error {
	// 清空当前消息
	state.GlobalState.ClearMessages()

	// 恢复会话ID
	if data.SessionID != "" {
		state.GlobalState.SetSessionID(data.SessionID)
	}

	// 设置对话ID
	state.GlobalState.ConversationID = data.ConversationID

	// 恢复消息
	for _, msg := range data.Messages {
		stateMsg := state.Message{
			UUID:    msg.UUID,
			Type:    msg.Type,
			Role:    msg.Role,
			Content: msg.Content,
		}
		state.GlobalState.AddMessage(stateMsg)
	}

	// 更新轮次计数
	state.GlobalState.TurnCount = len(data.Messages) / 2

	return nil
}

func init() { Register(NewLoadCommand()) }
