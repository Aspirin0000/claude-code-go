package commands

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/config"
)

// DoctorCommand diagnoses system and configuration issues
// Source: src/commands/doctor/
type DoctorCommand struct {
	*BaseCommand
}

// NewDoctorCommand creates the doctor command
func NewDoctorCommand() *DoctorCommand {
	return &DoctorCommand{
		BaseCommand: NewBaseCommand(
			"doctor",
			"Diagnose system and configuration issues",
			CategoryAdvanced,
		).
			WithHelp(`Run diagnostic checks to identify common issues.

Checks include:
- Go version
- Git installation
- API key configuration
- Environment variables
- Configuration files
- Network connectivity

Usage: /doctor`),
	}
}

// Execute runs diagnostics
func (c *DoctorCommand) Execute(ctx context.Context, args []string) error {
	fmt.Println()
	fmt.Println("🔍 Running Diagnostics")
	fmt.Println("═══════════════════════════════════════")

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{"Go Version", c.checkGoVersion},
		{"Git Installation", c.checkGit},
		{"Docker", c.checkDocker},
		{"Python 3", c.checkPython},
		{"Node.js", c.checkNode},
		{"NPM", c.checkNPM},
		{"Ripgrep", c.checkRipgrep},
		{"API Key", c.checkAPIKey},
		{"Anthropic API", c.checkAnthropicAPI},
		{"Environment Variables", c.checkEnv},
		{"Config File", c.checkConfig},
		{"Network Connection", c.checkNetwork},
	}

	allPassed := true
	for _, check := range checks {
		passed, msg := check.fn()
		status := "✅"
		if !passed {
			status = "❌"
			allPassed = false
		}
		fmt.Printf("%s %-20s %s\n", status, check.name+":", msg)
	}

	fmt.Println()
	if allPassed {
		fmt.Println("✓ All checks passed!")
	} else {
		fmt.Println("⚠️  Some issues found. Please check the output above.")
	}
	fmt.Println()

	return nil
}

func (c *DoctorCommand) checkGoVersion() (bool, string) {
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkGit() (bool, string) {
	cmd := exec.Command("git", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkDocker() (bool, string) {
	cmd := exec.Command("docker", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkPython() (bool, string) {
	cmd := exec.Command("python3", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkNode() (bool, string) {
	cmd := exec.Command("node", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkNPM() (bool, string) {
	cmd := exec.Command("npm", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	return true, strings.TrimSpace(string(out))
}

func (c *DoctorCommand) checkRipgrep() (bool, string) {
	cmd := exec.Command("rg", "--version")
	out, err := cmd.Output()
	if err != nil {
		return false, "not installed"
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return true, lines[0]
	}
	return true, "installed"
}

func (c *DoctorCommand) checkAPIKey() (bool, string) {
	// Check environment variable
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return true, "ANTHROPIC_API_KEY is set"
	}
	if os.Getenv("ANTHROPIC_AUTH_TOKEN") != "" {
		return true, "ANTHROPIC_AUTH_TOKEN is set"
	}

	// Check config file
	cfgPath := config.GetConfigPath()
	cfg, err := config.Load(cfgPath)
	if err == nil && cfg.APIKey != "" {
		return true, "API key found in config file"
	}

	return false, "No API key found (set ANTHROPIC_API_KEY or run /init)"
}

func (c *DoctorCommand) checkEnv() (bool, string) {
	required := []string{"HOME", "PATH"}
	missing := []string{}
	for _, env := range required {
		if os.Getenv(env) == "" {
			missing = append(missing, env)
		}
	}
	if len(missing) > 0 {
		return false, fmt.Sprintf("Missing: %s", strings.Join(missing, ", "))
	}
	return true, "All required variables set"
}

func (c *DoctorCommand) checkConfig() (bool, string) {
	configDir := getConfigDir()
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return true, "Config directory does not exist (will be created automatically)"
	}
	return true, "Config directory exists"
}

func (c *DoctorCommand) checkNetwork() (bool, string) {
	// Simple check: can we resolve a domain?
	cmd := exec.Command("ping", "-c", "1", "-W", "2", "github.com")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("ping", "-n", "1", "-w", "2000", "github.com")
	}
	err := cmd.Run()
	if err != nil {
		return false, "Unable to connect to network"
	}
	return true, "Network connection OK"
}

func (c *DoctorCommand) checkAnthropicAPI() (bool, string) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		cfgPath := config.GetConfigPath()
		cfg, err := config.Load(cfgPath)
		if err == nil {
			apiKey = cfg.APIKey
		}
	}

	if apiKey == "" {
		return false, "No API key available"
	}

	// Try a simple HEAD request to Anthropic API
	cmd := exec.Command("curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "-H", "x-api-key: "+apiKey, "-H", "anthropic-version: 2023-06-01", "https://api.anthropic.com/v1/models")
	out, err := cmd.Output()
	if err != nil {
		return false, "Unable to reach Anthropic API"
	}

	code := strings.TrimSpace(string(out))
	if code == "200" || code == "401" {
		// 401 means key is valid format but may lack permissions; still means API is reachable
		return true, "Anthropic API is reachable"
	}
	return false, fmt.Sprintf("Anthropic API returned HTTP %s", code)
}

func getConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "claude")
}

func init() {
	Register(NewDoctorCommand())
}
