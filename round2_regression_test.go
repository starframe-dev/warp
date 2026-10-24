package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type round2Panel struct {
	BasePanel
	resizes     []ResizeMsg
	mouse       []tea.MouseMsg
	keys        []tea.KeyMsg
	command     tea.Cmd
	windowSizes int
}

func (p *round2Panel) Elements(width, height int) []Element {
	return []Element{{Role: "panel", Name: "target", Bounds: Bounds{W: width, H: height}}}
}

func (p *round2Panel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ResizeMsg:
		p.resizes = append(p.resizes, msg)
		return p.command
	case tea.WindowSizeMsg:
		p.windowSizes++
		return p.command
	case tea.MouseMsg:
		p.mouse = append(p.mouse, msg)
	case tea.KeyMsg:
		p.keys = append(p.keys, msg)
	}
	return p.command
}

func TestTabGroupElementsExposeLeafViewport(t *testing.T) {
	tg := NewTabGroup(TabLeft)
	p := &round2Panel{}
	tg.ActiveTab().SetRootPanel(p)
	elems := tg.Elements(80, 12)
	element, ok := FindElement(elems, "panel", "target", "")
	if !ok || element.Bounds.W != tg.contentWidth(80) || element.Bounds.X != tg.verticalTabWidth {
		t.Fatalf("leaf element=%+v ok=%v contentWidth=%d elements=%+v", element, ok, tg.contentWidth(80), elems)
	}
}

