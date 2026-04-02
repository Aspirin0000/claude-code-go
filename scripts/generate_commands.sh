#!/bin/bash
# Batch command generator

cd /Users/hebing/Documents/Open Code/Project 4.1/cmd/claude/commands

# Create 50 more commands in batches

# Batch 1: System & Info (10 commands)
cat > touch.go << 'ENDOFFILE'
package commands
import ("context"; "fmt"; "os"; "time")
type TouchCommand struct{*BaseCommand}
func NewTouchCommand() *TouchCommand { return &TouchCommand{BaseCommand: NewBaseCommand("touch", "Create empty file or update timestamp", CategoryFiles).WithHelp("Create empty file or update timestamp")}}
func (c *TouchCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: /touch <file>") }
	for _, f := range args {
		file, err := os.OpenFile(f, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil { return err }
		file.Close()
		os.Chtimes(f, time.Now(), time.Now())
		fmt.Printf("✓ Touched: %s\n", f)
	}
	return nil
}
func init() { Register(NewTouchCommand()) }
ENDOFFILE

cat > chmod.go << 'ENDOFFILE'
package commands
import ("context"; "fmt"; "os"; "strconv")
type ChmodCommand struct{*BaseCommand}
func NewChmodCommand() *ChmodCommand { return &ChmodCommand{BaseCommand: NewBaseCommand("chmod", "Change file permissions", CategoryFiles).WithHelp("Change file permissions")}}
func (c *ChmodCommand) Execute(ctx context.Context, args []string) error {
	if len(args) < 2 { return fmt.Errorf("usage: /chmod <mode> <file>") }
	mode, _ := strconv.ParseInt(args[0], 8, 32)
	return os.Chmod(args[1], os.FileMode(mode))
}
func init() { Register(NewChmodCommand()) }
ENDOFFILE

echo "Created 2 commands: touch, chmod"
