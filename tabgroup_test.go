package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type tgTestPanel struct {
	BasePanel
	updated bool
	lastMsg tea.Msg
}

func (p *tgTestPanel) Update(msg tea.Msg) tea.Cmd {
	p.updated = true
	p.lastMsg = msg
	return nil
}

type tgTestElementPanel struct {
	BasePanel
}

func (tgTestElementPanel) Elements(w, h int) []Element {
	return []Element{
		{
			Role:   "test",
			Bounds: Bounds{X: 0, Y: 0, W: w, H: h},
			Children: []Element{
				{Bounds: Bounds{X: 0, Y: 0, W: 1, H: 1}},
			},
		},
	}
}

type customMsg struct{ v int }

func TestNewTabGroup(t *testing.T) {
	tg := NewTabGroup(TabTop)
	if tg == nil {
		t.Fatal("NewTabGroup returned nil")
	}
	if len(tg.tabs) != 1 {
		t.Fatalf("expected 1 tab, got %d", len(tg.tabs))
	}
	if got := tg.ActiveTab(); got == nil || got.name != "main" {
		t.Fatalf("expected active tab named main, got %v", got)
	}
}

func TestTabGroupNewTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tab := tg.NewTab("foo")
	if tab == nil {
		t.Fatal("NewTab returned nil")
	}
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tg.tabs))
	}
	if tg.activeTab != 1 {
		t.Fatalf("expected activeTab=1, got %d", tg.activeTab)
	}
	if got := tg.ActiveTab(); got != tab {
		t.Fatal("ActiveTab did not return the newly created tab")
	}
}

func TestActiveTabBounds(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.activeTab = -1
	if got := tg.ActiveTab(); got != nil {
		t.Fatalf("expected nil for negative activeTab, got %v", got)
	}
	tg.activeTab = 10
	if got := tg.ActiveTab(); got != nil {
		t.Fatalf("expected nil for out-of-range activeTab, got %v", got)
	}
}

func TestNextTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.NewTab("b")
	tg.activeTab = 0
	tg.NextTab()
	if tg.activeTab != 1 {
		t.Fatalf("expected activeTab=1, got %d", tg.activeTab)
	}
	tg.activeTab = 2
	tg.NextTab()
	if tg.activeTab != 0 {
		t.Fatalf("expected wrap to 0, got %d", tg.activeTab)
	}

	single := NewTabGroup(TabTop)
	single.NextTab()
	if single.activeTab != 0 {
		t.Fatalf("expected single tab to stay 0, got %d", single.activeTab)
	}
}

func TestPrevTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.activeTab = 0
	tg.PrevTab()
	if tg.activeTab != 1 {
		t.Fatalf("expected wrap to last, got %d", tg.activeTab)
	}
	tg.activeTab = 1
	tg.PrevTab()
	if tg.activeTab != 0 {
		t.Fatalf("expected activeTab=0, got %d", tg.activeTab)
	}

	single := NewTabGroup(TabTop)
	single.PrevTab()
	if single.activeTab != 0 {
		t.Fatalf("expected single tab to stay 0, got %d", single.activeTab)
	}
}

func TestTabGroupView(t *testing.T) {
	positions := []TabPosition{TabTop, TabBottom, TabLeft, TabRight, TabNone}
	for _, pos := range positions {
		tg := NewTabGroup(pos)
		out := tg.View(30, 5)
		if out == "" {
			t.Fatalf("expected non-empty view for position %d", pos)
		}
	}
}

func TestTabGroupViewEmptyTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.activeTab = -1
	out := tg.View(10, 3)
	if out == "" {
		t.Fatal("expected blank lines for missing active tab")
	}
}

func TestTabGroupElements(t *testing.T) {
	positions := []TabPosition{TabTop, TabLeft}
	for _, pos := range positions {
		tg := NewTabGroup(pos)
		_ = tg.View(30, 6)
		mock := tgTestElementPanel{}
		tg.ActiveTab().SetRootPanel(mock)
		elems := tg.Elements(30, 6)
		if len(elems) != 1 {
			t.Fatalf("position %d: expected 1 element, got %d", pos, len(elems))
		}
		if elems[0].Bounds.X == 0 && elems[0].Bounds.Y == 0 {
			t.Fatalf("position %d: expected element to be offset", pos)
		}
		if len(elems[0].Children) != 1 {
			t.Fatalf("position %d: expected 1 child, got %d", pos, len(elems[0].Children))
		}
		if elems[0].Children[0].Bounds.X != elems[0].Bounds.X || elems[0].Children[0].Bounds.Y != elems[0].Bounds.Y {
			t.Fatalf("position %d: child not shifted with parent", pos)
		}
	}
}

