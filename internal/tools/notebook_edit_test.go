package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNotebookEditToolAppendCreatesNew(t *testing.T) {
	tool := &NotebookEditTool{}
	tmpPath := filepath.Join(t.TempDir(), "test.ipynb")

	input, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"new_source":    "print('hello')",
		"cell_type":     "code",
		"edit_mode":     "append",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		t.Fatalf("notebook file not created: %v", err)
	}

	var notebook NotebookContent
	if err := json.Unmarshal(data, &notebook); err != nil {
		t.Fatalf("invalid notebook json: %v", err)
	}
	if len(notebook.Cells) != 1 {
		t.Fatalf("expected 1 cell, got %d", len(notebook.Cells))
	}
	if notebook.Cells[0].CellType != "code" {
		t.Errorf("expected code cell, got %s", notebook.Cells[0].CellType)
	}
	if notebook.Nbformat != 4 {
		t.Errorf("expected nbformat 4, got %d", notebook.Nbformat)
	}
}

func TestNotebookEditToolReplaceCell(t *testing.T) {
	tool := &NotebookEditTool{}
	tmpPath := filepath.Join(t.TempDir(), "test.ipynb")

	// Create initial notebook
	notebook := NotebookContent{
		Cells: []NotebookCell{
			{ID: "cell-1", CellType: "code", Source: []string{"old code\n"}},
		},
		Metadata:      map[string]interface{}{},
		Nbformat:      4,
		NbformatMinor: 5,
	}
	data, _ := json.MarshalIndent(notebook, "", "  ")
	os.WriteFile(tmpPath, data, 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"cell_id":       "cell-1",
		"new_source":    "new code",
		"cell_type":     "markdown",
		"edit_mode":     "replace",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}

	var updated NotebookContent
	updatedData, _ := os.ReadFile(tmpPath)
	json.Unmarshal(updatedData, &updated)

	if updated.Cells[0].CellType != "markdown" {
		t.Errorf("expected markdown after replace, got %s", updated.Cells[0].CellType)
	}
	if len(updated.Cells[0].Source) == 0 || updated.Cells[0].Source[0] != "new code" {
		t.Errorf("unexpected source after replace: %+v", updated.Cells[0].Source)
	}
}

func TestNotebookEditToolInsertAndDelete(t *testing.T) {
	tool := &NotebookEditTool{}
	tmpPath := filepath.Join(t.TempDir(), "test.ipynb")

	notebook := NotebookContent{
		Cells: []NotebookCell{
			{ID: "cell-a", CellType: "code", Source: []string{"a\n"}},
		},
		Metadata:      map[string]interface{}{},
		Nbformat:      4,
		NbformatMinor: 5,
	}
	data, _ := json.MarshalIndent(notebook, "", "  ")
	os.WriteFile(tmpPath, data, 0644)

	// Insert after cell-a
	input, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"edit_mode":     "insert",
		"insert_after":  "cell-a",
		"new_source":    "b",
		"cell_type":     "code",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)
	if !parsed["success"].(bool) {
		t.Fatalf("expected insert success, got: %+v", parsed)
	}

	var updated NotebookContent
	updatedData, _ := os.ReadFile(tmpPath)
	json.Unmarshal(updatedData, &updated)
	if len(updated.Cells) != 2 {
		t.Fatalf("expected 2 cells after insert, got %d", len(updated.Cells))
	}
	if updated.Cells[0].ID != "cell-a" || updated.Cells[1].Source[0] != "b" {
		t.Errorf("unexpected cell order: %+v", updated.Cells)
	}

	// Delete cell-a
	input2, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"edit_mode":     "delete",
		"cell_id":       "cell-a",
	})
	result2, err2 := tool.Call(context.Background(), input2)
	if err2 != nil {
		t.Fatalf("unexpected error on delete: %v", err2)
	}
	json.Unmarshal(result2, &parsed)
	if !parsed["success"].(bool) {
		t.Fatalf("expected delete success, got: %+v", parsed)
	}

	updatedData2, _ := os.ReadFile(tmpPath)
	json.Unmarshal(updatedData2, &updated)
	if len(updated.Cells) != 1 {
		t.Fatalf("expected 1 cell after delete, got %d", len(updated.Cells))
	}
	if updated.Cells[0].Source[0] != "b" {
		t.Errorf("unexpected remaining cell: %+v", updated.Cells[0])
	}
}

func TestNotebookEditToolInsertAtBeginning(t *testing.T) {
	tool := &NotebookEditTool{}
	tmpPath := filepath.Join(t.TempDir(), "test.ipynb")

	notebook := NotebookContent{
		Cells: []NotebookCell{
			{ID: "cell-1", CellType: "code", Source: []string{"first\n"}},
		},
		Metadata:      map[string]interface{}{},
		Nbformat:      4,
		NbformatMinor: 5,
	}
	data, _ := json.MarshalIndent(notebook, "", "  ")
	os.WriteFile(tmpPath, data, 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"edit_mode":     "insert",
		"new_source":    "zeroth",
		"cell_type":     "markdown",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)
	if !parsed["success"].(bool) {
		t.Fatalf("expected success, got: %+v", parsed)
	}

	var updated NotebookContent
	updatedData, _ := os.ReadFile(tmpPath)
	json.Unmarshal(updatedData, &updated)
	if len(updated.Cells) != 2 || updated.Cells[0].Source[0] != "zeroth" {
		t.Errorf("unexpected cells after insert at beginning: %+v", updated.Cells)
	}
}

func TestNotebookEditToolReplaceMissingCell(t *testing.T) {
	tool := &NotebookEditTool{}
	tmpPath := filepath.Join(t.TempDir(), "test.ipynb")

	notebook := NotebookContent{
		Cells: []NotebookCell{
			{ID: "cell-1", CellType: "code", Source: []string{"a\n"}},
		},
		Metadata:      map[string]interface{}{},
		Nbformat:      4,
		NbformatMinor: 5,
	}
	data, _ := json.MarshalIndent(notebook, "", "  ")
	os.WriteFile(tmpPath, data, 0644)

	input, _ := json.Marshal(map[string]interface{}{
		"notebook_path": tmpPath,
		"cell_id":       "missing",
		"new_source":    "x",
		"edit_mode":     "replace",
	})

	result, err := tool.Call(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var parsed map[string]interface{}
	json.Unmarshal(result, &parsed)
	if parsed["success"].(bool) {
		t.Error("expected failure for missing cell")
	}
}

func TestNotebookEditToolSplitSource(t *testing.T) {
	tool := &NotebookEditTool{}
	lines := tool.splitSource("line1\nline2")
	if len(lines) != 2 || lines[0] != "line1\n" || lines[1] != "line2" {
		t.Errorf("unexpected split result: %+v", lines)
	}

	lines = tool.splitSource("")
	if len(lines) != 0 {
		t.Errorf("expected empty split for empty string, got: %+v", lines)
	}
}
