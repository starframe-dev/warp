package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type commandResizePanel struct {
	BasePanel
	resizes        []ResizeMsg
	lastWindowSize tea.WindowSizeMsg
	command        tea.Cmd
	keys           []tea.KeyMsg
	mouse          []tea.MouseMsg
}

func (p *commandResizePanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ResizeMsg:
		p.resizes = append(p.resizes, msg)
		return p.command
	case tea.WindowSizeMsg:
		p.lastWindowSize = msg
	case tea.KeyMsg:
		p.keys = append(p.keys, msg)
	case tea.MouseMsg:
		p.mouse = append(p.mouse, msg)
	}
	return nil
}

func TestTabGroupWindowSizeUsesContentViewport(t *testing.T) {
	for _, pos := range []TabPosition{TabTop, TabBottom, TabLeft, TabRight, TabNone} {
		t.Run(map[TabPosition]string{TabTop: "top", TabBottom: "bottom", TabLeft: "left", TabRight: "right", TabNone: "none"}[pos], func(t *testing.T) {
			tg := NewTabGroup(pos)
			p := &commandResizePanel{}
			tg.ActiveTab().SetRootPanel(p)
			tg.Update(tea.WindowSizeMsg{Width: 30, Height: 12})
			if len(p.resizes) != 1 || p.resizes[0] != (ResizeMsg{Width: tg.contentWidth(30), Height: tg.contentHeight(12)}) {
				t.Fatalf("resize=%+v", p.resizes)
			}
			if p.lastWindowSize != (tea.WindowSizeMsg{Width: 30, Height: 12}) {
				t.Fatalf("legacy=%+v", p.lastWindowSize)
			}
		})
	}
}

func TestNestedTabGroupsResizeToContentViewport(t *testing.T) {
	leaf := &commandResizePanel{}
	inner := NewTabGroup(TabBottom)
	inner.ActiveTab().SetRootPanel(leaf)
	outer := NewTabGroup(TabTop)
	outer.ActiveTab().SetRootPanel(inner)
	outer.Update(tea.WindowSizeMsg{Width: 40, Height: 20})
	if len(leaf.resizes) != 1 || leaf.resizes[0] != (ResizeMsg{Width: 40, Height: 18}) {
		t.Fatalf("nested resize=%+v", leaf.resizes)
	}
}

func TestTabDragResizeCommandsAreReturned(t *testing.T) {
	for _, kind := range []string{"split", "flex"} {
		for _, action := range []tea.MouseAction{tea.MouseActionMotion, tea.MouseActionRelease} {
			t.Run(kind+map[tea.MouseAction]string{tea.MouseActionMotion: "-motion", tea.MouseActionRelease: "-release"}[action], func(t *testing.T) {
				p := &commandResizePanel{command: func() tea.Msg { return "resize" }}
				q := &commandResizePanel{command: p.command}
				tab := NewTab("drag")
				tab.SetRootPanel(p)
				if kind == "split" {
					tab.SplitVertical(p, .5, q)
					tab.dragging = tab.root.Split
				} else {
					tab.FlexRow(p, []FlexItemSpec{{Panel: p, Grow: 1}, {Panel: q, Grow: 1}})
					tab.flexDragging = tab.root.Flex
					tab.flexDragIdx = 0
				}
				tab.View(40, 10)
				cmd := tab.handleMouse(tea.MouseMsg{X: 22, Y: 4, Action: action, Button: tea.MouseButtonLeft}, 0, 0, 40, 10)
				if cmd == nil {
					t.Fatal("resize command discarded")
				}
				if _, ok := cmd().(tea.BatchMsg); !ok {
					t.Fatalf("command result %T", cmd())
				}
			})
		}
	}
}

func TestFloatResizePropagatesLocalDimensionsAndCommands(t *testing.T) {
	p := &commandResizePanel{command: func() tea.Msg { return "float" }}
	tab := NewTab("float")
	tab.Update(ResizeMsg{Width: 40, Height: 20})
	tab.Float(p, 2, 3, 18, 8)
	if len(p.resizes) != 1 || p.resizes[0] != (ResizeMsg{Width: 16, Height: 6}) {
		t.Fatalf("creation=%+v", p.resizes)
	}
	cmd := tab.Update(ResizeMsg{Width: 12, Height: 15})
	if len(p.resizes) != 2 || p.resizes[1] != (ResizeMsg{Width: 10, Height: 6}) || cmd == nil {
		t.Fatalf("resize=%+v cmd=%v", p.resizes, cmd)
	}
	n := len(p.resizes)
	if tab.Update(ResizeMsg{Width: 12, Height: 15}) != nil || len(p.resizes) != n {
		t.Fatal("unchanged geometry notified")
	}
}

