package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/services/analytics"
	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

// ChatRequest is the JSON body for POST /chat.
type ChatRequest struct {
	Prompt    string `json:"prompt"`
	System    string `json:"system,omitempty"`
	MaxRounds int    `json:"max_rounds,omitempty"`
}

// ChatResponse is the JSON body returned by POST /chat.
type ChatResponse struct {
	Success   bool           `json:"success"`
	Response  string         `json:"response,omitempty"`
	Messages  []JSONMessage  `json:"messages,omitempty"`
	ToolCalls []JSONToolCall `json:"tool_calls,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// HealthResponse is the JSON body returned by GET /health.
type HealthResponse struct {
	Status    string `json:"status"`
	Model     string `json:"model,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// Server wraps the HTTP server and dependencies.
type Server struct {
	client       *api.Client
	toolRegistry *tools.Registry
	model        string
}

// NewServer creates a new HTTP server handler wrapper.
func NewServer(client *api.Client, registry *tools.Registry, model string) *Server {
	return &Server{
		client:       client,
		toolRegistry: registry,
		model:        model,
	}
}

// RegisterRoutes registers all HTTP endpoints.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/tools", s.handleTools)
	mux.HandleFunc("/chat", s.handleChat)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	resp := HealthResponse{
		Status:    "ok",
		Model:     s.model,
		Timestamp: time.Now().Unix(),
	}
	writeJSON(w, resp)
}

func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	toolSchemas := s.toolRegistry.GetToolSchemas()
	writeJSON(w, map[string]interface{}{
		"tools": toolSchemas,
		"count": len(toolSchemas),
	})
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, ChatResponse{Success: false, Error: "invalid JSON: " + err.Error()})
		return
	}

	if req.Prompt == "" {
		writeJSON(w, ChatResponse{Success: false, Error: "missing prompt"})
		return
	}
	if req.MaxRounds <= 0 {
		req.MaxRounds = 10
	}

	if s.client == nil {
		writeJSON(w, ChatResponse{Success: false, Error: "AI client not initialized"})
		return
	}

	// Use a fresh app state per request so server is stateless
	origState := state.GlobalState
	state.GlobalState = state.NewAppState()
	defer func() { state.GlobalState = origState }()

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
		"mode": "http",
	})

	toolsList := buildToolsList(s.toolRegistry)
	var toolCalls []JSONToolCall
	var finalResponse string

	ctx := context.WithValue(r.Context(), tools.APIClientContextKey, s.client)

	for round := 0; round < req.MaxRounds; round++ {
		apiMessages := buildAPIMessagesFromState()
		resp, err := s.client.ChatWithBlocks(ctx, apiMessages, toolsList)
		if err != nil {
			writeJSON(w, ChatResponse{Success: false, Error: err.Error(), ToolCalls: toolCalls})
			return
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
			resultText, err := executeTool(ctx, s.toolRegistry, block.Name, block.Input)
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
			continue
		}
		jsonMessages = append(jsonMessages, JSONMessage{
			Role:    msg.Role,
			Content: msg.Content,
		})
	}

	writeJSON(w, ChatResponse{
		Success:   true,
		Response:  finalResponse,
		Messages:  jsonMessages,
		ToolCalls: toolCalls,
	})
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

// runServer starts the HTTP server mode.
func runServer(port string) error {
	app := NewApp()
	if app.apiClient == nil {
		return fmt.Errorf("AI client not initialized; please configure API key")
	}

	mux := http.NewServeMux()
	server := NewServer(app.apiClient, app.toolRegistry, app.config.Model)
	server.RegisterRoutes(mux)

	addr := ":" + port
	fmt.Printf("🌐 Claude Code HTTP server listening on http://localhost%s\n", addr)
	fmt.Println("Endpoints:")
	fmt.Println("  GET  /health  - Health check")
	fmt.Println("  GET  /tools   - List available tools")
	fmt.Println("  POST /chat    - Send a chat request")
	fmt.Println("Press Ctrl+C to stop")

	analytics.InitDefaultSink()
	analytics.LogEvent("session_started", analytics.LogEventMetadata{
		"mode": "http_server",
		"port": port,
	})

	return http.ListenAndServe(addr, mux)
}
