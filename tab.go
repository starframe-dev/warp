package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TabPosition defines where the tab bar is rendered.
type TabPosition int

const (
	TabTop TabPosition = iota
	TabBottom
	TabLeft
	TabRight
	TabNone
)

// Tab represents a single tab with its panel tree, float panes, and focus state.
type Tab struct {
	name    string
	root    *Node
	focused Panel
	floats  []*FloatPane
	parent  *TabGroup
	width   int
	height  int

	// Drag state
	dragging     *SplitConfig
	flexDragging *FlexConfig
	flexDragIdx  int // index of the border being dragged in flex
	lastBorders  []BorderHit
}

func newTab(name string, parent *TabGroup) *Tab {
	return &Tab{
		name:   name,
		root:   &Node{Panel: &emptyPanel{}},
		parent: parent,
	}
}

// NewTab creates a standalone Tab with no parent TabGroup.
// Useful for embedding a warp layout inside a Panel.
func NewTab(name string) *Tab {
	return &Tab{
		name: name,
		root: &Node{Panel: &emptyPanel{}},
	}
}

func (t *Tab) ensureRoot() {
	if t.root == nil {
		t.root = &Node{Panel: &emptyPanel{}}
	}
}

// RootPanel returns the root panel of this tab.
// Useful as the parent argument for the first Split call.
func (t *Tab) RootPanel() Panel {
	t.ensureRoot()
	return t.root.Panel
}

// SetRootPanel replaces the root panel of this tab.
func (t *Tab) SetRootPanel(panel Panel) {
	if isNilPanel(panel) {
		panel = &emptyPanel{}
	}
	t.setFocus(nil)
	t.root = &Node{Panel: panel}
}

// SplitVertical splits the panel vertically (left/right).
// fraction is the share for the left panel (0.0–1.0).
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel) {
	t.ensureRoot()
	if parent == nil {
		return
	}
	node := t.root.findNode(parent)
	if node == nil || node.Panel == nil {
		return
	}
	oldPanel := node.Panel
	if isNilPanel(newPanel) {
		newPanel = &emptyPanel{}
	}
	collapse := node.Collapse
	*node = Node{
		Split: &SplitConfig{
			Direction: Vertical,
			Fraction:  clampFraction(fraction),
			First:     &Node{Panel: oldPanel},
			Second:    &Node{Panel: newPanel},
		},
		Collapse: collapse,
	}
}

// SetSplitCollapse configures a collapse symbol on the border of a vertical split.
// The "<" symbol appears at the given row (0-indexed) instead of "│".
// When clicked, onCollapse is called. It should return a tea.Cmd to trigger re-render.
func (t *Tab) SetSplitCollapse(parent Panel, collapseRow int, onCollapse func() tea.Cmd) {
	// Find the parent node that has a Split containing the given panel.
	node := t.root.findSplitParent(parent)
	if node == nil || node.Split == nil {
		return
	}
	node.Split.CollapseRow = collapseRow
	// Wrap the callback to also toggle the node's Collapse state.
	// The first child of the split is the panel that should collapse.
	firstChild := node.Split.First
	node.Split.OnCollapse = func() tea.Cmd {
		if firstChild.Collapse == nil {
			firstChild.Collapse = &NodeCollapse{
				Active: true,
				Width:  1,
			}
		} else {
			firstChild.Collapse.Active = !firstChild.Collapse.Active
		}
		return onCollapse()
	}
}

// ToggleSplitCollapse toggles the collapse state of the split containing the given panel.
// This is used when the panel itself handles expand (e.g. clicking on a collapsed panel)
// and needs to sync the warp node's collapse state.
func (t *Tab) ToggleSplitCollapse(parent Panel) {
	node := t.root.findSplitParent(parent)
	if node == nil || node.Split == nil {
		return
	}
	firstChild := node.Split.First
	if firstChild.Collapse == nil {
		firstChild.Collapse = &NodeCollapse{
			Active: true,
			Width:  1,
		}
	} else {
		firstChild.Collapse.Active = !firstChild.Collapse.Active
	}
}

