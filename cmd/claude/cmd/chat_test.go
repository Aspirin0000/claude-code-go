package cmd

import (
	"encoding/json"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/state"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

func TestBuildAPIMessagesFromState(t *testing.T) {
	// Save and restore global state
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()

	state.GlobalState = state.NewAppState()

	state.GlobalState.AddMessage(state.Message{
		Role:    "user",
		Type:    "user",
		Content: "Hello",
	})
	state.GlobalState.AddMessage(state.Message{
		Role:    "assistant",
		Type:    "assistant",
		Content: "Hi there",
		Blocks: []state.ContentBlock{
			{Type: "text", Text: "Hi there"},
			{Type: "tool_use", ID: "tu_1", Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)},
		},
	})
	state.GlobalState.AddMessage(state.Message{
		Role: "user",
		Type: "user",
		Blocks: []state.ContentBlock{
			{Type: "tool_result", ToolUseID: "tu_1", Content: "file1.go\nfile2.go"},
		},
	})

	apiMessages := buildAPIMessagesFromState()

	if len(apiMessages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(apiMessages))
	}

	if apiMessages[0].Role != "user" || apiMessages[0].Content != "Hello" {
		t.Errorf("unexpected first message: %+v", apiMessages[0])
	}

	if apiMessages[1].Role != "assistant" || apiMessages[1].Content != "Hi there" {
		t.Errorf("unexpected second message: %+v", apiMessages[1])
	}
	if len(apiMessages[1].Blocks) != 2 {
		t.Fatalf("expected 2 blocks in assistant message, got %d", len(apiMessages[1].Blocks))
	}
	if apiMessages[1].Blocks[0].Type != "text" || apiMessages[1].Blocks[0].Text != "Hi there" {
		t.Errorf("unexpected first block: %+v", apiMessages[1].Blocks[0])
	}
	if apiMessages[1].Blocks[1].Type != "tool_use" || apiMessages[1].Blocks[1].ID != "tu_1" {
		t.Errorf("unexpected second block: %+v", apiMessages[1].Blocks[1])
	}

	if apiMessages[2].Role != "user" {
		t.Errorf("unexpected third message role: %s", apiMessages[2].Role)
	}
	if len(apiMessages[2].Blocks) != 1 || apiMessages[2].Blocks[0].Type != "tool_result" {
		t.Errorf("unexpected third message blocks: %+v", apiMessages[2].Blocks)
	}
	if apiMessages[2].Blocks[0].Content != "file1.go\nfile2.go" {
		t.Errorf("unexpected tool_result content: %s", apiMessages[2].Blocks[0].Content)
	}
}

func TestBuildToolsList(t *testing.T) {
	registry := tools.NewDefaultRegistry()
	toolsList := buildToolsList(registry)

	if len(toolsList) == 0 {
		t.Fatal("expected non-empty tool list")
	}

	hasBash := false
	hasFileRead := false
	for _, tool := range toolsList {
		if tool.Name == "bash" {
			hasBash = true
		}
		if tool.Name == "file_read" {
			hasFileRead = true
		}
	}

	if !hasBash {
		t.Error("expected bash tool in list")
	}
	if !hasFileRead {
		t.Error("expected file_read tool in list")
	}
}

func TestBuildAPIMessagesFromState_Empty(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()

	state.GlobalState = state.NewAppState()

	apiMessages := buildAPIMessagesFromState()
	if len(apiMessages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(apiMessages))
	}
}

func TestBuildAPIMessagesFromState_RoleFallback(t *testing.T) {
	origState := state.GlobalState
	defer func() { state.GlobalState = origState }()

	state.GlobalState = state.NewAppState()
	state.GlobalState.AddMessage(state.Message{
		Role:    "",
		Type:    "system",
		Content: "System message",
	})

	apiMessages := buildAPIMessagesFromState()
	if len(apiMessages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(apiMessages))
	}
	if apiMessages[0].Role != "system" {
		t.Errorf("expected role fallback to type, got %s", apiMessages[0].Role)
	}
}

func TestJSONRequestParsing(t *testing.T) {
	input := []byte(`{"prompt":"hello","system":"be helpful","max_rounds":5}`)
	var req JSONRequest
	if err := json.Unmarshal(input, &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if req.Prompt != "hello" {
		t.Errorf("unexpected prompt: %s", req.Prompt)
	}
	if req.System != "be helpful" {
		t.Errorf("unexpected system: %s", req.System)
	}
	if req.MaxRounds != 5 {
		t.Errorf("unexpected max_rounds: %d", req.MaxRounds)
	}
}

func TestJSONResponseSerialization(t *testing.T) {
	resp := JSONResponse{
		Success:   true,
		Response:  "hello",
		Messages:  []JSONMessage{{Role: "user", Content: "hi"}},
		ToolCalls: []JSONToolCall{{Name: "bash", Input: json.RawMessage(`{"command":"ls"}`)}},
	}
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	var parsed JSONResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !parsed.Success || parsed.Response != "hello" || len(parsed.Messages) != 1 || len(parsed.ToolCalls) != 1 {
		t.Errorf("unexpected parsed response: %+v", parsed)
	}
}

func TestNoTuiFlag(t *testing.T) {
	// Test that --no-tui flag exists
	if !rootCmd.Flags().Lookup("no-tui").Changed {
		// Flag exists but wasn't changed in this test
		// Just verify it's registered
		flag := rootCmd.Flags().Lookup("no-tui")
		if flag == nil {
			t.Fatal("--no-tui flag not registered")
		}
		if flag.DefValue != "false" {
			t.Errorf("expected default false, got %s", flag.DefValue)
		}
	}
}
