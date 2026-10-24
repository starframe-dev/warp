package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type round3Panel struct {
	BasePanel
	name     string
	resizes  []ResizeMsg
	windows  []tea.WindowSizeMsg
	views    []ResizeMsg
	mouse    []tea.MouseMsg
	unmounts int
}

type round3CommandResult struct {
	panel string
	kind  string
}

func (p *round3Panel) View(width, height int) string {
	p.views = append(p.views, ResizeMsg{Width: width, Height: height})
	line := strings.Repeat("x", max(0, width))
	rows := make([]string, max(0, height))
	for i := range rows {
		rows[i] = line
	}
	return strings.Join(rows, "\n")
}

func (p *round3Panel) Update(msg tea.Msg) tea.Cmd {
	kind := "other"
	switch msg := msg.(type) {
	case ResizeMsg:
		p.resizes = append(p.resizes, msg)
		kind = "resize"
	case tea.WindowSizeMsg:
		p.windows = append(p.windows, msg)
		kind = "window"
	case tea.MouseMsg:
		p.mouse = append(p.mouse, msg)
		kind = "mouse"
	}
	return func() tea.Msg { return round3CommandResult{panel: p.name, kind: kind} }
}

func (p *round3Panel) Elements(width, height int) []Element {
	return []Element{{Role: "panel", Name: p.name, Bounds: Bounds{W: width, H: height}}}
}

func (p *round3Panel) Unmount() { p.unmounts++ }

func runRound3Commands(cmd tea.Cmd, out *[]round3CommandResult) {
	if cmd == nil {
		return
	}
	switch msg := cmd().(type) {
	case tea.BatchMsg:
		for _, child := range msg {
			runRound3Commands(child, out)
		}
	case round3CommandResult:
		*out = append(*out, msg)
	}
}

func resetRound3Panel(p *round3Panel) {
	p.resizes = nil
	p.windows = nil
	p.views = nil
	p.mouse = nil
}

func TestNestedTabWindowSizePreservesLocalGeometry(t *testing.T) {
	cases := []struct {
		name  string
		wrap  func(*Tab, *Tab) // attach outer and nested tab
		want  ResizeMsg
		click tea.MouseMsg
	}{
		{
			name:  "tab-group-tab-leaf",
			wrap:  func(outer, inner *Tab) { outer.SetRootPanel(inner) },
			want:  ResizeMsg{Width: 80, Height: 23},
			click: tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		},
		{
			name:  "scrollable-tab-leaf",
			wrap:  func(outer, inner *Tab) { outer.SetRootPanel(NewScrollable(inner)) },
			want:  ResizeMsg{Width: 80, Height: 23},
			click: tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		},
		{
			name:  "collapsible-tab-leaf",
			wrap:  func(outer, inner *Tab) { outer.SetRootPanel(NewCollapsible("group", inner)) },
			want:  ResizeMsg{Width: 80, Height: 22},
			click: tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		},
		{
			name:  "float-tab-leaf",
			wrap:  func(outer, inner *Tab) { outer.Float(inner, 2, 2, 90, 30) },
			want:  ResizeMsg{Width: 78, Height: 21},
			click: tea.MouseMsg{X: 5, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			leaf := &round3Panel{name: "leaf"}
			inner := NewTab("inner")
			inner.SetRootPanel(leaf)
			outer := NewTabGroup(TabTop)
			outerTab := outer.ActiveTab()
			tc.wrap(outerTab, inner)
			resetRound3Panel(leaf) // discard Float's immediate initialization resize

			cmd := outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			if len(leaf.resizes) != 1 || leaf.resizes[0] != tc.want {
				t.Fatalf("local ResizeMsg sequence=%+v, want only %+v", leaf.resizes, tc.want)
			}
			if len(leaf.windows) != 1 || leaf.windows[0] != (tea.WindowSizeMsg{Width: 80, Height: 24}) {
				t.Fatalf("legacy WindowSizeMsg delivery=%+v, want one full-window message", leaf.windows)
			}
			var results []round3CommandResult
			runRound3Commands(cmd, &results)
			if len(results) != 2 {
				t.Fatalf("returned child commands=%+v, want ResizeMsg and WindowSizeMsg commands", results)
			}

			view := outer.View(80, 24)
			rows := strings.Split(view, "\n")
			if len(rows) != 24 {
				t.Fatalf("rendered rows=%d, want 24", len(rows))
			}
			for row, line := range rows {
				if got := ansi.StringWidth(line); got != 80 {
					t.Fatalf("rendered row %d has width %d, want 80", row, got)
				}
			}
			elements := outer.Elements(80, 24)
			element, ok := FindElement(elements, "panel", "leaf", "")
			if !ok || element.Bounds.W != tc.want.Width || element.Bounds.H > tc.want.Height {
				t.Fatalf("leaf semantic element=%+v, found=%v; local geometry=%+v", element, ok, tc.want)
			}

			outer.Update(tc.click)
			if len(leaf.mouse) != 1 {
				t.Fatalf("mouse hit-testing delivered %+v, want one event at leaf", leaf.mouse)
			}
		})
	}
}