func TestInputValueControlSequencesRenderAsPlainText(t *testing.T) {
	for _, value := range []string{"\x1b[31mred\x1b[0m", "\x1b]0;title\x07visible", "\x1b[31", "\x1b]0;partial", "\x1b\x07\x00", "👩‍💻e\u0301"} {
		input := NewInput("")
		input.SetValue(value)
		view := input.View(80, 1)
		plain := ansi.Strip(view)
		if strings.ContainsAny(plain, "\x1b\x00\x07") || input.Value != value {
			t.Fatalf("unsafe/mutated %q -> %q", value, view)
		}
	}
}

func TestFlexGrowZeroWeightNeverGetsWeightedRemainder(t *testing.T) {
	for _, grows := range [][]int{{1, 1, 0}, {0, 1, 1}, {1, 0, 1}} {
		items := make([]*FlexItem, len(grows))
		for i, grow := range grows {
			items[i] = &FlexItem{Node: &Node{Panel: BasePanel{}}, Basis: 1, Grow: grow}
		}
		sizes := computeFlexSizes(14, items)
		sum := 0
		for i, size := range sizes {
			sum += size
			if grows[i] == 0 && (grows[0]+grows[1]+grows[2]) > 0 && size != 1 {
				t.Fatalf("grows=%v sizes=%v zero-grow item %d got free space", grows, sizes, i)
			}
		}
		if sum != 14 {
			t.Fatalf("grows=%v sizes=%v sum=%d", grows, sizes, sum)
		}
	}
	items := []*FlexItem{{Node: &Node{Panel: BasePanel{}}, Basis: 1}, {Node: &Node{Panel: BasePanel{}}, Basis: 1}, {Node: &Node{Panel: BasePanel{}}, Basis: 1}}
	sizes := computeFlexSizes(14, items)
	if sizes[0]+sizes[1]+sizes[2] != 14 || sizes[2] <= 1 {
		t.Fatalf("zero-weight uniform sizes=%v", sizes)
	}
}

func TestPopoverClampsBorderWithinNarrowViewport(t *testing.T) {
	for _, width := range []int{1, 2, 3, 5, 8} {
		lines := make([]string, 4)
		for i := range lines {
			lines[i] = strings.Repeat(".", width)
		}
		p := &Popover{Width: 40, Items: []PopoverItem{{Name: "很长的菜单项目"}, {Name: "second"}}}
		out := p.Overlay(lines, width, len(lines))
		for i, line := range out {
			if ansi.StringWidth(line) != width {
				t.Fatalf("width %d row %d: %q", width, i, line)
			}
		}
		if p.boxW > width {
			t.Fatalf("box %d > viewport %d", p.boxW, width)
		}
	}
}

func TestSelectableForwardsUnsupportedShiftKeys(t *testing.T) {
	for _, key := range []tea.KeyMsg{{Type: tea.KeyShiftHome}, {Type: tea.KeyShiftEnd}, {Type: tea.KeyShiftTab}} {
		p := &commandResizePanel{}
		NewSelectable(p).Update(key)
		if len(p.keys) != 1 || p.keys[0].String() != key.String() {
			t.Errorf("%s forwarded keys=%+v", key.String(), p.keys)
		}
	}
}

func TestStandaloneTabRoutesKeyboardAndMouseToLeaf(t *testing.T) {
	first, second := &commandResizePanel{}, &commandResizePanel{}
	tab := NewTab("standalone")
	tab.SetRootPanel(first)
	tab.SplitVertical(first, .5, second)
	tab.setFocus(first)
	tab.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	if len(first.keys) != 1 || len(second.keys) != 0 {
		t.Fatalf("keyboard=%d/%d", len(first.keys), len(second.keys))
	}
	tab.handleMouse(tea.MouseMsg{X: 30, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, 0, 0, 40, 5)
	if len(second.mouse) != 1 || len(first.mouse) != 0 {
		t.Fatalf("mouse=%d/%d", len(first.mouse), len(second.mouse))
	}
}
