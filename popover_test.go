package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestPopoverRenderAndClick(t *testing.T) {
	itemClicked := ""
	closed := false

	popover := &Popover{
		Items: []PopoverItem{
			{Name: "Rename", Action: func() { itemClicked = "Rename" }},
			{Name: "Delete", Action: func() { itemClicked = "Delete" }},
		},
		X: 10,
		Y: 5,
		OnClose: func() { closed = true },
	}

	totalW, totalH := 80, 24
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", totalW)
	}
	lines = popover.Overlay(lines, totalW, totalH)

	t.Logf("Rendered popover:")
	for i, l := range lines {
		t.Logf("  %3d: %q (w=%d)", i, l, ansi.StringWidth(l))
	}
	t.Logf("boxW=%d boxH=%d", popover.boxW, popover.boxH)

	if popover.boxW == 0 {
		t.Fatal("boxW should be set after Overlay")
	}
	if popover.boxH == 0 {
		t.Fatal("boxH should be set after Overlay")
	}

	// The popover is at X=10, Y=5 (screen coords).
	// In lines coords: menuY = Y = 5 (header is in lines[0]).
	// box[0] = top border at line 5
	// box[1] = " Rename" at line 6
	// box[2] = " Delete" at line 7
	// box[3] = bottom border at line 8
	menuY := popover.Y // = 5
	menuX := popover.X // = 10

	t.Run("click first item", func(t *testing.T) {
		itemClicked = ""
		closed = false

		// Click on " Rename" (line 6 = menuY+1, col 11 = menuX+1)
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
}

func TestPopoverKeyboard(t *testing.T) {
	itemClicked := ""
	closed := false

	popover := &Popover{
		Items: []PopoverItem{
			{Name: "Rename", Action: func() { itemClicked = "Rename" }},
			{Name: "Delete", Action: func() { itemClicked = "Delete" }},
		},
		X: 10,
		Y: 5,
		OnClose: func() { closed = true },
	}

	// Must render first to set boxW.
	totalW, totalH := 80, 24
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = strings.Repeat(" ", totalW)
	}
	popover.Overlay(lines, totalW, totalH)

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

	// Re-render for next test.
	popover.Overlay(lines, totalW, totalH)

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

	// Re-render for next test.
	popover.Overlay(lines, totalW, totalH)

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
}

func TestPopoverEmpty(t *testing.T) {
	popover := &Popover{
		Items: []PopoverItem{},
		X:     10,
		Y:     5,
	}

	totalW, totalH := 80, 24
	lines := make([]string, totalH)
	for i := range lines {
		lines[i] = "test"
	}
	result := popover.Overlay(lines, totalW, totalH)
	if result[0] != "test" {
		t.Error("empty popover should not modify lines")
	}

	// HandleMouse on empty popover should be no-op.
	consumed := popover.HandleMouse(tea.MouseMsg{
		X:      0,
		Y:      0,
		Action: tea.MouseActionPress,
	})
	if consumed {
		t.Error("empty popover should not consume mouse")
	}
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

	// Content to the left of the popover should be preserved.
	// Popover at X=5, boxW ~22. Lines 0-2 and 7-9 should be untouched.
	// Lines 4-7 (popover rows) should have the popover at X=5.
	for i, l := range lines {
		vw := ansi.StringWidth(l)
		if vw != totalW {
			t.Errorf("line %d: expected visual width %d, got %d", i, totalW, vw)
		}
	}

	// Line 0 (before popover) should be unchanged.
	if lines[0] != "AAAAABBBBBCCCCCDDDDDEEEEEFFFFFGGGGGHHHHH" {
		t.Error("line 0 should be unchanged")
	}

	// Line 4 (top border of popover) should have the border at X=5.
	line4 := lines[4]
	left4 := ansi.Truncate(line4, 5, "")
	if ansi.StringWidth(left4) != 5 {
		t.Errorf("expected 5 chars left of popover on line 4, got %d", ansi.StringWidth(left4))
	}
}
