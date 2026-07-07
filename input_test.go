package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewInput(t *testing.T) {
	in := NewInput("$ ")
	if in.Value != "" {
		t.Errorf("expected empty value, got %q", in.Value)
	}
	if in.Cursor != 0 {
		t.Errorf("expected cursor 0, got %d", in.Cursor)
	}
	if in.Prompt != "$ " {
		t.Errorf("expected prompt '$ ', got %q", in.Prompt)
	}
	if in.Focused() {
		t.Errorf("expected new input to be blurred")
	}
}

func TestInputFocusBlur(t *testing.T) {
	in := NewInput("> ")
	if in.Focused() {
		t.Errorf("expected Focused to be false initially")
	}
	in.Focus()
	if !in.Focused() {
		t.Errorf("expected Focused to be true after Focus")
	}
	in.Blur()
	if in.Focused() {
		t.Errorf("expected Focused to be false after Blur")
	}
}

func TestInputSetValueAndCursor(t *testing.T) {
	in := NewInput("> ")
	in.SetValue("hello")
	if in.Value != "hello" {
		t.Errorf("expected value 'hello', got %q", in.Value)
	}
	if in.Cursor != 5 {
		t.Errorf("expected cursor at end (5), got %d", in.Cursor)
	}

	in.SetCursor(2)
	if in.Cursor != 2 {
		t.Errorf("expected cursor 2, got %d", in.Cursor)
	}

	in.SetCursor(-1)
	if in.Cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", in.Cursor)
	}

	in.SetCursor(100)
	if in.Cursor != 5 {
		t.Errorf("expected cursor clamped to 5, got %d", in.Cursor)
	}
}

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

	in.Update(tea.KeyMsg{Type: tea.KeyRight})
	if in.Cursor != 5 {
		t.Errorf("expected cursor 5 after right, got %d", in.Cursor)
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

func TestInputSpecialKeys(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	in.SetValue("abc")
	in.SetCursor(1)

	in.Update(tea.KeyMsg{Type: tea.KeyTab})
	if in.Value != "abc" || in.Cursor != 1 {
		t.Errorf("expected tab to be a no-op, got value %q cursor %d", in.Value, in.Cursor)
	}

	in.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if in.Value != "abc" || in.Cursor != 1 {
		t.Errorf("expected shift+tab to be a no-op, got value %q cursor %d", in.Value, in.Cursor)
	}

	in.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if in.Value != "abc" || in.Cursor != 1 {
		t.Errorf("expected enter to be a no-op, got value %q cursor %d", in.Value, in.Cursor)
	}

	in.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if in.Value != "abc" || in.Cursor != 1 {
		t.Errorf("expected unknown key to be a no-op, got value %q cursor %d", in.Value, in.Cursor)
	}
}

func TestInputFocusVisualBoxed(t *testing.T) {
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

func TestInputViewInline(t *testing.T) {
	in := NewInput("> ")
	in.SetValue("test")

	rendered := in.View(20, 1)
	lines := strings.Split(rendered, "\n")
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}

	if !strings.Contains(rendered, "test") {
		t.Errorf("expected value 'test' in rendered line, got %q", rendered)
	}
}

func TestInputEmptyValue(t *testing.T) {
	in := NewInput("> ")
	in.SetValue("")

	if in.Value != "" {
		t.Errorf("expected empty value, got %q", in.Value)
	}
	if in.Cursor != 0 {
		t.Errorf("expected cursor 0, got %d", in.Cursor)
	}

	rendered := in.View(20, 1)
	if rendered != "" {
		t.Errorf("expected empty rendered output, got %q", rendered)
	}
}

func TestInputLongValueTruncation(t *testing.T) {
	in := NewInput("> ")
	in.Focus()
	longValue := strings.Repeat("x", 100)
	in.SetValue(longValue)

	if in.Cursor != 100 {
		t.Errorf("expected cursor 100, got %d", in.Cursor)
	}

	rendered := in.View(10, 1)
	if strings.Count(rendered, "x") > 10 {
		t.Errorf("expected truncated value, got %q", rendered)
	}
}

func TestInputClampCursorOnSetValue(t *testing.T) {
	in := NewInput("> ")
	in.SetValue("abc")

	if in.Cursor != 3 {
		t.Errorf("expected cursor 3 after setValue, got %d", in.Cursor)
	}
}

func TestInputClampCursorOnSetCursor(t *testing.T) {
	in := NewInput("> ")
	in.SetValue("abc")

	in.SetCursor(100)
	if in.Cursor != 3 {
		t.Errorf("expected cursor clamped to 3, got %d", in.Cursor)
	}

	in.SetCursor(-5)
	if in.Cursor != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", in.Cursor)
	}
}