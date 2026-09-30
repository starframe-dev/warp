package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func TestDropdownKeyboardMouseHoverSelectionAndDismissal(t *testing.T) {
	const width, height = 24, 5
	menu := warp.NewDropdownMenu("Color", []warp.DropdownItem{
		{Label: "Red", Selected: true},
		{Label: "Green"},
		{Label: "Blue"},
	})
	var selected []int
	menu.OnSelect = func(index int) { selected = append(selected, index) }

	// Open the button and render the menu so its visible rows are established.
	menu.Update(tea.MouseMsg{X: 2, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	view := warp.StripANSI(menu.View(width, height))
	for _, label := range []string{"Color ▲", "Red", "Green", "Blue"} {
		if !strings.Contains(view, label) {
			t.Fatalf("open dropdown view %q does not contain %q", view, label)
		}
	}
	options := menu.Elements(width, height)
	if len(options) != 4 || options[1].Role != "option" || options[1].Name != "Red" {
		t.Fatalf("open dropdown semantic options = %+v", options)
	}

	// Arrow navigation moves the highlight without changing the current choice;
	// Enter selects Green.
	menu.Update(tea.KeyMsg{Type: tea.KeyDown})
	menu.Update(tea.KeyMsg{Type: tea.KeyDown})
	if menu.Hovered != 1 || !menu.Items[0].Selected {
		t.Fatalf("keyboard navigation changed selection or hovered wrong row: hovered=%d items=%+v", menu.Hovered, menu.Items)
	}
	menu.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if menu.Open || !menu.Items[1].Selected || menu.Items[0].Selected || menu.Items[2].Selected {
		t.Fatalf("keyboard selection state: open=%v items=%+v", menu.Open, menu.Items)
	}
	if len(selected) != 1 || selected[0] != 1 {
		t.Fatalf("selection callbacks = %v, want [1]", selected)
	}

	// Hovering Blue is non-selecting. Escape dismisses the open menu while
	// preserving Green as the selected option.
	menu.Update(tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	menu.View(width, height)
	menu.Update(tea.MouseMsg{X: 1, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	if menu.Hovered != 2 || !menu.Items[1].Selected || menu.Items[2].Selected {
		t.Fatalf("hover changed selection or hovered wrong option: hovered=%d items=%+v", menu.Hovered, menu.Items)
	}
	menu.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if menu.Open || !menu.Items[1].Selected || menu.Items[2].Selected || len(selected) != 1 {
		t.Fatalf("Escape changed dropdown selection: open=%v items=%+v callbacks=%v", menu.Open, menu.Items, selected)
	}

	// Reopen and choose Blue with the mouse; exactly one callback is added.
	menu.Update(tea.MouseMsg{X: 1, Y: 0, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	menu.View(width, height)
	menu.Update(tea.MouseMsg{X: 1, Y: 3, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	if menu.Open || !menu.Items[2].Selected || menu.Items[0].Selected || menu.Items[1].Selected {
		t.Fatalf("mouse selection state: open=%v items=%+v", menu.Open, menu.Items)
	}
	if len(selected) != 2 || selected[1] != 2 {
		t.Fatalf("selection callbacks = %v, want [1 2]", selected)
	}
}
