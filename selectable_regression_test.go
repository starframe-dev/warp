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

func TestSelectableEndExclusiveSelectionAcrossUnicode(t *testing.T) {
	texts := []string{"abc", "Привет", "界A", "👩‍💻x", "éx"}
	for _, text := range texts {
		text := text
		width := ansi.StringWidth(text)
		t.Run(text, func(t *testing.T) {
			newSelectable := func() *Selectable {
				selectable := NewSelectable(&selectableTextPanel{text: text})
				selectable.View(width, 1)
				return selectable
			}

			t.Run("select all", func(t *testing.T) {
				selectable := newSelectable()
				selectable.SelectAll(width, 1)
				selectable.View(width, 1)
				if got := selectable.SelectedText(); got != text {
					t.Fatalf("SelectAll selected %q, want %q", got, text)
				}
			})

			for _, direction := range []struct {
				name  string
				start int
				end   int
			}{
				{name: "forward mouse", start: 0, end: width - 1},
				{name: "reverse mouse", start: width - 1, end: 0},
			} {
				direction := direction
				t.Run(direction.name, func(t *testing.T) {
					selectable := newSelectable()
					selectable.Update(tea.MouseMsg{X: direction.start, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
					selectable.Update(tea.MouseMsg{X: direction.end, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
					selectable.Update(tea.MouseMsg{X: direction.end, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
					selectable.View(width, 1)
					if got := selectable.SelectedText(); got != text {
						t.Fatalf("mouse selection selected %q, want %q", got, text)
					}
				})
			}

			t.Run("Shift+Right to line end", func(t *testing.T) {
				selectable := newSelectable()
				for range width {
					selectable.Update(tea.KeyMsg{Type: tea.KeyShiftRight})
				}
				selectable.View(width, 1)
				if got := selectable.SelectedText(); got != text {
					t.Fatalf("Shift+Right selected %q, want %q", got, text)
				}
			})
		})
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
