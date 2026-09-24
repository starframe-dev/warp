package warp

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type rawTabGroupPanel struct {
	tgTestPanel
}

func (*rawTabGroupPanel) WantsRawKeys() bool { return true }

func TestRawKeyReceiverPrecedesTabShortcuts(t *testing.T) {
	keys := []tea.KeyMsg{
		{Type: tea.KeyCtrlC},
		{Type: tea.KeyCtrlT},
		{Type: tea.KeyCtrlW},
		{Type: tea.KeyRunes, Runes: []rune{'x'}},
	}
	for _, key := range keys {
		t.Run(key.String(), func(t *testing.T) {
			tg := NewTabGroup(TabTop)
			panel := &rawTabGroupPanel{}
			tg.ActiveTab().SetRootPanel(panel)
			tg.ActiveTab().setFocus(panel)

			cmd := tg.Update(key)
			if cmd != nil {
				t.Fatalf("raw receiver key %q returned a Warp command", key.String())
			}
			if !panel.updated {
				t.Fatalf("raw receiver did not receive key %q", key.String())
			}
			if got, ok := panel.lastMsg.(tea.KeyMsg); !ok || got.String() != key.String() {
				t.Fatalf("received message = %#v, want key %q", panel.lastMsg, key.String())
			}
			if len(tg.tabs) != 1 || tg.activeTab != 0 {
				t.Fatalf("raw key %q changed tabs: count=%d active=%d", key.String(), len(tg.tabs), tg.activeTab)
			}
		})
	}
}

func TestNormalFocusedPanelDoesNotSuppressCtrlC(t *testing.T) {
	tg := NewTabGroup(TabTop)
	panel := &tgTestPanel{}
	tg.ActiveTab().SetRootPanel(panel)
	tg.ActiveTab().setFocus(panel)

	cmd := tg.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("Ctrl+C should still quit when the focused panel does not request raw keys")
	}
	if panel.updated {
		t.Fatal("normal focused panel received Warp's Ctrl+C shortcut")
	}
}

func TestRightVerticalTabCloseUsesGlobalMouseX(t *testing.T) {
	tg := NewTabGroup(TabRight)
	tg.NewTab("second")
	tg.View(50, 5)
	region := tg.tabRegions[tg.activeTab]
	if region.closeX < 0 {
		t.Fatal("active right-side tab has no close cell")
	}
	globalX := tg.width - tg.verticalTabWidth + region.closeX
	tg.Update(tea.MouseMsg{
		X:      globalX,
		Y:      tg.activeTab,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	if len(tg.tabs) != 1 || tg.ActiveTab().name != "main" {
		t.Fatalf("right-side close left tabs=%v active=%v", tabNames(tg.tabs), tg.ActiveTab())
	}
}

func TestVerticalTabGeometryIsReadyOnFirstRender(t *testing.T) {
	tg := NewTabGroup(TabLeft)
	panel := &geometryTestPanel{name: "P"}
	tg.ActiveTab().SetRootPanel(panel)
	view := tg.View(40, 6)
	row := ansi.Strip(strings.Split(view, "\n")[0])
	marker := strings.Index(row, panel.name)
	if marker < 0 {
		t.Fatalf("rendered row %q does not contain content panel", row)
	}
	if got := ansi.StringWidth(row[:marker]); got != tg.verticalTabWidth {
		t.Fatalf("first rendered content offset=%d, vertical width=%d", got, tg.verticalTabWidth)
	}
	elements := tg.Elements(40, 6)
	if len(elements) != 1 || elements[0].Bounds.X != tg.verticalTabWidth {
		t.Fatalf("elements=%+v, vertical width=%d", elements, tg.verticalTabWidth)
	}
}

func TestClosingTabBeforeActivePreservesLogicalTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.NewTab("b")
	tg.NewTab("c")
	tg.switchTab(2)
	active := tg.ActiveTab()
	tg.closeTab(0)
	if tg.activeTab != 1 || tg.ActiveTab() != active {
		t.Fatalf("closing preceding tab selected index=%d tab=%v, want original active tab at index 1", tg.activeTab, tg.ActiveTab())
	}
}

func TestTabGroupTinyDimensionsStayNonNegative(t *testing.T) {
	for _, position := range []TabPosition{TabTop, TabBottom, TabLeft, TabRight, TabNone} {
		t.Run(fmt.Sprintf("position-%d", position), func(t *testing.T) {
			tg := NewTabGroup(position)
			for width := 0; width <= 3; width++ {
				for height := 0; height <= 3; height++ {
					tg.View(width, height)
					if got := tg.contentWidth(width); got < 0 {
						t.Fatalf("width=%d content width=%d", width, got)
					}
					if got := tg.contentHeight(height); got < 0 {
						t.Fatalf("height=%d content height=%d", height, got)
					}
					tg.Elements(width, height)
				}
			}
		})
	}
}

func tabNames(tabs []*Tab) []string {
	names := make([]string, len(tabs))
	for i, tab := range tabs {
		names[i] = tab.name
	}
	return names
}
