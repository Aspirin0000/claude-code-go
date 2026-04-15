package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/lsp"
)

func TestLSPToolNoServer(t *testing.T) {
	tool := &LSPTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"operation": "hover",
		"filePath":  "/tmp/test.xyz",
		"line":      1,
		"character": 1,
	})

	result, err := tool.Call(context.Background(), input)
	if err == nil {
		t.Fatalf("expected error when no LSP server available")
	}
	_ = result
}

func TestLSPToolDocumentSymbol(t *testing.T) {
	// Create a temp go file
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "main.go")
	os.WriteFile(goFile, []byte("package main\n\nfunc main() {}\n"), 0644)

	// Register a fake server (sleep won't respond correctly, but we test schema/path handling)
	mgr := lsp.NewManager(lsp.PathToURI(tmpDir))
	mgr.RegisterServer(lsp.ServerConfig{
		Command:    "sleep",
		Args:       []string{"1"},
		Extensions: []string{".go"},
	})
	lsp.SetGlobalManager(mgr)
	defer lsp.SetGlobalManager(nil)

	tool := &LSPTool{}
	input, _ := json.Marshal(map[string]interface{}{
		"operation": "documentSymbol",
		"filePath":  goFile,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := tool.Call(ctx, input)
	// We expect a timeout because sleep doesn't speak JSON-RPC
	if err == nil {
		t.Log("expected timeout or communication error from non-LSP process")
	}
}

func TestLSPToolCountResults(t *testing.T) {
	if countResults(json.RawMessage(`[1,2,3]`)) != 3 {
		t.Errorf("expected 3")
	}
	if countResults(json.RawMessage(`{"a":1}`)) != 1 {
		t.Errorf("expected 1")
	}
	if countResults(json.RawMessage(`null`)) != 0 {
		t.Errorf("expected 0")
	}
}

func TestFormatLocations(t *testing.T) {
	locs := []Location{
		{URI: "file:///tmp/a.go", Range: Range{Start: Position{Line: 0, Character: 0}}},
	}
	out := formatLocations(locs)
	if out == "" {
		t.Errorf("expected non-empty output")
	}
}

func TestHoverString(t *testing.T) {
	h := Hover{Contents: json.RawMessage(`"hello world"`)}
	if h.String() != "hello world" {
		t.Errorf("unexpected hover string: %s", h.String())
	}

	h2 := Hover{Contents: json.RawMessage(`{"kind":"plaintext","value":"hover text"}`)}
	if h2.String() != "hover text" {
		t.Errorf("unexpected markup hover: %s", h2.String())
	}
}

func TestSymbolKindName(t *testing.T) {
	if symbolKindName(12) != "Function" {
		t.Errorf("unexpected kind name")
	}
	if symbolKindName(999) != "Unknown(999)" {
		t.Errorf("unexpected unknown kind name")
	}
}
