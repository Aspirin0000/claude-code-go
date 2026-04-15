package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGrepCommand_NoArgs(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("no args should not error: %v", err)
	}
}

func TestGrepCommand_SingleFile(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := "hello world\nfoo bar\nhello again\n"
	_ = os.WriteFile(tmpFile, []byte(content), 0644)

	if err := cmd.Execute(ctx, []string{"hello", tmpFile}); err != nil {
		t.Fatalf("grep single file failed: %v", err)
	}
}

func TestGrepCommand_Directory(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("hello world"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.txt"), []byte("foo bar"), 0644)

	if err := cmd.Execute(ctx, []string{"hello", tmpDir}); err != nil {
		t.Fatalf("grep directory failed: %v", err)
	}
}

func TestGrepCommand_CaseInsensitive(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(tmpFile, []byte("HELLO World\n"), 0644)

	if err := cmd.Execute(ctx, []string{"-i", "hello", tmpFile}); err != nil {
		t.Fatalf("grep case insensitive failed: %v", err)
	}
}

func TestGrepCommand_Recursive(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "root.txt"), []byte("find me"), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "nested.txt"), []byte("find me too"), 0644)

	if err := cmd.Execute(ctx, []string{"-r", "find", tmpDir}); err != nil {
		t.Fatalf("grep recursive failed: %v", err)
	}
}

func TestGrepCommand_InvalidRegex(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(tmpFile, []byte("content"), 0644)

	// Invalid regex should not return error, just print message and return nil
	if err := cmd.Execute(ctx, []string{"[invalid", tmpFile}); err != nil {
		t.Fatalf("invalid regex should not error: %v", err)
	}
}

func TestGrepCommand_NoMatches(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	_ = os.WriteFile(tmpFile, []byte("hello world\n"), 0644)

	if err := cmd.Execute(ctx, []string{"xyz123", tmpFile}); err != nil {
		t.Fatalf("no matches should not error: %v", err)
	}
}

func TestGrepCommand_GlobPattern(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("search target"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "b.go"), []byte("no match here"), 0644)

	pattern := filepath.Join(tmpDir, "*.txt")
	if err := cmd.Execute(ctx, []string{"target", pattern}); err != nil {
		t.Fatalf("grep glob pattern failed: %v", err)
	}
}

func TestGrepCommand_NonexistentFile(t *testing.T) {
	cmd := NewGrepCommand()
	ctx := context.Background()

	// Should handle nonexistent file gracefully
	if err := cmd.Execute(ctx, []string{"pattern", "/nonexistent/file.txt"}); err != nil {
		t.Fatalf("nonexistent file should not error: %v", err)
	}
}

func TestGrepCommand_Aliases(t *testing.T) {
	cmd := NewGrepCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"grep-files": true}
	for _, alias := range aliases {
		if !expected[alias] {
			t.Errorf("unexpected alias: %s", alias)
		}
		delete(expected, alias)
	}
	if len(expected) > 0 {
		t.Errorf("missing expected aliases: %v", expected)
	}
}
