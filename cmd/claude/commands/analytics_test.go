package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAnalyticsCommand_Status(t *testing.T) {
	cmd := NewAnalyticsCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"status"})
	})

	if !strings.Contains(out, "Analytics Status") {
		t.Errorf("expected header, got: %s", out)
	}
}

func TestAnalyticsCommand_EnableDisable(t *testing.T) {
	// Use temp config dir
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer func() {
		if origHome == "" {
			os.Unsetenv("HOME")
		} else {
			os.Setenv("HOME", origHome)
		}
	}()

	cmd := NewAnalyticsCommand()

	// Disable
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"disable"})
	})
	if !strings.Contains(out, "Analytics disabled") {
		t.Errorf("expected disable confirmation, got: %s", out)
	}

	// Check that disable was successful (file should exist somewhere under tmpDir)
	var found bool
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == ".analytics-disabled" {
			found = true
		}
		return nil
	})
	if !found {
		t.Error("expected disable file to exist somewhere in temp dir")
	}

	// Enable
	out = captureOutput(func() {
		_ = cmd.Execute(nil, []string{"enable"})
	})
	if !strings.Contains(out, "Analytics enabled") {
		t.Errorf("expected enable confirmation, got: %s", out)
	}

	// Verify file is removed
	found = false
	filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == ".analytics-disabled" {
			found = true
		}
		return nil
	})
	if found {
		t.Error("expected disable file to be removed")
	}
}

func TestAnalyticsCommand_InvalidAction(t *testing.T) {
	cmd := NewAnalyticsCommand()
	err := cmd.Execute(nil, []string{"invalid"})
	if err == nil || !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("expected unknown action error, got: %v", err)
	}
}

func TestAnalyticsCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewAnalyticsCommand())
	if _, ok := reg.Get("analytics"); !ok {
		t.Error("analytics command not registered")
	}
}
