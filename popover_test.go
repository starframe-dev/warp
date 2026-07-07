package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func blankLines(totalW, totalH int) []string {
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", totalW)
	}
	return lines
}

func TestPopoverOverlay(t *testing.T) {
	t.Run("renders and sets dimensions", func(t *testing.T) {
		popover := &Popover{
			Items: []PopoverItem{
				{Name: "Rename", Action: func() {}},
				{Name: "Delete", Action: func() {}},
			},
			X: 10, Y: 5,
		}
		lines := blankLines(80, 24)
		popover.Overlay(lines, 80, 24)
		if popover.boxW == 0 {
			t.Fatal("boxW should be set after Overlay")
		}
		if popover.boxH == 0 {
			t.Fatal("boxH should be set after Overlay")
		}
	})

	t.Run("empty items returns lines unchanged", func(t *testing.T) {
		popover := &Popover{Items: []PopoverItem{}, X: 10, Y: 5}
		lines := blankLines(80, 24)
		result := popover.Overlay(lines, 80, 24)
		if result[0] != strings.Repeat(" ", 80) {
			t.Error("empty popover should not modify lines")
		}
	})

	t.Run("auto width default", func(t *testing.T) {
		popover := &Popover{
			Items: []PopoverItem{{Name: "A", Action: func() {}}},
			X: 0, Y: 0, Width: 0,
		}
		lines := blankLines(80, 10)
		popover.Overlay(lines, 80, 10)
		if popover.boxW <= 0 || popover.boxH <= 0 {
			t.Fatal("expected dimensions to be set with default width")
		}
	})

	t.Run("width clamps to total width", func(t *testing.T) {
		popover := &Popover{
			Items: []PopoverItem{{Name: "Long item name", Action: func() {}}},
			X: 0, Y: 0, Width: 100,
		}
		lines := blankLines(30, 10)
		popover.Overlay(lines, 30, 10)
		if popover.boxW <= 0 || popover.boxH <= 0 {
			t.Fatal("expected dimensions to be set with clamped width")
		}
	})

	t.Run("position clamps to screen", func(t *testing.T) {
		popover := &Popover{
			Items: []PopoverItem{
				{Name: "One", Action: func() {}},
				{Name: "Two", Action: func() {}},
			},
			X: 70, Y: 20,
		}
		lines := blankLines(80, 24)
		result := popover.Overlay(lines, 80, 24)
		// All lines should still have total visual width.
		for i, l := range result {
			if vw := ansi.StringWidth(l); vw != 80 {
				t.Errorf("line %d: expected width 80, got %d", i, vw)
			}
		}
	})
}

func TestPopoverContentPreserved(t *testing.T) {
	popover := &Popover{
		Items: []PopoverItem{
			{Name: "Test Item", Action: func() {}},
		},
		X: 5,
		Y: 3,
	}

	totalW, totalH := 40, 10
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = "AAAAABBBBBCCCCCDDDDDEEEEEFFFFFGGGGGHHHHH"
	}
	lines = popover.Overlay(lines, totalW, totalH)

	for i, l := range lines {
		vw := ansi.StringWidth(l)
		if vw != totalW {
			t.Errorf("line %d: expected visual width %d, got %d", i, totalW, vw)
		}
	}

	// Line before popover should be unchanged.
	if lines[0] != "AAAAABBBBBCCCCCDDDDDEEEEEFFFFFGGGGGHHHHH" {
		t.Error("line 0 should be unchanged")
	}

	// Line inside popover should have content preserved to the left.
	line4 := lines[4]
	left4 := ansi.Truncate(line4, 5, "")
	if ansi.StringWidth(left4) != 5 {
		t.Errorf("expected 5 visual columns left of popover, got %d", ansi.StringWidth(left4))
	}
}

