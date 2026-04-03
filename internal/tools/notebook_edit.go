// Package tools 提供 NotebookEdit 工具
// 来源: src/tools/NotebookEditTool/NotebookEditTool.ts (490行)
// 重构: Go NotebookEdit 工具（完整实现）
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// NotebookEditTool Notebook 编辑工具
type NotebookEditTool struct{}

// Name 返回工具名称
func (n *NotebookEditTool) Name() string {
	return "notebook_edit"
}

// Description 返回工具描述
func (n *NotebookEditTool) Description() string {
	return "Edit Jupyter Notebook files (.ipynb) - add, modify, or delete cells"
}

// IsReadOnly 是否只读
func (n *NotebookEditTool) IsReadOnly() bool {
	return false
}

// IsDestructive 是否有破坏性
func (n *NotebookEditTool) IsDestructive() bool {
	return true
}

// InputSchema 输入参数模式
func (n *NotebookEditTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"notebook_path": {
				"type": "string",
				"description": "Absolute path to the notebook file"
			},
			"cell_id": {
				"type": "string",
				"description": "ID of the cell to edit (required for replace/delete modes)"
			},
			"new_source": {
				"type": "string",
				"description": "New content for the cell"
			},
			"cell_type": {
				"type": "string",
				"enum": ["code", "markdown"],
				"description": "Type of cell (code or markdown)"
			},
			"edit_mode": {
				"type": "string",
				"enum": ["replace", "insert", "delete", "append"],
				"description": "Edit mode: replace existing, insert new, delete cell, or append to end"
			},
			"insert_after": {
				"type": "string",
				"description": "Cell ID to insert after (for insert mode)"
			}
		},
		"required": ["notebook_path", "new_source"]
	}`)
}

// NotebookCell Notebook 单元格 (Jupyter format)
type NotebookCell struct {
	ID             string                 `json:"id"`
	CellType       string                 `json:"cell_type"`
	Source         []string               `json:"source"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	ExecutionCount *int                   `json:"execution_count,omitempty"`
	Outputs        []interface{}          `json:"outputs,omitempty"`
}

// NotebookContent Notebook 内容 (Jupyter format)
type NotebookContent struct {
	Cells         []NotebookCell         `json:"cells"`
	Metadata      map[string]interface{} `json:"metadata"`
	Nbformat      int                    `json:"nbformat"`
	NbformatMinor int                    `json:"nbformat_minor"`
}

// Call 执行工具
func (n *NotebookEditTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		NotebookPath string `json:"notebook_path"`
		CellID       string `json:"cell_id,omitempty"`
		NewSource    string `json:"new_source"`
		CellType     string `json:"cell_type,omitempty"`
		EditMode     string `json:"edit_mode,omitempty"`
		InsertAfter  string `json:"insert_after,omitempty"`
	}

	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("failed to parse parameters: %w", err)
	}

	// Set defaults
	if params.CellType == "" {
		params.CellType = "code"
	}
	if params.EditMode == "" {
		if params.CellID != "" {
			params.EditMode = "replace"
		} else {
			params.EditMode = "append"
		}
	}

	// Check if file exists
	fileExists := true
	if _, err := os.Stat(params.NotebookPath); os.IsNotExist(err) {
		fileExists = false
	}

	var notebook NotebookContent

	if fileExists {
		// Read existing notebook
		content, err := os.ReadFile(params.NotebookPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read notebook: %w", err)
		}

		if err := json.Unmarshal(content, &notebook); err != nil {
			return nil, fmt.Errorf("failed to parse notebook: %w", err)
		}
	} else {
		// Create new notebook
		notebook = NotebookContent{
			Cells:         []NotebookCell{},
			Metadata:      map[string]interface{}{},
			Nbformat:      4,
			NbformatMinor: 5,
		}
	}

	// Perform edit operation
	var result map[string]interface{}

	switch params.EditMode {
	case "replace":
		result = n.replaceCell(&notebook, params.CellID, params.NewSource, params.CellType)
	case "insert":
		result = n.insertCell(&notebook, params.InsertAfter, params.NewSource, params.CellType)
	case "delete":
		result = n.deleteCell(&notebook, params.CellID)
	case "append":
		result = n.appendCell(&notebook, params.NewSource, params.CellType)
	default:
		return nil, fmt.Errorf("invalid edit_mode: %s", params.EditMode)
	}

	if result["success"].(bool) {
		// Write back to file
		output, err := json.MarshalIndent(notebook, "", "  ")
		if err != nil {
			return nil, fmt.Errorf("failed to serialize notebook: %w", err)
		}

		if err := os.WriteFile(params.NotebookPath, output, 0644); err != nil {
			return nil, fmt.Errorf("failed to write notebook: %w", err)
		}
	}

	return json.Marshal(result)
}

