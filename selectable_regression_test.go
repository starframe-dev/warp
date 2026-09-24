package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type selectableTextPanel struct{ text string }

func (p *selectableTextPanel) View(_, _ int) string { return p.text }
func (*selectableTextPanel) Update(tea.Msg) tea.Cmd { return nil }

func TestSelectableUsesTerminalCellsForCJKAndEmoji(t *testing.T) {
	content := &selectableTextPanel{text: "界🙂e\u0301"}
	selectable := NewSelectable(content)
	view := selectable.View(8, 1)
	if got := ansi.StringWidth(view); got != ansi.StringWidth(content.text) {
		t.Fatalf("view width=%d, want content width %d", got, ansi.StringWidth(content.text))
	}

	selectable.AnchorX, selectable.AnchorY = 1, 0
	selectable.CursorX, selectable.CursorY = 2, 0
	selectable.HasSelection = true
	selected := selectable.SelectedText()
	if selected != "界" {
		t.Fatalf("selection of one CJK cell returned %q, want complete grapheme 界", selected)
	}
	highlighted := ansi.Strip(selectable.View(8, 1))
	if highlighted != content.text {
		t.Fatalf("highlighting changed visible Unicode text: %q", highlighted)
	}

	selectable.AnchorX, selectable.CursorX = 2, 4
	if got := selectable.SelectedText(); got != "🙂" {
		t.Fatalf("emoji selection=%q, want 🙂", got)
	}

	selectable.AnchorX, selectable.CursorX = 4, 5
	if got := selectable.SelectedText(); got != "e\u0301" {
		t.Fatalf("combining selection=%q, want é", got)
	}
}

func TestSelectableClampsZeroAndNegativeViewports(t *testing.T) {
	selectable := NewSelectable(&geometryTestPanel{name: "界🙂"})
	selectable.HasSelection = true
	selectable.AnchorX, selectable.CursorX = 1, 2
	if got := selectable.View(0, 0); got != "" {
		t.Fatalf("zero viewport returned %q", got)
	}
	selectable.SelectAll(0, 2)
	if selectable.HasSelection || selectable.Selecting {
		t.Fatal("SelectAll should clear selection for an empty viewport")
	}
	if got := selectable.View(-1, -1); got != "" {
		t.Fatalf("negative viewport returned %q", got)
	}
}