// SplitHorizontal splits the panel horizontally (top/bottom).
// fraction is the share for the top panel (0.0–1.0).
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel) {
	t.ensureRoot()
	if parent == nil {
		return
	}
	node := t.root.findNode(parent)
	if node == nil || node.Panel == nil {
		return
	}
	oldPanel := node.Panel
	if isNilPanel(newPanel) {
		newPanel = &emptyPanel{}
	}
	collapse := node.Collapse
	*node = Node{
		Split: &SplitConfig{
			Direction: Horizontal,
			Fraction:  clampFraction(fraction),
			First:     &Node{Panel: oldPanel},
			Second:    &Node{Panel: newPanel},
		},
		Collapse: collapse,
	}
}

// FlexItemSpec describes a panel and its flex-grow weight.
type FlexItemSpec struct {
	Panel Panel
	Grow  int
}

// FlexRow replaces the parent panel with a horizontal flex layout.
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec) {
	t.ensureRoot()
	node := t.root.findNode(parent)
	if node == nil {
		return
	}
	if len(items) == 0 {
		return
	}
	flexItems := make([]*FlexItem, len(items))
	for i, spec := range items {
		grow := spec.Grow
		if grow < 0 {
			grow = 0
		}
		panel := spec.Panel
		if isNilPanel(panel) {
			panel = &emptyPanel{}
		}
		flexItems[i] = &FlexItem{
			Node: &Node{Panel: panel},
			Grow: grow,
		}
	}
	collapse := node.Collapse
	*node = Node{
		Flex:     &FlexConfig{Direction: Horizontal, Items: flexItems},
		Collapse: collapse,
	}
}

// FlexColumn replaces the parent panel with a vertical flex layout.
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec) {
	t.ensureRoot()
	node := t.root.findNode(parent)
	if node == nil {
		return
	}
	if len(items) == 0 {
		return
	}
	flexItems := make([]*FlexItem, len(items))
	for i, spec := range items {
		grow := spec.Grow
		if grow < 0 {
			grow = 0
		}
		panel := spec.Panel
		if isNilPanel(panel) {
			panel = &emptyPanel{}
		}
		flexItems[i] = &FlexItem{
			Node: &Node{Panel: panel},
			Grow: grow,
		}
	}
	collapse := node.Collapse
	*node = Node{
		Flex:     &FlexConfig{Direction: Vertical, Items: flexItems},
		Collapse: collapse,
	}
}

// Float makes a panel floating above the layout.
func (t *Tab) Float(panel Panel, x, y, width, height int) {
	if isNilPanel(panel) || width <= 0 || height <= 0 {
		return
	}
	width = max(floatMinWidth, width)
	height = max(floatMinHeight, height)
	fp := &FloatPane{
		Panel:           panel,
		X:               max(0, x),
		Y:               max(0, y),
		Width:           width,
		Height:          height,
		Title:           "Float",
		preferredWidth:  width,
		preferredHeight: height,
	}
	fp.clampPosition(t.width, t.height)
	t.floats = append(t.floats, fp)
}

// CloseFloat removes a floating pane.
func (t *Tab) CloseFloat(fp *FloatPane) {
	for i, f := range t.floats {
		if f == fp {
			t.floats = append(t.floats[:i], t.floats[i+1:]...)
			return
		}
	}
}

// Focus returns the currently focused panel.
func (t *Tab) Focus() Panel {
	return t.focused
}

// SetFocus sets the focused panel and returns a command from its Update.
func (t *Tab) SetFocus(panel Panel) tea.Cmd {
	t.setFocus(panel)
	return nil
}

// FocusNext moves focus to the next focusable panel in the layout tree.
// The developer decides which key triggers this; warp does not bind Tab.
func (t *Tab) FocusNext() {
	t.focusStep(1)
}

// FocusPrev moves focus to the previous focusable panel.
func (t *Tab) FocusPrev() {
	t.focusStep(-1)
}

