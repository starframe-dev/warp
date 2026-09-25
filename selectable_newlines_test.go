package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type lineStructurePanel struct {
	BasePanel
	content string
}

func (p *lineStructurePanel) View(_, _ int) string {
	return p.content
}

func TestSelectableSelectAllPreservesEmptyRows(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		want  string
		width int
	}{
		{name: "single empty middle row", text: "a\n\nb", want: "a\n\nb", width: 1},
		{name: "multiple empty middle rows", text: "foo\n\n\nbar", want: "foo\n\n\nbar", width: 3},
		{name: "leading empty row", text: "\nabc", want: "\nabc", width: 3},
		{name: "trailing empty row", text: "abc\n", want: "abc\n", width: 3},
		{name: "wide graphemes", text: "界\n\n👩‍💻", want: "界\n\n👩‍💻", width: 2},
		{name: "ANSI-only middle row", text: "a\n\x1b[31m\x1b[0m\nb", want: "a\n\nb", width: 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			lines := strings.Split(tc.text, "\n")
			selectable := NewSelectable(&lineStructurePanel{content: tc.text})
			selectable.SelectAll(tc.width, len(lines))

			if got := selectable.SelectedText(); got != tc.want {
				t.Fatalf("SelectedText()=%q, want %q", got, tc.want)
			}
		})
	}
}

func TestSelectableMouseSelectionPreservesEmptyRows(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		want  string
		width int
	}{
		{name: "single empty middle row", text: "a\n\nb", want: "a\n\nb", width: 1},
		{name: "multiple empty middle rows", text: "foo\n\n\nbar", want: "foo\n\n\nbar", width: 3},
		{name: "leading empty row", text: "\nabc", want: "\nabc", width: 3},
		{name: "trailing empty row", text: "abc\n", want: "abc\n", width: 3},
		{name: "wide graphemes", text: "界\n\n👩‍💻", want: "界\n\n👩‍💻", width: 2},
		{name: "ANSI-only middle row", text: "a\n\x1b[31m\x1b[0m\nb", want: "a\n\nb", width: 1},
	}

	for _, tc := range cases {
		for _, reverse := range []bool{false, true} {
			direction := "forward"
			if reverse {
				direction = "reverse"
			}
			t.Run(tc.name+"/"+direction, func(t *testing.T) {
				lines := strings.Split(tc.text, "\n")
				selectable := NewSelectable(&lineStructurePanel{content: tc.text})
				selectable.View(tc.width, len(lines))

				startX, startY := 0, 0
				endX, endY := tc.width-1, len(lines)-1
				if reverse {
					startX, startY, endX, endY = endX, endY, startX, startY
				}
				selectable.Update(tea.MouseMsg{X: startX, Y: startY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
				selectable.Update(tea.MouseMsg{X: endX, Y: endY, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
				selectable.Update(tea.MouseMsg{X: endX, Y: endY, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})

				if got := selectable.SelectedText(); got != tc.want {
					t.Fatalf("SelectedText()=%q, want %q", got, tc.want)
				}
			})
		}
	}
}

func TestSelectablePartialRowsPreserveBlankMiddle(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		direction := "forward"
		if reverse {
			direction = "reverse"
		}
		t.Run(direction, func(t *testing.T) {
			selectable := NewSelectable(&lineStructurePanel{content: "abc\n\nxyz"})
			selectable.View(3, 3)
			startX, startY := 1, 0
			endX, endY := 1, 2
			if reverse {
				startX, startY, endX, endY = endX, endY, startX, startY
			}
			selectable.Update(tea.MouseMsg{X: startX, Y: startY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
			selectable.Update(tea.MouseMsg{X: endX, Y: endY, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
			selectable.Update(tea.MouseMsg{X: endX, Y: endY, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
			if got := selectable.SelectedText(); got != "bc\n\nxy" {
				t.Fatalf("partial multiline SelectedText()=%q, want %q", got, "bc\n\nxy")
			}
		})
	}
}

func TestSelectableKeyboardSelectionPreservesBlankRows(t *testing.T) {
	selectable := NewSelectable(&lineStructurePanel{content: "a\n\nb"})
	selectable.View(1, 3)
	selectable.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	selectable.Update(tea.KeyMsg{Type: tea.KeyShiftDown})
	selectable.Update(tea.KeyMsg{Type: tea.KeyShiftRight})

	if got := selectable.SelectedText(); got != "a\n\nb" {
		t.Fatalf("SelectedText()=%q, want %q", got, "a\n\nb")
	}
}
