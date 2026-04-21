package mcp

import (
	"testing"
)

// ============================================================================
// LRU Cache Tests
// ============================================================================

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache[string](3)
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if cache.capacity != 3 {
		t.Errorf("expected capacity 3, got %d", cache.capacity)
	}
	if cache.Len() != 0 {
		t.Errorf("expected empty cache, got %d items", cache.Len())
	}
}

func TestLRUCacheSetAndGet(t *testing.T) {
	cache := NewLRUCache[string](3)

	// Set values
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Get existing values
	val, found := cache.Get("key1")
	if !found {
		t.Error("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}

	val2, found2 := cache.Get("key2")
	if !found2 {
		t.Error("expected to find key2")
	}
	if val2 != "value2" {
		t.Errorf("expected value2, got %s", val2)
	}

	// Get non-existing value
	_, found3 := cache.Get("key3")
	if found3 {
		t.Error("expected not to find key3")
	}
}

func TestLRUCacheUpdateExisting(t *testing.T) {
	cache := NewLRUCache[string](3)

	cache.Set("key1", "value1")
	cache.Set("key1", "updated")

	val, found := cache.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != "updated" {
		t.Errorf("expected updated, got %s", val)
	}

	if cache.Len() != 1 {
		t.Errorf("expected 1 item, got %d", cache.Len())
	}
}

func TestLRUCacheEviction(t *testing.T) {
	cache := NewLRUCache[string](2)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3") // Should evict key1

	if cache.Len() != 2 {
		t.Errorf("expected 2 items after eviction, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be evicted")
	}

	val, found := cache.Get("key2")
	if !found {
		t.Fatal("expected to find key2")
	}
	if val != "value2" {
		t.Errorf("expected value2, got %s", val)
	}
}

func TestLRUCacheLRUOrder(t *testing.T) {
	cache := NewLRUCache[string](2)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	// Access key1 to make it recently used
	cache.Get("key1")

	// Add key3, should evict key2 (least recently used)
	cache.Set("key3", "value3")

	_, found := cache.Get("key2")
	if found {
		t.Error("expected key2 to be evicted (LRU)")
	}

	val, found := cache.Get("key1")
	if !found {
		t.Fatal("expected to find key1")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %s", val)
	}
}

func TestLRUCacheDelete(t *testing.T) {
	cache := NewLRUCache[string](3)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	deleted := cache.Delete("key1")
	if !deleted {
		t.Error("expected Delete to return true")
	}

	if cache.Len() != 1 {
		t.Errorf("expected 1 item after delete, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected key1 to be deleted")
	}

	// Delete non-existing key
	deleted2 := cache.Delete("key3")
	if deleted2 {
		t.Error("expected Delete to return false for non-existing key")
	}
}

func TestLRUCacheClear(t *testing.T) {
	cache := NewLRUCache[string](3)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	if cache.Len() != 0 {
		t.Errorf("expected 0 items after clear, got %d", cache.Len())
	}

	_, found := cache.Get("key1")
	if found {
		t.Error("expected cache to be empty after clear")
	}
}

func TestLRUCacheConcurrency(t *testing.T) {
	cache := NewLRUCache[int](10)
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			cache.Set(string(rune('a'+i%26)), i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			cache.Get(string(rune('a' + i%26)))
		}
		done <- true
	}()

	<-done
	<-done
}

// ============================================================================
// MCP Cache Tests
// ============================================================================

func TestNewMCPCache(t *testing.T) {
	cache := NewMCPCache()
	if cache == nil {
		t.Fatal("expected non-nil MCP cache")
	}
	if cache.toolsCache == nil {
		t.Error("expected non-nil tools cache")
	}
	if cache.resourcesCache == nil {
		t.Error("expected non-nil resources cache")
	}
	if cache.promptsCache == nil {
		t.Error("expected non-nil prompts cache")
	}
}

func TestGetGlobalMCPCache(t *testing.T) {
	cache1 := GetGlobalMCPCache()
	cache2 := GetGlobalMCPCache()

	if cache1 == nil {
		t.Fatal("expected non-nil global cache")
	}
	if cache1 != cache2 {
		t.Error("expected same global cache instance")
	}
}

func TestClearClientCache(t *testing.T) {
	// Use a fresh cache for this test
	globalMCPCache = NewMCPCache()
	globalMCPCache.toolsCache.Set("test-client", []ToolInfo{{Name: "tool1"}})
	globalMCPCache.resourcesCache.Set("test-client", []ResourceInfo{{Name: "res1"}})
	globalMCPCache.promptsCache.Set("test-client", []PromptDetail{{Name: "prompt1"}})

	ClearClientCache("test-client")

	_, found := globalMCPCache.toolsCache.Get("test-client")
	if found {
		t.Error("expected tools cache to be cleared")
	}

	_, found = globalMCPCache.resourcesCache.Get("test-client")
	if found {
		t.Error("expected resources cache to be cleared")
	}

	_, found = globalMCPCache.promptsCache.Get("test-client")
	if found {
		t.Error("expected prompts cache to be cleared")
	}
}

func TestClearAllCache(t *testing.T) {
	// Use a fresh cache for this test
	globalMCPCache = NewMCPCache()
	globalMCPCache.toolsCache.Set("client1", []ToolInfo{{Name: "tool1"}})
	globalMCPCache.toolsCache.Set("client2", []ToolInfo{{Name: "tool2"}})

	ClearAllCache()

	_, found := globalMCPCache.toolsCache.Get("client1")
	if found {
		t.Error("expected all tools cache to be cleared")
	}

	_, found = globalMCPCache.toolsCache.Get("client2")
	if found {
		t.Error("expected all tools cache to be cleared")
	}
}

// ============================================================================
// Helper Functions Tests
// ============================================================================

func TestShouldSkipSDKPrefix(t *testing.T) {
	// Default should be false
	if shouldSkipSDKPrefix("test-client") {
		t.Error("expected false by default")
	}
}

func TestFormatToolName(t *testing.T) {
	// Tool name without prefix
	result := formatToolName("myTool", "myServer")
	expected := "mcp__myServer__myTool"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	// Tool name already with prefix
	result2 := formatToolName("mcp__server__tool", "myServer")
	expected2 := "mcp__server__tool"
	if result2 != expected2 {
		t.Errorf("expected %s, got %s", expected2, result2)
	}
}

func TestFormatPromptName(t *testing.T) {
	// Prompt name without prefix
	result := formatPromptName("myPrompt", "myServer")
	expected := "mcp__myServer__myPrompt"
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	// Prompt name already with prefix
	result2 := formatPromptName("mcp__server__prompt", "myServer")
	expected2 := "mcp__server__prompt"
	if result2 != expected2 {
		t.Errorf("expected %s, got %s", expected2, result2)
	}
}

func TestAddServerNameToResources(t *testing.T) {
	resources := []ResourceInfo{
		{Name: "resource1"},
		{Name: "resource2"},
	}

	result := addServerNameToResources(resources, "myServer")

	if len(result) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(result))
	}

	if result[0].Name != "myServer/resource1" {
		t.Errorf("expected myServer/resource1, got %s", result[0].Name)
	}

	if result[1].Name != "myServer/resource2" {
		t.Errorf("expected myServer/resource2, got %s", result[1].Name)
	}
}