// replaceCell replaces an existing cell
func (n *NotebookEditTool) replaceCell(notebook *NotebookContent, cellID string, newSource string, cellType string) map[string]interface{} {
	if cellID == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "cell_id is required for replace mode",
		}
	}

	for i := range notebook.Cells {
		if notebook.Cells[i].ID == cellID {
			notebook.Cells[i].CellType = cellType
			notebook.Cells[i].Source = n.splitSource(newSource)
			return map[string]interface{}{
				"success":   true,
				"message":   fmt.Sprintf("Cell %s replaced successfully", cellID),
				"cell_id":   cellID,
				"cell_type": cellType,
			}
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf("Cell %s not found", cellID),
	}
}

// insertCell inserts a new cell after a specified cell
func (n *NotebookEditTool) insertCell(notebook *NotebookContent, insertAfter string, newSource string, cellType string) map[string]interface{} {
	newCell := NotebookCell{
		ID:       n.generateCellID(),
		CellType: cellType,
		Source:   n.splitSource(newSource),
		Metadata: map[string]interface{}{},
	}

	if insertAfter == "" {
		// Insert at beginning
		notebook.Cells = append([]NotebookCell{newCell}, notebook.Cells...)
		return map[string]interface{}{
			"success":   true,
			"message":   "Cell inserted at beginning",
			"cell_id":   newCell.ID,
			"cell_type": cellType,
		}
	}

	// Find position and insert after
	for i := range notebook.Cells {
		if notebook.Cells[i].ID == insertAfter {
			// Insert after position i
			notebook.Cells = append(
				notebook.Cells[:i+1],
				append([]NotebookCell{newCell}, notebook.Cells[i+1:]...)...,
			)
			return map[string]interface{}{
				"success":   true,
				"message":   fmt.Sprintf("Cell inserted after %s", insertAfter),
				"cell_id":   newCell.ID,
				"cell_type": cellType,
			}
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf("Cell %s not found", insertAfter),
	}
}

// deleteCell deletes a cell
func (n *NotebookEditTool) deleteCell(notebook *NotebookContent, cellID string) map[string]interface{} {
	if cellID == "" {
		return map[string]interface{}{
			"success": false,
			"error":   "cell_id is required for delete mode",
		}
	}

	for i := range notebook.Cells {
		if notebook.Cells[i].ID == cellID {
			// Remove cell at position i
			notebook.Cells = append(notebook.Cells[:i], notebook.Cells[i+1:]...)
			return map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("Cell %s deleted successfully", cellID),
				"cell_id": cellID,
			}
		}
	}

	return map[string]interface{}{
		"success": false,
		"error":   fmt.Sprintf("Cell %s not found", cellID),
	}
}

// appendCell appends a cell to the end
func (n *NotebookEditTool) appendCell(notebook *NotebookContent, newSource string, cellType string) map[string]interface{} {
	newCell := NotebookCell{
		ID:       n.generateCellID(),
		CellType: cellType,
		Source:   n.splitSource(newSource),
		Metadata: map[string]interface{}{},
	}

	notebook.Cells = append(notebook.Cells, newCell)

	return map[string]interface{}{
		"success":   true,
		"message":   "Cell appended successfully",
		"cell_id":   newCell.ID,
		"cell_type": cellType,
	}
}

// splitSource splits source string into lines (Jupyter format)
func (n *NotebookEditTool) splitSource(source string) []string {
	if source == "" {
		return []string{}
	}

	// Jupyter notebook stores source as array of lines
	lines := []string{}
	current := ""
	for _, char := range source {
		if char == '\n' {
			lines = append(lines, current+"\n")
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

// generateCellID generates a unique cell ID
func (n *NotebookEditTool) generateCellID() string {
	return fmt.Sprintf("cell-%d", time.Now().UnixNano())
}

// RegisterNotebookEditTool 注册 NotebookEdit 工具到注册表
func RegisterNotebookEditTool(registry *Registry) {
	registry.Register(&NotebookEditTool{})
}
