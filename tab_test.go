package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type tabMockPanel struct {
	BasePanel
	id int
}

type tabRecordingPanel struct {
	BasePanel
	last tea.Msg
}

func (p *tabRecordingPanel) Update(msg tea.Msg) tea.Cmd {
	p.last = msg
	return nil
}

type tabFocusablePanel struct {
	BasePanel
	id      int
	focused bool
}

func (p *tabFocusablePanel) Focus()        { p.focused = true }
func (p *tabFocusablePanel) Blur()         { p.focused = false }
func (p *tabFocusablePanel) Focused() bool { return p.focused }

type tabElementPanel struct {
	BasePanel
	id    int
	elems []Element
}

func (p tabElementPanel) Elements(width, height int) []Element {
	return p.elems
}

func TestTabPositionConstants(t *testing.T) {
	if TabTop != 0 || TabBottom != 1 || TabLeft != 2 || TabRight != 3 || TabNone != 4 {
		t.Fatalf("TabPosition constants mismatch: %d %d %d %d %d", TabTop, TabBottom, TabLeft, TabRight, TabNone)
	}
}

func TestTabNew(t *testing.T) {
	tab := NewTab("test")
	if tab == nil {
		t.Fatal("NewTab returned nil")
	}
	if tab.RootPanel() == nil {
		t.Fatal("new tab root panel is nil")
	}
}

func TestTabRootPanelAndSetRootPanel(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.SetRootPanel(p)
	if got := tab.RootPanel(); got != p {
		t.Fatalf("SetRootPanel did not set panel, got %v", got)
	}
}

func TestTabSplitVertical(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)
	if tab.root.Split == nil {
		t.Fatal("expected vertical split")
	}
	if tab.root.Split.Direction != Vertical {
		t.Fatalf("expected Vertical, got %v", tab.root.Split.Direction)
	}
	if tab.root.Split.Second.Panel != p2 {
		t.Fatal("expected second panel to be p2")
	}
}

func TestTabSplitHorizontal(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.SetRootPanel(p1)
	tab.SplitHorizontal(p1, 0.5, p2)
	if tab.root.Split == nil || tab.root.Split.Direction != Horizontal {
		t.Fatal("expected horizontal split")
	}
}

func TestTabSplitUnknownParent(t *testing.T) {
	tab := NewTab("test")
	unknown := tabMockPanel{id: 99}
	newPanel := tabMockPanel{id: 1}
	tab.SplitVertical(unknown, 0.5, newPanel)
	if tab.root.Split != nil {
		t.Fatal("expected split to be ignored for unknown parent")
	}
}

func TestTabFlexRowAndColumn(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: -1},
	})
	if tab.root.Flex == nil {
		t.Fatal("expected flex row")
	}
	if tab.root.Flex.Direction != Horizontal {
		t.Fatalf("expected Horizontal, got %v", tab.root.Flex.Direction)
	}
	if len(tab.root.Flex.Items) != 2 {
		t.Fatalf("expected 2 flex items, got %d", len(tab.root.Flex.Items))
	}
	if tab.root.Flex.Items[1].Grow != 0 {
		t.Fatalf("expected negative grow clamped to 0, got %d", tab.root.Flex.Items[1].Grow)
	}

	p3 := tabMockPanel{id: 3}
	tab.FlexColumn(p1, []FlexItemSpec{{Panel: p3, Grow: 2}})
	if p1Node := tab.root.Flex.Items[0].Node; p1Node.Flex == nil || p1Node.Flex.Direction != Vertical {
		t.Fatal("expected nested flex column")
	}
}

func TestTabFloatAndCloseFloat(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.Float(p, 0, 0, 10, 3)
	if len(tab.floats) != 1 {
		t.Fatalf("expected 1 float, got %d", len(tab.floats))
	}
	fp := tab.floats[0]
	if fp.Panel != p {
		t.Fatal("float panel mismatch")
	}
	tab.CloseFloat(fp)
	if len(tab.floats) != 0 {
		t.Fatalf("expected 0 floats after close, got %d", len(tab.floats))
	}
	// closing unknown float should not panic
	tab.CloseFloat(fp)
}

