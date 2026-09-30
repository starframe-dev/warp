package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type contentWrappersPanel struct {
	lines []string
}

func (p *contentWrappersPanel) View(width, height int) string {
	if height <= 0 {
		return ""
	}
	lines := make([]string, height)
	for i := range lines {
		if i < len(p.lines) {
			line := p.lines[i]
			if len(line) > width {
				line = line[:width]
			}
			lines[i] = line
		}
	}
	return strings.Join(lines, "\n")
}

func (*contentWrappersPanel) Update(tea.Msg) tea.Cmd { return nil }

func (p *contentWrappersPanel) ContentHeight(int) (int, bool) {
	return len(p.lines), true
}

func TestContentWrappersCollapseScrollAndSelectText(t *testing.T) {
	const width, height = 24, 4
	longContent := &contentWrappersPanel{lines: []string{
		"select this text",
		"second content row",
		"third content row",
		"fourth content row",
		"fifth content row",
		"sixth content row",
		"seventh content row",
		"eighth content row",
		"ninth content row",
		"tenth content row",
		"eleventh content row",
		"twelfth content row",
	}}

	section := warp.NewCollapsible("Long section", longContent)
	if view := warp.StripANSI(section.View(width, height)); !strings.Contains(view, "▼ Long section") || !strings.Contains(view, "select this text") {
		t.Fatalf("expanded collapsible view = %q", view)
	}
	section.Update(tea.MouseMsg{X: 2, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if !section.Collapsed {
		t.Fatal("clicking the collapsible title did not collapse it")
	}
	if view := warp.StripANSI(section.View(width, height)); !strings.Contains(view, "▶ Long section") || strings.Contains(view, "select this text") {
		t.Fatalf("collapsed collapsible view = %q", view)
	}
	section.Update(tea.MouseMsg{X: 2, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if section.Collapsed {
		t.Fatal("clicking the collapsible title did not expand it")
	}

	scroller := warp.NewScrollable(longContent)
	scroller.Update(tea.WindowSizeMsg{Width: width, Height: height})
	scroller.View(width, height)
	scroller.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
	if scroller.Offset != 3 {
		t.Fatalf("wheel-down offset = %d, want 3", scroller.Offset)
	}
	if view := warp.StripANSI(scroller.View(width, height)); !strings.Contains(view, "fourth content row") {
		t.Fatalf("view after wheel scrolling = %q", view)
	}
	scroller.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if scroller.Offset != 8 {
		t.Fatalf("page-down offset = %d, want the maximum viewport offset 8", scroller.Offset)
	}
	if view := warp.StripANSI(scroller.View(width, height)); !strings.Contains(view, "ninth content row") {
		t.Fatalf("view after page-down = %q", view)
	}
	scroller.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	if scroller.Offset != 0 {
		t.Fatalf("page-up offset = %d, want 0", scroller.Offset)
	}

	selectable := warp.NewSelectable(&contentWrappersPanel{lines: []string{"select this text", "another row"}})
	selectable.View(width, height)
	selectable.Update(tea.MouseMsg{X: 0, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	selectable.Update(tea.MouseMsg{X: 5, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	selectable.Update(tea.MouseMsg{X: 5, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	if got := selectable.SelectedText(); got != "select" {
		t.Fatalf("mouse-selected text = %q, want %q", got, "select")
	}
	selectable.ClearSelection()
	if selectable.HasSelection || selectable.Selecting || selectable.SelectedText() != "" {
		t.Fatal("ClearSelection did not clear mouse selection")
	}

	selectable = warp.NewSelectable(&contentWrappersPanel{lines: []string{"select this text", "another row"}})
	selectable.View(width, height)
	selectable.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	selectable.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
	if got := selectable.SelectedText(); got != "se" {
		t.Fatalf("keyboard-selected text = %q, want %q", got, "se")
	}
	selectable.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if selectable.HasSelection || selectable.SelectedText() != "" {
		t.Fatal("Escape did not clear keyboard selection")
	}
}