// FocusFirst moves focus to the first focusable panel.
func (t *Tab) FocusFirst() {
	focusables := collectFocusables(t.root)
	if len(focusables) == 0 {
		return
	}
	t.setFocusedFocusable(focusables[0])
}

// FocusPanel sets focus to a specific panel if it is focusable.
func (t *Tab) FocusPanel(panel Panel) {
	if f, ok := isFocusable(panel); ok {
		t.setFocusedFocusable(f)
	}
}

func (t *Tab) focusStep(delta int) {
	focusables := collectFocusables(t.root)
	if len(focusables) == 0 {
		return
	}
	idx := focusIndex(focusables, t.focused)
	idx += delta
	if idx < 0 {
		idx = len(focusables) - 1
	} else if idx >= len(focusables) {
		idx = 0
	}
	t.setFocusedFocusable(focusables[idx])
}

func (t *Tab) setFocusedFocusable(next Focusable) {
	t.setFocus(next)
}

func (t *Tab) setFocus(panel Panel) {
	if isNilPanel(panel) {
		panel = nil
	}
	if samePanel(t.focused, panel) {
		return
	}
	if current, ok := isFocusable(t.focused); ok {
		current.Blur()
	}
	if next, ok := isFocusable(panel); ok {
		next.Focus()
	}
	t.focused = panel
}

// View implements warp.Panel for Tab so it can be queried as an element provider.
func (t *Tab) View(width, height int) string {
	return t.renderContent(width, height)
}

// Update implements warp.Panel for Tab.
func (t *Tab) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.clampFloatsToViewport()
		// Send the panel-specific ResizeMsg to all leaves and collect any
		// commands they produce (e.g. starting side panels like ai-knowledge).
		resizeCmds := t.broadcastResize(t.root, 0, 0, t.width, t.height)
		// Keep backward compatibility for panels that still expect WindowSizeMsg.
		return tea.Batch(append(resizeCmds, t.broadcastMsg(msg)...)...)
	case ResizeMsg:
		// Container.View sends ResizeMsg to ensure child panels receive their
		// allocated size. Broadcast it to all leaf panels so they can start
		// their processes (e.g. ContextPanel starting ai-knowledge).
		t.width = msg.Width
		t.height = msg.Height
		t.clampFloatsToViewport()
		resizeCmds := t.broadcastResize(t.root, 0, 0, t.width, t.height)
		return tea.Batch(resizeCmds...)
	default:
		// Forward unknown messages (PtyOutputMsg, PtyReadyMsg, etc.) to all
		// leaf panels so emulators can receive them.
		return tea.Batch(t.broadcastMsg(msg)...)
	}
}

// renderContent renders the panel tree at the given dimensions.
func (t *Tab) renderContent(w, h int) string {
	t.clampFloats(w, h)
	layout := newLayout(t.root, layoutRect{w: max(0, w), h: max(0, h)})
	t.lastBorders = collectLayoutBorders(layout)
	lines := renderLayout(layout)

	// Render floats on top
	for _, fp := range t.floats {
		overlayFloat(lines, fp, w, h)
	}

	return strings.Join(lines, "\n")
}

// Elements returns elements from the panel tree, recursively accounting for
// splits/flex layouts so coordinates are relative to the tab content area.
func (t *Tab) clampFloatsToViewport() {
	t.clampFloats(t.width, t.height)
}

func (t *Tab) clampFloats(width, height int) {
	for _, fp := range t.floats {
		fp.clampPosition(width, height)
	}
}

func (t *Tab) Elements(w, h int) []Element {
	layout := newLayout(t.root, layoutRect{w: max(0, w), h: max(0, h)})
	return elementsFromLayout(layout)
}

func (t *Tab) elementsNode(node *Node, x, y, w, h int) []Element {
	layout := newLayout(node, layoutRect{x: x, y: y, w: max(0, w), h: max(0, h)})
	return elementsFromLayout(layout)
}

func (t *Tab) elementsFlex(flex *FlexConfig, x, y, w, h int) []Element {
	return t.elementsNode(&Node{Flex: flex}, x, y, w, h)
}

