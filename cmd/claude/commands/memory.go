package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MemoryCommand manages persistent memory notes
type MemoryCommand struct {
	*BaseCommand
	getMemoryFilePath func() string
}

// MemoryEntry represents a single memory entry
type MemoryEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MemoryStore holds all memory entries
type MemoryStore struct {
	Entries []MemoryEntry `json:"entries"`
}

// NewMemoryCommand creates the memory command
func NewMemoryCommand() *MemoryCommand {
	cmd := &MemoryCommand{
		BaseCommand: NewBaseCommand(
			"memory",
			"Manage persistent memory notes",
			CategoryAdvanced,
		).WithAliases("mem", "memo").
			WithHelp(`Usage: /memory [action] [args]

Manage persistent memory notes that survive across sessions.

Actions:
  set <key> <value>  Save a memory note
  get <key>          Retrieve a memory note
  list               List all memory notes
  delete <key>       Delete a memory note
  search <term>      Search memory notes
  clear              Delete all memory notes

Examples:
  /memory set project_language Go
  /memory get project_language
  /memory list
  /memory search project
  /memory delete project_language
  /memory clear`),
	}
	cmd.getMemoryFilePath = func() string {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, ".config", "claude", "memory.json")
	}
	return cmd
}

// Execute runs the memory command
func (c *MemoryCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.listMemories()
	}

	action := strings.ToLower(args[0])

	switch action {
	case "set", "add", "save":
		if len(args) < 3 {
			return fmt.Errorf("usage: /memory set <key> <value>")
		}
		return c.setMemory(args[1], strings.Join(args[2:], " "))
	case "get", "read", "show":
		if len(args) < 2 {
			return fmt.Errorf("usage: /memory get <key>")
		}
		return c.getMemory(args[1])
	case "list", "ls", "all":
		return c.listMemories()
	case "delete", "rm", "remove":
		if len(args) < 2 {
			return fmt.Errorf("usage: /memory delete <key>")
		}
		return c.deleteMemory(args[1])
	case "search", "find":
		if len(args) < 2 {
			return fmt.Errorf("usage: /memory search <term>")
		}
		return c.searchMemories(strings.Join(args[1:], " "))
	case "clear", "purge":
		return c.clearMemories()
	default:
		// If two or more args and no action specified, treat as set
		if len(args) >= 2 {
			return c.setMemory(args[0], strings.Join(args[1:], " "))
		}
		return fmt.Errorf("unknown action: %s", action)
	}
}

func (c *MemoryCommand) loadMemories() (*MemoryStore, error) {
	memPath := c.getMemoryFilePath()

	data, err := os.ReadFile(memPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &MemoryStore{Entries: []MemoryEntry{}}, nil
		}
		return nil, err
	}

	var store MemoryStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	return &store, nil
}

func (c *MemoryCommand) saveMemories(store *MemoryStore) error {
	memPath := c.getMemoryFilePath()

	dir := filepath.Dir(memPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(memPath, data, 0644)
}

func (c *MemoryCommand) setMemory(key, value string) error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	now := time.Now()
	found := false

	for i := range store.Entries {
		if store.Entries[i].Key == key {
			store.Entries[i].Value = value
			store.Entries[i].UpdatedAt = now
			found = true
			break
		}
	}

	if !found {
		store.Entries = append(store.Entries, MemoryEntry{
			Key:       key,
			Value:     value,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if err := c.saveMemories(store); err != nil {
		return fmt.Errorf("failed to save memory: %w", err)
	}

	fmt.Printf("✓ Memory '%s' saved.\n", key)
	return nil
}

func (c *MemoryCommand) getMemory(key string) error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	for _, entry := range store.Entries {
		if entry.Key == key {
			fmt.Printf("%s = %s\n", entry.Key, entry.Value)
			return nil
		}
	}

	return fmt.Errorf("memory '%s' not found", key)
}

func (c *MemoryCommand) listMemories() error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	if len(store.Entries) == 0 {
		fmt.Println("No memories stored yet.")
		fmt.Println("Use '/memory set <key> <value>' to add one.")
		return nil
	}

	fmt.Printf("\n%-20s %-40s %-16s\n", "Key", "Value", "Updated")
	fmt.Println(strings.Repeat("-", 80))

	for _, entry := range store.Entries {
		value := entry.Value
		if len(value) > 38 {
			value = value[:35] + "..."
		}
		fmt.Printf("%-20s %-40s %-16s\n",
			entry.Key,
			value,
			entry.UpdatedAt.Format("2006-01-02 15:04"),
		)
	}

	fmt.Println()
	fmt.Printf("Total: %d memory note(s)\n", len(store.Entries))
	return nil
}

func (c *MemoryCommand) deleteMemory(key string) error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	var newEntries []MemoryEntry
	found := false
	for _, entry := range store.Entries {
		if entry.Key == key {
			found = true
			continue
		}
		newEntries = append(newEntries, entry)
	}

	if !found {
		return fmt.Errorf("memory '%s' not found", key)
	}

	store.Entries = newEntries
	if err := c.saveMemories(store); err != nil {
		return fmt.Errorf("failed to save memories: %w", err)
	}

	fmt.Printf("✓ Memory '%s' deleted.\n", key)
	return nil
}

func (c *MemoryCommand) searchMemories(term string) error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	termLower := strings.ToLower(term)
	var matches []MemoryEntry

	for _, entry := range store.Entries {
		if strings.Contains(strings.ToLower(entry.Key), termLower) ||
			strings.Contains(strings.ToLower(entry.Value), termLower) {
			matches = append(matches, entry)
		}
	}

	if len(matches) == 0 {
		fmt.Printf("No memories found matching '%s'.\n", term)
		return nil
	}

	fmt.Printf("\nFound %d match(es) for '%s':\n\n", len(matches), term)
	for _, entry := range matches {
		fmt.Printf("  %s = %s\n", entry.Key, entry.Value)
	}
	fmt.Println()
	return nil
}

func (c *MemoryCommand) clearMemories() error {
	store, err := c.loadMemories()
	if err != nil {
		return fmt.Errorf("failed to load memories: %w", err)
	}

	if len(store.Entries) == 0 {
		fmt.Println("No memories to clear.")
		return nil
	}

	fmt.Printf("WARNING: This will delete all %d memory notes.\n", len(store.Entries))
	fmt.Print("Are you sure? Type 'yes' to confirm: ")

	var response string
	fmt.Scanln(&response)

	if response != "yes" {
		fmt.Println("Cancelled.")
		return nil
	}

	store.Entries = []MemoryEntry{}
	if err := c.saveMemories(store); err != nil {
		return fmt.Errorf("failed to clear memories: %w", err)
	}

	fmt.Println("✓ All memories cleared.")
	return nil
}

func init() {
	Register(NewMemoryCommand())
}
