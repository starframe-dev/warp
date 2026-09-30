package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type selectableFixturePanel struct {
	updates []tea.Msg
}

func (*selectableFixturePanel) View(_, _ int) string {
	return "\x1b[31mhé界x\x1b[0m\n第二行"
}

func (p *selectableFixturePanel) Update(msg tea.Msg) tea.Cmd {
	p.updates = append(p.updates, msg)
	return nil
}

func selectableMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func TestSelectableMouseKeyboardRenderingExtractionAndForwarding(t *testing.T) {
	content := &selectableFixturePanel{}
	selected := warp.NewSelectable(content)

	// Render first so pointer coordinates are clamped to the visible panel and
	// ANSI-colored text is selected using terminal-cell coordinates.
	selected.View(8, 2)
	selected.Update(selectableMouse(0, 0, tea.MouseActionPress))
	selected.Update(selectableMouse(2, 0, tea.MouseActionMotion))
	selected.Update(selectableMouse(2, 0, tea.MouseActionRelease))

	if !selected.HasSelection || selected.Selecting {
		t.Fatalf("mouse selection state: HasSelection=%v Selecting=%v", selected.HasSelection, selected.Selecting)
	}
	if got, want := selected.SelectedText(), "hé界"; got != want {
		t.Fatalf("mouse SelectedText()=%q, want %q", got, want)
	}
	rendered := selected.View(8, 2)
	if !strings.Contains(rendered, "\x1b[7m") || !strings.Contains(rendered, "hé界") {
		t.Fatalf("selected rendering lacks reverse-video highlight or text: %q", rendered)
	}

	selected.ClearSelection()
	if selected.HasSelection || selected.Selecting || selected.SelectedText() != "" {
		t.Fatal("ClearSelection did not clear the selection")
	}

	// Shift+right starts a keyboard selection at the current cursor position.
	selected.CursorX, selected.CursorY = 0, 1
	selected.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if !selected.HasSelection {
		t.Fatal("Shift+Right did not create a selection")
	}
	if got, want := selected.SelectedText(), "第"; got != want {
		t.Fatalf("keyboard SelectedText()=%q, want %q", got, want)
	}
	if rendered := selected.View(8, 2); !strings.Contains(rendered, "\x1b[7m第\x1b[0m") {
		t.Fatalf("keyboard selection not highlighted: %q", rendered)
	}

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'z'}}
	selected.Update(key)
	if len(content.updates) != 4 {
		t.Fatalf("mouse and unhandled key forwarding: updates=%#v", content.updates)
	}
	forwarded, ok := content.updates[len(content.updates)-1].(tea.KeyMsg)
	if !ok || forwarded.Type != key.Type || string(forwarded.Runes) != string(key.Runes) {
		t.Fatalf("forwarded message=%#v, want %#v", content.updates[len(content.updates)-1], key)
	}

	selected.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if selected.HasSelection || selected.SelectedText() != "" {
		t.Fatal("Escape did not clear the selection")
	}
	if len(content.updates) != 4 {
		t.Fatalf("handled Escape was forwarded to content; updates=%#v", content.updates)
	}
}
