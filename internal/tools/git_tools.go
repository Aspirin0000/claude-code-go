// Package tools provides concrete tool implementations
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// GitStatusTool Git status tool
type GitStatusTool struct{}

func (g *GitStatusTool) Name() string        { return "git_status" }
func (g *GitStatusTool) Description() string { return "Check git repository status" }
func (g *GitStatusTool) IsReadOnly() bool    { return true }
func (g *GitStatusTool) IsDestructive() bool { return false }

func (g *GitStatusTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"}
		}
	}`)
}

func (g *GitStatusTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "status", "--porcelain", "-b")
	output, err := cmd.CombinedOutput()

	result := struct {
		Status string `json:"status"`
		Branch string `json:"branch"`
		Path   string `json:"path"`
		Error  string `json:"error,omitempty"`
	}{
		Status: string(output),
		Path:   path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	// Parse branch from first line
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			result.Branch = strings.TrimPrefix(line, "## ")
			break
		}
	}

	return json.Marshal(result)
}

// GitDiffTool Git diff tool
type GitDiffTool struct{}

func (g *GitDiffTool) Name() string        { return "git_diff" }
func (g *GitDiffTool) Description() string { return "Show git diff for the repository" }
func (g *GitDiffTool) IsReadOnly() bool    { return true }
func (g *GitDiffTool) IsDestructive() bool { return false }

func (g *GitDiffTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"staged": {"type": "boolean", "description": "Show staged changes"}
		}
	}`)
}

func (g *GitDiffTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Staged bool   `json:"staged"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "diff"}
	if params.Staged {
		args = append(args, "--staged")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Diff  string `json:"diff"`
		Path  string `json:"path"`
		Error string `json:"error,omitempty"`
	}{
		Diff: string(output),
		Path: path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitLogTool Git log tool
type GitLogTool struct{}

func (g *GitLogTool) Name() string        { return "git_log" }
func (g *GitLogTool) Description() string { return "Show recent git commit history" }
func (g *GitLogTool) IsReadOnly() bool    { return true }
func (g *GitLogTool) IsDestructive() bool { return false }

func (g *GitLogTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"count": {"type": "number", "description": "Number of commits to show (default: 10)"}
		}
	}`)
}

func (g *GitLogTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path  string `json:"path"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}
	count := params.Count
	if count <= 0 {
		count = 10
	}

	cmd := exec.CommandContext(ctx, "git", "-C", path, "log", "--oneline", "-n", fmt.Sprintf("%d", count))
	output, err := cmd.CombinedOutput()

	result := struct {
		Log   string `json:"log"`
		Path  string `json:"path"`
		Count int    `json:"count"`
		Error string `json:"error,omitempty"`
	}{
		Log:   string(output),
		Path:  path,
		Count: count,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitCommitTool Git commit tool
type GitCommitTool struct{}

func (g *GitCommitTool) Name() string        { return "git_commit" }
func (g *GitCommitTool) Description() string { return "Create a git commit with a message" }
func (g *GitCommitTool) IsReadOnly() bool    { return false }
func (g *GitCommitTool) IsDestructive() bool { return false }

func (g *GitCommitTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"message": {"type": "string", "description": "Commit message"},
			"all": {"type": "boolean", "description": "Stage all modified/deleted files before commit"}
		},
		"required": ["message"]
	}`)
}

func (g *GitCommitTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path    string `json:"path"`
		Message string `json:"message"`
		All     bool   `json:"all"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "commit", "-m", params.Message}
	if params.All {
		args = append(args, "-a")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitBranchTool Git branch tool
type GitBranchTool struct{}

func (g *GitBranchTool) Name() string        { return "git_branch" }
func (g *GitBranchTool) Description() string { return "List, create, or delete git branches" }
func (g *GitBranchTool) IsReadOnly() bool    { return true }
func (g *GitBranchTool) IsDestructive() bool { return false }

func (g *GitBranchTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"create": {"type": "string", "description": "Create a new branch with this name"},
			"delete": {"type": "string", "description": "Delete a branch with this name"},
			"force_delete": {"type": "boolean", "description": "Force delete the branch"}
		}
	}`)
}

