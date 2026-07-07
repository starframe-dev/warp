package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewDropdownMenu(t *testing.T) {
	items := []DropdownItem{{Label: "A"}, {Label: "B"}}
	d := NewDropdownMenu("Label", items)

	if d == nil {
		t.Fatal("NewDropdownMenu returned nil")
	}
	if d.Label != "Label" {
		t.Errorf("Label = %q, want %q", d.Label, "Label")
	}
	if len(d.Items) != 2 {
		t.Errorf("Items length = %d, want 2", len(d.Items))
	}
	if d.Open {
		t.Error("new dropdown should be closed")
	}
	if d.Hovered != -1 {
		t.Errorf("Hovered = %d, want -1", d.Hovered)
	}
}

func TestDropdownMenu_ViewClosed(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{{Label: "A"}})
	view := d.View(20, 5)
	if view == "" {
		t.Error("View returned empty string")
	}
	if strings.Contains(view, "▲") {
		t.Error("closed dropdown should not show open arrow")
	}
	if !strings.Contains(view, "▼") {
		t.Error("closed dropdown should show down arrow")
	}
}

func TestDropdownMenu_ViewOpen(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One"},
		{Label: "Two"},
	})
	d.Open = true
	view := d.View(20, 10)
	if view == "" {
		t.Fatal("View returned empty string")
	}
	if !strings.Contains(view, "▲") {
		t.Error("open dropdown should show up arrow")
	}
	if !strings.Contains(view, "One") {
		t.Error("open dropdown should show first item")
	}
	if !strings.Contains(view, "Two") {
		t.Error("open dropdown should show second item")
	}
}

func TestDropdownMenu_ViewOpenWithSelectionAndHover(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One", Selected: true},
		{Label: "Two"},
	})
	d.Open = true
	d.Hovered = 1
	view := d.View(20, 10)
	if !strings.Contains(view, "One") {
		t.Error("view should show selected item")
	}
	if !strings.Contains(view, "Two") {
		t.Error("view should show hovered item")
	}
}

func TestDropdownMenu_ViewTruncation(t *testing.T) {
	d := NewDropdownMenu("VeryLongLabel", []DropdownItem{})
	view := d.View(5, 5)
	if !strings.Contains(view, "…") {
		t.Error("long label should be truncated with ellipsis")
	}
}

func TestDropdownMenu_UpdateMouseOpenAndClose(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One"},
		{Label: "Two"},
	})

	// Click closed button at Y=0 opens menu
	d.Update(tea.MouseMsg{Action: tea.MouseActionPress, Y: 0})
	if !d.Open {
		t.Error("click on closed button should open menu")
	}

	// Click open button at Y=0 closes menu
	d.Update(tea.MouseMsg{Action: tea.MouseActionPress, Y: 0})
	if d.Open {
		t.Error("click on open button should close menu")
	}
}

func TestDropdownMenu_UpdateMouseSelect(t *testing.T) {
	selected := -1
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One"},
		{Label: "Two"},
	})
	d.OnSelect = func(idx int) { selected = idx }
	d.Open = true

	d.Update(tea.MouseMsg{Action: tea.MouseActionPress, Y: 2})
	if selected != 1 {
		t.Errorf("OnSelect called with %d, want 1", selected)
	}
	if d.Open {
		t.Error("selecting with mouse should close menu")
	}
	if !d.Items[1].Selected {
		t.Error("selected item should be marked selected")
	}
	if d.Items[0].Selected {
		t.Error("non-selected item should not be marked selected")
	}
}

func TestDropdownMenu_UpdateMouseOutOfBounds(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One"},
	})
	d.Open = true

	// Click outside item range should not panic
	d.Update(tea.MouseMsg{Action: tea.MouseActionPress, Y: 5})
	if !d.Open {
		t.Error("click out of range should not close menu")
	}

	// Non-press mouse actions should be ignored
	d.Update(tea.MouseMsg{Action: tea.MouseActionRelease, Y: 1})
	if !d.Open {
		t.Error("non-press action should not close menu")
	}
}

func TestDropdownMenu_UpdateKeyboard(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{
		{Label: "One"},
		{Label: "Two"},
		{Label: "Three"},
	})
	d.Open = true

	// down should move hover to 0
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.Hovered != 0 {
		t.Errorf("after down Hovered = %d, want 0", d.Hovered)
	}

	// down again should move to 1
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.Hovered != 1 {
		t.Errorf("after second down Hovered = %d, want 1", d.Hovered)
	}

	// up should move back to 0
	d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.Hovered != 0 {
		t.Errorf("after up Hovered = %d, want 0", d.Hovered)
	}

	// up at 0 should stay 0
	d.Update(tea.KeyMsg{Type: tea.KeyUp})
	if d.Hovered != 0 {
		t.Errorf("up at first item Hovered = %d, want 0", d.Hovered)
	}

	// enter should select hovered item
	d.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if d.Open {
		t.Error("enter should close menu")
	}
	if !d.Items[0].Selected {
		t.Error("enter should select hovered item")
	}
}

func TestDropdownMenu_UpdateEsc(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{{Label: "One"}})
	d.Open = true

	d.Update(tea.KeyMsg{Type: tea.KeyEscape})
	if d.Open {
		t.Error("esc should close menu")
	}
}

func TestDropdownMenu_UpdateKeyboardWhenClosed(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{{Label: "One"}})

	// Key presses when closed should be ignored and not panic
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.Open {
		t.Error("closed dropdown should stay closed on key down")
	}
	if d.Hovered != -1 {
		t.Errorf("Hoved = %d, want -1", d.Hovered)
	}
}

func TestDropdownMenu_Close(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{{Label: "One"}})
	d.Open = true
	d.Close()
	if d.Open {
		t.Error("Close should set Open to false")
	}
}

func TestDropdownMenu_KeyboardBounds(t *testing.T) {
	d := NewDropdownMenu("Choose", []DropdownItem{{Label: "One"}})
	d.Open = true

	// Move down to last item
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.Hovered != 0 {
		t.Errorf("Hovered = %d, want 0", d.Hovered)
	}

	// Down again should not exceed last index
	d.Update(tea.KeyMsg{Type: tea.KeyDown})
	if d.Hovered != 0 {
		t.Errorf("down at last item Hovered = %d, want 0", d.Hovered)
	}
}
