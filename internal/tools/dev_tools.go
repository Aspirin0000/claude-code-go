package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Aspirin0000/claude-code-go/internal/state"
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

// DockerExecTool Docker exec tool
type DockerExecTool struct{}

func (d *DockerExecTool) Name() string { return "docker_exec" }
func (d *DockerExecTool) Description() string {
	return "Execute a command inside a running Docker container"
}
func (d *DockerExecTool) IsReadOnly() bool    { return false }
func (d *DockerExecTool) IsDestructive() bool { return false }

func (d *DockerExecTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"container": {"type": "string", "description": "Container ID or name"},
 			"command": {"type": "string", "description": "Command to execute inside the container"},
 			"workdir": {"type": "string", "description": "Working directory inside the container"}
 		},
 		"required": ["container", "command"]
 	}`)
}

func (d *DockerExecTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Container string `json:"container"`
		Command   string `json:"command"`
		Workdir   string `json:"workdir"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	args := []string{"exec"}
	if params.Workdir != "" {
		args = append(args, "-w", params.Workdir)
	}
	args = append(args, params.Container, "sh", "-c", params.Command)

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

// DockerBuildTool Docker build tool
type DockerBuildTool struct{}

func (d *DockerBuildTool) Name() string        { return "docker_build" }
func (d *DockerBuildTool) Description() string { return "Build a Docker image from a Dockerfile" }
func (d *DockerBuildTool) IsReadOnly() bool    { return false }
func (d *DockerBuildTool) IsDestructive() bool { return false }

func (d *DockerBuildTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"path": {"type": "string", "description": "Build context path (default: current directory)"},
 			"tag": {"type": "string", "description": "Image tag to apply"},
 			"dockerfile": {"type": "string", "description": "Path to Dockerfile (default: PATH/Dockerfile)"}
 		}
 	}`)
}

func (d *DockerBuildTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path       string `json:"path"`
		Tag        string `json:"tag"`
		Dockerfile string `json:"dockerfile"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	workDir := params.Path
	if workDir == "" {
		workDir = "."
	}

	args := []string{"build", workDir}
	if params.Tag != "" {
		args = append(args, "-t", params.Tag)
	}
	if params.Dockerfile != "" {
		args = append(args, "-f", params.Dockerfile)
	}

	cmd := exec.CommandContext(ctx, "docker", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// SedReplaceTool sed-like replacement tool
type SedReplaceTool struct{}

func (s *SedReplaceTool) Name() string { return "sed_replace" }
func (s *SedReplaceTool) Description() string {
	return "Replace text in a file using regex (sed-style)"
}
func (s *SedReplaceTool) IsReadOnly() bool    { return false }
func (s *SedReplaceTool) IsDestructive() bool { return true }

func (s *SedReplaceTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"file_path": {"type": "string", "description": "Path to the file"},
 			"pattern": {"type": "string", "description": "Regular expression pattern to match"},
 			"replacement": {"type": "string", "description": "Replacement string"},
 			"all": {"type": "boolean", "description": "Replace all occurrences (default: first only)"}
 		},
 		"required": ["file_path", "pattern", "replacement"]
 	}`)
}

func (s *SedReplaceTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		FilePath    string `json:"file_path"`
		Pattern     string `json:"pattern"`
		Replacement string `json:"replacement"`
		All         bool   `json:"all"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	content, err := os.ReadFile(params.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	re, err := regexp.Compile(params.Pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var newContent string
	if params.All {
		newContent = re.ReplaceAllString(string(content), params.Replacement)
	} else {
		newContent = re.ReplaceAllStringFunc(string(content), func(s string) string {
			// Only replace first match by tracking state
			// Actually ReplaceAllStringFunc replaces all matches.
			// For single replacement, use ReplaceAllString with limit=1 workaround.
			return s
		})
		// Correct single replacement approach
		loc := re.FindStringIndex(string(content))
		if loc == nil {
			return nil, fmt.Errorf("pattern not found in file")
		}
		newContent = string(content)[:loc[0]] + re.ReplaceAllString(string(content)[loc[0]:loc[1]], params.Replacement) + string(content)[loc[1]:]
	}

	if !params.All {
		// Check if pattern was found
		if re.FindStringIndex(string(content)) == nil {
			return nil, fmt.Errorf("pattern not found in file")
		}
	}

	if err := os.WriteFile(params.FilePath, []byte(newContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	state.GlobalState.AddEdit(state.Edit{
		Tool:        "sed_replace",
		FilePath:    params.FilePath,
		Operation:   "edit",
		Description: "Edited file (sed replacement)",
	})

	return json.Marshal(struct {
		Success bool `json:"success"`
	}{
		Success: true,
	})
}

// JSONQueryTool JSON query tool
type JSONQueryTool struct{}

func (j *JSONQueryTool) Name() string        { return "json_query" }
func (j *JSONQueryTool) Description() string { return "Query JSON data using a dot-notation path" }
func (j *JSONQueryTool) IsReadOnly() bool    { return true }
func (j *JSONQueryTool) IsDestructive() bool { return false }

func (j *JSONQueryTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"json": {"type": "string", "description": "JSON string to query"},
 			"path": {"type": "string", "description": "Dot-notation path (e.g., user.name or items.0.id)"}
 		},
 		"required": ["json", "path"]
 	}`)
}

func (j *JSONQueryTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		JSON string `json:"json"`
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	var data interface{}
	if err := json.Unmarshal([]byte(params.JSON), &data); err != nil {
		return nil, fmt.Errorf("invalid json: %w", err)
	}

	parts := strings.Split(params.Path, ".")
	current := data
	for _, part := range parts {
		if part == "" {
			continue
		}
		switch v := current.(type) {
		case map[string]interface{}:
			if next, ok := v[part]; ok {
				current = next
			} else {
				return nil, fmt.Errorf("path not found: %s", params.Path)
			}
		case []interface{}:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, fmt.Errorf("invalid array index in path: %s", part)
			}
			current = v[idx]
		default:
			return nil, fmt.Errorf("cannot traverse path: %s", params.Path)
		}
	}

	result := struct {
		Value interface{} `json:"value"`
		Found bool        `json:"found"`
	}{
		Value: current,
		Found: true,
	}
	return json.Marshal(result)
}

// EnvGetTool Environment variable get tool
type EnvGetTool struct{}

func (e *EnvGetTool) Name() string        { return "env_get" }
func (e *EnvGetTool) Description() string { return "Get environment variables" }
func (e *EnvGetTool) IsReadOnly() bool    { return true }
func (e *EnvGetTool) IsDestructive() bool { return false }

func (e *EnvGetTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"names": {"type": "array", "items": {"type": "string"}, "description": "Variable names to get (default: all)"}
 		}
 	}`)
}

