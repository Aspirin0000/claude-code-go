// Package tools LSP tool implementation
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Aspirin0000/claude-code-go/internal/lsp"
)

// LSPTool provides language-server-protocol operations.
type LSPTool struct{}

func (l *LSPTool) Name() string { return "lsp" }
func (l *LSPTool) Description() string {
	return "Query language servers for definitions, references, hover, and symbols"
}
func (l *LSPTool) IsReadOnly() bool    { return true }
func (l *LSPTool) IsDestructive() bool { return false }

func (l *LSPTool) InputSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"operation": {
				"type": "string",
				"enum": ["goToDefinition", "findReferences", "hover", "documentSymbol", "workspaceSymbol"],
				"description": "LSP operation to perform"
			},
			"filePath": {"type": "string", "description": "Target file path"},
			"line": {"type": "number", "description": "1-based line number"},
			"character": {"type": "number", "description": "1-based character position"},
			"query": {"type": "string", "description": "Query string for workspaceSymbol"}
		},
		"required": ["operation", "filePath"]
	}`)
}

func (l *LSPTool) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var params struct {
		Operation string `json:"operation"`
		FilePath  string `json:"filePath"`
		Line      int    `json:"line"`
		Character int    `json:"character"`
		Query     string `json:"query"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, err
	}

	mgr := lsp.GetGlobalManager()

	// Ensure server is started
	server, err := mgr.EnsureStarted(ctx, params.FilePath)
	if err != nil {
		return nil, fmt.Errorf("no LSP server available for %s: %w", params.FilePath, err)
	}

	// Open file if not already open
	content, _ := os.ReadFile(params.FilePath)
	_ = mgr.OpenFile(params.FilePath, 1, string(content))

	uri := lsp.PathToURI(params.FilePath)
	line := params.Line
	char := params.Character
	if line > 0 {
		line-- // convert to 0-based
	}
	if char > 0 {
		char-- // convert to 0-based
	}

	var result json.RawMessage
	var method string

	switch params.Operation {
	case "goToDefinition":
		method = "textDocument/definition"
		result, err = server.Request(ctx, method, map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": uri},
			"position":     map[string]interface{}{"line": line, "character": char},
		})
	case "findReferences":
		method = "textDocument/references"
		result, err = server.Request(ctx, method, map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": uri},
			"position":     map[string]interface{}{"line": line, "character": char},
			"context":      map[string]interface{}{"includeDeclaration": true},
		})
	case "hover":
		method = "textDocument/hover"
		result, err = server.Request(ctx, method, map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": uri},
			"position":     map[string]interface{}{"line": line, "character": char},
		})
	case "documentSymbol":
		method = "textDocument/documentSymbol"
		result, err = server.Request(ctx, method, map[string]interface{}{
			"textDocument": map[string]interface{}{"uri": uri},
		})
	case "workspaceSymbol":
		method = "workspace/symbol"
		query := params.Query
		if query == "" {
			query = "*"
		}
		result, err = server.Request(ctx, method, map[string]interface{}{"query": query})
	default:
		return nil, fmt.Errorf("unsupported LSP operation: %s", params.Operation)
	}

	if err != nil {
		return nil, fmt.Errorf("LSP request failed: %w", err)
	}

	formatted := formatLSPResult(params.Operation, result, uri)

	return json.Marshal(map[string]interface{}{
		"operation":   params.Operation,
		"result":      formatted,
		"filePath":    params.FilePath,
		"resultCount": countResults(result),
	})
}

// formatLSPResult converts raw LSP result into human-readable text.
func formatLSPResult(operation string, raw json.RawMessage, baseURI string) string {
	switch operation {
	case "goToDefinition", "findReferences":
		var locations []Location
		if err := json.Unmarshal(raw, &locations); err == nil && len(locations) > 0 {
			return formatLocations(locations)
		}
		var single Location
		if err := json.Unmarshal(raw, &single); err == nil {
			return formatLocations([]Location{single})
		}
	case "hover":
		var hover Hover
		if err := json.Unmarshal(raw, &hover); err == nil {
			return hover.String()
		}
	case "documentSymbol":
		var symbols []DocumentSymbol
		if err := json.Unmarshal(raw, &symbols); err == nil {
			return formatDocumentSymbols(symbols, 0)
		}
		var symInfos []SymbolInformation
		if err := json.Unmarshal(raw, &symInfos); err == nil {
			return formatSymbolInformation(symInfos)
		}
	case "workspaceSymbol":
		var symInfos []SymbolInformation
		if err := json.Unmarshal(raw, &symInfos); err == nil {
			return formatSymbolInformation(symInfos)
		}
	}
	return string(raw)
}

