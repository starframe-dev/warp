package warp

import (
	"math"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestSharedGeometryAgreesForCollapsedAndExpandedSplit(t *testing.T) {
	for _, collapsed := range []bool{false, true} {
		t.Run(map[bool]string{false: "expanded", true: "collapsed"}[collapsed], func(t *testing.T) {
			first := &geometryTestPanel{name: "A"}
			second := &geometryTestPanel{name: "B"}
			tab := NewTab("geometry")
			tab.SetRootPanel(first)
			tab.SplitVertical(first, 0.5, second)
			if collapsed {
				tab.root.Split.First.Collapse = &NodeCollapse{Active: true, Width: 1}
			}
			checkGeometryAgreement(t, tab, second, 15, 5, map[bool]int{false: 8, true: 1}[collapsed])
			if collapsed && len(tab.lastBorders) != 0 {
				t.Fatalf("collapsed split has %d visible borders", len(tab.lastBorders))
			}
			if !collapsed && (len(tab.lastBorders) != 1 || tab.lastBorders[0].X != 7) {
				t.Fatalf("expanded split borders = %+v, want border at x=7", tab.lastBorders)
			}
		})
	}
}

func TestSharedGeometryAgreesForCollapsedAndExpandedFlex(t *testing.T) {
	for _, collapsed := range []bool{false, true} {
		t.Run(map[bool]string{false: "expanded", true: "collapsed"}[collapsed], func(t *testing.T) {
			first := &geometryTestPanel{name: "A"}
			second := &geometryTestPanel{name: "B"}
			tab := NewTab("geometry")
			tab.SetRootPanel(first)
			tab.FlexRow(first, []FlexItemSpec{{Panel: first, Grow: 1}, {Panel: second, Grow: 1}})
			if collapsed {
				tab.root.Flex.Items[0].Collapsed = true
			}
			checkGeometryAgreement(t, tab, second, 15, 5, map[bool]int{false: 8, true: 1}[collapsed])
			if collapsed && len(tab.lastBorders) != 0 {
				t.Fatalf("collapsed flex has %d visible borders", len(tab.lastBorders))
			}
			if !collapsed && (len(tab.lastBorders) != 1 || tab.lastBorders[0].X != 7) {
				t.Fatalf("expanded flex borders = %+v, want border at x=7", tab.lastBorders)
			}
		})
	}
}

func checkGeometryAgreement(t *testing.T, tab *Tab, panel *geometryTestPanel, width, height, wantX int) {
	t.Helper()
	view := tab.View(width, height)
	lines := strings.Split(view, "\n")
	if len(lines) != height {
		t.Fatalf("render has %d lines, want %d", len(lines), height)
	}
	plain := ansi.Strip(lines[0])
	markerByte := strings.Index(plain, panel.name)
	if markerByte < 0 {
		t.Fatalf("rendered row %q does not contain panel marker %q", plain, panel.name)
	}
	markerX := ansi.StringWidth(plain[:markerByte])
	if markerX != wantX {
		t.Fatalf("rendered panel x=%d, want %d (%q)", markerX, wantX, plain)
	}

	var bounds *Bounds
	for _, element := range tab.Elements(width, height) {
		if element.Name == panel.name {
			copy := element.Bounds
			bounds = &copy
			break
		}
	}
	if bounds == nil {
		t.Fatalf("Elements does not contain panel %q", panel.name)
	}
	if bounds.X != wantX {
		t.Fatalf("element x=%d, want %d", bounds.X, wantX)
	}

	hit := tab.panelAt(wantX, 0, width, height)
	if hit == nil || !samePanel(hit.Node.Panel, panel) {
		t.Fatalf("panelAt(%d,0) = %+v, want panel %q", wantX, hit, panel.name)
	}

	tab.Update(ResizeMsg{Width: width, Height: height})
	if panel.lastResize.Width != bounds.W || panel.lastResize.Height != bounds.H {
		t.Fatalf("ResizeMsg=%+v, element size=%dx%d", panel.lastResize, bounds.W, bounds.H)
	}
}

func TestNestedFlexBordersUseActualChildBounds(t *testing.T) {
	for _, collapsed := range []bool{false, true} {
		t.Run(map[bool]string{false: "expanded", true: "collapsed"}[collapsed], func(t *testing.T) {
			a := &geometryTestPanel{name: "A"}
			b := &geometryTestPanel{name: "B"}
			c := &geometryTestPanel{name: "C"}
			tab := NewTab("nested-flex")
			tab.SetRootPanel(a)
			tab.FlexRow(a, []FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 1}})
			tab.SplitVertical(b, 0.5, c)
			if collapsed {
				tab.root.Flex.Items[0].Collapsed = true
			}
			tab.View(31, 9)

			want := []struct {
				x      int
				y      int
				length int
			}{{x: 15, y: 0, length: 9}}
			if collapsed {
				want = []struct {
					x      int
					y      int
					length int
				}{{x: 15, y: 0, length: 9}}
			} else {
				want = append(want, struct {
					x      int
					y      int
					length int
				}{x: 23, y: 0, length: 9})
			}
			if len(tab.lastBorders) != len(want) {
				t.Fatalf("got borders %+v, want %d", tab.lastBorders, len(want))
			}
			for i, border := range tab.lastBorders {
				if border.Direction != Vertical || border.X != want[i].x || border.Y != want[i].y || border.Length != want[i].length {
					t.Errorf("border %d = %+v, want vertical x=%d y=%d length=%d", i, border, want[i].x, want[i].y, want[i].length)
				}
			}
		})
	}
}

