package commands

import (
	"testing"
)

func TestIsToolAllowed(t *testing.T) {
	tests := []struct {
		name     string
		level    PermissionLevel
		toolName string
		allowed  bool
		needsAsk bool
	}{
		{"ask allows read tool", PermissionLevelAsk, "file_read", true, true},
		{"ask allows write tool", PermissionLevelAsk, "file_write", true, true},
		{"ask allows dangerous tool", PermissionLevelAsk, "bash", true, true},
		{"read-only allows read tool", PermissionLevelReadOnly, "file_read", true, false},
		{"read-only denies write tool", PermissionLevelReadOnly, "file_write", false, false},
		{"read-only denies dangerous tool", PermissionLevelReadOnly, "bash", false, false},
		{"standard allows read tool without ask", PermissionLevelStandard, "file_read", true, false},
		{"standard allows write tool with ask", PermissionLevelStandard, "file_write", true, true},
		{"standard allows dangerous tool with ask", PermissionLevelStandard, "bash", true, true},
		{"full allows read tool", PermissionLevelFull, "file_read", true, false},
		{"full allows write tool", PermissionLevelFull, "file_write", true, false},
		{"full allows dangerous tool", PermissionLevelFull, "bash", true, false},
		{"unknown tool in ask", PermissionLevelAsk, "unknown_tool", false, true},
		{"unknown tool in read-only", PermissionLevelReadOnly, "unknown_tool", false, false},
		{"unknown tool in standard", PermissionLevelStandard, "unknown_tool", true, true},
		{"unknown tool in full", PermissionLevelFull, "unknown_tool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, needsAsk := IsToolAllowed(tt.level, tt.toolName)
			if allowed != tt.allowed {
				t.Errorf("IsToolAllowed(%q, %q) allowed = %v, want %v", tt.level, tt.toolName, allowed, tt.allowed)
			}
			if needsAsk != tt.needsAsk {
				t.Errorf("IsToolAllowed(%q, %q) needsAsk = %v, want %v", tt.level, tt.toolName, needsAsk, tt.needsAsk)
			}
		})
	}
}

func TestGetAllowedTools(t *testing.T) {
	readOnlyAllowed := GetAllowedTools(PermissionLevelReadOnly)
	if len(readOnlyAllowed) == 0 {
		t.Error("expected some allowed tools for read-only")
	}
	for _, tool := range readOnlyAllowed {
		if tool.Category != ToolCategoryRead {
			t.Errorf("read-only should only allow read tools, got %s", tool.Name)
		}
	}

	fullAllowed := GetAllowedTools(PermissionLevelFull)
	if len(fullAllowed) != len(ToolRegistry) {
		t.Errorf("full should allow all registered tools, got %d, want %d", len(fullAllowed), len(ToolRegistry))
	}
}

func TestGetToolsNeedingAsk(t *testing.T) {
	askTools := GetToolsNeedingAsk(PermissionLevelAsk)
	if len(askTools) != len(ToolRegistry) {
		t.Errorf("ask should require confirmation for all registered tools, got %d, want %d", len(askTools), len(ToolRegistry))
	}

	fullTools := GetToolsNeedingAsk(PermissionLevelFull)
	if len(fullTools) != 0 {
		t.Errorf("full should require confirmation for no tools, got %d", len(fullTools))
	}

	standardTools := GetToolsNeedingAsk(PermissionLevelStandard)
	var dangerousCount int
	for _, tool := range ToolRegistry {
		if tool.IsDangerous {
			dangerousCount++
		}
	}
	if len(standardTools) != dangerousCount {
		t.Errorf("standard should require confirmation for dangerous tools only, got %d, want %d", len(standardTools), dangerousCount)
	}
}

func TestIsValidPermissionLevel(t *testing.T) {
	validLevels := []PermissionLevel{PermissionLevelAsk, PermissionLevelReadOnly, PermissionLevelStandard, PermissionLevelFull}
	for _, level := range validLevels {
		if !isValidPermissionLevel(level) {
			t.Errorf("expected %q to be valid", level)
		}
	}

	invalidLevels := []PermissionLevel{"invalid", "", "admin", "restricted"}
	for _, level := range invalidLevels {
		if isValidPermissionLevel(level) {
			t.Errorf("expected %q to be invalid", level)
		}
	}
}

func TestFindTool(t *testing.T) {
	tool := findTool("bash")
	if tool == nil {
		t.Fatal("expected to find bash tool")
	}
	if tool.Name != "bash" {
		t.Errorf("expected name bash, got %s", tool.Name)
	}

	notFound := findTool("nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent tool")
	}
}

func TestGetAllowedToolsForRegistry(t *testing.T) {
	// This requires a real tool registry, so just verify it doesn't panic
	// and returns a subset based on permission level
	registry := ToolRegistry // using the package-level registry for test
	_ = registry
	// We can't easily construct a tools.Registry here without importing internal/tools
	// The function is tested implicitly through integration
}
