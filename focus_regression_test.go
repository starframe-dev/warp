package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestMouseClickChangesRealInputFocus(t *testing.T) {
	first := NewInput("A: ")
	second := NewInput("B: ")
	tg := NewTabGroup(TabNone)
	tab := tg.ActiveTab()
	tab.SetRootPanel(first)
	tab.SplitVertical(first, 0.5, second)
	tg.Update(tea.WindowSizeMsg{Width: 20, Height: 6})
	tg.View(20, 6)

	tg.Update(tea.MouseMsg{X: 1, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !first.Focused() || second.Focused() {
		t.Fatalf("left click focus: first=%v second=%v", first.Focused(), second.Focused())
	}
	tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if first.Value != "a" {
		t.Fatalf("first input received value %q, want a", first.Value)
	}

	tg.Update(tea.MouseMsg{X: 10, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if first.Focused() || !second.Focused() {
		t.Fatalf("right click focus: first=%v second=%v", first.Focused(), second.Focused())
	}
	tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if second.Value != "b" {
		t.Fatalf("second input received value %q, want b", second.Value)
	}
	if first.Value != "a" {
		t.Fatalf("switching focus changed first input to %q", first.Value)
	}
}
