package ui

import (
	"strings"
	"testing"
)

func TestNewTerminalUI(t *testing.T) {
	ui := NewTerminalUI()
	if ui.width <= 0 || ui.height <= 0 {
		t.Errorf("unexpected dimensions: %dx%d", ui.width, ui.height)
	}
}

func TestSetSize(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(100, 50)
	if ui.width != 100 || ui.height != 50 {
		t.Errorf("unexpected dimensions after set: %dx%d", ui.width, ui.height)
	}
}

func TestWrapText(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(10, 10)

	wrapped := ui.WrapText("hello world test")
	lines := strings.Split(wrapped, "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "hello worl" {
		t.Errorf("unexpected first line: %s", lines[0])
	}
	if lines[1] != "d test" {
		t.Errorf("unexpected second line: %s", lines[1])
	}
}

func TestWrapTextMultiline(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(20, 10)

	wrapped := ui.WrapText("line1\nthis is a very long line that needs wrapping")
	lines := strings.Split(wrapped, "\n")
	if len(lines) < 3 {
		t.Errorf("expected at least 3 lines, got %d", len(lines))
	}
}

func TestWrapTextShort(t *testing.T) {
	ui := NewTerminalUI()
	ui.SetSize(50, 10)

	wrapped := ui.WrapText("short")
	if wrapped != "short" {
		t.Errorf("unexpected wrap result: %s", wrapped)
	}
}