// countResults tries to count the number of items in an LSP result.
func countResults(raw json.RawMessage) int {
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err == nil {
		return len(arr)
	}
	if len(raw) > 0 && raw[0] == '{' {
		return 1
	}
	return 0
}

// Location LSP location.
type Location struct {
	URI   string `json:"uri"`
	Range Range  `json:"range"`
}

// Range LSP range.
type Range struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

// Position LSP position.
type Position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

// Hover LSP hover result.
type Hover struct {
	Contents json.RawMessage `json:"contents"`
	Range    *Range          `json:"range,omitempty"`
}

func (h Hover) String() string {
	var s string
	// Try string first
	if err := json.Unmarshal(h.Contents, &s); err == nil {
		return s
	}
	// Try MarkupContent
	var mc MarkupContent
	if err := json.Unmarshal(h.Contents, &mc); err == nil {
		return mc.Value
	}
	// Try array of MarkedString
	var arr []json.RawMessage
	if err := json.Unmarshal(h.Contents, &arr); err == nil {
		var parts []string
		for _, item := range arr {
			var str string
			if err := json.Unmarshal(item, &str); err == nil {
				parts = append(parts, str)
				continue
			}
			var m MarkupContent
			if err := json.Unmarshal(item, &m); err == nil {
				parts = append(parts, m.Value)
			}
		}
		return strings.Join(parts, "\n")
	}
	return string(h.Contents)
}

// MarkupContent LSP markup content.
type MarkupContent struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// DocumentSymbol LSP document symbol.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           int              `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// SymbolInformation LSP symbol information.
type SymbolInformation struct {
	Name          string   `json:"name"`
	Kind          int      `json:"kind"`
	ContainerName string   `json:"containerName,omitempty"`
	Location      Location `json:"location"`
}

func formatLocations(locs []Location) string {
	var lines []string
	for _, loc := range locs {
		path := strings.TrimPrefix(loc.URI, "file://")
		lines = append(lines, fmt.Sprintf("%s:%d:%d", path, loc.Range.Start.Line+1, loc.Range.Start.Character+1))
	}
	return strings.Join(lines, "\n")
}

func formatDocumentSymbols(syms []DocumentSymbol, indent int) string {
	var lines []string
	prefix := strings.Repeat("  ", indent)
	for _, s := range syms {
		line := fmt.Sprintf("%s%s (%s) at L%d", prefix, s.Name, symbolKindName(s.Kind), s.Range.Start.Line+1)
		if s.Detail != "" {
			line += " - " + s.Detail
		}
		lines = append(lines, line)
		if len(s.Children) > 0 {
			lines = append(lines, formatDocumentSymbols(s.Children, indent+1))
		}
	}
	return strings.Join(lines, "\n")
}

func formatSymbolInformation(syms []SymbolInformation) string {
	var lines []string
	for _, s := range syms {
		path := strings.TrimPrefix(s.Location.URI, "file://")
		line := fmt.Sprintf("%s (%s) in %s", s.Name, symbolKindName(s.Kind), s.ContainerName)
		if line == "" {
			line = s.Name
		}
		lines = append(lines, fmt.Sprintf("%s at %s:%d:%d", line, path, s.Location.Range.Start.Line+1, s.Location.Range.Start.Character+1))
	}
	return strings.Join(lines, "\n")
}

func symbolKindName(kind int) string {
	names := map[int]string{
		1: "File", 2: "Module", 3: "Namespace", 4: "Package", 5: "Class",
		6: "Method", 7: "Property", 8: "Field", 9: "Constructor", 10: "Enum",
		11: "Interface", 12: "Function", 13: "Variable", 14: "Constant", 15: "String",
		16: "Number", 17: "Boolean", 18: "Array", 19: "Object", 20: "Key",
		21: "Null", 22: "EnumMember", 23: "Struct", 24: "Event", 25: "Operator",
		26: "TypeParameter",
	}
	if n, ok := names[kind]; ok {
		return n
	}
	return "Unknown(" + strconv.Itoa(kind) + ")"
}

// PathToURI exported helper wraps lsp.PathToURI for use by tests/tools.
func PathToURI(path string) string {
	return lsp.PathToURI(path)
}
