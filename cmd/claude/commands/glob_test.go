package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGlobCommand_NoArgs(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	// No args should print help and return nil
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("no args should not error: %v", err)
	}
}

func TestGlobCommand_SimplePattern(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	// Create temp directory with files
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "test2.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "readme.md"), []byte("# readme"), 0644)

	if err := cmd.Execute(ctx, []string{"*.go", tmpDir}); err != nil {
		t.Fatalf("simple glob failed: %v", err)
	}
}

func TestGlobCommand_RecursivePattern(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.MkdirAll(subDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "root.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(subDir, "nested.go"), []byte("package sub"), 0644)

	if err := cmd.Execute(ctx, []string{"**/*.go", tmpDir}); err != nil {
		t.Fatalf("recursive glob failed: %v", err)
	}
}

func TestGlobCommand_NoMatches(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("text"), 0644)

	if err := cmd.Execute(ctx, []string{"*.go", tmpDir}); err != nil {
		t.Fatalf("no matches should not error: %v", err)
	}
}

func TestGlobCommand_InvalidDirectory(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"*.go", "/nonexistent/path/12345"}); err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestGlobCommand_FileAsDirectory(t *testing.T) {
	cmd := NewGlobCommand()
	ctx := context.Background()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "notadir.txt")
	_ = os.WriteFile(tmpFile, []byte("text"), 0644)

	if err := cmd.Execute(ctx, []string{"*.go", tmpFile}); err == nil {
		t.Fatal("expected error when passing a file as directory")
	}
}

func TestGlobCommand_Aliases(t *testing.T) {
	cmd := NewGlobCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"find": true, "files": true}
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