// HandleMouse processes mouse events for this tab with no offset and the
// tab's current content dimensions. This is the public entry point for
// embedded tabs (e.g. Container's innerTab) that need border dragging.
func (t *Tab) HandleMouse(msg tea.MouseMsg) tea.Cmd {
	return t.handleMouse(msg, 0, 0, t.width, t.height)
}

// handleMouse processes mouse events for this tab.
// offsetX, offsetY account for tab bar position.
// cw, ch are the content area dimensions.
func (t *Tab) handleMouse(msg tea.MouseMsg, offsetX, offsetY, cw, ch int) tea.Cmd {
	mx := msg.X - offsetX
	my := msg.Y - offsetY

	// Refresh the shared layout before processing mouse events so geometry changes are immediate.
	layout := newLayout(t.root, layoutRect{w: max(0, cw), h: max(0, ch)})
	t.lastBorders = collectLayoutBorders(layout)

	// Check float panes first (top z-order)
	hitFloat := false
	for i := len(t.floats) - 1; i >= 0; i-- {
		fp := t.floats[i]

		// Detect if cursor is inside this float before calling handleMouse
		// (handleMouse skips bounds check during active drag/resize)
		fp.clampPosition(cw, ch)
		inside := mx >= fp.X && mx < fp.X+fp.Width && my >= fp.Y && my < fp.Y+fp.Height

		cmd := fp.handleMouseWithin(msg, mx, my, cw, ch)
		if fp.CloseRequested {
			t.CloseFloat(fp)
			return nil
		}
		if cmd != nil {
			hitFloat = true
			// Bring to top on press inside float content
			if msg.Action == tea.MouseActionPress {
				t.floats = append(t.floats[:i], t.floats[i+1:]...)
				t.floats = append(t.floats, fp)
				t.setFocus(fp.Panel)
			}
			return cmd
		}
		if inside {
			hitFloat = true
			// Bring to top and consume the event when clicking inside
			// (title bar drag, edge resize, or close button)
			if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
				t.floats = append(t.floats[:i], t.floats[i+1:]...)
				t.floats = append(t.floats, fp)
				t.setFocus(fp.Panel)
			}
			return nil
		}
	}

	// Close floats that want auto-close on outside click
	if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && !hitFloat {
		for i := len(t.floats) - 1; i >= 0; i-- {
			if t.floats[i].CloseOnOutsideClick {
				t.CloseFloat(t.floats[i])
			}
		}
	}

	// Check border dragging
	switch msg.Button {
	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			// Check collapse symbol on borders first
			for _, bh := range t.lastBorders {
				if bh.Split != nil && bh.Split.CollapseRow >= 0 && bh.Split.OnCollapse != nil {
					if mx == bh.X && my == bh.Y+bh.Split.CollapseRow {
						return bh.Split.OnCollapse()
					}
				}
			}
			// Check border dragging
			for _, bh := range t.lastBorders {
				if t.hitBorder(bh, mx, my) {
					if bh.Split != nil {
						t.dragging = bh.Split
						t.dragging.Dragging = true
					} else if bh.Flex != nil {
						t.flexDragging = bh.Flex
						t.flexDragIdx = bh.FlexIndex
						t.flexDragging.Dragging = true
					}
					return nil
				}
			}
			// Click on a panel — focus it, toggle collapsible, then forward
			if hit := t.panelAt(mx, my, cw, ch); hit != nil {
				t.setFocus(hit.Node.Panel)
				// Toggle collapsible on title bar click
				if c, ok := hit.Node.Panel.(*Collapsible); ok {
					if my == hit.Y { // Click on the first line (title bar)
						c.Toggle()
						t.updateFlexCollapsed(hit.Node)
						return nil
					}
				}
				// Forward mouse event to the panel
				relMsg := tea.MouseMsg{
					X:      mx - hit.X,
					Y:      my - hit.Y,
					Action: msg.Action,
					Button: msg.Button,
					Type:   msg.Type,
					Alt:    msg.Alt,
				}
				if hit.Node.Panel != nil {
					return hit.Node.Panel.Update(relMsg)
				}
			}
		case tea.MouseActionMotion:
			if t.dragging != nil || t.flexDragging != nil {
				t.updateDrag(mx, my, cw, ch)
				return nil
			}
		case tea.MouseActionRelease:
			wasDragging := t.dragging != nil || t.flexDragging != nil
			if t.dragging != nil {
				t.dragging.Dragging = false
				t.dragging = nil
			}
			if t.flexDragging != nil {
				t.flexDragging.Dragging = false
				t.flexDragging = nil
				t.flexDragIdx = -1
			}
			// Do not forward the release event to a panel if it ends a drag.
			if wasDragging {
				t.broadcastResize(t.root, 0, 0, cw, ch)
				return nil
			}
		}
	}

	// Forward mouse to panel under cursor (relative coordinates)
	if msg.Action == tea.MouseActionPress || msg.Action == tea.MouseActionMotion || msg.Action == tea.MouseActionRelease {
		if hit := t.panelAt(mx, my, cw, ch); hit != nil && hit.Node.Panel != nil {
			relMsg := tea.MouseMsg{
				X:      mx - hit.X,
				Y:      my - hit.Y,
				Action: msg.Action,
				Button: msg.Button,
				Type:   msg.Type,
				Alt:    msg.Alt,
			}
			return hit.Node.Panel.Update(relMsg)
		}
	}
	return nil
}

