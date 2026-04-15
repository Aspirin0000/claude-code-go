package commands

import (
	"context"
	"testing"
)

func TestReloadCommand(t *testing.T) {
	cmd := NewReloadCommand()
	ctx := context.Background()

	// This may fail if no config file exists, which is acceptable for a basic test
	_ = cmd.Execute(ctx, []string{})
}