func TestNestedTabSplitAndFlexPreserveLocalWindowGeometry(t *testing.T) {
	for _, kind := range []string{"split", "flex"} {
		t.Run(kind, func(t *testing.T) {
			left := &round3Panel{name: "left"}
			right := &round3Panel{name: "right"}
			inner := NewTab("inner-layout")
			inner.SetRootPanel(left)
			if kind == "split" {
				inner.SplitVertical(left, 0.5, right)
			} else {
				inner.FlexRow(left, []FlexItemSpec{{Panel: left, Grow: 1}, {Panel: right, Grow: 1}})
			}
			outer := NewTabGroup(TabTop)
			outer.ActiveTab().SetRootPanel(inner)
			cmd := outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			for _, panel := range []*round3Panel{left, right} {
				if len(panel.resizes) != 1 || panel.resizes[0].Height != 23 || panel.resizes[0].Width >= 80 {
					t.Fatalf("%s panel %q local resize=%+v", kind, panel.name, panel.resizes)
				}
				if len(panel.windows) != 1 || panel.windows[0].Width != 80 || panel.windows[0].Height != 24 {
					t.Fatalf("%s panel %q legacy messages=%+v", kind, panel.name, panel.windows)
				}
			}
			var results []round3CommandResult
			runRound3Commands(cmd, &results)
			if len(results) != 4 {
				t.Fatalf("returned commands=%+v, want two per leaf", results)
			}
			view := outer.View(80, 24)
			if len(strings.Split(view, "\n")) != 24 {
				t.Fatalf("%s rendered view did not preserve parent height", kind)
			}
			elements := outer.Elements(80, 24)
			for _, panel := range []*round3Panel{left, right} {
				element, ok := FindElement(elements, "panel", panel.name, "")
				if !ok || element.Bounds.W != panel.resizes[0].Width || element.Bounds.H != 23 {
					t.Fatalf("%s element %q = %+v, found=%v; resize=%+v", kind, panel.name, element, ok, panel.resizes[0])
				}
			}
			outer.Update(tea.MouseMsg{X: 60, Y: 5, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
			if len(right.mouse) != 1 || len(left.mouse) != 0 {
				t.Fatalf("%s mouse hit left/right=%d/%d, want target right leaf only", kind, len(left.mouse), len(right.mouse))
			}
		})
	}
}

func TestAsPanelReturnsStableAdapterIdentity(t *testing.T) {
	inner := New()
	first := inner.AsPanel()
	for i := 0; i < 16; i++ {
		if adapter := inner.AsPanel(); adapter != first {
			t.Fatalf("AsPanel call %d returned a different adapter identity", i)
		}
	}
}

func TestMultipleAsPanelAdaptersRetainSharedOwnershipUntilLastRemoval(t *testing.T) {
	leaf := &round3Panel{name: "resource"}
	inner := New()
	inner.SetRoot(leaf)
	first, second := inner.AsPanel(), inner.AsPanel()

	outer := New()
	firstTab := outer.ActiveTab()
	firstTab.SetRootPanel(first)
	firstTab.SplitVertical(first, 0.5, second)
	destination := outer.NewTab("moved-adapter")
	destination.SetRootPanel(first)
	firstTab.SetRootPanel(second) // move first to another tab within the same ownership domain
	destination.SetRootPanel(&round3Panel{name: "placeholder"})

	if inner.ownership != outer.ownership || inner.ownsDomainRoot {
		t.Errorf("nested Warp lost shared ownership after first adapter removal: ownership=%p outer=%p ownsRoot=%v", inner.ownership, outer.ownership, inner.ownsDomainRoot)
	}
	if leaf.unmounts != 0 {
		t.Fatalf("resource unmounted while second adapter remains: %d", leaf.unmounts)
	}

	outer.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if len(leaf.resizes) == 0 {
		t.Fatal("remaining adapter stopped forwarding resize to its root")
	}
	if view := second.View(40, 12); view == "" {
		t.Fatal("remaining adapter stopped rendering its Warp")
	}
	if len(leaf.views) == 0 {
		t.Fatal("remaining adapter did not render the shared resource")
	}
	if err := inner.Close(); err != nil {
		t.Fatal(err)
	}
	if leaf.unmounts != 0 {
		t.Errorf("Close unmounted resource while another adapter remained: %d", leaf.unmounts)
	}
	if view := second.View(40, 12); view == "" {
		t.Error("remaining adapter stopped rendering after Close while embedded")
	}

	firstTab.SetRootPanel(&round3Panel{name: "replacement"})
	if leaf.unmounts != 1 {
		t.Fatalf("resource unmount count after removing last adapter=%d, want 1", leaf.unmounts)
	}
	for i := 0; i < 2; i++ {
		if err := outer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := inner.Close(); err != nil {
			t.Fatal(err)
		}
	}
	if leaf.unmounts != 1 {
		t.Fatalf("resource unmounted %d times after repeated Close, want exactly once", leaf.unmounts)
	}
}

func BenchmarkOwnershipDetachWithRetainedAsPanel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		leaf := &round3Panel{name: "resource"}
		inner := New()
		inner.SetRoot(leaf)
		first, second := inner.AsPanel(), inner.AsPanel()
		outer := New()
		tab := outer.ActiveTab()
		tab.SetRootPanel(first)
		tab.SplitVertical(first, 0.5, second)
		tab.SetRootPanel(second)
	}
}

