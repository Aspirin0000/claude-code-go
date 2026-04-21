package ui

import (
	"strings"
	"testing"
)

// ============================================================================
// TerminalUI Print Methods Tests
// ============================================================================

func TestTerminalUIPrintWelcome(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.PrintWelcome()
}

func TestTerminalUIPrintPrompt(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.PrintPrompt()
}

func TestTerminalUIPrintMessage(t *testing.T) {
	ui := NewTerminalUI()

	// Test user message
	ui.PrintMessage("user", "Hello")

	// Test assistant message
	ui.PrintMessage("assistant", "Hi there")

	// Test system message
	ui.PrintMessage("system", "System notification")

	// Test default message
	ui.PrintMessage("unknown", "Unknown role")
}

func TestTerminalUIPrintError(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.PrintError("Something went wrong")
}

func TestTerminalUIPrintSuccess(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.PrintSuccess("Operation completed")
}

func TestTerminalUIPrintToolUse(t *testing.T) {
	ui := NewTerminalUI()
	// Test with input
	ui.PrintToolUse("bash", "ls -la")

	// Test without input
	ui.PrintToolUse("bash", "")
}

func TestTerminalUIPrintToolResult(t *testing.T) {
	ui := NewTerminalUI()

	// Test success with output
	ui.PrintToolResult(true, "line1\nline2\nline3")

	// Test success with no output
	ui.PrintToolResult(true, "")

	// Test failure
	ui.PrintToolResult(false, "Error occurred")

	// Test success with many lines (should truncate)
	longOutput := "line1\nline2\nline3\nline4\nline5\nline6\nline7"
	ui.PrintToolResult(true, longOutput)
}

func TestTerminalUIPrintHelp(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.PrintHelp()
}

func TestTerminalUIClearScreen(t *testing.T) {
	ui := NewTerminalUI()
	// Just ensure it doesn't panic
	ui.ClearScreen()
}

// ============================================================================
// WrapText Extended Tests
// ============================================================================

func TestWrapTextEmpty(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(10, 10)

	result := ui.WrapText("")
	if result != "" {
		t.Errorf("expected empty string, got %q", result)
	}
}

func TestWrapTextExactWidth(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(10, 10)

	// Text exactly 10 characters (width)
	result := ui.WrapText("0123456789")
	if result != "0123456789" {
		t.Errorf("expected '0123456789', got %q", result)
	}
}

func TestWrapTextMultipleLines(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(10, 10)

	// Text with multiple newlines
	result := ui.WrapText("line1\nline2\nline3")
	lines := strings.Split(result, "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestWrapTextLongLine(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(5, 10)

	// Line longer than width
	result := ui.WrapText("0123456789")
	lines := strings.Split(result, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "01234" {
		t.Errorf("expected '01234', got %q", lines[0])
	}
	if lines[1] != "56789" {
		t.Errorf("expected '56789', got %q", lines[1])
	}
}

func TestWrapTextVeryLongLine(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(3, 10)

	// Line much longer than width
	result := ui.WrapText("0123456789")
	lines := strings.Split(result, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(lines))
	}
}

// ============================================================================
// Terminal Dimensions Tests
// ============================================================================

func TestTerminalUIDimensions(t *testing.T) {
	ui := NewTerminalUI()

	// Default dimensions should be set
	if ui.width <= 0 {
		t.Errorf("expected positive width, got %d", ui.width)
	}
	if ui.height <= 0 {
		t.Errorf("expected positive height, got %d", ui.height)
	}
}

func TestSetSizeZero(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(0, 0)
	if ui.width != 0 {
		t.Errorf("expected width 0, got %d", ui.width)
	}
	if ui.height != 0 {
		t.Errorf("expected height 0, got %d", ui.height)
	}
}

func TestSetSizeLarge(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(9999, 9999)
	if ui.width != 9999 {
		t.Errorf("expected width 9999, got %d", ui.width)
	}
	if ui.height != 9999 {
		t.Errorf("expected height 9999, got %d", ui.height)
	}
}
