package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestInputRendersUnicodeAtTerminalCellWidth(t *testing.T) {
	input := NewInput("> ")
	input.Focus()
	input.SetValue("A界🙂e\u0301")
	input.SetCursor(len([]rune(input.Value)))

	for _, width := range []int{0, 1, 2, 3, 4, 5, 7, 8, 12} {
		view := input.View(width, 1)
		if got := ansi.StringWidth(view); got != width {
			t.Errorf("width=%d produced %d cells: %q", width, got, view)
		}
	}
}

func TestInputAcceptsMultiRuneKeyEvents(t *testing.T) {
	input := NewInput("")
	input.Focus()
	text := "👩‍💻"
	input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(text)})
	if input.Value != text || input.Cursor != len([]rune(text)) {
		t.Fatalf("multi-rune input value=%q cursor=%d", input.Value, input.Cursor)
	}
}

func TestInputNarrowPromptAndBoxStayWithinCellBounds(t *testing.T) {
	input := NewInput("界🙂 ")
	for width := 0; width <= 8; width++ {
		for height := 0; height <= 5; height++ {
			lines := strings.Split(input.View(width, height), "\n")
			if height == 0 {
				if input.View(width, height) != "" {
					t.Fatalf("zero-height input returned content")
				}
				continue
			}
			if len(lines) != height {
				t.Fatalf("size %dx%d produced %d lines", width, height, len(lines))
			}
			for y, line := range lines {
				if got := ansi.StringWidth(line); got != width {
					t.Errorf("size %dx%d line %d width=%d", width, height, y, got)
				}
			}
		}
	}
}
