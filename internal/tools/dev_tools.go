package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

// DockerPsTool Docker ps tool
type DockerPsTool struct{}

func (d *DockerPsTool) Name() string        { return "docker_ps" }
func (d *DockerPsTool) Description() string { return "List running Docker containers" }
func (d *DockerPsTool) IsReadOnly() bool    { return true }
func (d *DockerPsTool) IsDestructive() bool { return false }

func (d *DockerPsTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"all": {"type": "boolean", "description": "Show all containers (including stopped)"}
		}
	}`)
}

func (d *DockerPsTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		All bool `json:"all"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	args := []string{"ps"}
	if params.All {
		args = append(args, "-a")
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Output string `json:"output"`
		Error  string `json:"error,omitempty"`
	}{
		Output: string(output),
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// DockerLogsTool Docker logs tool
type DockerLogsTool struct{}

func (d *DockerLogsTool) Name() string        { return "docker_logs" }
func (d *DockerLogsTool) Description() string { return "Fetch logs from a Docker container" }
func (d *DockerLogsTool) IsReadOnly() bool    { return true }
func (d *DockerLogsTool) IsDestructive() bool { return false }

func (d *DockerLogsTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"container": {"type": "string", "description": "Container ID or name"},
			"tail": {"type": "number", "description": "Number of lines to show from the end"}
		},
		"required": ["container"]
	}`)
}

func (d *DockerLogsTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Container string `json:"container"`
		Tail      int    `json:"tail"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	args := []string{"logs"}
	if params.Tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", params.Tail))
	}
	args = append(args, params.Container)

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Output string `json:"output"`
		Error  string `json:"error,omitempty"`
	}{
		Output: string(output),
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// NpmInstallTool npm install tool
type NpmInstallTool struct{}

func (n *NpmInstallTool) Name() string        { return "npm_install" }
func (n *NpmInstallTool) Description() string { return "Install npm packages" }
func (n *NpmInstallTool) IsReadOnly() bool    { return false }
func (n *NpmInstallTool) IsDestructive() bool { return false }

func (n *NpmInstallTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Working directory (default: current directory)"},
			"packages": {"type": "array", "items": {"type": "string"}, "description": "Packages to install (default: all dependencies)"},
			"dev": {"type": "boolean", "description": "Install as dev dependencies"}
		}
	}`)
}

func (n *NpmInstallTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path     string   `json:"path"`
		Packages []string `json:"packages"`
		Dev      bool     `json:"dev"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	workDir := params.Path
	if workDir == "" {
		workDir = "."
	}

	args := []string{"install"}
	if params.Dev {
		args = append(args, "--save-dev")
	}
	args = append(args, params.Packages...)

	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    workDir,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// NpmRunTool npm run tool
type NpmRunTool struct{}

func (n *NpmRunTool) Name() string        { return "npm_run" }
func (n *NpmRunTool) Description() string { return "Run an npm script" }
func (n *NpmRunTool) IsReadOnly() bool    { return true }
func (n *NpmRunTool) IsDestructive() bool { return false }

func (n *NpmRunTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Working directory (default: current directory)"},
			"script": {"type": "string", "description": "Script name to run"},
			"args": {"type": "array", "items": {"type": "string"}, "description": "Additional arguments to pass to the script"}
		},
		"required": ["script"]
	}`)
}

func (n *NpmRunTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string   `json:"path"`
		Script string   `json:"script"`
		Args   []string `json:"args"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	workDir := params.Path
	if workDir == "" {
		workDir = "."
	}

	args := append([]string{"run", params.Script}, params.Args...)
	cmd := exec.CommandContext(ctx, "npm", args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := struct {
		Output string `json:"output"`
		Path   string `json:"path"`
		Error  string `json:"error,omitempty"`
	}{
		Output: string(output),
		Path:   workDir,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GoBuildTool Go build tool
type GoBuildTool struct{}

func (g *GoBuildTool) Name() string        { return "go_build" }
func (g *GoBuildTool) Description() string { return "Build a Go project" }
func (g *GoBuildTool) IsReadOnly() bool    { return true }
func (g *GoBuildTool) IsDestructive() bool { return false }

func (g *GoBuildTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Working directory (default: current directory)"},
			"output": {"type": "string", "description": "Output binary name"}
		}
	}`)
}

func (g *GoBuildTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Output string `json:"output"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	workDir := params.Path
	if workDir == "" {
		workDir = "."
	}

	args := []string{"build"}
	if params.Output != "" {
		args = append(args, "-o", params.Output)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    workDir,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GoTestTool Go test tool
type GoTestTool struct{}

func (g *GoTestTool) Name() string        { return "go_test" }
func (g *GoTestTool) Description() string { return "Run Go tests" }
func (g *GoTestTool) IsReadOnly() bool    { return true }
func (g *GoTestTool) IsDestructive() bool { return false }

func (g *GoTestTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Working directory or package path (default: current directory)"},
			"verbose": {"type": "boolean", "description": "Enable verbose output"},
			"run": {"type": "string", "description": "Run only tests matching this pattern"}
		}
	}`)
}

func (g *GoTestTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path    string `json:"path"`
		Verbose bool   `json:"verbose"`
		Run     string `json:"run"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	workDir := params.Path
	if workDir == "" {
		workDir = "."
	}

	args := []string{"test"}
	if params.Verbose {
		args = append(args, "-v")
	}
	if params.Run != "" {
		args = append(args, "-run", params.Run)
	}

	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    workDir,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// PythonRunTool Python run tool
type PythonRunTool struct{}

func (p *PythonRunTool) Name() string        { return "python_run" }
func (p *PythonRunTool) Description() string { return "Run a Python script or command" }
func (p *PythonRunTool) IsReadOnly() bool    { return true }
func (p *PythonRunTool) IsDestructive() bool { return false }

func (p *PythonRunTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"code": {"type": "string", "description": "Python code to execute"},
			"file": {"type": "string", "description": "Python file to run"},
			"args": {"type": "array", "items": {"type": "string"}, "description": "Arguments to pass when running a file"}
		}
	}`)
}

func (p *PythonRunTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Code string   `json:"code"`
		File string   `json:"file"`
		Args []string `json:"args"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	var cmd *exec.Cmd
	if params.File != "" {
		args := append([]string{params.File}, params.Args...)
		cmd = exec.CommandContext(ctx, "python3", args...)
	} else if params.Code != "" {
		cmd = exec.CommandContext(ctx, "python3", "-c", params.Code)
	} else {
		return nil, fmt.Errorf("either 'code' or 'file' must be provided")
	}

	output, err := cmd.CombinedOutput()

	result := struct {
		Output string `json:"output"`
		Error  string `json:"error,omitempty"`
	}{
		Output: string(output),
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}
