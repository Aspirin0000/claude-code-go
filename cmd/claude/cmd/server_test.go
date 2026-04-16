package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/api"
	"github.com/Aspirin0000/claude-code-go/internal/tools"
)

func TestServerHealth(t *testing.T) {
	client := api.NewClient("test-key", "claude-test")
	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Status != "ok" {
		t.Errorf("expected status ok, got %s", resp.Status)
	}
	if resp.Model != "claude-test" {
		t.Errorf("unexpected model: %s", resp.Model)
	}
}

func TestServerTools(t *testing.T) {
	client := api.NewClient("test-key", "claude-test")
	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/tools", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["count"].(float64) == 0 {
		t.Error("expected non-zero tool count")
	}
}

func TestServerChatMissingPrompt(t *testing.T) {
	client := api.NewClient("test-key", "claude-test")
	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(map[string]interface{}{})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Error("expected failure for missing prompt")
	}
}

func TestServerChatNoClient(t *testing.T) {
	registry := tools.NewDefaultRegistry()
	server := NewServer(nil, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	body, _ := json.Marshal(map[string]interface{}{"prompt": "hello"})
	req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	var resp ChatResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp.Success {
		t.Error("expected failure when client is nil")
	}
}

func TestServerMethodNotAllowed(t *testing.T) {
	client := api.NewClient("test-key", "claude-test")
	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestServerModels(t *testing.T) {
	client := api.NewClient("test-key", "claude-test")
	registry := tools.NewDefaultRegistry()
	server := NewServer(client, registry, "claude-test")

	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/models", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if resp["count"].(float64) == 0 {
		t.Error("expected non-zero model count")
	}
	models, ok := resp["models"].([]interface{})
	if !ok || len(models) == 0 {
		t.Error("expected models list")
	}
}