func (g *GitBranchTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path        string `json:"path"`
		Create      string `json:"create"`
		Delete      string `json:"delete"`
		ForceDelete bool   `json:"force_delete"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "branch"}
	if params.Create != "" {
		args = append(args, params.Create)
	} else if params.Delete != "" {
		if params.ForceDelete {
			args = append(args, "-D", params.Delete)
		} else {
			args = append(args, "-d", params.Delete)
		}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitCheckoutTool Git checkout tool
type GitCheckoutTool struct{}

func (g *GitCheckoutTool) Name() string        { return "git_checkout" }
func (g *GitCheckoutTool) Description() string { return "Checkout a git branch or create a new one" }
func (g *GitCheckoutTool) IsReadOnly() bool    { return false }
func (g *GitCheckoutTool) IsDestructive() bool { return false }

func (g *GitCheckoutTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"branch": {"type": "string", "description": "Branch name to checkout"},
			"create": {"type": "boolean", "description": "Create the branch if it doesn't exist (-b)"}
		},
		"required": ["branch"]
	}`)
}

func (g *GitCheckoutTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Branch string `json:"branch"`
		Create bool   `json:"create"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	path := params.Path
	if path == "" {
		path = "."
	}

	args := []string{"-C", path, "checkout"}
	if params.Create {
		args = append(args, "-b", params.Branch)
	} else {
		args = append(args, params.Branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    path,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitAddTool Git add tool
type GitAddTool struct{}

func (g *GitAddTool) Name() string        { return "git_add" }
func (g *GitAddTool) Description() string { return "Stage files for git commit" }
func (g *GitAddTool) IsReadOnly() bool    { return false }
func (g *GitAddTool) IsDestructive() bool { return false }

func (g *GitAddTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"files": {"type": "array", "items": {"type": "string"}, "description": "Files to stage (default: [\".\"])"}
		}
	}`)
}

func (g *GitAddTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path  string   `json:"path"`
		Files []string `json:"files"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}
	files := params.Files
	if len(files) == 0 {
		files = []string{"."}
	}

	args := append([]string{"-C", repoPath, "add"}, files...)
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}

	if err != nil {
		result.Error = err.Error()
	}

	return json.Marshal(result)
}

// GitPushTool Git push tool
type GitPushTool struct{}

func (g *GitPushTool) Name() string        { return "git_push" }
func (g *GitPushTool) Description() string { return "Push local commits to a remote repository" }
func (g *GitPushTool) IsReadOnly() bool    { return false }
func (g *GitPushTool) IsDestructive() bool { return false }

func (g *GitPushTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"remote": {"type": "string", "description": "Remote name (default: origin)"},
			"branch": {"type": "string", "description": "Branch name (default: current branch)"},
			"force": {"type": "boolean", "description": "Force push"}
		}
	}`)
}

func (g *GitPushTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Remote string `json:"remote"`
		Branch string `json:"branch"`
		Force  bool   `json:"force"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}
	remote := params.Remote
	if remote == "" {
		remote = "origin"
	}

	args := []string{"-C", repoPath, "push", remote}
	if params.Force {
		args = append(args, "--force")
	}
	if params.Branch != "" {
		args = append(args, params.Branch)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitPullTool Git pull tool
type GitPullTool struct{}

func (g *GitPullTool) Name() string        { return "git_pull" }
func (g *GitPullTool) Description() string { return "Pull changes from a remote repository" }
func (g *GitPullTool) IsReadOnly() bool    { return false }
func (g *GitPullTool) IsDestructive() bool { return false }

func (g *GitPullTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"remote": {"type": "string", "description": "Remote name (default: origin)"},
			"branch": {"type": "string", "description": "Branch name (default: current branch)"},
			"rebase": {"type": "boolean", "description": "Use rebase instead of merge"}
		}
	}`)
}

func (g *GitPullTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Remote string `json:"remote"`
		Branch string `json:"branch"`
		Rebase bool   `json:"rebase"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}

	args := []string{"-C", repoPath, "pull"}
	if params.Rebase {
		args = append(args, "--rebase")
	}
	if params.Remote != "" {
		args = append(args, params.Remote)
		if params.Branch != "" {
			args = append(args, params.Branch)
		}
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitResetTool Git reset tool
type GitResetTool struct{}

func (g *GitResetTool) Name() string        { return "git_reset" }
func (g *GitResetTool) Description() string { return "Reset current HEAD to a specified state" }
func (g *GitResetTool) IsReadOnly() bool    { return false }
func (g *GitResetTool) IsDestructive() bool { return true }

func (g *GitResetTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"mode": {"type": "string", "enum": ["soft", "mixed", "hard"], "description": "Reset mode"},
			"commit": {"type": "string", "description": "Commit to reset to (default: HEAD)"}
		},
		"required": ["mode"]
	}`)
}

