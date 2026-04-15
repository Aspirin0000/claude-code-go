package commands

import (
	"context"
	"testing"
	"time"
)

func TestBashCommand_ParseArgs(t *testing.T) {
	cmd := NewBashCommand()

	tests := []struct {
		name     string
		args     []string
		expected string
		dryRun   bool
	}{
		{"simple echo", []string{"echo", "hello"}, "echo hello", false},
		{"dry run", []string{"--dry-run", "echo", "hello"}, "echo hello", true},
		{"with -c", []string{"-c", "echo", "hello"}, "echo hello", false},
		{"timeout", []string{"--timeout", "5", "echo", "hello"}, "echo hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := cmd.parseArgs(tt.args)
			if err != nil {
				t.Fatalf("parseArgs failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("parseArgs() = %q, want %q", result, tt.expected)
			}
			if cmd.dryRun != tt.dryRun {
				t.Errorf("dryRun = %v, want %v", cmd.dryRun, tt.dryRun)
			}
		})
		// Reset dryRun for next test
		cmd.dryRun = false
	}
}

func TestBashCommand_ParseArgs_InvalidTimeout(t *testing.T) {
	cmd := NewBashCommand()
	_, err := cmd.parseArgs([]string{"--timeout", "invalid", "echo", "hello"})
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
}

func TestBashCommand_ValidateCommand(t *testing.T) {
	cmd := NewBashCommand()

	// Dangerous commands should be blocked
	dangerous := []string{
		"rm -rf /",
		"dd if=/dev/zero of=/dev/sda",
		":(){ :|:& };:",
	}
	for _, command := range dangerous {
		if err := cmd.validateCommand(command); err == nil {
			t.Errorf("expected dangerous command to be blocked: %s", command)
		}
	}

	// Safe commands should pass
	safe := []string{
		"echo hello",
		"ls -la",
		"pwd",
	}
	for _, command := range safe {
		if err := cmd.validateCommand(command); err != nil {
			t.Errorf("expected safe command to pass: %s", command)
		}
	}
}

func TestBashCommand_IsReadonlySafe(t *testing.T) {
	cmd := NewBashCommand()

	safe := []string{"ls", "cat file.txt", "pwd", "echo hello"}
	for _, command := range safe {
		if !cmd.isReadonlySafe(command) {
			t.Errorf("expected command to be readonly safe: %s", command)
		}
	}

	unsafe := []string{"rm file", "mv a b", "cp a b", "chmod 755 file"}
	for _, command := range unsafe {
		if cmd.isReadonlySafe(command) {
			t.Errorf("expected command to be unsafe: %s", command)
		}
	}
}

func TestBashCommand_IsDangerous(t *testing.T) {
	cmd := NewBashCommand()

	dangerous := []string{"rm file", "mv a b", "cp a b", "echo hello > file.txt"}
	for _, command := range dangerous {
		if !cmd.isDangerous(command) {
			t.Errorf("expected command to be dangerous: %s", command)
		}
	}

	safe := []string{"ls", "cat file.txt", "pwd", "echo hello"}
	for _, command := range safe {
		if cmd.isDangerous(command) {
			t.Errorf("expected command to be safe: %s", command)
		}
	}
}

func TestBashCommand_CheckPermissions_ReadOnlyBlocksBash(t *testing.T) {
	cmd := NewBashCommand()

	// Temporarily set permission level to read-only
	oldLevel := GetCurrentPermissionLevel()
	_ = SetPermissionLevel(PermissionLevelReadOnly)
	defer SetPermissionLevel(oldLevel)

	if err := cmd.checkPermissions("echo hello"); err == nil {
		t.Error("expected bash to be blocked in read-only mode")
	}
}

func TestBashCommand_CheckPermissions_ReadOnlyBlocksBashEvenIfSafe(t *testing.T) {
	cmd := NewBashCommand()

	oldLevel := GetCurrentPermissionLevel()
	_ = SetPermissionLevel(PermissionLevelReadOnly)
	defer SetPermissionLevel(oldLevel)

	// In read-only mode, the bash tool itself is blocked regardless of command safety
	if err := cmd.checkPermissions("ls"); err == nil {
		t.Error("expected bash tool to be blocked in read-only mode even for safe commands")
	}
}

func TestBashCommand_DryRunExecution(t *testing.T) {
	cmd := NewBashCommand()
	cmd.SetDryRun(true)
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{"echo", "hello"}); err != nil {
		t.Fatalf("dry run execution failed: %v", err)
	}
}

func TestBashCommand_RealSafeExecution(t *testing.T) {
	cmd := NewBashCommand()
	cmd.SetTimeout(5 * time.Second)
	ctx := context.Background()

	// Use full permission level to avoid confirmation prompt
	oldLevel := GetCurrentPermissionLevel()
	_ = SetPermissionLevel(PermissionLevelFull)
	defer SetPermissionLevel(oldLevel)

	if err := cmd.Execute(ctx, []string{"echo", "hello"}); err != nil {
		t.Fatalf("safe execution failed: %v", err)
	}
}

func TestBashCommand_NoArgs(t *testing.T) {
	cmd := NewBashCommand()
	ctx := context.Background()

	if err := cmd.Execute(ctx, []string{}); err != nil {
		t.Fatalf("no args should not error: %v", err)
	}
}

func TestBashCommand_HelpFlag(t *testing.T) {
	cmd := NewBashCommand()
	ctx := context.Background()

	// --help prints help and returns "no command specified" error
	if err := cmd.Execute(ctx, []string{"--help"}); err == nil {
		t.Fatal("expected error for help flag (no command specified)")
	}
}

func TestBashCommand_Aliases(t *testing.T) {
	cmd := NewBashCommand()
	aliases := cmd.Aliases()

	expected := map[string]bool{"sh": true, "shell": true, "exec": true}
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
