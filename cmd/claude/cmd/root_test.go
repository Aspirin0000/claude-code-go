package cmd

import (
	"testing"
)

func TestRootCommand_Name(t *testing.T) {
	if rootCmd.Name() != "claude" {
		t.Errorf("expected command name 'claude', got %q", rootCmd.Name())
	}
}

func TestRootCommand_Flags(t *testing.T) {
	persistentFlags := []string{"api-key", "model", "verbose"}
	for _, flag := range persistentFlags {
		if rootCmd.PersistentFlags().Lookup(flag) == nil {
			t.Errorf("expected persistent flag %q to be registered", flag)
		}
	}

	flags := []string{"prompt", "json", "serve", "port"}
	for _, flag := range flags {
		if rootCmd.Flags().Lookup(flag) == nil {
			t.Errorf("expected flag %q to be registered", flag)
		}
	}
}

func TestRootCommand_DefaultPort(t *testing.T) {
	port, err := rootCmd.Flags().GetString("port")
	if err != nil {
		t.Fatalf("failed to get port flag: %v", err)
	}
	if port != "8080" {
		t.Errorf("expected default port 8080, got %s", port)
	}
}