func (g *GitResetTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Mode   string `json:"mode"`
		Commit string `json:"commit"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}
	commit := params.Commit
	if commit == "" {
		commit = "HEAD"
	}

	args := []string{"-C", repoPath, "reset", "--" + params.Mode, commit}
	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitStashTool Git stash tool
type GitStashTool struct{}

func (g *GitStashTool) Name() string        { return "git_stash" }
func (g *GitStashTool) Description() string { return "Stash or unstash changes in a git repository" }
func (g *GitStashTool) IsReadOnly() bool    { return false }
func (g *GitStashTool) IsDestructive() bool { return false }

func (g *GitStashTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"action": {"type": "string", "enum": ["push", "pop", "list", "drop"], "description": "Stash action"},
			"message": {"type": "string", "description": "Stash message (for push)"}
		},
		"required": ["action"]
	}`)
}

func (g *GitStashTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path    string `json:"path"`
		Action  string `json:"action"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}

	args := []string{"-C", repoPath, "stash", params.Action}
	if params.Action == "push" && params.Message != "" {
		args = append(args, "-m", params.Message)
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitRemoteTool Git remote tool
type GitRemoteTool struct{}

func (g *GitRemoteTool) Name() string        { return "git_remote" }
func (g *GitRemoteTool) Description() string { return "List, add, or remove git remotes" }
func (g *GitRemoteTool) IsReadOnly() bool    { return true }
func (g *GitRemoteTool) IsDestructive() bool { return false }

func (g *GitRemoteTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"action": {"type": "string", "enum": ["list", "add", "remove"], "description": "Remote action"},
			"name": {"type": "string", "description": "Remote name"},
			"url": {"type": "string", "description": "Remote URL (for add)"}
		},
		"required": ["action"]
	}`)
}

func (g *GitRemoteTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Action string `json:"action"`
		Name   string `json:"name"`
		URL    string `json:"url"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}

	args := []string{"-C", repoPath, "remote"}
	if params.Action == "add" {
		args = append(args, "add", params.Name, params.URL)
	} else if params.Action == "remove" {
		args = append(args, "remove", params.Name)
	} else {
		args = append(args, "-v")
	}

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitMergeTool Git merge tool
type GitMergeTool struct{}

func (g *GitMergeTool) Name() string        { return "git_merge" }
func (g *GitMergeTool) Description() string { return "Merge a branch into the current branch" }
func (g *GitMergeTool) IsReadOnly() bool    { return false }
func (g *GitMergeTool) IsDestructive() bool { return false }

func (g *GitMergeTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"branch": {"type": "string", "description": "Branch to merge"},
			"message": {"type": "string", "description": "Merge commit message"},
			"squash": {"type": "boolean", "description": "Squash merge"}
		},
		"required": ["branch"]
	}`)
}

func (g *GitMergeTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path    string `json:"path"`
		Branch  string `json:"branch"`
		Message string `json:"message"`
		Squash  bool   `json:"squash"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}

	args := []string{"-C", repoPath, "merge"}
	if params.Squash {
		args = append(args, "--squash")
	}
	if params.Message != "" {
		args = append(args, "-m", params.Message)
	}
	args = append(args, params.Branch)

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Success bool   `json:"success"`
		Output  string `json:"output"`
		Path    string `json:"path"`
		Error   string `json:"error,omitempty"`
	}{
		Success: err == nil,
		Output:  string(output),
		Path:    repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}

// GitShowTool Git show tool
type GitShowTool struct{}

func (g *GitShowTool) Name() string        { return "git_show" }
func (g *GitShowTool) Description() string { return "Show details of a git commit" }
func (g *GitShowTool) IsReadOnly() bool    { return true }
func (g *GitShowTool) IsDestructive() bool { return false }

func (g *GitShowTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"path": {"type": "string", "description": "Repository path (default: current directory)"},
			"commit": {"type": "string", "description": "Commit hash or reference (default: HEAD)"},
			"stat": {"type": "boolean", "description": "Show diffstat"}
		}
	}`)
}

func (g *GitShowTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Path   string `json:"path"`
		Commit string `json:"commit"`
		Stat   bool   `json:"stat"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	repoPath := params.Path
	if repoPath == "" {
		repoPath = "."
	}
	commit := params.Commit
	if commit == "" {
		commit = "HEAD"
	}

	args := []string{"-C", repoPath, "show"}
	if params.Stat {
		args = append(args, "--stat")
	}
	args = append(args, commit)

	cmd := exec.CommandContext(ctx, "git", args...)
	output, err := cmd.CombinedOutput()

	result := struct {
		Output string `json:"output"`
		Path   string `json:"path"`
		Error  string `json:"error,omitempty"`
	}{
		Output: string(output),
		Path:   repoPath,
	}
	if err != nil {
		result.Error = err.Error()
	}
	return json.Marshal(result)
}
