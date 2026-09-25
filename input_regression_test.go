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

func TestInputEditingUsesGraphemeBoundaries(t *testing.T) {
	clusters := []string{"e\u0301", "👩‍💻", "👨‍👩‍👧‍👦", "🇫🇮"}
	for _, cluster := range clusters {
		runeCount := len([]rune(cluster))
		input := NewInput("")
		input.Focus()
		input.SetValue("A" + cluster + "B")

		input.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if input.Cursor != 1+runeCount {
			t.Fatalf("left before B for %q: cursor=%d, want %d", cluster, input.Cursor, 1+runeCount)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if input.Cursor != 1 {
			t.Fatalf("left before %q: cursor=%d, want 1", cluster, input.Cursor)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyLeft})
		if input.Cursor != 0 {
			t.Fatalf("left before start for %q: cursor=%d, want 0", cluster, input.Cursor)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyRight})
		if input.Cursor != 1 {
			t.Fatalf("right before %q: cursor=%d, want 1", cluster, input.Cursor)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyRight})
		if input.Cursor != 1+runeCount {
			t.Fatalf("right after %q: cursor=%d, want %d", cluster, input.Cursor, 1+runeCount)
		}

		input.Update(tea.KeyMsg{Type: tea.KeyHome})
		if input.Cursor != 0 {
			t.Fatalf("home for %q: cursor=%d, want 0", cluster, input.Cursor)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyEnd})
		if input.Cursor != 2+runeCount {
			t.Fatalf("end for %q: cursor=%d, want %d", cluster, input.Cursor, 2+runeCount)
		}

		input.SetCursor(2)
		if input.Cursor != 1+runeCount {
			t.Fatalf("SetCursor inside %q: cursor=%d, want %d", cluster, input.Cursor, 1+runeCount)
		}
		input.Update(tea.KeyMsg{Type: tea.KeyBackspace})
		if input.Value != "AB" || input.Cursor != 1 {
			t.Fatalf("backspace after %q: value=%q cursor=%d", cluster, input.Value, input.Cursor)
		}

		input.SetValue("A" + cluster + "B")
		input.SetCursor(1)
		input.Update(tea.KeyMsg{Type: tea.KeyDelete})
		if input.Value != "AB" || input.Cursor != 1 {
			t.Fatalf("delete before %q: value=%q cursor=%d", cluster, input.Value, input.Cursor)
		}

		input.SetValue(cluster)
		input.SetCursor(0)
		input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'X'}})
		if input.Value != "X"+cluster || input.Cursor != 1 {
			t.Fatalf("insert before %q: value=%q cursor=%d", cluster, input.Value, input.Cursor)
		}
		input.SetCursor(1 + runeCount)
		input.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Y'}})
		if input.Value != "X"+cluster+"Y" || input.Cursor != 2+runeCount {
			t.Fatalf("insert after %q: value=%q cursor=%d", cluster, input.Value, input.Cursor)
		}

		for width := 0; width <= 6; width++ {
			if got := ansi.StringWidth(input.View(width, 1)); got != width {
				t.Errorf("narrow view after editing %q at width %d has %d cells", cluster, width, got)
			}
		}
	}
}

func TestInputBoxedValueAppearsInOneInteriorRow(t *testing.T) {
	input := NewInput("")
	input.Focus()
	input.SetValue("boxed-value")

	for height := 3; height <= 9; height++ {
		view := input.View(30, height)
		lines := strings.Split(ansi.Strip(view), "\n")
		if len(lines) != height {
			t.Fatalf("boxed input height=%d produced %d rows", height, len(lines))
		}
		if got := strings.Count(ansi.Strip(view), "boxed-value"); got != 1 {
			t.Fatalf("boxed input height=%d rendered value %d times", height, got)
		}
		contentRow := 1 + (height-3)/2
		if !strings.Contains(lines[contentRow], "boxed-value") {
			t.Fatalf("height=%d content row %d=%q does not contain the value", height, contentRow, lines[contentRow])
		}
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