func TestTabGroupUpdateKeyCtrlC(t *testing.T) {
	tg := NewTabGroup(TabTop)
	cmd := tg.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("expected quit command for ctrl+c")
	}
}

func TestTabGroupUpdateKeyForward(t *testing.T) {
	tg := NewTabGroup(TabTop)
	p := &tgTestPanel{}
	tg.ActiveTab().focused = p
	tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !p.updated {
		t.Fatal("expected key message to be forwarded to focused panel")
	}
	if _, ok := p.lastMsg.(tea.KeyMsg); !ok {
		t.Fatal("expected forwarded message to be a KeyMsg")
	}
}

func TestTabGroupUpdateCtrlT(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.Update(tea.KeyMsg{Type: tea.KeyCtrlT})
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs after ctrl+t, got %d", len(tg.tabs))
	}
}

func TestTabGroupUpdateCtrlW(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs, got %d", len(tg.tabs))
	}
	tg.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	if len(tg.tabs) != 1 {
		t.Fatalf("expected 1 tab after ctrl+w, got %d", len(tg.tabs))
	}
}

func TestTabGroupUpdateMouseTabBarClick(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("second")
	tg.NewTab("third")
	_ = tg.View(80, 5)
	tg.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if tg.activeTab != 0 {
		t.Fatalf("expected click on first tab to switch active, got %d", tg.activeTab)
	}
}

func TestTabGroupUpdateMouseNewTabClick(t *testing.T) {
	tg := NewTabGroup(TabTop)
	_ = tg.View(80, 5)
	if tg.newTabRegion == nil {
		t.Fatal("newTabRegion was not set")
	}
	tg.Update(tea.MouseMsg{X: tg.newTabRegion.startX, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if len(tg.tabs) != 2 {
		t.Fatalf("expected new tab to be created, got %d tabs", len(tg.tabs))
	}
}

func TestTabGroupUpdateMouseVerticalTabBar(t *testing.T) {
	tg := NewTabGroup(TabLeft)
	tg.NewTab("a")
	_ = tg.View(30, 6)
	if len(tg.tabRegions) < 2 {
		t.Fatalf("expected at least 2 vertical tab regions, got %d", len(tg.tabRegions))
	}
	tg.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if tg.activeTab != 0 {
		t.Fatalf("expected click on first vertical tab to switch active, got %d", tg.activeTab)
	}
}

func TestTabGroupUpdateMouseContent(t *testing.T) {
	tg := NewTabGroup(TabTop)
	p := &tgTestPanel{}
	tg.ActiveTab().SetRootPanel(p)
	_ = tg.View(30, 6)
	tg.Update(tea.MouseMsg{X: 0, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if !p.updated {
		t.Fatal("expected mouse press to be forwarded to content panel")
	}
}

func TestTabGroupUpdateWindowSize(t *testing.T) {
	tg := NewTabGroup(TabTop)
	_ = tg.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	if tg.width != 100 || tg.height != 40 {
		t.Fatalf("expected width=100 height=40, got width=%d height=%d", tg.width, tg.height)
	}
}

func TestTabGroupUpdateResizeMsg(t *testing.T) {
	tg := NewTabGroup(TabTop)
	_ = tg.Update(ResizeMsg{Width: 80, Height: 24})
	if tg.width != 80 || tg.height != 24 {
		t.Fatalf("expected width=80 height=24, got width=%d height=%d", tg.width, tg.height)
	}
}

func TestTabGroupUpdateUnknownMsg(t *testing.T) {
	tg := NewTabGroup(TabTop)
	p := &tgTestPanel{}
	tg.ActiveTab().SetRootPanel(p)
	tg.Update(customMsg{42})
	if !p.updated {
		t.Fatal("expected unknown message to be broadcast to root panel")
	}
	if _, ok := p.lastMsg.(customMsg); !ok {
		t.Fatal("expected broadcast message to be customMsg")
	}
}
