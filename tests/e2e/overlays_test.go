package e2e_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

func TestModalAndPopoverOverlayInteractions(t *testing.T) {
	const width, height = 80, 24
	background := make([]string, height)
	for i := range background {
		background[i] = strings.Repeat("existing content ", 5)
	}

	modalButtons := 0
	modalClosed := 0
	modal := warp.NewModal("Confirm", "Apply these changes?", []warp.ModalButton{
		{Label: "Apply", Action: func() { modalButtons++ }},
	}, func() { modalClosed++ })
	modalView := modal.Overlay(append([]string(nil), background...), width, height)
	if len(modalView) != height || !strings.Contains(warp.StripANSI(strings.Join(modalView, "\n")), "Confirm") {
		t.Fatal("modal was not rendered over the existing content")
	}
	if !strings.Contains(warp.StripANSI(strings.Join(modalView, "\n")), "existing content") {
		t.Fatal("modal overlay did not preserve surrounding content")
	}

	// Activate the modal button before dragging the dialog by its top padding.
	buttonX := modal.StartX() + 4 // left border/padding followed by the opening bracket
	buttonY := modal.StartY() + 4
	if !modal.HandleMouse(tea.MouseMsg{X: buttonX, Y: buttonY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}) || modalButtons != 1 {
		t.Fatalf("modal button activation: consumed/action count = %v/%d", modalButtons == 1, modalButtons)
	}
	startX, startY := modal.StartX(), modal.StartY()
	dragX, dragY := startX+8, startY+1
	if !modal.HandleMouse(tea.MouseMsg{X: dragX, Y: dragY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}) {
		t.Fatal("modal drag did not start from the top padding")
	}
	modal.HandleMouse(tea.MouseMsg{X: dragX + 5, Y: dragY + 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	modal.HandleMouse(tea.MouseMsg{X: dragX + 5, Y: dragY + 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	if modal.StartX() != startX+5 || modal.StartY() != startY+2 {
		t.Fatalf("modal position after drag = (%d,%d), want (%d,%d)", modal.StartX(), modal.StartY(), startX+5, startY+2)
	}
	modal.Overlay(append([]string(nil), background...), width, height)
	closeX, closeY := modal.StartX()+modal.BoxWidth()-4, modal.StartY()+2
	if !modal.HandleMouse(tea.MouseMsg{X: closeX, Y: closeY, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}) || modalClosed != 1 {
		t.Fatalf("modal close interaction: callback count = %d", modalClosed)
	}

	menuActions, menuClosed := 0, 0
	popover := &warp.Popover{
		X: 12, Y: 4, Width: 18,
		Items: []warp.PopoverItem{
			{Name: "Open item", Action: func() { menuActions++ }},
			{Name: "Other action"},
		},
		OnClose: func() { menuClosed++ },
	}
	popoverView := popover.Overlay(append([]string(nil), background...), width, height)
	if len(popoverView) != height || !strings.Contains(warp.StripANSI(strings.Join(popoverView, "\n")), "Open item") || !strings.Contains(warp.StripANSI(strings.Join(popoverView, "\n")), "existing content") {
		t.Fatal("popover was not rendered over and alongside existing content")
	}
	if !popover.HandleMouse(tea.MouseMsg{X: popover.X + 2, Y: popover.Y + 1, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}) || menuActions != 1 || menuClosed != 1 {
		t.Fatalf("popover menu action/close callbacks = %d/%d", menuActions, menuClosed)
	}

	// Reopen the menu and dismiss it explicitly with Escape.
	popover.OnClose = func() { menuClosed++ }
	popover.Overlay(append([]string(nil), background...), width, height)
	if !popover.HandleKey(tea.KeyMsg{Type: tea.KeyEsc}) || menuClosed != 2 {
		t.Fatalf("popover Escape dismissal callback count = %d", menuClosed)
	}
}