func TestVerticalTabGroupRecomputesWidthAfterViewportChanges(t *testing.T) {
	for _, pos := range []TabPosition{TabLeft, TabRight} {
		t.Run(map[TabPosition]string{TabLeft: "left", TabRight: "right"}[pos], func(t *testing.T) {
			tg := NewTabGroup(pos)
			p := &round2Panel{}
			tg.ActiveTab().SetRootPanel(p)
			tg.ActiveTab().name = "a very long first tab"
			second := tg.NewTab("another long tab name")
			second.name = "another long tab name"
			tg.switchTab(0)
			for _, width := range []int{80, 10, 80} {
				tg.switchTab(0)
				tg.width = width
				tg.height = 12
				tg.Update(tea.WindowSizeMsg{Width: width, Height: 12})
				wantW := max(0, width-tg.verticalTabBarWidth(width))
				if len(p.resizes) == 0 || p.resizes[len(p.resizes)-1].Width != wantW {
					t.Fatalf("pre-View width=%d resize=%+v want content width %d active=%p target=%p root=%T tabs=%d index=%d", width, p.resizes, wantW, tg.ActiveTab(), tg.tabs[0], tg.ActiveTab().RootPanel(), len(tg.tabs), tg.activeTab)
				}
				elements := tg.Elements(width, 12)
				element, ok := FindElement(elements, "panel", "target", "")
				if !ok || element.Bounds.W != wantW {
					t.Fatalf("Elements=%+v want leaf width %d", elements, wantW)
				}
				wantX := 0
				if pos == TabLeft {
					wantX = width - wantW
				}
				if element.Bounds.X != wantX {
					t.Fatalf("Elements x=%d want %d", element.Bounds.X, wantX)
				}
				clickX := 0
				if pos == TabRight {
					clickX = width - tg.verticalTabBarWidth(width)
				}
				tg.Update(tea.MouseMsg{X: clickX, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
				if len(p.mouse) != 0 {
					t.Fatalf("tab-bar click leaked to panel: %+v", p.mouse)
				}
				view := tg.View(width, 12)
				if pos == TabLeft {
					off, _ := tg.contentOffset()
					if off != width-tg.contentWidth(width) {
						t.Fatalf("left offset=%d content width=%d", off, wantW)
					}
				}
				if pos == TabRight {
					off, _ := tg.contentOffset()
					if off != 0 {
						t.Fatalf("right offset=%d", off)
					}
				}
				if len(view) == 0 {
					t.Fatalf("width=%d rendered empty view", width)
				}
				for row, line := range strings.Split(view, "\n") {
					if got := ansi.StringWidth(line); got != width {
						t.Fatalf("width=%d rendered row=%d width=%d: %q", width, row, got, line)
					}
				}
			}
		})
	}
}

func TestGlobalWindowSizeDoesNotOverrideNestedTabGroupBehindWrappers(t *testing.T) {
	cases := []struct {
		name string
		wrap func(Panel) Panel
	}{
		{"scrollable", func(p Panel) Panel { return NewScrollable(p) }},
		{"collapsible", func(p Panel) Panel { return NewCollapsible("wrap", p) }},
		{"selectable", func(p Panel) Panel { return NewSelectable(p) }},
		{"scrollable-collapsible", func(p Panel) Panel { return NewScrollable(NewCollapsible("wrap", p)) }},
		{"nested-warp", func(p Panel) Panel {
			inner := New()
			inner.SetRoot(NewTabGroup(TabTop))
			inner.ActiveTab().SetRootPanel(p)
			return inner.AsPanel()
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			leaf := &round2Panel{}
			inner := NewTabGroup(TabTop)
			inner.ActiveTab().SetRootPanel(leaf)
			outer := NewTabGroup(TabTop)
			outer.ActiveTab().SetRootPanel(tc.wrap(inner))
			outer.Update(tea.WindowSizeMsg{Width: 50, Height: 20})
			wantHeight := 18
			if tc.name == "collapsible" || tc.name == "scrollable-collapsible" || tc.name == "nested-warp" {
				wantHeight = 17
			}
			if len(leaf.resizes) == 0 || leaf.resizes[len(leaf.resizes)-1] != (ResizeMsg{Width: 50, Height: wantHeight}) {
				t.Fatalf("leaf sizes=%+v want 50x%d", leaf.resizes, wantHeight)
			}
		})
	}
}

func TestLegacyWindowSizeTraversalHandlesEmptyNodesAndWrappers(t *testing.T) {
	tab := NewTab("empty")
	commands := []tea.Cmd{}
	msg := tea.WindowSizeMsg{Width: 10, Height: 5}
	tab.appendLegacyWindowSizePanel(&commands, nil, msg)
	tab.appendLegacyWindowSizePanel(&commands, &Scrollable{}, msg)
	tab.appendLegacyWindowSizePanel(&commands, &Collapsible{}, msg)
	tab.appendLegacyWindowSizePanel(&commands, &Selectable{}, msg)
	tab.appendLegacyWindowSizePanel(&commands, &warpPanel{}, msg)
	tab.appendLegacyWindowSizePanel(&commands, &TabGroup{tabs: []*Tab{nil, {floats: []*FloatPane{nil}}}}, msg)
	tab.appendLegacyWindowSizeNode(&commands, nil, msg)
	if len(commands) != 0 {
		t.Fatalf("empty wrappers emitted %d commands", len(commands))
	}
	legacy := &round2Panel{command: func() tea.Msg { return "legacy-resize" }}
	tab.appendLegacyWindowSizePanel(&commands, legacy, msg)
	if len(commands) != 1 || commands[0]() != "legacy-resize" {
		t.Fatalf("legacy WindowSize command not preserved: %d commands", len(commands))
	}
}

func TestGlobalWindowSizeRoutesThroughNestedLayoutNodes(t *testing.T) {
	for _, kind := range []string{"split", "flex"} {
		t.Run(kind, func(t *testing.T) {
			leaf := &round2Panel{}
			inner := NewTabGroup(TabTop)
			anchor := &round2Panel{}
			innerTab := inner.ActiveTab()
			innerTab.SetRootPanel(anchor)
			if kind == "split" {
				innerTab.SplitVertical(anchor, .5, leaf)
			} else {
				innerTab.FlexRow(anchor, []FlexItemSpec{{Panel: anchor, Grow: 1}, {Panel: leaf, Grow: 1}})
			}
			outer := NewTabGroup(TabTop)
			outer.ActiveTab().SetRootPanel(inner)
			outer.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
			if len(leaf.resizes) == 0 {
				t.Fatalf("%s nested leaf received no local resize", kind)
			}
			got := leaf.resizes[len(leaf.resizes)-1]
			if got.Height != 18 || got.Width >= 60 {
				t.Fatalf("%s nested leaf resize=%+v; expected allocated local size, not global 60x20", kind, got)
			}
		})
	}
}

func TestFloatPanelsReceiveUnknownBroadcastMessages(t *testing.T) {
	panel := &round2Panel{command: func() tea.Msg { return "float-message" }}
	tab := NewTab("float-broadcast")
	tab.Float(panel, 1, 1, 12, 5)
	commands := tab.broadcastMsg(struct{}{})
	if len(commands) != 1 || commands[0]() != "float-message" {
		t.Fatalf("float broadcast commands=%d", len(commands))
	}
}

func TestLegacyWindowSizeStillReachesOrdinaryNestedFloatPanels(t *testing.T) {
	leaf := &round2Panel{}
	inner := NewTabGroup(TabNone)
	inner.ActiveTab().Float(leaf, 2, 2, 12, 5)
	outer := NewTabGroup(TabTop)
	outer.ActiveTab().SetRootPanel(inner)
	outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if leaf.windowSizes != 1 {
		t.Fatalf("ordinary float leaf got %d WindowSizeMsg deliveries, want 1", leaf.windowSizes)
	}
}

func TestFloatGlobalWindowSizeDoesNotOverrideNestedTabGroup(t *testing.T) {
	leaf := &round2Panel{}
	inner := NewTabGroup(TabTop)
	inner.ActiveTab().SetRootPanel(leaf)
	outer := NewTabGroup(TabNone)
	tab := outer.ActiveTab()
	outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	tab.Float(inner, 2, 2, 30, 10)
	outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if len(leaf.resizes) == 0 || leaf.resizes[len(leaf.resizes)-1] != (ResizeMsg{Width: 28, Height: 7}) {
		t.Fatalf("float nested sizes=%+v want 28x7", leaf.resizes)
	}
}

func TestFloatCreationCommandHasBubbleTeaCompatibleAPI(t *testing.T) {
	type floatWithCommand interface {
		FloatWithCmd(Panel, int, int, int, int) tea.Cmd
	}
	instance, ok := interface{}(NewTab("float-command")).(floatWithCommand)
	if !ok {
		t.Fatal("Tab has no compatible method to return Float initializer command")
	}
	p := &round2Panel{command: func() tea.Msg { return "float-init" }}
	if cmd := instance.FloatWithCmd(nil, 0, 0, 20, 8); cmd != nil {
		t.Fatal("nil panel unexpectedly returned a command")
	}
	if cmd := instance.FloatWithCmd(p, 0, 0, 0, 8); cmd != nil {
		t.Fatal("invalid dimensions unexpectedly returned a command")
	}
	cmd := instance.FloatWithCmd(p, 0, 0, 20, 8)
	if cmd == nil {
		t.Fatal("FloatWithCmd returned nil command")
	}
	if got := cmd(); got != "float-init" {
		t.Fatalf("command result=%v want float-init", got)
	}
}

func TestNestedWarpSetRootAfterOuterCloseDoesNotDoubleUnmount(t *testing.T) {
	old := &ownershipTestPanel{}
	next := &ownershipTestPanel{}
	inner := New()
	inner.ActiveTab().SetRootPanel(old)
	outer := New()
	outer.ActiveTab().SetRootPanel(inner.AsPanel())
	_ = outer.Close()
	if old.unmounts != 1 {
		t.Fatalf("outer close old unmounts=%d", old.unmounts)
	}
	inner.ActiveTab().SetRootPanel(next)
	if old.unmounts != 1 {
		t.Fatalf("detach replacement old unmounts=%d", old.unmounts)
	}
	_ = inner.Close()
	if old.unmounts != 1 || next.unmounts != 1 {
		t.Fatalf("after inner close old/new unmounts=%d/%d", old.unmounts, next.unmounts)
	}
}

func TestTabMouseUpdateAndTabGroupResizeCommands(t *testing.T) {
	panel := &round2Panel{command: func() tea.Msg { return "resize" }}
	group := NewTabGroup(TabNone)
	tab := group.ActiveTab()
	tab.SetRootPanel(panel)
	tab.Update(ResizeMsg{Width: 20, Height: 8})
	tab.Update(tea.MouseMsg{X: 1, Y: 1, Action: tea.MouseActionMotion})
	if len(panel.mouse) != 1 {
		t.Fatalf("Tab.Update did not route mouse: %+v", panel.mouse)
	}
	if cmd := group.Update(tea.WindowSizeMsg{Width: 20, Height: 8}); cmd == nil {
		t.Fatal("WindowSize resize command was lost")
	}
	group.tabs = append(group.tabs, nil)
	group.Update(tea.WindowSizeMsg{Width: 20, Height: 8})
	group.tabs = group.tabs[:1]
	if cmd := group.Update(ResizeMsg{Width: 20, Height: 8}); cmd == nil {
		t.Fatal("ResizeMsg command was lost")
	}
}

func TestStandaloneTabWithoutFocusDoesNotBroadcastKeyboard(t *testing.T) {
	first, second := &round2Panel{}, &round2Panel{}
	tab := NewTab("no-focus")
	tab.SetRootPanel(first)
	tab.SplitVertical(first, .5, second)
	tab.setFocus(nil)
	tab.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if len(first.keys) != 0 || len(second.keys) != 0 {
		t.Fatalf("unfocused keyboard delivered to leaves: %d/%d", len(first.keys), len(second.keys))
	}
}