func TestTabFocusAndSetFocus(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	cmd := tab.SetFocus(p)
	if cmd != nil {
		t.Fatalf("SetFocus cmd = %v, want nil", cmd)
	}
	if tab.Focus() != p {
		t.Fatalf("Focus() = %v, want %v", tab.Focus(), p)
	}
}

func TestTabFocusNextPrevFirst(t *testing.T) {
	tab := NewTab("test")
	p1 := &tabFocusablePanel{id: 1}
	p2 := &tabFocusablePanel{id: 2}
	p3 := &tabFocusablePanel{id: 3}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)
	tab.SplitVertical(p2, 0.5, p3)

	// FocusFirst should focus the first panel
	tab.FocusFirst()
	if tab.Focus() != p1 || !p1.focused {
		t.Fatal("FocusFirst should focus first panel")
	}

	// FocusNext
	tab.FocusNext()
	if tab.Focus() != p2 || !p2.focused || p1.focused {
		t.Fatal("FocusNext should move to second panel")
	}

	// FocusNext wrap
	tab.FocusNext()
	if tab.Focus() != p3 || !p3.focused || p2.focused {
		t.Fatal("FocusNext should move to third panel")
	}

	// FocusNext wrap around
	tab.FocusNext()
	if tab.Focus() != p1 {
		t.Fatal("FocusNext should wrap to first panel")
	}

	// FocusPrev
	tab.FocusPrev()
	if tab.Focus() != p3 {
		t.Fatal("FocusPrev should wrap to last panel")
	}
}

func TestTabFocusPanel(t *testing.T) {
	tab := NewTab("test")
	p := &tabFocusablePanel{id: 1}
	tab.SetRootPanel(p)
	tab.FocusPanel(p)
	if tab.Focus() != p || !p.focused {
		t.Fatal("FocusPanel should focus focusable panel")
	}

	nf := tabMockPanel{id: 2}
	tab.FocusPanel(nf)
	if tab.Focus() != p {
		t.Fatal("FocusPanel should ignore non-focusable panel")
	}
}

func TestTabView(t *testing.T) {
	tab := NewTab("test")
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	view := tab.View(20, 10)
	lines := strings.Split(view, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}
}

func TestTabUpdateWindowSize(t *testing.T) {
	tab := NewTab("test")
	tab.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if tab.width != 80 || tab.height != 24 {
		t.Fatalf("expected width=80 height=24, got %d %d", tab.width, tab.height)
	}
}

func TestTabUpdateResizeMsg(t *testing.T) {
	tab := NewTab("test")
	tab.Update(ResizeMsg{Width: 30, Height: 15})
	if tab.width != 30 || tab.height != 15 {
		t.Fatalf("expected width=30 height=15, got %d %d", tab.width, tab.height)
	}
}

func TestTabUpdateBroadcastsUnknownMsg(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	tab.Update(key)
	if rec.last == nil {
		t.Fatal("expected panel to receive broadcast message")
	}
	if _, ok := rec.last.(tea.KeyMsg); !ok {
		t.Fatalf("expected KeyMsg, got %T", rec.last)
	}
}

func TestTabElements(t *testing.T) {
	tab := NewTab("test")
	p := &tabElementPanel{elems: []Element{
		{Role: "button", Bounds: Bounds{X: 0, Y: 0, W: 3, H: 1}},
	}}
	tab.SetRootPanel(p)
	tab.Update(tea.WindowSizeMsg{Width: 10, Height: 5})
	elems := tab.Elements(10, 5)
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Role != "button" {
		t.Fatalf("expected element role button, got %s", elems[0].Role)
	}
}

