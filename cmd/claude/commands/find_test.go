package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindCommand_MissingTerm(t *testing.T) {
	cmd := NewFindCommand()
	err := cmd.Execute(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "missing search term") {
		t.Errorf("expected missing search term error, got: %v", err)
	}
}

func TestFindCommand_Search(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.WriteFile(filepath.Join(tmpDir, "main.go"), []byte("package main"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "main_test.go"), []byte("package main"), 0644)
	_ = os.MkdirAll(filepath.Join(tmpDir, "cmd", "sub"), 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "cmd", "sub", "helper.go"), []byte("package sub"), 0644)

	// Change to temp dir
	origWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(origWd) }()

	cmd := NewFindCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"main"})
	})

	if !strings.Contains(out, "main.go") {
		t.Errorf("expected main.go in output, got: %s", out)
	}
	if !strings.Contains(out, "main_test.go") {
		t.Errorf("expected main_test.go in output, got: %s", out)
	}
	if strings.Contains(out, "helper.go") {
		t.Errorf("did not expect helper.go in output, got: %s", out)
	}
}

func TestFindCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewFindCommand())
	if _, ok := reg.Get("find"); !ok {
		t.Error("find command not registered")
	}
}
