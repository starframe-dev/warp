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

func TestTabGroupElementsNoTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.activeTab = -1
	if got := tg.Elements(10, 3); got != nil {
		t.Fatalf("expected nil elements for missing active tab, got %v", got)
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

func TestTabGroupUpdateKeyNoActiveTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.activeTab = -1
	cmd := tg.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if cmd != nil {
		t.Fatal("expected nil when no active tab to forward keys to")
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

func TestTabGroupUpdateMouseCloseTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.NewTab("b")
	_ = tg.View(80, 5)
	closeX := tg.tabRegions[tg.activeTab].closeX
	if closeX < 0 {
		t.Fatal("expected close button X to be set for active tab")
	}
	tg.Update(tea.MouseMsg{X: closeX, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs after closing, got %d", len(tg.tabs))
	}
	if tg.activeTab != 1 {
		t.Fatalf("expected activeTab=1 after close, got %d", tg.activeTab)
	}
}

func TestTabGroupUpdateMouseCloseVerticalTab(t *testing.T) {
	tg := NewTabGroup(TabLeft)
	tg.NewTab("a")
	tg.NewTab("b")
	_ = tg.View(80, 5)
	if tg.activeTab < 0 || tg.activeTab >= len(tg.tabRegions) {
		t.Fatal("invalid active tab index")
	}
	r := tg.tabRegions[tg.activeTab]
	if r.closeX < 0 {
		t.Fatal("expected close X for vertical active tab")
	}
	tg.Update(tea.MouseMsg{X: r.closeX, Y: tg.activeTab, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs after close, got %d", len(tg.tabs))
	}
}

func TestTabGroupUpdateMouseNotOnTabBar(t *testing.T) {
	tg := NewTabGroup(TabNone)
	p := &tgTestPanel{}
	tg.ActiveTab().SetRootPanel(p)
	_ = tg.View(30, 6)
	tg.Update(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
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

func TestContentWidth(t *testing.T) {
	cases := []struct {
		pos   TabPosition
		v     TabGroup
		wantW int
	}{
		{TabTop, TabGroup{verticalTabWidth: 4}, 40},
		{TabLeft, TabGroup{verticalTabWidth: 4}, 36},
		{TabRight, TabGroup{verticalTabWidth: 4}, 36},
		{TabNone, TabGroup{}, 40},
	}
	for _, c := range cases {
		tg := c.v
		tg.tabPosition = c.pos
		if got := tg.contentWidth(40); got != c.wantW {
			t.Errorf("position %d: contentWidth(40)=%d, want %d", c.pos, got, c.wantW)
		}
	}
}

func TestContentHeight(t *testing.T) {
	cases := []struct {
		pos   TabPosition
		wantH int
	}{
		{TabTop, 19},
		{TabBottom, 19},
		{TabNone, 20},
	}
	for _, c := range cases {
		tg := &TabGroup{tabPosition: c.pos}
		if got := tg.contentHeight(20); got != c.wantH {
			t.Errorf("position %d: contentHeight(20)=%d, want %d", c.pos, got, c.wantH)
		}
	}
}

func TestContentOffset(t *testing.T) {
	tg := &TabGroup{tabPosition: TabTop, verticalTabWidth: 3}
	x, y := tg.contentOffset()
	if x != 0 || y != 1 {
		t.Fatalf("TabTop: got (%d,%d), want (0,1)", x, y)
	}
	tg.tabPosition = TabLeft
	x, y = tg.contentOffset()
	if x != 3 || y != 0 {
		t.Fatalf("TabLeft: got (%d,%d), want (3,0)", x, y)
	}
	tg.tabPosition = TabBottom
	x, y = tg.contentOffset()
	if x != 0 || y != 0 {
		t.Fatalf("TabBottom: got (%d,%d), want (0,0)", x, y)
	}
}

func TestRenderTabBarHorizontalLongName(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("thisisareallylongtabname")
	out := tg.View(60, 5)
	if out == "" {
		t.Fatal("expected non-empty view")
	}
	if len(tg.tabRegions) == 0 {
		t.Fatal("expected tabRegions to be set")
	}
}

func TestRenderTabBarVerticalLongName(t *testing.T) {
	tg := NewTabGroup(TabLeft)
	tg.NewTab("thisisareallylongtabname")
	out := tg.View(60, 5)
	if out == "" {
		t.Fatal("expected non-empty view")
	}
	if len(tg.tabRegions) == 0 {
		t.Fatal("expected tabRegions to be set")
	}
}

func TestIsOnTabBar(t *testing.T) {
	tg := &TabGroup{tabPosition: TabTop, width: 10, height: 10}
	if !tg.isOnTabBar(0, 0) {
		t.Fatal("expected TabTop y==0 to be on tab bar")
	}
	tg.tabPosition = TabBottom
	if !tg.isOnTabBar(0, tg.height-1) {
		t.Fatal("expected TabBottom y==height-1 to be on tab bar")
	}
	tg.tabPosition = TabLeft
	tg.verticalTabWidth = 2
	if !tg.isOnTabBar(0, 0) {
		t.Fatal("expected TabLeft x<verticalTabWidth to be on tab bar")
	}
	if !tg.isOnTabBar(1, 0) {
		t.Fatal("expected TabLeft x<verticalTabWidth to be on tab bar")
	}
	tg.tabPosition = TabRight
	tg.verticalTabWidth = 2
	if !tg.isOnTabBar(tg.width-1, 0) {
		t.Fatal("expected TabRight x>=width-verticalTabWidth to be on tab bar")
	}
	tg.tabPosition = TabNone
	if tg.isOnTabBar(0, 0) {
		t.Fatal("expected TabNone to never be on tab bar")
	}
}

func TestHandleTabBarClickNoAction(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	_ = tg.View(80, 5)
	cmd := tg.handleTabBarClick(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	if cmd != nil {
		t.Fatal("expected nil for non-press action")
	}
	cmd = tg.handleTabBarClick(tea.MouseMsg{X: 0, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonRight})
	if cmd != nil {
		t.Fatal("expected nil for right button")
	}
}

func TestSwitchTab(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.NewTab("b")
	tg.activeTab = 0
	tg.switchTab(2)
	if tg.activeTab != 2 {
		t.Fatalf("expected switchTab to set 2, got %d", tg.activeTab)
	}
	tg.switchTab(-1)
	if tg.activeTab != 2 {
		t.Fatalf("expected switchTab(-1) to be no-op, activeTab=%d", tg.activeTab)
	}
	tg.switchTab(5)
	if tg.activeTab != 2 {
		t.Fatalf("expected switchTab(5) out-of-range to be no-op, activeTab=%d", tg.activeTab)
	}
}

func TestCloseTabDirect(t *testing.T) {
	tg := NewTabGroup(TabTop)
	tg.NewTab("a")
	tg.NewTab("b")
	tg.activeTab = 0
	tg.closeTab(1)
	if len(tg.tabs) != 2 {
		t.Fatalf("expected 2 tabs after close, got %d", len(tg.tabs))
	}
	tg.closeTab(-1)
	if len(tg.tabs) != 2 {
		t.Fatalf("expected closeTab(-1) to be no-op, got %d tabs", len(tg.tabs))
	}
	tg.closeTab(10)
	if len(tg.tabs) != 2 {
		t.Fatalf("expected closeTab(10) to be no-op, got %d tabs", len(tg.tabs))
	}
	single := NewTabGroup(TabTop)
	single.closeTab(0)
	if len(single.tabs) != 1 {
		t.Fatalf("expected closeTab to leave single tab, got %d", len(single.tabs))
	}
}

func TestPadRight(t *testing.T) {
	if got := padRight("ab", 5); got != "ab   " {
		t.Fatalf("expected padded to 5, got %q", got)
	}
	if got := padRight("abcd", 5); got != "abcd " {
		t.Fatalf("expected padding to width 5, got %q", got)
	}
	if got := padRight("abcd", 2); got != "abcd" {
		t.Fatalf("expected unchanged, got %q", got)
	}
}