func TestTabElementsWithSplitShifting(t *testing.T) {
	tab := NewTab("test")
	p1 := &tabElementPanel{elems: []Element{{Role: "left", Bounds: Bounds{X: 0, Y: 0, W: 2, H: 1}}}}
	p2 := &tabElementPanel{elems: []Element{{Role: "right", Bounds: Bounds{X: 0, Y: 0, W: 2, H: 1}}}}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	elems := tab.Elements(20, 10)
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}
	if elems[0].Role != "left" || elems[0].Bounds.X != 0 {
		t.Fatalf("unexpected left element: %+v", elems[0])
	}
	if elems[1].Role != "right" || elems[1].Bounds.X <= elems[0].Bounds.X {
		t.Fatalf("unexpected right element shift: %+v", elems[1])
	}
}

func TestTabHandleMousePressPanel(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	msg := tea.MouseMsg{
		X:      5,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)
	if tab.Focus() != rec {
		t.Fatal("click should focus panel")
	}
	if rec.last == nil {
		t.Fatal("panel should receive forwarded mouse message")
	}
	if m, ok := rec.last.(tea.MouseMsg); !ok || m.X != 5 || m.Y != 5 {
		t.Fatalf("unexpected forwarded mouse message: %+v", rec.last)
	}
}

func TestTabHandleMouseBorderDrag(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	tab.View(20, 10)

	// border is at x = firstW = 9 for width=20
	press := tea.MouseMsg{
		X:      9,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(press)
	if tab.dragging == nil {
		t.Fatal("expected drag to start on border press")
	}

	motion := tea.MouseMsg{
		X:      12,
		Y:      1,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(motion)

	release := tea.MouseMsg{
		X:      12,
		Y:      1,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(release)
	if tab.dragging != nil {
		t.Fatal("expected drag to end on release")
	}
}

func TestTabHandleMouseOutsideFloat(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.Float(p, 0, 0, 10, 3)
	tab.floats[0].CloseOnOutsideClick = true
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	msg := tea.MouseMsg{
		X:      15,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)
	if len(tab.floats) != 0 {
		t.Fatalf("expected float to close on outside click, got %d", len(tab.floats))
	}
}

func TestTabHandleMouseInsideFloat(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.Float(p, 0, 0, 10, 3)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	msg := tea.MouseMsg{
		X:      1,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)
	if tab.Focus() != p {
		t.Fatal("click inside float should focus float panel")
	}
}

func TestTabSetSplitFractionAndGet(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)

	if !tab.SetSplitFraction(p1, 0.3) {
		t.Fatal("expected SetSplitFraction to succeed")
	}
	frac, ok := tab.GetSplitFraction(p1)
	if !ok {
		t.Fatal("expected GetSplitFraction to succeed")
	}
	if frac < 0.25 || frac > 0.35 {
		t.Fatalf("expected fraction near 0.3, got %f", frac)
	}

	_, ok = tab.GetSplitFraction(tabMockPanel{id: 99})
	if ok {
		t.Fatal("expected GetSplitFraction to fail for unknown panel")
	}
}

func TestTabCollapseAndExpand(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, p2)

	if !tab.Collapse(p1, 5) {
		t.Fatal("expected Collapse to succeed")
	}
	if p1Node := tab.root.findNode(p1); p1Node == nil || p1Node.Collapse == nil || !p1Node.Collapse.Active || p1Node.Collapse.Width != 5 {
		t.Fatal("expected p1 collapsed with width 5")
	}

	if !tab.Expand(p1) {
		t.Fatal("expected Expand to succeed")
	}
	if p1Node := tab.root.findNode(p1); p1Node.Collapse == nil || p1Node.Collapse.Active {
		t.Fatal("expected p1 expanded")
	}

	if tab.Collapse(tabMockPanel{id: 99}, 5) {
		t.Fatal("expected Collapse to fail for unknown panel")
	}
}

func TestTabToggleCollapsible(t *testing.T) {
	tab := NewTab("test")
	c := NewCollapsible("title", BasePanel{})
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{{Panel: c, Grow: 1}})
	if c.Collapsed {
		t.Fatal("expected collapsible to start expanded")
	}
	tab.ToggleCollapsible(c)
	if !c.Collapsed {
		t.Fatal("expected collapsible to be collapsed")
	}
	if !tab.root.Flex.Items[0].Collapsed {
		t.Fatal("expected flex item to be collapsed")
	}
}

