package integration_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type focusRoutingPanel struct {
	focusCount int
	blurCount  int
	keys       []tea.KeyMsg
	rawKeys    bool
}

func (*focusRoutingPanel) View(_, _ int) string { return "panel" }
func (p *focusRoutingPanel) Update(msg tea.Msg) tea.Cmd {
	if key, ok := msg.(tea.KeyMsg); ok {
		p.keys = append(p.keys, key)
	}
	return nil
}
func (p *focusRoutingPanel) Focus()             { p.focusCount++ }
func (p *focusRoutingPanel) Blur()              { p.blurCount++ }
func (p *focusRoutingPanel) Focused() bool      { return p.focusCount > p.blurCount }
func (p *focusRoutingPanel) WantsRawKeys() bool { return p.rawKeys }

func TestFocusTraversalIncludesLayoutAndFloatPanels(t *testing.T) {
	tab := warp.NewTab("focus")
	first := &focusRoutingPanel{}
	second := &focusRoutingPanel{}
	floating := &focusRoutingPanel{}
	tab.SetRootPanel(first)
	tab.SplitVertical(first, 0.5, second)
	tab.Float(floating, 2, 2, 20, 8)

	tab.FocusFirst()
	if tab.Focus() != first || first.focusCount != 1 {
		t.Fatalf("FocusFirst focused %T, want first layout panel", tab.Focus())
	}
	tab.FocusNext()
	if tab.Focus() != second || first.blurCount != 1 || second.focusCount != 1 {
		t.Fatal("FocusNext did not traverse the layout panels in order")
	}
	tab.FocusNext()
	if tab.Focus() != floating || second.blurCount != 1 || floating.focusCount != 1 {
		t.Fatal("FocusNext did not traverse from the layout to the float panel")
	}
	tab.FocusNext()
	if tab.Focus() != first || floating.blurCount != 1 {
		t.Fatal("focus traversal did not wrap to the first layout panel")
	}
	tab.FocusPrev()
	if tab.Focus() != floating {
		t.Fatal("FocusPrev did not wrap from the first layout panel to the last float")
	}
}

func TestTabSwitchSuspendsAndRestoresFocus(t *testing.T) {
	tg := warp.NewTabGroup(warp.TabNone)
	firstTab := tg.ActiveTab()
	firstPanel := &focusRoutingPanel{}
	firstTab.SetRootPanel(firstPanel)
	firstTab.FocusFirst()
	if firstPanel.focusCount != 1 {
		t.Fatal("initial tab panel was not focused")
	}

	secondTab := tg.NewTab("second")
	secondPanel := &focusRoutingPanel{}
	secondTab.SetRootPanel(secondPanel)
	secondTab.FocusFirst()
	if firstTab.Focus() != firstPanel || firstPanel.blurCount != 1 {
		t.Fatal("switching tabs did not suspend the old tab's focus")
	}
	if secondPanel.focusCount != 1 {
		t.Fatal("new active tab's focused panel was not focused")
	}

	tg.PrevTab()
	if tg.ActiveTab() != firstTab || firstPanel.focusCount != 2 {
		t.Fatal("switching back did not restore the suspended focus")
	}
	if secondPanel.blurCount != 1 {
		t.Fatal("switching away did not blur the second tab's focused panel")
	}
	tg.NextTab()
	if tg.ActiveTab() != secondTab || secondPanel.focusCount != 2 {
		t.Fatal("switching forward did not restore the second tab's focus")
	}
}

func TestRawKeyReceiverPrecedesTabGroupShortcuts(t *testing.T) {
	tg := warp.NewTabGroup(warp.TabNone)
	mainTab := tg.ActiveTab()
	receiver := &focusRoutingPanel{rawKeys: true}
	mainTab.SetRootPanel(receiver)
	mainTab.FocusFirst()
	otherTab := tg.NewTab("other")
	tg.PrevTab()
	if tg.ActiveTab() != mainTab {
		t.Fatal("failed to activate tab with raw key receiver")
	}

	tg.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if tg.ActiveTab() != mainTab || len(receiver.keys) != 1 || receiver.keys[0].Type != tea.KeyCtrlT {
		t.Fatal("raw receiver did not receive Ctrl+T before the TabGroup shortcut")
	}
	cmd := tg.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd != nil {
		t.Fatal("Ctrl+C was interpreted as quit despite the raw receiver requesting keys")
	}
	if len(receiver.keys) != 2 || receiver.keys[1].Type != tea.KeyCtrlC {
		t.Fatal("raw receiver did not receive Ctrl+C")
	}

	receiver.rawKeys = false
	tg.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if tg.ActiveTab() == otherTab {
		t.Fatal("TabGroup Ctrl+T shortcut did not create and activate a tab when raw keys were not requested")
	}
}