// handleKeys forwards key messages to the focused panel.
// Warp does NOT intercept Tab/Shift+Tab for focus traversal automatically.
// Use tab.FocusNext() / tab.FocusPrev() explicitly if you want keyboard focus switching.
func (t *Tab) handleKeys(msg tea.KeyMsg) tea.Cmd {
	if t.focused != nil {
		return t.focused.Update(msg)
	}
	return nil
}

// SetSplitFraction updates the fraction of the split that contains the given
// panel. It returns true if the panel was found inside a split.
func (t *Tab) SetSplitFraction(panel Panel, fraction float64) bool {
	t.ensureRoot()
	target := t.root.findNode(panel)
	if target == nil {
		return false
	}
	return t.setSplitFractionNode(t.root, target, clampFraction(fraction))
}

// GetSplitFraction returns the current fraction of the split that contains the
// given panel. It returns 0 and false if the panel is not inside a split.
func (t *Tab) GetSplitFraction(panel Panel) (float64, bool) {
	t.ensureRoot()
	target := t.root.findNode(panel)
	if target == nil {
		return 0, false
	}
	return t.getSplitFractionNode(t.root, target)
}

// Collapse shrinks the panel to a fixed size inside its parent split.
// width is the desired size in cells for vertical splits; for horizontal splits
// it is interpreted as height. The previous split fraction is saved so Expand
// can restore it. It returns true if the panel was found.
func (t *Tab) Collapse(panel Panel, size int) bool {
	t.ensureRoot()
	target := t.root.findNode(panel)
	if target == nil {
		return false
	}
	return t.collapseNode(t.root, target, size)
}

// Expand restores the panel to its saved split fraction.
// It returns true if the panel was found and was collapsed.
func (t *Tab) Expand(panel Panel) bool {
	t.ensureRoot()
	target := t.root.findNode(panel)
	if target == nil {
		return false
	}
	return t.expandNode(t.root, target)
}

func (t *Tab) collapseNode(parent, target *Node, size int) bool {
	if parent == nil {
		return false
	}
	if parent.Split != nil {
		if parent.Split.First == target {
			t.saveCollapse(target, size, 0)
			return true
		}
		if parent.Split.Second == target {
			t.saveCollapse(target, 0, size)
			return true
		}
		if t.collapseNode(parent.Split.First, target, size) {
			return true
		}
		return t.collapseNode(parent.Split.Second, target, size)
	}
	if parent.Flex != nil {
		for _, item := range parent.Flex.Items {
			if item.Node == target {
				item.Collapsed = true
				if c, ok := target.Panel.(*Collapsible); ok {
					c.Toggle()
				}
				return true
			}
			if t.collapseNode(item.Node, target, size) {
				return true
			}
		}
	}
	return false
}