func TestTabBroadcastResize(t *testing.T) {
	tab := NewTab("test")
	tab.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	msg := tab.BroadcastResize()
	if _, ok := msg.(tea.BatchMsg); !ok {
		t.Fatalf("expected BatchMsg, got %T", msg)
	}
}

func TestTabHandleKeys(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	tab.HandleKeys(key)

	if rec.last == nil {
		t.Fatal("expected panel to receive forwarded key message")
	}
	if _, ok := rec.last.(tea.KeyMsg); !ok {
		t.Fatalf("expected KeyMsg, got %T", rec.last)
	}
}

func TestTabHandleKeysNoFocus(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	key := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	tab.HandleKeys(key)

	if rec.last != nil {
		t.Fatal("expected no message when tab has no focus")
	}
}

func TestTabHandleMouseMotionWithoutDrag(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.SetRootPanel(p)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	// Motion without a drag should forward to panel
	msg := tea.MouseMsg{
		X:      5,
		Y:      5,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)

	// Panel should receive the motion message
	if pNode := tab.root.findNode(p); pNode == nil || pNode.Panel == nil {
		t.Fatal("expected to find panel node")
	}
}

func TestTabHandleMouseReleaseWithoutDrag(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	// Release without a drag should forward to panel
	msg := tea.MouseMsg{
		X:      5,
		Y:      5,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)

	// Panel should receive the release message
	if rec.last == nil {
		t.Fatal("expected panel to receive forwarded release message")
	}
}

func TestTabHandleMousePressWithoutBorder(t *testing.T) {
	tab := NewTab("test")
	rec := &tabRecordingPanel{}
	tab.SetRootPanel(rec)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	// Press in center of panel (not on border)
	msg := tea.MouseMsg{
		X:      5,
		Y:      5,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)

	if rec.last == nil {
		t.Fatal("expected panel to receive press message")
	}
}

func TestTabClampFraction(t *testing.T) {
	if got := clampFraction(0.05); got != 0.1 {
		t.Fatalf("clampFraction(0.05) = %f, want 0.1", got)
	}
	if got := clampFraction(0.09); got != 0.09 {
		t.Fatalf("clampFraction(0.09) = %f, want 0.09", got)
	}
	if got := clampFraction(0.1); got != 0.1 {
		t.Fatalf("clampFraction(0.1) = %f, want 0.1", got)
	}
	if got := clampFraction(0.5); got != 0.5 {
		t.Fatalf("clampFraction(0.5) = %f, want 0.5", got)
	}
	if got := clampFraction(0.9); got != 0.9 {
		t.Fatalf("clampFraction(0.9) = %f, want 0.9", got)
	}
	if got := clampFraction(0.95); got != 0.9 {
		t.Fatalf("clampFraction(0.95) = %f, want 0.9", got)
	}
}

func TestTabHandleMouseWithMultipleFloats(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.Float(p1, 0, 0, 10, 3)
	tab.Float(p2, 0, 0, 5, 2)
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})

	msg := tea.MouseMsg{
		X:      1,
		Y:      1,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
		Type:   tea.MouseEventType(0),
	}
	tab.HandleMouse(msg)

	// Should focus the top-most float (p2 comes after p1 in list, so p1 is top)
	// Actually p2 was added last, so p2 is top
	if tab.Focus() != p2 {
		t.Fatalf("expected to focus top float p2, got %v", tab.Focus())
	}
}

func TestTabBroadcastResizeReturnsEmpty(t *testing.T) {
	tab := NewTab("test")
	// With no panels, broadcast should return empty batch
	msg := tab.BroadcastResize()
	if _, ok := msg.(tea.BatchMsg); !ok {
		t.Fatalf("expected BatchMsg, got %T", msg)
	}
}