func (e *EnvGetTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Names []string `json:"names"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	env := make(map[string]string)
	if len(params.Names) == 0 {
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				env[e[:i]] = e[i+1:]
			}
		}
	} else {
		for _, name := range params.Names {
			env[name] = os.Getenv(name)
		}
	}

	return json.Marshal(struct {
		Env map[string]string `json:"env"`
	}{
		Env: env,
	})
}

// EnvSetTool Environment variable set tool
type EnvSetTool struct{}

func (e *EnvSetTool) Name() string        { return "env_set" }
func (e *EnvSetTool) Description() string { return "Set environment variables for the current process" }
func (e *EnvSetTool) IsReadOnly() bool    { return false }
func (e *EnvSetTool) IsDestructive() bool { return false }

func (e *EnvSetTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"vars": {"type": "object", "description": "Key-value pairs to set"}
 		},
 		"required": ["vars"]
 	}`)
}

func (e *EnvSetTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Vars map[string]string `json:"vars"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	for k, v := range params.Vars {
		os.Setenv(k, v)
	}

	return json.Marshal(struct {
		Success bool `json:"success"`
		Count   int  `json:"count"`
	}{
		Success: true,
		Count:   len(params.Vars),
	})
}

// FileInfoTool File info tool
type FileInfoTool struct{}

func (f *FileInfoTool) Name() string { return "file_info" }
func (f *FileInfoTool) Description() string {
	return "Get detailed information about a file or directory"
}
func (f *FileInfoTool) IsReadOnly() bool    { return true }
func (f *FileInfoTool) IsDestructive() bool { return false }

func (f *FileInfoTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
 		"type": "object",
 		"properties": {
 			"path": {"type": "string", "description": "File or directory path"}
 		},
 		"required": ["path"]
 	}`)
}

func (f *FileInfoTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	info, err := os.Stat(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat path: %w", err)
	}

	return json.Marshal(struct {
		Path    string `json:"path"`
		Size    int64  `json:"size"`
		IsDir   bool   `json:"is_dir"`
		Mode    string `json:"mode"`
		ModTime string `json:"mod_time"`
	}{
		Path:    params.Path,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().Format(time.RFC3339),
	})
}
