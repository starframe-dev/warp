package e2e_test

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type layoutE2EPanel struct {
	name        string
	width       int
	height      int
	lastContent string
}

func (p *layoutE2EPanel) View(width, height int) string {
	p.width, p.height = width, height
	p.lastContent = fmt.Sprintf("[%s]", p.name)
	return p.lastContent
}

func (*layoutE2EPanel) Update(tea.Msg) tea.Cmd { return nil }

func TestSplitAndFlexLayoutsResizeCollapseAndRestore(t *testing.T) {
	const width, height = 60, 16

	left := &layoutE2EPanel{name: "left"}
	right := &layoutE2EPanel{name: "right"}
	tab := warp.NewTab("split")
	tab.SetRootPanel(left)
	tab.SplitVertical(left, 0.5, right)
	tab.Update(tea.WindowSizeMsg{Width: width, Height: height})
	tab.View(width, height)

	if left.width <= 0 || right.width <= 0 || left.height != height || right.height != height {
		t.Fatalf("split panels rendered at left=%dx%d and right=%dx%d; want positive widths and height %d", left.width, left.height, right.width, right.height, height)
	}
	if left.width >= right.width+2 || right.width >= left.width+2 {
		t.Fatalf("initial split is not approximately even: left=%d right=%d", left.width, right.width)
	}

	// Drag the visible split boundary to the right and verify the panels resize.
	tab.HandleMouse(tea.MouseMsg{X: left.width, Y: height / 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionPress})
	tab.HandleMouse(tea.MouseMsg{X: 41, Y: height / 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionMotion})
	tab.HandleMouse(tea.MouseMsg{X: 41, Y: height / 2, Button: tea.MouseButtonLeft, Action: tea.MouseActionRelease})
	tab.View(width, height)
	if left.width <= right.width {
		t.Fatalf("dragging divider right did not enlarge first panel: left=%d right=%d", left.width, right.width)
	}
	fraction, ok := tab.GetSplitFraction(left)
	if !ok || fraction <= 0.5 {
		t.Fatalf("split fraction after drag = %v, found=%v; want > 0.5", fraction, ok)
	}

	// Collapse the first side to a narrow fixed width, then restore its saved fraction.
	if !tab.Collapse(left, 2) {
		t.Fatal("Collapse(left, 2) did not find the split")
	}
	tab.View(width, height)
	collapsedWidth := left.width
	if collapsedWidth > 3 || right.width < width-6 {
		t.Fatalf("collapsed panel sizes left=%d right=%d; want narrow first side and expanded second side", left.width, right.width)
	}
	if !tab.Expand(left) {
		t.Fatal("Expand(left) did not restore the split")
	}
	tab.View(width, height)
	if left.width <= collapsedWidth || left.width <= right.width {
		t.Fatalf("expanded sizes left=%d right=%d; want saved larger first side restored", left.width, right.width)
	}

	// A flex row distributes its available width according to grow weights.
	one := &layoutE2EPanel{name: "one"}
	two := &layoutE2EPanel{name: "two"}
	three := &layoutE2EPanel{name: "three"}
	flex := warp.NewTab("flex")
	flex.SetRootPanel(one)
	flex.FlexRow(one, []warp.FlexItemSpec{
		{Panel: one, Grow: 1},
		{Panel: two, Grow: 2},
		{Panel: three, Grow: 1},
	})
	flex.Update(tea.WindowSizeMsg{Width: width, Height: height})
	flex.View(width, height)
	if one.width <= 0 || two.width <= 0 || three.width <= 0 {
		t.Fatalf("flex children must all occupy space: widths=%d,%d,%d", one.width, two.width, three.width)
	}
	if two.width <= one.width || two.width <= three.width {
		t.Fatalf("flex grow weights were not reflected in widths: %d,%d,%d", one.width, two.width, three.width)
	}
	if one.height != height || two.height != height || three.height != height {
		t.Fatalf("flex child heights = %d,%d,%d; want %d", one.height, two.height, three.height, height)
	}
}
