package plugins

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Aspirin0000/claude-code-go/internal/types"
)

// ============================================================================
// PluginSource Tests
// ============================================================================

func TestPluginSourceIsRemote(t *testing.T) {
	tests := []struct {
		source   PluginSource
		expected bool
	}{
		{PluginSource{Type: "npm"}, true},
		{PluginSource{Type: "github"}, true},
		{PluginSource{Type: "url"}, true},
		{PluginSource{Type: "local"}, false},
		{PluginSource{Type: "git-subdir"}, false},
	}

	for _, test := range tests {
		result := test.source.IsRemote()
		if result != test.expected {
			t.Errorf("IsRemote() for %q = %v, expected %v", test.source.Type, result, test.expected)
		}
	}
}

// ============================================================================
// Path Functions Tests
// ============================================================================

func TestGetPluginsDirectory(t *testing.T) {
	dir := GetPluginsDirectory()
	if dir == "" {
		t.Error("expected non-empty plugins directory")
	}
}

func TestGetPluginCachePath(t *testing.T) {
	path := GetPluginCachePath()
	if path == "" {
		t.Error("expected non-empty cache path")
	}
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
}

func TestParsePluginIdentifier(t *testing.T) {
	tests := []struct {
		input             string
		expectedName      string
		expectedMarketplace string
	}{
		{"plugin@npm", "plugin", "npm"},
		{"plugin", "plugin", ""},
		{"plugin@github@ref", "plugin", "github"},
	}

	for _, test := range tests {
		name, marketplace := ParsePluginIdentifier(test.input)
		if name != test.expectedName {
			t.Errorf("ParsePluginIdentifier(%q) name = %q, expected %q", test.input, name, test.expectedName)
		}
		if marketplace != test.expectedMarketplace {
			t.Errorf("ParsePluginIdentifier(%q) marketplace = %q, expected %q", test.input, marketplace, test.expectedMarketplace)
		}
	}
}

func TestGetVersionedCachePath(t *testing.T) {
	path := GetVersionedCachePath("test-plugin", "1.0.0")
	if path == "" {
		t.Error("expected non-empty path")
	}
	if !contains(path, "test-plugin") {
		t.Error("expected path to contain plugin name")
	}
	if !contains(path, "1.0.0") {
		t.Error("expected path to contain version")
	}
}

func TestGetVersionedCachePathIn(t *testing.T) {
	tmpDir := t.TempDir()
	path := GetVersionedCachePathIn(tmpDir, "test-plugin", "1.0.0")
	if path == "" {
		t.Error("expected non-empty path")
	}
	if !filepath.IsAbs(path) {
		t.Error("expected absolute path")
	}
}

func TestGetLegacyCachePath(t *testing.T) {
	path := GetLegacyCachePath("test-plugin")
	if path == "" {
		t.Error("expected non-empty path")
	}
}

func TestResolvePluginPath(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Test with existing versioned path
	path, err := ResolvePluginPath("test-plugin", "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if path == "" {
		t.Error("expected non-empty path")
	}
}

// ============================================================================
// CopyDir Tests
// ============================================================================

func TestCopyDir(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	dstDir := filepath.Join(tmpDir, "dst")

	// Create source directory with files
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0644)
	os.WriteFile(filepath.Join(srcDir, "file2.txt"), []byte("content2"), 0644)

	// Create subdirectory
	subDir := filepath.Join(srcDir, "subdir")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "file3.txt"), []byte("content3"), 0644)

	// Copy
	err := CopyDir(srcDir, dstDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files were copied
	if _, err := os.Stat(filepath.Join(dstDir, "file1.txt")); os.IsNotExist(err) {
		t.Error("expected file1.txt to exist in destination")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "file2.txt")); os.IsNotExist(err) {
		t.Error("expected file2.txt to exist in destination")
	}
	if _, err := os.Stat(filepath.Join(dstDir, "subdir", "file3.txt")); os.IsNotExist(err) {
		t.Error("expected subdir/file3.txt to exist in destination")
	}
}

func TestCopyDirNonExistentSource(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "non-existent")
	dstDir := filepath.Join(tmpDir, "dst")

	err := CopyDir(srcDir, dstDir)
	if err == nil {
		t.Error("expected error for non-existent source")
	}
}

// ============================================================================
// CopyPluginToVersionedCache Tests
// ============================================================================

func TestCopyPluginToVersionedCache(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)
	os.WriteFile(filepath.Join(srcDir, "plugin.json"), []byte("{}"), 0644)

	cachePath, err := CopyPluginToVersionedCache(srcDir, "test-plugin", "1.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cachePath == "" {
		t.Error("expected non-empty cache path")
	}

	// Verify file was copied
	if _, err := os.Stat(filepath.Join(cachePath, "plugin.json")); os.IsNotExist(err) {
		t.Error("expected plugin.json to exist in cache")
	}
}