func TestConvertPromptsToCommands(t *testing.T) {
	prompts := []PromptDetail{
		{Name: "prompt1", Description: "desc1"},
		{Name: "prompt2", Description: "desc2"},
	}

	commands := convertPromptsToCommands(prompts, "myServer")

	if len(commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(commands))
	}

	if commands[0].Name != "mcp__myServer__prompt1" {
		t.Errorf("expected mcp__myServer__prompt1, got %s", commands[0].Name)
	}
	if commands[0].Description != "desc1" {
		t.Errorf("expected desc1, got %s", commands[0].Description)
	}
	if commands[0].Source != "mcp" {
		t.Errorf("expected source 'mcp', got %s", commands[0].Source)
	}
	if commands[0].ServerName != "myServer" {
		t.Errorf("expected server name 'myServer', got %s", commands[0].ServerName)
	}
}

func TestFilterToolsByPermission(t *testing.T) {
	tools := []ToolInfo{
		{Name: "allowed-tool"},
		{Name: "mcp__ide__executeCode"},
		{Name: "mcp__ide__forbidden"},
		{Name: "mcp_sdk_test"},
	}

	filtered := filterToolsByPermission(tools, "test-client")

	// Should include allowed-tool, mcp__ide__executeCode, and mcp_sdk_test (SDK prefix not skipped by default)
	// Should exclude mcp__ide__forbidden
	if len(filtered) != 3 {
		t.Errorf("expected 3 filtered tools, got %d", len(filtered))
	}
}
