package main

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/starframe-dev/warp"
)

type statusPanel struct {
	msg string
}

func (p *statusPanel) View(w, h int) string {
	line := padRight(p.msg, w)
	lines := make([]string, h)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}

func (p *statusPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *statusPanel) Set(msg string) {
	p.msg = msg
}

func TestPadRight(t *testing.T) {
	cases := []struct {
		input string
		width int
		want  int
	}{
		{"hello", 10, 10},
		{"hello world", 5, 5},
		{"", 3, 3},
		{"exact", 5, 5},
	}
	for _, c := range cases {
		got := padRight(c.input, c.width)
		if len(got) != c.want {
			t.Errorf("padRight(%q, %d) length = %d, want %d", c.input, c.width, len(got), c.want)
		}
	}
}

func TestDemoPanel(t *testing.T) {
	p := &demoPanel{name: "Test", count: 0}

	view := p.View(20, 4)
	if !strings.Contains(view, "Test") {
		t.Errorf("View missing panel name, got %q", view)
	}
	if !strings.Contains(view, "clicks: 0") {
		t.Errorf("View missing initial click count, got %q", view)
	}

	if cmd := p.Update(tea.MouseMsg{}); cmd != nil {
		t.Errorf("Update returned non-nil command: %v", cmd)
	}
	if p.count != 1 {
		t.Errorf("count = %d after mouse message, want 1", p.count)
	}

	view2 := p.View(20, 4)
	if !strings.Contains(view2, "clicks: 1") {
		t.Errorf("View missing updated click count, got %q", view2)
	}
}

func TestTextPanel(t *testing.T) {
	text := "This is a sample text used for the text panel wrapping behavior."
	p := newTextPanel(text, 20)

	if len(p.lines) == 0 {
		t.Fatal("newTextPanel produced no lines")
	}

	view := p.View(20, 5)
	if !strings.Contains(view, p.lines[0]) {
		t.Errorf("View missing first line, got %q", view)
	}

	if cmd := p.Update(tea.MouseMsg{}); cmd != nil {
		t.Errorf("Update returned non-nil command: %v", cmd)
	}

	empty := newTextPanel("", 20)
	viewEmpty := empty.View(10, 3)
	if viewEmpty == "" {
		t.Error("empty textPanel View returned empty string")
	}
}

func TestStatusPanel(t *testing.T) {
	p := &statusPanel{msg: "ready"}

	view := p.View(10, 3)
	if !strings.Contains(view, "ready") {
		t.Errorf("View missing initial message, got %q", view)
	}

	p.Set("updated")
	if p.msg != "updated" {
		t.Errorf("msg = %q after Set, want updated", p.msg)
	}

	view2 := p.View(10, 3)
	if !strings.Contains(view2, "updated") {
		t.Errorf("View missing updated message, got %q", view2)
	}

	if cmd := p.Update(tea.MouseMsg{}); cmd != nil {
		t.Errorf("Update returned non-nil command: %v", cmd)
	}
}

func TestAppRoot(t *testing.T) {
	tg := warp.NewTabGroup(warp.TabTop)
	root := &appRoot{tg: tg}

	view := root.View(40, 10)
	if view == "" {
		t.Error("appRoot.View returned empty string")
	}

	if cmd := root.Update(tea.KeyMsg{Type: tea.KeyTab}); cmd != nil {
		t.Errorf("tab key Update returned non-nil command: %v", cmd)
	}

	if cmd := root.Update(tea.KeyMsg{Type: tea.KeyShiftTab}); cmd != nil {
		t.Errorf("shift+tab key Update returned non-nil command: %v", cmd)
	}

	if cmd := root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}); cmd != nil {
		t.Errorf("non-tab key Update returned non-nil command: %v", cmd)
	}

	if cmd := root.Update(tea.MouseMsg{}); cmd != nil {
		t.Errorf("mouse Update returned non-nil command: %v", cmd)
	}
}