func (t *Tab) expandNode(parent, target *Node) bool {
	if parent == nil {
		return false
	}
	if parent.Split != nil {
		if parent.Split.First == target || parent.Split.Second == target {
			t.restoreCollapse(target)
			return true
		}
		if t.expandNode(parent.Split.First, target) {
			return true
		}
		return t.expandNode(parent.Split.Second, target)
	}
	if parent.Flex != nil {
		for _, item := range parent.Flex.Items {
			if item.Node == target {
				item.Collapsed = false
				if c, ok := target.Panel.(*Collapsible); ok {
					c.Toggle()
				}
				return true
			}
			if t.expandNode(item.Node, target) {
				return true
			}
		}
	}
	return false
}

func (t *Tab) saveCollapse(node *Node, width, height int) {
	if node.Collapse == nil {
		node.Collapse = &NodeCollapse{}
	}
	node.Collapse.Active = true
	node.Collapse.Width = width
	node.Collapse.Height = height
	// Save the current fraction from the parent split if possible.
	t.saveFractionForNode(t.root, node)
}

func (t *Tab) saveFractionForNode(parent, target *Node) {
	if parent == nil {
		return
	}
	if parent.Split != nil {
		if parent.Split.First == target {
			if target.Collapse != nil {
				target.Collapse.Saved = parent.Split.Fraction
			}
			return
		}
		if parent.Split.Second == target {
			if target.Collapse != nil {
				target.Collapse.Saved = 1 - parent.Split.Fraction
			}
			return
		}
		t.saveFractionForNode(parent.Split.First, target)
		t.saveFractionForNode(parent.Split.Second, target)
	}
	if parent.Flex != nil {
		for _, item := range parent.Flex.Items {
			t.saveFractionForNode(item.Node, target)
		}
	}
}

func (t *Tab) restoreCollapse(node *Node) {
	if node.Collapse == nil {
		return
	}
	node.Collapse.Active = false
	if node.Collapse.Saved > 0 {
		t.setSplitFractionNode(t.root, node, clampFraction(node.Collapse.Saved))
	}
}

func (t *Tab) setSplitFractionNode(parent, target *Node, fraction float64) bool {
	if parent == nil {
		return false
	}
	if parent.Split != nil {
		if parent.Split.First == target {
			parent.Split.Fraction = fraction
			return true
		}
		if parent.Split.Second == target {
			parent.Split.Fraction = 1 - fraction
			return true
		}
		if t.setSplitFractionNode(parent.Split.First, target, fraction) {
			return true
		}
		return t.setSplitFractionNode(parent.Split.Second, target, fraction)
	}
	if parent.Flex != nil {
		for _, item := range parent.Flex.Items {
			if t.setSplitFractionNode(item.Node, target, fraction) {
				return true
			}
		}
	}
	return false
}

func (t *Tab) getSplitFractionNode(parent, target *Node) (float64, bool) {
	if parent == nil {
		return 0, false
	}
	if parent.Split != nil {
		if parent.Split.First == target {
			return parent.Split.Fraction, true
		}
		if parent.Split.Second == target {
			return 1 - parent.Split.Fraction, true
		}
		if f, ok := t.getSplitFractionNode(parent.Split.First, target); ok {
			return f, true
		}
		return t.getSplitFractionNode(parent.Split.Second, target)
	}
	if parent.Flex != nil {
		for _, item := range parent.Flex.Items {
			if f, ok := t.getSplitFractionNode(item.Node, target); ok {
				return f, true
			}
		}
	}
	return 0, false
}