func TestVisualBytePos(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		pos := visualBytePos("hello world", 5)
		if pos != 6 {
			t.Errorf("expected pos 6, got %d", pos)
		}
	})

	t.Run("zero column", func(t *testing.T) {
		pos := visualBytePos("hello world", 0)
		if pos != 1 {
			t.Errorf("expected pos 1, got %d", pos)
		}
	})

	t.Run("beyond string length", func(t *testing.T) {
		pos := visualBytePos("hi", 100)
		if pos != 2 {
			t.Errorf("expected pos 2, got %d", pos)
		}
	})

	t.Run("contains ANSI codes", func(t *testing.T) {
		// ANSI escape code \x1b[31m is 4 chars but renders as 1 visual char
		redText := "\x1b[31mhello\x1b[0m"
		pos := visualBytePos(redText, 4)
		// Should skip ANSI codes, so pos 4 should be after 'hello'
		// The ANSI codes are between 'h' and 'e', so visual counting skips them
		if pos != 5 {
			t.Errorf("expected pos 5 (skipping ANSI), got %d", pos)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		pos := visualBytePos("", 5)
		if pos != 0 {
			t.Errorf("expected pos 0, got %d", pos)
		}
	})

	t.Run("single character", func(t *testing.T) {
		pos := visualBytePos("a", 0)
		if pos != 1 {
			t.Errorf("expected pos 1, got %d", pos)
		}
	})
}

func TestPopoverHandleMouse(t *testing.T) {
	itemClicked := ""
	closed := false

	popover := &Popover{
		Items: []PopoverItem{
			{Name: "Rename", Action: func() { itemClicked = "Rename" }},
			{Name: "Delete", Action: func() { itemClicked = "Delete" }},
		},
		X: 10, Y: 5,
		OnClose: func() { closed = true },
	}

	totalW, totalH := 80, 24
	popover.Overlay(blankLines(totalW, totalH), totalW, totalH)

	menuY := popover.Y
	menuX := popover.X

	t.Run("click first item", func(t *testing.T) {
		itemClicked = ""
		closed = false

		consumed := popover.HandleMouse(tea.MouseMsg{
			X:      menuX + 1,
			Y:      menuY + 1,
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
		})
		if !consumed {
			t.Error("click should be consumed")
		}
		if itemClicked != "Rename" {
			t.Errorf("expected Rename, got %q", itemClicked)
		}
		if !closed {
			t.Error("OnClose should be called")
		}
	})

	t.Run("click outside closes", func(t *testing.T) {
		closed = false

		consumed := popover.HandleMouse(tea.MouseMsg{
			X:      0,
			Y:      0,
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
		})
		if !consumed {
			t.Error("outside click should be consumed")
		}
		if !closed {
			t.Error("OnClose should be called on outside click")
		}
	})

	t.Run("motion updates selection", func(t *testing.T) {
		popover.selected = 0

		consumed := popover.HandleMouse(tea.MouseMsg{
			X:      menuX + 1,
			Y:      menuY + 2,
			Action: tea.MouseActionMotion,
		})
		if !consumed {
			t.Error("motion should be consumed")
		}
		if popover.selected != 1 {
			t.Errorf("expected selected=1, got %d", popover.selected)
		}
	})

	t.Run("release is consumed", func(t *testing.T) {
		consumed := popover.HandleMouse(tea.MouseMsg{
			X:      menuX + 1,
			Y:      menuY + 1,
			Action: tea.MouseActionRelease,
			Button: tea.MouseButtonLeft,
		})
		if !consumed {
			t.Error("release should be consumed")
		}
	})

	t.Run("click on border does not trigger item", func(t *testing.T) {
		itemClicked = ""
		closed = false

		consumed := popover.HandleMouse(tea.MouseMsg{
			X:      menuX + 1,
			Y:      menuY, // top border
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
		})
		if !consumed {
			t.Error("border click should be consumed")
		}
		if itemClicked != "" {
			t.Errorf("expected no item click, got %q", itemClicked)
		}
	})

	t.Run("not consumed before overlay", func(t *testing.T) {
		fresh := &Popover{
			Items: []PopoverItem{{Name: "A", Action: func() {}}},
			X: 10, Y: 5,
		}
		consumed := fresh.HandleMouse(tea.MouseMsg{
			X:      10,
			Y:      5,
			Action: tea.MouseActionPress,
			Button: tea.MouseButtonLeft,
		})
		if consumed {
			t.Error("mouse should not be consumed before Overlay")
		}
	})
}

