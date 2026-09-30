package integration_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type tabInteractionPanel struct {
	lastMouse tea.MouseMsg
	mouseSeen bool
}

func (p *tabInteractionPanel) View(_, _ int) string { return "content" }
func (p *tabInteractionPanel) Update(msg tea.Msg) tea.Cmd {
	if mouse, ok := msg.(tea.MouseMsg); ok {
		p.lastMouse = mouse
		p.mouseSeen = true
	}
	return nil
}
func (p *tabInteractionPanel) Elements(_, _ int) []warp.Element {
	return []warp.Element{{Role: "status", Name: "content", Bounds: warp.Bounds{X: 0, Y: 0, W: 1, H: 1}}}
}

func interactionMouse(x, y int) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress}
}

func interactionKey(key tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: key} }

func findInteractionElement(t *testing.T, elements []warp.Element, role, name, action string) warp.Element {
	t.Helper()
	el, ok := warp.FindElement(elements, role, name, action)
	if !ok {
		t.Fatalf("element %q/%q action %q missing: %+v", role, name, action, elements)
	}
	return el
}

func TestTabGroupKeyboardAndMouseInteractions(t *testing.T) {
	tg := warp.NewTabGroup(warp.TabTop)
	main := tg.ActiveTab()
	if main == nil {
		t.Fatal("NewTabGroup has no initial tab")
	}
	tg.Update(interactionKey(tea.KeyCtrlT))
	second := tg.ActiveTab()
	if second == main {
		t.Fatal("Ctrl+T did not create and activate a tab")
	}
	tg.Update(interactionKey(tea.KeyCtrlT))
	third := tg.ActiveTab()
	if third == second {
		t.Fatal("second Ctrl+T did not create a tab")
	}
	tg.NextTab()
	if tg.ActiveTab() != main {
		t.Fatal("NextTab did not cycle from the last tab to the first")
	}
	tg.PrevTab()
	if tg.ActiveTab() != third {
		t.Fatal("PrevTab did not cycle from the first tab to the last")
	}

	const width, height = 60, 12
	tg.View(width, height)
	elements := tg.Elements(width, height)
	mainTab := findInteractionElement(t, elements, "tab", "main", "activate-tab")
	x, y := mainTab.Bounds.Center()
	tg.Update(interactionMouse(x, y))
	if tg.ActiveTab() != main {
		t.Fatal("clicking an inactive tab did not activate it")
	}

	// Clicking the semantic new-tab target must create and activate a tab.
	tg.View(width, height)
	elements = tg.Elements(width, height)
	newButton := findInteractionElement(t, elements, "button", "New tab", "new-tab")
	x, y = newButton.Bounds.Center()
	tg.Update(interactionMouse(x, y))
	created := tg.ActiveTab()
	if created == main || created == second || created == third {
		t.Fatal("clicking + did not create a new tab")
	}

	// The close affordance is a child of the active tab element.
	tg.View(width, height)
	elements = tg.Elements(width, height)
	closeButton := findInteractionElement(t, elements, "button", "Close tab", "close-tab")
	closeX, closeY := closeButton.Bounds.Center()
	tg.Update(interactionMouse(closeX, closeY))
	if tg.ActiveTab() == created {
		t.Fatal("clicking the active tab's close button did not close it")
	}

	// Closing the remaining non-main tabs must leave the final tab intact.
	tg.Update(interactionKey(tea.KeyCtrlW))
	if tg.ActiveTab() == third {
		t.Fatal("Ctrl+W did not close the remaining active tab")
	}
	tg.Update(interactionKey(tea.KeyCtrlW))
	if tg.ActiveTab() != main {
		t.Fatal("Ctrl+W did not select the final remaining tab")
	}
	tg.Update(interactionKey(tea.KeyCtrlW))
	if tg.ActiveTab() != main {
		t.Fatal("Ctrl+W closed the last tab")
	}
}

func TestTabGroupTabBarPositionsAndContentMouseCoordinates(t *testing.T) {
	positions := []struct {
		name string
		pos  warp.TabPosition
	}{
		{name: "top", pos: warp.TabTop},
		{name: "bottom", pos: warp.TabBottom},
		{name: "left", pos: warp.TabLeft},
		{name: "right", pos: warp.TabRight},
	}
	for _, test := range positions {
		t.Run(test.name, func(t *testing.T) {
			const width, height = 40, 10
			tg := warp.NewTabGroup(test.pos)
			panel := &tabInteractionPanel{}
			tg.ActiveTab().SetRootPanel(panel)
			tg.View(width, height) // Establish dimensions and mouse tab-bar regions.
			elements := tg.Elements(width, height)
			tab := findInteractionElement(t, elements, "tab", "main", "activate-tab")
			if tab.Bounds.W == 0 || tab.Bounds.H == 0 {
				t.Fatalf("tab bar element has empty bounds: %+v", tab)
			}
			content := findInteractionElement(t, elements, "status", "content", "")
			clickX, clickY := content.Bounds.X, content.Bounds.Y
			if clickX < 0 || clickY < 0 || clickX >= width || clickY >= height {
				t.Fatalf("content test point out of bounds: %+v", content.Bounds)
			}
			tg.Update(interactionMouse(clickX, clickY))
			if !panel.mouseSeen {
				t.Fatalf("content click at (%d,%d) was not forwarded", clickX, clickY)
			}
			if panel.lastMouse.X != 0 || panel.lastMouse.Y != 0 {
				t.Fatalf("forwarded mouse coordinates=(%d,%d), want content-relative (0,0)", panel.lastMouse.X, panel.lastMouse.Y)
			}
		})
	}
}