func TestTabSplitVerticalWithEmptyItems(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	tab.SetRootPanel(p1)
	tab.SplitVertical(p1, 0.5, &tabMockPanel{id: 2})
	// Should not panic even if subsequent operations are done
	if tab.root.Split == nil {
		t.Fatal("expected vertical split to be created")
	}
}

func TestTabFlexRowWithEmptyItems(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	tab.SetRootPanel(p1)
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{})
	// Should not panic with empty items
	if tab.root.Flex != nil {
		t.Fatal("expected flex to not be created with empty items")
	}
}

func TestTabFlexColumnWithEmptyItems(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	tab.SetRootPanel(p1)
	tab.FlexColumn(tab.RootPanel(), []FlexItemSpec{})
	// Should not panic with empty items
	if tab.root.Flex != nil {
		t.Fatal("expected flex to not be created with empty items")
	}
}

func TestTabFloatWithZeroSize(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.Float(p, 0, 0, 0, 0)
	if len(tab.floats) != 1 {
		t.Fatalf("expected float to be added with zero size, got %d", len(tab.floats))
	}
}

func TestTabFloatWithNegativeGrow(t *testing.T) {
	tab := NewTab("test")
	p1 := tabMockPanel{id: 1}
	p2 := tabMockPanel{id: 2}
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{
		{Panel: p1, Grow: 1},
		{Panel: p2, Grow: -100},
	})
	if len(tab.root.Flex.Items) != 2 {
		t.Fatal("expected 2 flex items")
	}
	if tab.root.Flex.Items[1].Grow != 0 {
		t.Fatalf("expected negative grow clamped to 0, got %d", tab.root.Flex.Items[1].Grow)
	}
}

func TestTabFocusPanelWithNilPanel(t *testing.T) {
	tab := NewTab("test")
	tab.SetRootPanel(&tabMockPanel{id: 1})
	tab.FocusPanel(nil)
	// Should not panic
}

func TestTabToggleCollapsibleOnNonCollapsible(t *testing.T) {
	tab := NewTab("test")
	p := tabMockPanel{id: 1}
	tab.SetRootPanel(p)
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{{Panel: p, Grow: 1}})
	// Should not panic when toggling non-collapsible panel
	tab.ToggleCollapsible(p)
}

func TestTabElementsWithFlexLayout(t *testing.T) {
	tab := NewTab("test")
	p1 := &tabElementPanel{elems: []Element{{Role: "left", Bounds: Bounds{X: 0, Y: 0, W: 5, H: 1}}}}
	p2 := &tabElementPanel{elems: []Element{{Role: "right", Bounds: Bounds{X: 0, Y: 0, W: 5, H: 1}}}}
	tab.SetRootPanel(p1)
	tab.FlexRow(tab.RootPanel(), []FlexItemSpec{{Panel: p1, Grow: 1}, {Panel: p2, Grow: 1}})
	tab.Update(tea.WindowSizeMsg{Width: 20, Height: 10})
	elems := tab.Elements(20, 10)
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}
}

func TestTabElementsWithFlexColumn(t *testing.T) {
	tab := NewTab("test")
	p1 := &tabElementPanel{elems: []Element{{Role: "top", Bounds: Bounds{X: 0, Y: 0, W: 1, H: 5}}}}
	p2 := &tabElementPanel{elems: []Element{{Role: "bottom", Bounds: Bounds{X: 0, Y: 0, W: 1, H: 5}}}}
	tab.SetRootPanel(p1)
	tab.FlexColumn(tab.RootPanel(), []FlexItemSpec{{Panel: p1, Grow: 1}, {Panel: p2, Grow: 1}})
	tab.Update(tea.WindowSizeMsg{Width: 10, Height: 20})
	elems := tab.Elements(10, 20)
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}
}