func TestNestedWarpAdaptersRetainOwnershipAtEachLevel(t *testing.T) {
	leaf := &round3Panel{name: "deep-resource"}
	inner := New()
	inner.SetRoot(leaf)
	middle := New()
	middle.SetRoot(inner.AsPanel())
	first, second := middle.AsPanel(), middle.AsPanel()
	outer := New()
	tab := outer.ActiveTab()
	tab.SetRootPanel(first)
	tab.SplitVertical(first, 0.5, second)

	tab.SetRootPanel(second)
	if middle.ownership != outer.ownership || inner.ownership != outer.ownership || leaf.unmounts != 0 {
		t.Errorf("nested ownership detached early: middle=%p inner=%p outer=%p unmounts=%d", middle.ownership, inner.ownership, outer.ownership, leaf.unmounts)
	}
	if err := middle.Close(); err != nil {
		t.Fatal(err)
	}
	if leaf.unmounts != 0 {
		t.Errorf("Close released nested resource while another adapter remained: %d", leaf.unmounts)
	}
	tab.SetRootPanel(&round3Panel{name: "replacement"})
	if leaf.unmounts != 1 {
		t.Fatalf("nested resource unmounted %d times after last adapter removal, want 1", leaf.unmounts)
	}
	_ = outer.Close()
	_ = middle.Close()
	_ = inner.Close()
	if leaf.unmounts != 1 {
		t.Fatalf("nested resource unmounted %d times after Close, want exactly once", leaf.unmounts)
	}
}
