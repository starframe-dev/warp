package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputTyping(t *testing.T) {
	in := NewInput("> ")
	in.Focus()

	for _, r := range "hi" {
		in.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	if in.Value != "hi" {
		t.Errorf("expected value 'hi', got %q", in.Value)
	}
	if in.Cursor != 2 {
		t.Errorf("expected cursor 2, got %d", in.Cursor)
	}
}

func TestInputBackspace(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("abc")
	in.SetCursor(2)

	in.Update(tea.KeyMsg{Type: tea.KeyBackspace})

	if in.Value != "ac" {
		t.Errorf("expected value 'ac', got %q", in.Value)
	}
	if in.Cursor != 1 {
		t.Errorf("expected cursor 1, got %d", in.Cursor)
	}
}

func TestInputDelete(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("abc")
	in.SetCursor(1)

	in.Update(tea.KeyMsg{Type: tea.KeyDelete})

	if in.Value != "ac" {
		t.Errorf("expected value 'ac', got %q", in.Value)
	}
	if in.Cursor != 1 {
		t.Errorf("expected cursor 1, got %d", in.Cursor)
	}
}

func TestInputCursorMovement(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("hello")
	in.SetCursor(5)

	in.Update(tea.KeyMsg{Type: tea.KeyLeft})
	if in.Cursor != 4 {
		t.Errorf("expected cursor 4 after left, got %d", in.Cursor)
	}

	in.Update(tea.KeyMsg{Type: tea.KeyHome})
	if in.Cursor != 0 {
		t.Errorf("expected cursor 0 after home, got %d", in.Cursor)
	}

	in.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if in.Cursor != 5 {
		t.Errorf("expected cursor 5 after end, got %d", in.Cursor)
	}
}

func TestInputDoesNotTypeWhenBlurred(t *testing.T) {
	in := NewInput("> ")
	in.Blur()

	in.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})

	if in.Value != "" {
		t.Errorf("expected no input when blurred, got %q", in.Value)
	}
}

func TestInputFocusVisual(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("ok")

	rendered := in.View(20, 3)
	lines := strings.Split(rendered, "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if !strings.Contains(lines[1], "o") || !strings.Contains(lines[1], "k") {
		t.Errorf("expected value 'ok' in rendered line, got %q", lines[1])
	}
}

func TestInputCursorRender(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("ab")
	in.SetCursor(1)

	rendered := in.View(20, 1)
	if !strings.Contains(rendered, "\x1b[7m") {
		t.Errorf("expected cursor highlight ANSI code, got %q", rendered)
	}
}
