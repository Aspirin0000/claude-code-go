package lsp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestClientStartAndClose(t *testing.T) {
	// Use a simple echo-like process that won't crash immediately
	client, err := NewClient("sleep", "10")
	if err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Close()

	if client.cmd == nil || client.cmd.Process == nil {
		t.Errorf("expected process to be started")
	}
}

func TestClientRequestTimeout(t *testing.T) {
	// Start a process that won't respond to JSON-RPC
	client, err := NewClient("sleep", "10")
	if err != nil {
		t.Fatalf("failed to start client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = client.Request(ctx, "initialize", map[string]interface{}{})
	if err == nil {
		t.Errorf("expected timeout error")
	}
}

func TestManagerRegisterAndGetServer(t *testing.T) {
	mgr := NewManager("file:///tmp")
	mgr.RegisterServer(ServerConfig{
		Command:    "sleep",
		Args:       []string{"10"},
		Extensions: []string{".go"},
	})

	server, err := mgr.getServerForPath("/tmp/main.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server == nil {
		t.Fatalf("expected server")
	}

	_, err = mgr.getServerForPath("/tmp/main.py")
	if err == nil {
		t.Errorf("expected error for unregistered extension")
	}
}

func TestManagerOpenFileNoServer(t *testing.T) {
	mgr := NewManager("file:///tmp")
	err := mgr.OpenFile("/tmp/test.go", 1, "package main")
	if err == nil {
		t.Errorf("expected error when no server registered")
	}
}

func TestPathToURI(t *testing.T) {
	cwd, _ := os.Getwd()
	uri := PathToURI(cwd)
	if !json.Valid([]byte(fmt.Sprintf(`"%s"`, uri))) {
		// Not testing JSON validity of the URI itself, just that it starts with file://
	}
	if uri[:7] != "file://" {
		t.Errorf("expected URI to start with file://, got %s", uri)
	}
}

func TestLanguageIDFromExt(t *testing.T) {
	tests := map[string]string{
		".go":  "go",
		".py":  "python",
		".rs":  "rust",
		".cpp": "cpp",
		".xyz": "xyz",
	}
	for ext, expected := range tests {
		if got := languageIDFromExt(ext); got != expected {
			t.Errorf("languageIDFromExt(%q) = %q, want %q", ext, got, expected)
		}
	}
}

func TestServerInstanceStartNoBinary(t *testing.T) {
	server := NewServerInstance("/nonexistent/binary", nil, []string{".go"}, "file:///tmp")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	err := server.Start(ctx)
	if err == nil {
		t.Errorf("expected error starting nonexistent binary")
	}
	if server.State != "error" {
		t.Errorf("expected state error, got %s", server.State)
	}
}
