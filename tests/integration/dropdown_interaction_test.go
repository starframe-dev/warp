package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func dropdownMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func dropdownKey(key tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: key} }

func TestDropdownMouseAndKeyboardSelectionExposeSemantics(t *testing.T) {
	menu := warp.NewDropdownMenu("Colors", []warp.DropdownItem{
		{Label: "Red"},
		{Label: "Green"},
		{Label: "Blue"},
	})
	var selected []int
	menu.OnSelect = func(idx int) { selected = append(selected, idx) }

	closed := menu.Elements(20, 5)
	if len(closed) != 1 || closed[0].Role != "combobox" || closed[0].Name != "Colors" || closed[0].Action != "toggle" || closed[0].Bounds != (warp.Bounds{W: 20, H: 1}) {
		t.Fatalf("closed dropdown semantics = %+v", closed)
	}

	menu.Update(dropdownMouse(2, 0, tea.MouseActionPress))
	if !menu.Open {
		t.Fatal("mouse press on the button did not open dropdown")
	}
	view := warp.StripANSI(menu.View(20, 4))
	for _, label := range []string{"Colors ▲", "Red", "Green", "Blue"} {
		if !strings.Contains(view, label) {
			t.Errorf("open view %q does not contain %q", view, label)
		}
	}
	elements := menu.Elements(20, 4)
	if len(elements) != 4 {
		t.Fatalf("open semantic elements = %+v, want combobox and three options", elements)
	}
	for i, label := range []string{"Red", "Green", "Blue"} {
		option := elements[i+1]
		if option.Role != "option" || option.Name != label || option.Action != "select" || option.Bounds != (warp.Bounds{Y: i + 1, W: 20, H: 1}) {
			t.Errorf("option %d semantics = %+v", i, option)
		}
	}

	// Keyboard navigation operates on the open menu; selection updates the
	// selected item, closes the menu, and reports its index exactly once.
	menu.Update(dropdownKey(tea.KeyDown))
	menu.Update(dropdownKey(tea.KeyDown))
	if menu.Hovered != 1 {
		t.Fatalf("keyboard navigation hovered index %d, want 1", menu.Hovered)
	}
	menu.Update(dropdownKey(tea.KeyEnter))
	if menu.Open || !menu.Items[1].Selected || menu.Items[0].Selected || menu.Items[2].Selected {
		t.Fatalf("keyboard selection state: open=%v items=%+v", menu.Open, menu.Items)
	}
	if len(selected) != 1 || selected[0] != 1 {
		t.Fatalf("selection callbacks = %v, want [1]", selected)
	}

	// Reopen by mouse and select the visible third item by its row.
	menu.Update(dropdownMouse(2, 0, tea.MouseActionPress))
	menu.View(20, 4)
	menu.Update(dropdownMouse(2, 3, tea.MouseActionPress))
	if menu.Open || !menu.Items[2].Selected || menu.Items[1].Selected {
		t.Fatalf("mouse selection state: open=%v items=%+v", menu.Open, menu.Items)
	}
	if len(selected) != 2 || selected[1] != 2 {
		t.Fatalf("selection callbacks = %v, want [1 2]", selected)
	}
}

func TestDropdownClipsOptionsToViewportHeight(t *testing.T) {
	menu := warp.NewDropdownMenu("Choices", []warp.DropdownItem{
		{Label: "First"}, {Label: "Second"}, {Label: "Third"}, {Label: "Fourth"},
	})
	callbackCount := 0
	menu.OnSelect = func(int) { callbackCount++ }
	menu.Update(dropdownMouse(0, 0, tea.MouseActionPress))

	const width, height = 18, 3 // button and only two option rows fit
	view := warp.StripANSI(menu.View(width, height))
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("clipped view has %d lines, want %d: %q", len(lines), height, view)
	}
	if !strings.Contains(view, "First") || !strings.Contains(view, "Second") || strings.Contains(view, "Third") || strings.Contains(view, "Fourth") {
		t.Fatalf("menu did not render only viewport-visible options: %q", view)
	}
	elements := menu.Elements(width, height)
	if len(elements) != 3 {
		t.Fatalf("clipped semantic elements = %+v, want button and two visible options", elements)
	}
	for _, name := range []string{"First", "Second"} {
		if _, ok := warp.FindElement(elements, "option", name, "select"); !ok {
			t.Errorf("visible option %q missing from semantic tree %+v", name, elements)
		}
	}
	for _, name := range []string{"Third", "Fourth"} {
		if _, ok := warp.FindElement(elements, "option", name, "select"); ok {
			t.Errorf("clipped option %q unexpectedly exposed: %+v", name, elements)
		}
	}

	// Neither a mouse press on a clipped row nor keyboard navigation can
	// reach an option beyond the rendered viewport.
	menu.Update(dropdownMouse(0, 3, tea.MouseActionPress))
	if !menu.Open || callbackCount != 0 {
		t.Fatalf("clipped mouse row changed menu: open=%v callbacks=%d", menu.Open, callbackCount)
	}
	menu.Update(dropdownKey(tea.KeyDown))
	menu.Update(dropdownKey(tea.KeyDown))
	menu.Update(dropdownKey(tea.KeyDown))
	if menu.Hovered != 1 {
		t.Fatalf("keyboard navigation escaped visible options: hovered=%d", menu.Hovered)
	}
	menu.Update(dropdownKey(tea.KeyEnter))
	if menu.Open || !menu.Items[1].Selected || menu.Items[2].Selected || callbackCount != 1 {
		t.Fatalf("clipped keyboard selection state: open=%v items=%+v callbacks=%d", menu.Open, menu.Items, callbackCount)
	}
}