func TestPopoverHandleKey(t *testing.T) {
	itemClicked := ""
	closed := false

	popover := &Popover{
		Items: []PopoverItem{
			{Name: "Rename", Action: func() { itemClicked = "Rename" }},
			{Name: "Delete", Action: func() { itemClicked = "Delete" }},
		},
		X: 10, Y: 5,
		OnClose: func() { closed = true },
	}

	totalW, totalH := 80, 24
	popover.Overlay(blankLines(totalW, totalH), totalW, totalH)

	t.Run("Esc closes", func(t *testing.T) {
		closed = false
		consumed := popover.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
		if !consumed {
			t.Error("Esc should be consumed")
		}
		if !closed {
			t.Error("OnClose should be called on Esc")
		}
	})

	// Re-render to reset box state after closing.
	popover.Overlay(blankLines(totalW, totalH), totalW, totalH)

	t.Run("Enter selects first item", func(t *testing.T) {
		itemClicked = ""
		closed = false
		popover.selected = 0

		consumed := popover.HandleKey(tea.KeyMsg{Type: tea.KeyEnter})
		if !consumed {
			t.Error("Enter should be consumed")
		}
		if itemClicked != "Rename" {
			t.Errorf("expected Rename, got %q", itemClicked)
		}
		if !closed {
			t.Error("OnClose should be called on Enter")
		}
	})

	popover.Overlay(blankLines(totalW, totalH), totalW, totalH)

	t.Run("Down arrow moves selection", func(t *testing.T) {
		popover.selected = 0

		consumed := popover.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
		if !consumed {
			t.Error("Down should be consumed")
		}
		if popover.selected != 1 {
			t.Errorf("expected selected=1, got %d", popover.selected)
		}
	})

	t.Run("Up arrow moves selection", func(t *testing.T) {
		popover.selected = 1

		consumed := popover.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
		if !consumed {
			t.Error("Up should be consumed")
		}
		if popover.selected != 0 {
			t.Errorf("expected selected=0, got %d", popover.selected)
		}
	})

	t.Run("Up at top stays at 0", func(t *testing.T) {
		popover.selected = 0

		popover.HandleKey(tea.KeyMsg{Type: tea.KeyUp})
		if popover.selected != 0 {
			t.Errorf("expected selected=0, got %d", popover.selected)
		}
	})

	t.Run("Down at bottom stays at last", func(t *testing.T) {
		popover.selected = 1

		popover.HandleKey(tea.KeyMsg{Type: tea.KeyDown})
		if popover.selected != 1 {
			t.Errorf("expected selected=1, got %d", popover.selected)
		}
	})

	t.Run("unknown key is not consumed", func(t *testing.T) {
		consumed := popover.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
		if consumed {
			t.Error("unknown key should not be consumed")
		}
	})

	t.Run("not consumed before overlay", func(t *testing.T) {
		fresh := &Popover{
			Items: []PopoverItem{{Name: "A", Action: func() {}}},
			X: 10, Y: 5,
		}
		consumed := fresh.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
		if consumed {
			t.Error("key should not be consumed before Overlay")
		}
	})
}

func TestPopoverOnCloseNil(t *testing.T) {
	popover := &Popover{
		Items: []PopoverItem{
			{Name: "A", Action: func() {}},
		},
		X: 10, Y: 5,
		OnClose: nil,
	}

	totalW, totalH := 80, 24
	popover.Overlay(blankLines(totalW, totalH), totalW, totalH)

	// These should not panic even when OnClose is nil.
	popover.HandleKey(tea.KeyMsg{Type: tea.KeyEsc})
	popover.HandleMouse(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress})
	popover.HandleMouse(tea.MouseMsg{X: 11, Y: 6, Action: tea.MouseActionPress})
}