func TestNestedFlexDragUsesLocalBoundaryIndexAndPersists(t *testing.T) {
	a := &geometryTestPanel{name: "A"}
	b := &geometryTestPanel{name: "B"}
	right := &geometryTestPanel{name: "R"}
	tab := NewTab("nested-flex-drag")
	tab.SetRootPanel(a)
	tab.SplitVertical(a, 0.5, right)
	tab.FlexRow(a, []FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 1}})
	tab.Update(tea.WindowSizeMsg{Width: 40, Height: 10})
	tab.View(40, 10)

	var boundary BorderHit
	found := false
	for _, border := range tab.lastBorders {
		if border.Flex != nil {
			boundary = border
			found = true
			break
		}
	}
	if !found {
		t.Fatal("nested flex boundary not found")
	}
	if boundary.FlexIndex != 0 || boundary.Bounds.W != 19 {
		t.Fatalf("nested boundary = %+v, want local index 0 in width 19", boundary)
	}

	tab.HandleMouse(tea.MouseMsg{X: boundary.X, Y: 2, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if tab.flexDragging != boundary.Flex || tab.flexDragIdx != 0 {
		t.Fatalf("drag target flex=%p index=%d, want %p index 0", tab.flexDragging, tab.flexDragIdx, boundary.Flex)
	}
	tab.HandleMouse(tea.MouseMsg{X: boundary.X + 5, Y: 2, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft})
	if b.lastResize.Width != 4 || a.lastResize.Width != 14 {
		t.Fatalf("after drag resize A=%+v B=%+v, want widths 14 and 4", a.lastResize, b.lastResize)
	}
	tab.HandleMouse(tea.MouseMsg{X: boundary.X + 5, Y: 2, Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft})
	tab.View(40, 10)
	plain := ansi.Strip(strings.Split(tab.View(40, 10), "\n")[0])
	markerByte := strings.Index(plain, "B")
	if !strings.HasPrefix(plain, "A") || markerByte < 0 || ansi.StringWidth(plain[:markerByte]) != 15 {
		t.Fatalf("flex drag did not persist on subsequent render: %q", plain)
	}
}

func TestNestedSplitDraggingUsesItsOwnRectangle(t *testing.T) {
	builders := []struct {
		name  string
		setup func() (*Tab, *SplitConfig, int, int)
	}{
		{
			name: "split inside split",
			setup: func() (*Tab, *SplitConfig, int, int) {
				a, b, c := geometryPanels()
				tab := NewTab("split")
				tab.SetRootPanel(a)
				tab.SplitVertical(a, 0.5, b)
				tab.SplitVertical(b, 0.5, c)
				return tab, tab.root.Split.Second.Split, 40, 20
			},
		},
		{
			name: "split inside flex row",
			setup: func() (*Tab, *SplitConfig, int, int) {
				a, b, c := geometryPanels()
				tab := NewTab("flex-row")
				tab.SetRootPanel(a)
				tab.FlexRow(a, []FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 1}})
				tab.SplitVertical(b, 0.5, c)
				return tab, tab.root.Flex.Items[1].Node.Split, 40, 20
			},
		},
		{
			name: "split inside flex column",
			setup: func() (*Tab, *SplitConfig, int, int) {
				a, b, c := geometryPanels()
				tab := NewTab("flex-column")
				tab.SetRootPanel(a)
				tab.FlexColumn(a, []FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 1}})
				tab.SplitHorizontal(b, 0.5, c)
				return tab, tab.root.Flex.Items[1].Node.Split, 20, 40
			},
		},
	}

	for _, builder := range builders {
		t.Run(builder.name, func(t *testing.T) {
			tab, target, width, height := builder.setup()
			tab.Update(tea.WindowSizeMsg{Width: width, Height: height})
			tab.View(width, height)
			var dragged BorderHit
			for _, border := range tab.lastBorders {
				if border.Split == target {
					dragged = border
					break
				}
			}
			if dragged.Split == nil {
				t.Fatal("nested split border not found")
			}

			x, y := dragged.X, dragged.Y
			if dragged.Direction == Vertical {
				y += dragged.Length / 2
			} else {
				x += dragged.Length / 2
			}
			tab.HandleMouse(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
			if tab.dragging != target {
				t.Fatal("nested split drag did not start")
			}

			if dragged.Direction == Vertical {
				x++
			} else {
				y++
			}
			expected := float64(x-dragged.Bounds.X) / float64(max(1, dragged.Bounds.W-1))
			if dragged.Direction == Horizontal {
				expected = float64(y-dragged.Bounds.Y) / float64(max(1, dragged.Bounds.H-1))
			}
			tab.HandleMouse(tea.MouseMsg{X: x, Y: y, Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft})
			if math.Abs(target.Fraction-clampFraction(expected)) > 1e-9 {
				t.Fatalf("nested fraction=%f, want %f from rectangle %+v", target.Fraction, clampFraction(expected), dragged.Bounds)
			}
		})
	}
}

func geometryPanels() (*geometryTestPanel, *geometryTestPanel, *geometryTestPanel) {
	return &geometryTestPanel{name: "A"}, &geometryTestPanel{name: "B"}, &geometryTestPanel{name: "C"}
}