// broadcastMsg sends a message to all panels in this tab (tree + floats).
func (t *Tab) broadcastMsg(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, t.broadcastNode(t.root, msg)...)
	for _, fp := range t.floats {
		if fp.Panel != nil {
			if cmd := fp.Panel.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return cmds
}

func (t *Tab) broadcastResize(node *Node, x, y, w, h int) []tea.Cmd {
	layout := newLayout(node, layoutRect{x: x, y: y, w: max(0, w), h: max(0, h)})
	return t.broadcastLayoutResize(layout)
}

func (t *Tab) broadcastNode(node *Node, msg tea.Msg) []tea.Cmd {
	if node == nil {
		return nil
	}
	if node.IsLeaf() && node.Panel != nil {
		if cmd := node.Panel.Update(msg); cmd != nil {
			return []tea.Cmd{cmd}
		}
		return nil
	}
	var cmds []tea.Cmd
	if node.Split != nil {
		cmds = append(cmds, t.broadcastNode(node.Split.First, msg)...)
		cmds = append(cmds, t.broadcastNode(node.Split.Second, msg)...)
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			cmds = append(cmds, t.broadcastNode(item.Node, msg)...)
		}
	}
	return cmds
}

// BroadcastResize sends ResizeMsg with each leaf panel's current content size.
// It is called automatically by Tab on WindowSizeMsg and after a drag ends.
func (t *Tab) BroadcastResize() tea.Msg {
	cmds := t.broadcastResize(t.root, 0, 0, t.width, t.height)
	return tea.BatchMsg(cmds)
}

func (t *Tab) hitBorder(bh BorderHit, mx, my int) bool {
	switch bh.Direction {
	case Vertical:
		return mx == bh.X && my >= bh.Y && my < bh.Y+bh.Length
	case Horizontal:
		return my == bh.Y && mx >= bh.X && mx < bh.X+bh.Length
	}
	return false
}

// panelHit describes a panel and its bounds.
type panelHit struct {
	Node *Node
	X, Y int
	W, H int
}

func (t *Tab) panelAt(mx, my, cw, ch int) *panelHit {
	layout := newLayout(t.root, layoutRect{w: max(0, cw), h: max(0, ch)})
	return findLayoutPanel(layout, mx, my)
}

func (t *Tab) panelAtNode(node *Node, x, y, w, h int, mx, my int) *panelHit {
	layout := newLayout(node, layoutRect{x: x, y: y, w: max(0, w), h: max(0, h)})
	return findLayoutPanel(layout, mx, my)
}

func (t *Tab) panelAtFlex(flex *FlexConfig, x, y, w, h int, mx, my int) *panelHit {
	return t.panelAtNode(&Node{Flex: flex}, x, y, w, h, mx, my)
}

// ToggleCollapsible toggles the collapsed state of a collapsible panel.
// It also updates the flex item if the panel is inside a flex layout.
func (t *Tab) ToggleCollapsible(panel Panel) {
	t.toggleCollapsibleNode(t.root, panel)
}

func (t *Tab) toggleCollapsibleNode(node *Node, panel Panel) bool {
	if node == nil {
		return false
	}
	if node.IsLeaf() && samePanel(node.Panel, panel) {
		if c, ok := panel.(*Collapsible); ok {
			c.Toggle()
		}
		return true
	}
	if node.Split != nil {
		if t.toggleCollapsibleNode(node.Split.First, panel) {
			return true
		}
		return t.toggleCollapsibleNode(node.Split.Second, panel)
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			if item == nil || item.Node == nil {
				continue
			}
			if item.Node.IsLeaf() && samePanel(item.Node.Panel, panel) {
				if c, ok := panel.(*Collapsible); ok {
					c.Toggle()
					item.Collapsed = c.Collapsed
				}
				return true
			}
			if t.toggleCollapsibleNode(item.Node, panel) {
				return true
			}
		}
	}
	return false
}

func (t *Tab) updateDrag(mx, my, cw, ch int) {
	if t.dragging != nil {
		t.updateSplitDrag(mx, my, cw, ch)
	}
	if t.flexDragging != nil {
		t.updateFlexDrag(mx, my, cw, ch)
	}
	// Send live ResizeMsg so panels update while dragging.
	t.broadcastResize(t.root, 0, 0, cw, ch)
}

