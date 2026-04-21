package commands

import (
	"os"
	"strings"
	"testing"
)

func TestEnvCommand_ShowAll(t *testing.T) {
	cmd := NewEnvCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, nil)
	})

	if !strings.Contains(out, "Environment Variables") {
		t.Errorf("expected header, got: %s", out)
	}
	if !strings.Contains(out, "PATH=") {
		t.Error("expected PATH to be shown")
	}
}

func TestEnvCommand_ShowOne(t *testing.T) {
	os.Setenv("TEST_VAR", "test_value")
	defer os.Unsetenv("TEST_VAR")

	cmd := NewEnvCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"TEST_VAR"})
	})

	if !strings.Contains(out, "TEST_VAR=test_value") {
		t.Errorf("expected TEST_VAR value, got: %s", out)
	}
}

func TestEnvCommand_Set(t *testing.T) {
	cmd := NewEnvCommand()
	out := captureOutput(func() {
		_ = cmd.Execute(nil, []string{"NEW_TEST_VAR", "new_value"})
	})

	if !strings.Contains(out, "Set NEW_TEST_VAR=new_value") {
		t.Errorf("expected set confirmation, got: %s", out)
	}

	if os.Getenv("NEW_TEST_VAR") != "new_value" {
		t.Error("expected variable to be set")
	}

	os.Unsetenv("NEW_TEST_VAR")
}

func TestEnvCommand_Registered(t *testing.T) {
	reg := NewRegistry()
	_ = reg.Register(NewEnvCommand())
	if _, ok := reg.Get("env"); !ok {
		t.Error("env command not registered")
	}
}
