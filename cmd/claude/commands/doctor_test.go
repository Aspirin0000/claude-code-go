package commands

import (
	"context"
	"strings"
	"testing"
)

func TestDoctorCommandExecute(t *testing.T) {
	cmd := NewDoctorCommand()
	err := cmd.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoctorCommandCheckGoVersion(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkGoVersion()
	if !passed && msg != "not installed" {
		t.Errorf("unexpected message when failed: %s", msg)
	}
}

func TestDoctorCommandCheckGit(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkGit()
	if !passed && msg != "not installed" {
		t.Errorf("unexpected message when failed: %s", msg)
	}
	if passed && !strings.Contains(msg, "git version") {
		t.Logf("git check returned: %s", msg)
	}
}

func TestDoctorCommandCheckRipgrep(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkRipgrep()
	if !passed && msg != "not installed" {
		t.Errorf("unexpected message when failed: %s", msg)
	}
	if passed && msg == "" {
		t.Errorf("expected non-empty message for ripgrep")
	}
}

func TestDoctorCommandCheckEnv(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkEnv()
	if !passed {
		t.Errorf("expected env check to pass, got: %s", msg)
	}
}

func TestDoctorCommandCheckConfig(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, _ := cmd.checkConfig()
	// Should pass whether config dir exists or not
	if !passed {
		t.Error("expected config check to pass")
	}
}

func TestDoctorCommandCheckAPIKey(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkAPIKey()
	// Result depends on environment; just ensure it doesn't panic
	_ = passed
	_ = msg
}

func TestDoctorCommandCheckNetwork(t *testing.T) {
	cmd := NewDoctorCommand()
	passed, msg := cmd.checkNetwork()
	// Network-dependent test
	_ = passed
	_ = msg
}