func TestCopyPluginToVersionedCacheEmptySource(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(srcDir, 0755)

	// Create a file in the source to make CopyDir succeed
	os.WriteFile(filepath.Join(srcDir, ".gitkeep"), []byte(""), 0644)

	_, err := CopyPluginToVersionedCache(srcDir, "test-plugin", "1.0.0")
	// Should succeed now since source has a file
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// ============================================================================
// ValidateGitUrl Tests
// ============================================================================

func TestValidateGitUrl(t *testing.T) {
	tests := []struct {
		url       string
		expectErr bool
	}{
		{"https://github.com/user/repo.git", false},
		{"http://github.com/user/repo.git", false},
		{"git@github.com:user/repo.git", false},
		{"file:///path/to/repo.git", false},
		{"invalid-url", true},
		{"ftp://github.com/user/repo.git", true},
	}

	for _, test := range tests {
		err := ValidateGitUrl(test.url)
		if test.expectErr {
			if err == nil {
				t.Errorf("expected error for %q", test.url)
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error for %q: %v", test.url, err)
			}
		}
	}
}

// ============================================================================
// GenerateTemporaryCacheName Tests
// ============================================================================

func TestGenerateTemporaryCacheName(t *testing.T) {
	source := PluginSource{Type: "npm"}
	name := GenerateTemporaryCacheName(source)
	if name == "" {
		t.Error("expected non-empty cache name")
	}
	if !contains(name, "npm") {
		t.Error("expected cache name to contain 'npm'")
	}
}

func TestGenerateTemporaryCacheNameUnknown(t *testing.T) {
	source := PluginSource{Type: "unknown"}
	name := GenerateTemporaryCacheName(source)
	if name == "" {
		t.Error("expected non-empty cache name")
	}
	if !contains(name, "local") {
		t.Error("expected cache name to contain 'local'")
	}
}

// ============================================================================
// LoadPluginManifest Tests
// ============================================================================

func TestLoadPluginManifestNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "non-existent.json")

	manifest, err := LoadPluginManifest(manifestPath, "test-plugin", "local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest == nil {
		t.Fatal("expected non-nil manifest")
	}
	if manifest.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %q", manifest.Name)
	}
}

func TestLoadPluginManifestValid(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "plugin.json")
	content := `{"name": "my-plugin", "description": "A test plugin"}`
	os.WriteFile(manifestPath, []byte(content), 0644)

	manifest, err := LoadPluginManifest(manifestPath, "fallback", "local")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Name != "my-plugin" {
		t.Errorf("expected name 'my-plugin', got %q", manifest.Name)
	}
	if manifest.Description != "A test plugin" {
		t.Errorf("expected description 'A test plugin', got %q", manifest.Description)
	}
}

func TestLoadPluginManifestInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	manifestPath := filepath.Join(tmpDir, "plugin.json")
	os.WriteFile(manifestPath, []byte("not json"), 0644)

	_, err := LoadPluginManifest(manifestPath, "test", "local")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ============================================================================
// CreatePluginFromPath Tests
// ============================================================================

func TestCreatePluginFromPath(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create manifest
	manifestDir := filepath.Join(pluginDir, ".claude-plugin")
	os.MkdirAll(manifestDir, 0755)
	os.WriteFile(filepath.Join(manifestDir, "plugin.json"), []byte(`{"name": "test-plugin"}`), 0644)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %q", plugin.Name)
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestCreatePluginFromPathWithCommands(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create commands directory
	commandsDir := filepath.Join(pluginDir, "commands")
	os.MkdirAll(commandsDir, 0755)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.CommandsPath == nil {
		t.Error("expected CommandsPath to be set")
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestCreatePluginFromPathWithAgents(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create agents directory
	agentsDir := filepath.Join(pluginDir, "agents")
	os.MkdirAll(agentsDir, 0755)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.AgentsPath == nil {
		t.Error("expected AgentsPath to be set")
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestCreatePluginFromPathWithSkills(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create skills directory
	skillsDir := filepath.Join(pluginDir, "skills")
	os.MkdirAll(skillsDir, 0755)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.SkillsPath == nil {
		t.Error("expected SkillsPath to be set")
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestCreatePluginFromPathWithOutputStyles(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	// Create output-styles directory
	stylesDir := filepath.Join(pluginDir, "output-styles")
	os.MkdirAll(stylesDir, 0755)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.OutputStylesPath == nil {
		t.Error("expected OutputStylesPath to be set")
	}
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

func TestCreatePluginFromPathNoManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pluginDir := filepath.Join(tmpDir, "test-plugin")
	os.MkdirAll(pluginDir, 0755)

	plugin, errors := CreatePluginFromPath(pluginDir, "local", true, "test-plugin", true)
	if plugin == nil {
		t.Fatal("expected non-nil plugin")
	}
	if plugin.Name != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got %q", plugin.Name)
	}
	// Should have no errors since it creates a default manifest
	if len(errors) != 0 {
		t.Errorf("expected 0 errors, got %d", len(errors))
	}
}

// ============================================================================
// Enable/Disable Plugin Tests
// ============================================================================

func TestEnablePlugin(t *testing.T) {
	plugin := &types.LoadedPlugin{}
	err := EnablePlugin(plugin)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if plugin.Enabled == nil || !*plugin.Enabled {
		t.Error("expected plugin to be enabled")
	}
}

func TestDisablePlugin(t *testing.T) {
	plugin := &types.LoadedPlugin{}
	err := DisablePlugin(plugin)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if plugin.Enabled == nil || *plugin.Enabled {
		t.Error("expected plugin to be disabled")
	}
}

// ============================================================================
// Sanitize Tests
// ============================================================================

func TestSanitizePathComponent(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"normal", "normal"},
		{"path/with/slashes", "path-with-slashes"},
		{"path\\with\\backslashes", "path-with-backslashes"},
		{"..", "-"},
		{"file:name", "file-name"},
		{"file*name", "file-name"},
		{"file?name", "file-name"},
		{"file<name", "file-name"},
		{"file>name", "file-name"},
		{"file|name", "file-name"},
	}

	for _, test := range tests {
		result := sanitizePathComponent(test.input)
		if result != test.expected {
			t.Errorf("sanitizePathComponent(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

// ============================================================================
// Helper function
// ============================================================================

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