func (t *Tab) updateSplitDrag(mx, my, _, _ int) {
	for _, border := range t.lastBorders {
		if border.Split != t.dragging {
			continue
		}
		switch border.Direction {
		case Vertical:
			avail := max(0, border.Bounds.W-1)
			if avail > 0 {
				fraction := float64(mx-border.Bounds.X) / float64(avail)
				t.dragging.Fraction = clampFraction(fraction)
			}
		case Horizontal:
			avail := max(0, border.Bounds.H-1)
			if avail > 0 {
				fraction := float64(my-border.Bounds.Y) / float64(avail)
				t.dragging.Fraction = clampFraction(fraction)
			}
		}
		return
	}
}

func (t *Tab) updateFlexDrag(mx, my, _, _ int) {
	if t.flexDragging == nil || t.flexDragIdx < 0 || t.flexDragIdx >= len(t.flexDragging.Items)-1 {
		return
	}

	var hit *BorderHit
	for i := range t.lastBorders {
		border := &t.lastBorders[i]
		if border.Flex == t.flexDragging && border.FlexIndex == t.flexDragIdx {
			hit = border
			break
		}
	}
	if hit == nil {
		return
	}

	flexLayout := newLayout(&Node{Flex: t.flexDragging}, layoutRect{
		x: hit.Bounds.X,
		y: hit.Bounds.Y,
		w: hit.Bounds.W,
		h: hit.Bounds.H,
	})
	if len(flexLayout.children) != len(t.flexDragging.Items) {
		return
	}

	vertical := t.flexDragging.Direction == Vertical
	sizes := make([]int, len(flexLayout.children))
	for i, child := range flexLayout.children {
		if vertical {
			sizes[i] = child.bounds.h
		} else {
			sizes[i] = child.bounds.w
		}
	}
	pairSize := sizes[t.flexDragIdx] + sizes[t.flexDragIdx+1]
	if pairSize <= 0 {
		return
	}

	position := mx
	borderPosition := hit.X
	if vertical {
		position = my
		borderPosition = hit.Y
	}
	firstSize := sizes[t.flexDragIdx] + position - borderPosition
	minimum := 0
	if pairSize >= 2*MinPanelSize {
		minimum = MinPanelSize
	}
	firstSize = min(max(firstSize, minimum), pairSize-minimum)
	sizes[t.flexDragIdx] = firstSize
	sizes[t.flexDragIdx+1] = pairSize - firstSize

	for i, item := range t.flexDragging.Items {
		if item == nil || item.Collapsed {
			continue
		}
		item.Basis = max(1, sizes[i])
		item.Grow = max(1, sizes[i])
	}
}

// updateFlexCollapsed updates the FlexItem.Collapsed flag for a collapsible panel.
func (t *Tab) updateFlexCollapsed(target *Node) {
	t.updateFlexCollapsedNode(t.root, target)
}

func (t *Tab) updateFlexCollapsedNode(node *Node, target *Node) bool {
	if node == nil {
		return false
	}
	if node.Flex != nil {
		for _, item := range node.Flex.Items {
			if item.Node == target {
				if c, ok := target.Panel.(*Collapsible); ok {
					item.Collapsed = c.Collapsed
				}
				return true
			}
			if t.updateFlexCollapsedNode(item.Node, target) {
				return true
			}
		}
	}
	if node.Split != nil {
		if t.updateFlexCollapsedNode(node.Split.First, target) {
			return true
		}
		return t.updateFlexCollapsedNode(node.Split.Second, target)
	}
	return false
}

func clampFraction(f float64) float64 {
	if f < 0.1 {
		return 0.1
	}
	if f > 0.9 {
		return 0.9
	}
	return f
}

// emptyPanel is used as a placeholder when a tab has no user panels yet.
type emptyPanel struct{}

func (emptyPanel) View(width, height int) string {
	return ""
}

func (emptyPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}
