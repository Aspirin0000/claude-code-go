package commands

import (
	"context"
	"path/filepath"
	"testing"
)

func TestSkillsCommand_ListEmpty(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()
	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("list empty skills failed: %v", err)
	}
}

func TestSkillsCommand_AddAndShow(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"add", "review", "Review this code for bugs."}); err != nil {
		t.Fatalf("add skill failed: %v", err)
	}

	if err := cmd.Execute(ctx, []string{"show", "review"}); err != nil {
		t.Fatalf("show skill failed: %v", err)
	}
}

func TestSkillsCommand_Use(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()
	_ = cmd.Execute(ctx, []string{"add", "test", "Test prompt"})

	if err := cmd.Execute(ctx, []string{"use", "test"}); err != nil {
		t.Fatalf("use skill failed: %v", err)
	}
}

func TestSkillsCommand_Edit(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()
	_ = cmd.Execute(ctx, []string{"add", "editme", "Original prompt"})

	if err := cmd.Execute(ctx, []string{"edit", "editme", "Updated prompt"}); err != nil {
		t.Fatalf("edit skill failed: %v", err)
	}
}

func TestSkillsCommand_Remove(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()
	_ = cmd.Execute(ctx, []string{"add", "removeme", "Prompt"})

	if err := cmd.Execute(ctx, []string{"remove", "removeme"}); err != nil {
		t.Fatalf("remove skill failed: %v", err)
	}
}

func TestSkillsCommand_DuplicateAdd(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()
	_ = cmd.Execute(ctx, []string{"add", "dup", "Prompt"})

	if err := cmd.Execute(ctx, []string{"add", "dup", "Another"}); err == nil {
		t.Fatal("expected error for duplicate skill name")
	}
}

func TestSkillsCommand_NotFound(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"show", "missing"}); err == nil {
		t.Fatal("expected error for missing skill")
	}

	if err := cmd.Execute(ctx, []string{"use", "missing"}); err == nil {
		t.Fatal("expected error for missing skill")
	}

	if err := cmd.Execute(ctx, []string{"remove", "missing"}); err == nil {
		t.Fatal("expected error for missing skill")
	}

	if err := cmd.Execute(ctx, []string{"edit", "missing", "new"}); err == nil {
		t.Fatal("expected error for missing skill")
	}
}

func TestSkillsCommand_InvalidSubcommand(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"unknown"}); err == nil {
		t.Fatal("expected error for unknown subcommand")
	}
}

func TestSkillsCommand_MissingArgs(t *testing.T) {
	cmd := NewSkillsCommand()
	tmpDir := t.TempDir()
	cmd.getSkillsFilePath = func() string {
		return filepath.Join(tmpDir, "skills.json")
	}

	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"add"}); err == nil {
		t.Fatal("expected error for missing add args")
	}

	if err := cmd.Execute(ctx, []string{"show"}); err == nil {
		t.Fatal("expected error for missing show arg")
	}

	if err := cmd.Execute(ctx, []string{"use"}); err == nil {
		t.Fatal("expected error for missing use arg")
	}

	if err := cmd.Execute(ctx, []string{"remove"}); err == nil {
		t.Fatal("expected error for missing remove arg")
	}

	if err := cmd.Execute(ctx, []string{"edit"}); err == nil {
		t.Fatal("expected error for missing edit args")
	}
}
