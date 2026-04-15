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

// Skill represents a reusable prompt template
type Skill struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Prompt      string    `json:"prompt"`
	CreatedAt   time.Time `json:"created_at"`
}

// SkillsData holds all user skills
type SkillsData struct {
	Skills    []Skill   `json:"skills"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SkillsCommand manages reusable prompt templates
// Source: src/commands/skills/
type SkillsCommand struct {
	*BaseCommand
	getSkillsFilePath func() string
}

// NewSkillsCommand creates the /skills command
func NewSkillsCommand() *SkillsCommand {
	cmd := &SkillsCommand{
		BaseCommand: NewBaseCommand(
			"skills",
			"Manage reusable prompt templates (skills)",
			CategoryAdvanced,
		).WithHelp(`Usage: /skills [subcommand] [args...]

Manage reusable prompt templates (skills) for common tasks.

Subcommands:
  (no args)              List all skills
  add <name> <prompt>    Add a new skill
  show <name>            Display a skill's prompt
  use <name>             Copy a skill prompt to clipboard (or print it)
  remove <name>          Remove a skill
  edit <name> <prompt>   Update an existing skill

Examples:
  /skills
  /skills add review "Review this code for bugs, performance, and style issues."
  /skills show review
  /skills use review
  /skills remove review`),
		getSkillsFilePath: func() string {
			configDir, _ := os.UserConfigDir()
			return filepath.Join(configDir, "claude", "skills.json")
		},
	}
	return cmd
}

// Execute runs the skills command
func (c *SkillsCommand) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return c.listSkills()
	}

	switch args[0] {
	case "add", "create":
		if len(args) < 3 {
			return fmt.Errorf("usage: /skills add <name> <prompt>")
		}
		return c.addSkill(args[1], strings.Join(args[2:], " "))
	case "show", "view":
		if len(args) < 2 {
			return fmt.Errorf("usage: /skills show <name>")
		}
		return c.showSkill(args[1])
	case "use", "run":
		if len(args) < 2 {
			return fmt.Errorf("usage: /skills use <name>")
		}
		return c.useSkill(args[1])
	case "remove", "rm", "delete":
		if len(args) < 2 {
			return fmt.Errorf("usage: /skills remove <name>")
		}
		return c.removeSkill(args[1])
	case "edit", "update":
		if len(args) < 3 {
			return fmt.Errorf("usage: /skills edit <name> <prompt>")
		}
		return c.editSkill(args[1], strings.Join(args[2:], " "))
	default:
		return fmt.Errorf("unknown subcommand: %s", args[0])
	}
}

func (c *SkillsCommand) loadSkills() (*SkillsData, error) {
	path := c.getSkillsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &SkillsData{Skills: []Skill{}}, nil
		}
		return nil, err
	}

	var skillsData SkillsData
	if err := json.Unmarshal(data, &skillsData); err != nil {
		return nil, err
	}
	return &skillsData, nil
}

func (c *SkillsCommand) saveSkills(data *SkillsData) error {
	path := c.getSkillsFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data.UpdatedAt = time.Now()
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, out, 0644)
}

func (c *SkillsCommand) listSkills() error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════╗")
	fmt.Println("║              Reusable Skills                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════╝")
	fmt.Println()

	if len(data.Skills) == 0 {
		fmt.Println("No skills saved yet.")
		fmt.Println()
		fmt.Println("Use '/skills add <name> <prompt>' to create one.")
		fmt.Println()
		return nil
	}

	for i, skill := range data.Skills {
		fmt.Printf("%d. %s\n", i+1, skill.Name)
		if skill.Description != "" {
			fmt.Printf("   %s\n", skill.Description)
		}
		preview := skill.Prompt
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		fmt.Printf("   %s\n", preview)
	}

	fmt.Printf("\nTotal: %d skill(s)\n", len(data.Skills))
	fmt.Println()
	return nil
}

func (c *SkillsCommand) addSkill(name, prompt string) error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	for _, skill := range data.Skills {
		if skill.Name == name {
			return fmt.Errorf("skill '%s' already exists. Use 'edit' to update it", name)
		}
	}

	data.Skills = append(data.Skills, Skill{
		Name:      name,
		Prompt:    prompt,
		CreatedAt: time.Now(),
	})

	if err := c.saveSkills(data); err != nil {
		return fmt.Errorf("failed to save skills: %w", err)
	}

	fmt.Printf("✅ Skill '%s' added successfully.\n", name)
	return nil
}

func (c *SkillsCommand) showSkill(name string) error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	for _, skill := range data.Skills {
		if skill.Name == name {
			fmt.Println()
			fmt.Printf("📝 Skill: %s\n", skill.Name)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println(skill.Prompt)
			fmt.Println()
			return nil
		}
	}

	return fmt.Errorf("skill '%s' not found", name)
}

func (c *SkillsCommand) useSkill(name string) error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	for _, skill := range data.Skills {
		if skill.Name == name {
			fmt.Println()
			fmt.Printf("📝 Using skill: %s\n", skill.Name)
			fmt.Println(strings.Repeat("─", 50))
			fmt.Println(skill.Prompt)
			fmt.Println()
			return nil
		}
	}

	return fmt.Errorf("skill '%s' not found", name)
}

func (c *SkillsCommand) removeSkill(name string) error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	found := false
	newSkills := make([]Skill, 0, len(data.Skills))
	for _, skill := range data.Skills {
		if skill.Name == name {
			found = true
			continue
		}
		newSkills = append(newSkills, skill)
	}

	if !found {
		return fmt.Errorf("skill '%s' not found", name)
	}

	data.Skills = newSkills
	if err := c.saveSkills(data); err != nil {
		return fmt.Errorf("failed to save skills: %w", err)
	}

	fmt.Printf("✅ Skill '%s' removed successfully.\n", name)
	return nil
}

func (c *SkillsCommand) editSkill(name, prompt string) error {
	data, err := c.loadSkills()
	if err != nil {
		return fmt.Errorf("failed to load skills: %w", err)
	}

	found := false
	for i := range data.Skills {
		if data.Skills[i].Name == name {
			data.Skills[i].Prompt = prompt
			data.Skills[i].CreatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("skill '%s' not found", name)
	}

	if err := c.saveSkills(data); err != nil {
		return fmt.Errorf("failed to save skills: %w", err)
	}

	fmt.Printf("✅ Skill '%s' updated successfully.\n", name)
	return nil
}

func init() {
	Register(NewSkillsCommand())
}
