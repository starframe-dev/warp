package integration_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type layoutResizePanel struct {
	name    string
	width   int
	height  int
	resizes []warp.ResizeMsg
	mouse   []tea.MouseMsg
}

func (p *layoutResizePanel) View(width, height int) string {
	p.width, p.height = width, height
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := make([]string, height)
	for i := range lines {
		lines[i] = p.name + strings.Repeat(" ", max(0, width-len(p.name)))
	}
	return strings.Join(lines, "\n")
}

func (p *layoutResizePanel) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case warp.ResizeMsg:
		p.resizes = append(p.resizes, msg)
	case tea.MouseMsg:
		p.mouse = append(p.mouse, msg)
	}
	return nil
}

func (p *layoutResizePanel) Elements(width, height int) []warp.Element {
	return []warp.Element{{Role: "status", Name: p.name, Bounds: warp.Bounds{W: width, H: height}}}
}

func layoutMouse(x, y int, action tea.MouseAction) tea.MouseMsg {
	return tea.MouseMsg{X: x, Y: y, Button: tea.MouseButtonLeft, Action: action}
}

func TestNestedSplitFlexResizeDragCollapseAndHitTesting(t *testing.T) {
	tg := warp.NewTabGroup(warp.TabNone)
	tab := tg.ActiveTab()
	left := &layoutResizePanel{name: "left"}
	top := &layoutResizePanel{name: "top"}
	bottom := &layoutResizePanel{name: "bottom"}
	tab.SetRootPanel(left)
	tab.SplitVertical(left, .5, top)
	tab.FlexColumn(top, []warp.FlexItemSpec{{Panel: top, Grow: 1}, {Panel: bottom, Grow: 1}})

	const width, height = 40, 16
	_ = tg.View(width, height)
	if left.width != 19 || left.height != height {
		t.Fatalf("initial left size=%dx%d, want 19x%d", left.width, left.height, height)
	}
	if top.width != 20 || bottom.width != 20 || top.height+bottom.height != height-1 {
		t.Fatalf("nested flex sizes top=%dx%d bottom=%dx%d, want widths 20 and one border row", top.width, top.height, bottom.width, bottom.height)
	}

	elements := tab.Elements(width, height)
	var leftElement, topElement, bottomElement warp.Element
	for _, element := range elements {
		switch element.Name {
		case "left":
			leftElement = element
		case "top":
			topElement = element
		case "bottom":
			bottomElement = element
		}
	}
	if leftElement.Bounds != (warp.Bounds{W: 19, H: height}) || topElement.Bounds != (warp.Bounds{X: 20, W: 20, H: 7}) || bottomElement.Bounds != (warp.Bounds{X: 20, Y: 8, W: 20, H: 8}) {
		t.Fatalf("element bounds disagree with layout: left=%+v top=%+v bottom=%+v", leftElement.Bounds, topElement.Bounds, bottomElement.Bounds)
	}
	view := tg.View(width, height)
	if !strings.Contains(view, "│") || !strings.Contains(view, "─") {
		t.Fatalf("render omitted split/flex boundaries: %q", view)
	}

	// A click in each content rectangle must hit the corresponding panel with
	// coordinates translated to that panel's local origin.
	tg.Update(layoutMouse(0, 0, tea.MouseActionPress))
	tg.Update(layoutMouse(21, 0, tea.MouseActionPress))
	tg.Update(layoutMouse(21, 9, tea.MouseActionPress))
	if len(left.mouse) == 0 || left.mouse[len(left.mouse)-1].X != 0 || left.mouse[len(left.mouse)-1].Y != 0 {
		t.Fatal("left panel hit test did not report panel-relative coordinates")
	}
	if len(top.mouse) == 0 || len(bottom.mouse) == 0 {
		t.Fatal("nested flex hit testing did not reach both panels")
	}

	// Drag the root separator right. Live layout updates and resize messages
	// must describe the same leaf allocations.
	tg.Update(layoutMouse(19, 4, tea.MouseActionPress))
	tg.Update(layoutMouse(25, 4, tea.MouseActionMotion))
	if len(left.resizes) == 0 || len(top.resizes) == 0 || len(bottom.resizes) == 0 {
		t.Fatal("drag did not deliver ResizeMsg to every leaf panel")
	}
	_ = tab.View(width, height)
	if left.width != 25 || top.width != 14 || bottom.width != 14 {
		t.Fatalf("dragged sizes left=%d top=%d bottom=%d, want 25,14,14", left.width, top.width, bottom.width)
	}
	tg.Update(layoutMouse(25, 4, tea.MouseActionRelease))
	if len(left.resizes) < 2 {
		t.Fatal("drag release did not broadcast final resize")
	}

	if !tab.Collapse(left, 2) {
		t.Fatal("Collapse failed for split leaf")
	}
	_ = tg.View(width, height)
	if left.width != 2 {
		t.Fatalf("collapsed panel width=%d, want 2", left.width)
	}
	if !tab.Expand(left) {
		t.Fatal("Expand failed for collapsed split leaf")
	}
	_ = tg.View(width, height)
	if left.width != 25 || top.width != 14 {
		t.Fatalf("expand did not restore dragged allocation: left=%d top=%d", left.width, top.width)
	}
}
