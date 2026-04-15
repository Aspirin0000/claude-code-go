package commands

import (
	"context"
	"strings"
	"testing"
)

func TestGitCommand_IsGitRepo(t *testing.T) {
	cmd := NewGitCommand()
	if !cmd.IsGitRepo() {
		t.Fatal("expected current directory to be a git repository")
	}
}

func TestGitCommand_GetCurrentBranch(t *testing.T) {
	cmd := NewGitCommand()
	branch := cmd.GetCurrentBranch()
	if branch == "" || branch == "unknown" {
		t.Errorf("expected valid branch name, got %q", branch)
	}
}

func TestGitCommand_GetLastCommit(t *testing.T) {
	cmd := NewGitCommand()
	commit := cmd.GetLastCommit()
	if commit == "" || commit == "unknown" {
		t.Errorf("expected valid commit info, got %q", commit)
	}
}

func TestGitCommand_GetRepoStatus(t *testing.T) {
	cmd := NewGitCommand()
	status := cmd.GetRepoStatus()
	if status == "" {
		t.Error("expected non-empty repo status")
	}
	if !strings.Contains(status, cmd.GetCurrentBranch()) {
		t.Error("expected repo status to contain branch name")
	}
}

func TestGitCommand_Status(t *testing.T) {
	cmd := NewGitCommand()
	if err := cmd.status(); err != nil {
		t.Fatalf("git status failed: %v", err)
	}
}

func TestGitCommand_Log(t *testing.T) {
	cmd := NewGitCommand()
	if err := cmd.log([]string{"5"}); err != nil {
		t.Fatalf("git log failed: %v", err)
	}
}

func TestGitCommand_Diff(t *testing.T) {
	cmd := NewGitCommand()
	if err := cmd.diff([]string{}); err != nil {
		t.Fatalf("git diff failed: %v", err)
	}
}

func TestGitCommand_Branch(t *testing.T) {
	cmd := NewGitCommand()
	if err := cmd.branch([]string{}); err != nil {
		t.Fatalf("git branch failed: %v", err)
	}
}

func TestGitCommand_RunGitCommand(t *testing.T) {
	cmd := NewGitCommand()
	ctx := context.Background()

	// Run a safe, read-only git command
	if err := cmd.Execute(ctx, []string{"remote", "-v"}); err != nil {
		t.Fatalf("git remote failed: %v", err)
	}
}

func TestGitCommand_Help(t *testing.T) {
	cmd := NewGitCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"help"}); err != nil {
		t.Fatalf("git help failed: %v", err)
	}
}

func TestGitCommand_NoArgs(t *testing.T) {
	cmd := NewGitCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("git no args should not error: %v", err)
	}
}

func TestGitCommand_Aliases(t *testing.T) {
	cmd := NewGitCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"g": true, "vcs": true}
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

func TestGitCommand_ColorizeStatusLine(t *testing.T) {
	cmd := NewGitCommand()

	// Just verify it doesn't panic on various line types
	lines := []string{
		"## main...origin/main",
		" M modified.go",
		"A  added.go",
		" D deleted.go",
		"?? untracked.go",
		"  normal line",
	}
	for _, line := range lines {
		cmd.colorizeStatusLine(line)
	}
}
