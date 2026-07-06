package warp

import (
    "fmt"
    "os"
    "strings"

    tea "github.com/charmbracelet/bubbletea"
)

func tabDebugLog(format string, args ...interface{}) {
    f, err := os.OpenFile("/tmp/warp-mouse.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
    if err != nil {
        return
    }
    defer f.Close()
    fmt.Fprintf(f, format, args...)
}

// TabPosition defines where the tab bar is rendered.
type TabPosition int

const (
    TabTop    TabPosition = iota
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
    t.ensureRoot()
    t.root.Panel = panel
}

// SplitVertical splits the panel vertically (left/right).
// fraction is the share for the left panel (0.0–1.0).
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    oldPanel := node.Panel
    node.Panel = nil
    node.Split = &SplitConfig{
        Direction: Vertical,
        Fraction:  clampFraction(fraction),
        First:     &Node{Panel: oldPanel},
        Second:    &Node{Panel: newPanel},
    }
}

// SplitHorizontal splits the panel horizontally (top/bottom).
// fraction is the share for the top panel (0.0–1.0).
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel) {
    t.ensureRoot()
    node := t.root.findNode(parent)
    if node == nil {
        return
    }
    oldPanel := node.Panel
    node.Panel = nil
    node.Split = &SplitConfig{
        Direction: Horizontal,
        Fraction:  clampFraction(fraction),
        First:     &Node{Panel: oldPanel},
        Second:    &Node{Panel: newPanel},
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
    node.Panel = nil
    node.Split = nil
    flexItems := make([]*FlexItem, len(items))
    for i, spec := range items {
        grow := spec.Grow
        if grow < 0 {
            grow = 0
        }
        flexItems[i] = &FlexItem{
            Node: &Node{Panel: spec.Panel},
            Grow: grow,
        }
    }
    node.Flex = &FlexConfig{
        Direction: Horizontal,
        Items:     flexItems,
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
    node.Panel = nil
    node.Split = nil
    flexItems := make([]*FlexItem, len(items))
    for i, spec := range items {
        grow := spec.Grow
        if grow < 0 {
            grow = 0
        }
        flexItems[i] = &FlexItem{
            Node: &Node{Panel: spec.Panel},
            Grow: grow,
        }
    }
    node.Flex = &FlexConfig{
        Direction: Vertical,
        Items:     flexItems,
    }
}

// Float makes a panel floating above the layout.
func (t *Tab) Float(panel Panel, x, y, width, height int) {
    fp := &FloatPane{
        Panel:  panel,
        X:      x,
        Y:      y,
        Width:  width,
        Height: height,
        Title:  "Float",
    }
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
    t.focused = panel
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
    if f, ok := panel.(Focusable); ok {
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
    if current, ok := t.focused.(Focusable); ok && current != nil {
        applyFocus(current, next)
    } else if next != nil {
        next.Focus()
    }
    if next != nil {
        t.focused = next
    }
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
    t.lastBorders = findBorders(t.root, 0, 0, w, h)
    lines := renderNode(t.root, w, h)

    // Render floats on top
    for _, fp := range t.floats {
        overlayFloat(lines, fp, w, h)
    }

    return strings.Join(lines, "\n")
}

// Elements returns elements from the panel tree, recursively accounting for
// splits/flex layouts so coordinates are relative to the tab content area.
func (t *Tab) Elements(w, h int) []Element {
    return t.elementsNode(t.root, 0, 0, w, h)
}

func (t *Tab) elementsNode(node *Node, x, y, w, h int) []Element {
    if node == nil {
        return nil
    }
    if node.IsLeaf() {
        elems := collectElements(node.Panel, w, h)
        for i := range elems {
            elems[i].Bounds.X += x
            elems[i].Bounds.Y += y
            shiftElements(elems[i].Children, x, y)
        }
        return elems
    }

    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW, secondW := computeSplitSizes(availW, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Vertical), node.Split.Second.CollapsedSize(Vertical))
            var elems []Element
            elems = append(elems, t.elementsNode(node.Split.First, x, y, firstW, h)...)
            elems = append(elems, t.elementsNode(node.Split.Second, x+firstW+borderW, y, secondW, h)...)
            return elems
        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH, secondH := computeSplitSizes(availH, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Horizontal), node.Split.Second.CollapsedSize(Horizontal))
            var elems []Element
            elems = append(elems, t.elementsNode(node.Split.First, x, y, w, firstH)...)
            elems = append(elems, t.elementsNode(node.Split.Second, x, y+firstH+borderH, w, secondH)...)
            return elems
        }
    }

    if node.Flex != nil {
        return t.elementsFlex(node.Flex, x, y, w, h)
    }

    return nil
}

func (t *Tab) elementsFlex(flex *FlexConfig, x, y, w, h int) []Element {
    if len(flex.Items) == 0 {
        return nil
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1

    switch flex.Direction {
    case Horizontal:
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        var elems []Element
        cx := x
        for i, item := range flex.Items {
            if i > 0 {
                cx += borderSize
            }
            elems = append(elems, t.elementsNode(item.Node, cx, y, sizes[i], h)...)
            cx += sizes[i]
        }
        return elems
    case Vertical:
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        var elems []Element
        cy := y
        for i, item := range flex.Items {
            if i > 0 {
                cy += borderSize
            }
            elems = append(elems, t.elementsNode(item.Node, x, cy, w, sizes[i])...)
            cy += sizes[i]
        }
        return elems
    }

    return nil
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

    // Check float panes first (top z-order)
    hitFloat := false
    for i := len(t.floats) - 1; i >= 0; i-- {
        fp := t.floats[i]

        // Detect if cursor is inside this float before calling handleMouse
        // (handleMouse skips bounds check during active drag/resize)
        inside := mx >= fp.X && mx < fp.X+fp.Width && my >= fp.Y && my < fp.Y+fp.Height

        cmd := fp.handleMouse(msg, mx, my)
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
                t.focused = fp.Panel
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
                t.focused = fp.Panel
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
            for i, bh := range t.lastBorders {
                if t.hitBorder(bh, mx, my) {
                    tabDebugLog("MOUSE press on border mx=%d my=%d\n", mx, my)
                    if bh.Split != nil {
                        t.dragging = bh.Split
                        t.dragging.Dragging = true
                    } else if bh.Flex != nil {
                        t.flexDragging = bh.Flex
                        t.flexDragIdx = i
                        t.flexDragging.Dragging = true
                    }
                    return nil
                }
            }
            // Click on a panel — focus it, toggle collapsible
            if hit := t.panelAt(mx, my, cw, ch); hit != nil {
                t.focused = hit.Node.Panel
                // Toggle collapsible on title bar click
                if c, ok := hit.Node.Panel.(*Collapsible); ok {
                    if my == hit.Y { // Click on the first line (title bar)
                        c.Toggle()
                        t.updateFlexCollapsed(hit.Node)
                        return nil
                    }
                }
            }
        case tea.MouseActionMotion:
            if t.dragging != nil || t.flexDragging != nil {
                tabDebugLog("MOUSE motion mx=%d my=%d dragging=%v\n", mx, my, t.dragging != nil)
                t.updateDrag(mx, my, cw, ch)
                return nil
            }
        case tea.MouseActionRelease:
            wasDragging := t.dragging != nil || t.flexDragging != nil
            tabDebugLog("MOUSE release mx=%d my=%d wasDragging=%v dragging=%v flexDragging=%v\n", mx, my, wasDragging, t.dragging != nil, t.flexDragging != nil)
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
    var cmds []tea.Cmd
    if node == nil {
        return cmds
    }
    if node.IsLeaf() && node.Panel != nil {
        if cmd := node.Panel.Update(ResizeMsg{Width: w, Height: h}); cmd != nil {
            cmds = append(cmds, cmd)
        }
        return cmds
    }
    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW, secondW := computeSplitSizes(availW, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Vertical), node.Split.Second.CollapsedSize(Vertical))
            cmds = append(cmds, t.broadcastResize(node.Split.First, x, y, firstW, h)...)
            cmds = append(cmds, t.broadcastResize(node.Split.Second, x+firstW+borderW, y, secondW, h)...)
        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH, secondH := computeSplitSizes(availH, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Horizontal), node.Split.Second.CollapsedSize(Horizontal))
            cmds = append(cmds, t.broadcastResize(node.Split.First, x, y, w, firstH)...)
            cmds = append(cmds, t.broadcastResize(node.Split.Second, x, y+firstH+borderH, w, secondH)...)
        }
        return cmds
    }
    if node.Flex != nil {
        borderSize := 1
        numBorders := len(node.Flex.Items) - 1
        switch node.Flex.Direction {
        case Horizontal:
            availW := w - numBorders*borderSize
            sizes := computeFlexSizes(availW, node.Flex.Items)
            cx := x
            for i, item := range node.Flex.Items {
                if i > 0 {
                    cx += borderSize
                }
                cmds = append(cmds, t.broadcastResize(item.Node, cx, y, sizes[i], h)...)
                cx += sizes[i]
            }
        case Vertical:
            availH := h - numBorders*borderSize
            sizes := computeFlexSizes(availH, node.Flex.Items)
            cy := y
            for i, item := range node.Flex.Items {
                if i > 0 {
                    cy += borderSize
                }
                cmds = append(cmds, t.broadcastResize(item.Node, x, cy, w, sizes[i])...)
                cy += sizes[i]
            }
        }
    }
    return cmds
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
    return t.panelAtNode(t.root, 0, 0, cw, ch, mx, my)
}

func (t *Tab) panelAtNode(node *Node, x, y, w, h int, mx, my int) *panelHit {
    if node == nil {
        return nil
    }
    if node.IsLeaf() {
        if mx >= x && mx < x+w && my >= y && my < y+h {
            return &panelHit{Node: node, X: x, Y: y, W: w, H: h}
        }
        return nil
    }

    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW, secondW := computeSplitSizes(availW, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Vertical), node.Split.Second.CollapsedSize(Vertical))
            if mx < x+firstW {
                return t.panelAtNode(node.Split.First, x, y, firstW, h, mx, my)
            }
            if mx >= x+firstW+borderW {
                return t.panelAtNode(node.Split.Second, x+firstW+borderW, y, secondW, h, mx, my)
            }
            return nil // Border area
        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH, secondH := computeSplitSizes(availH, node.Split.Fraction, node.Split.First.IsCollapsed(), node.Split.Second.IsCollapsed(), node.Split.First.CollapsedSize(Horizontal), node.Split.Second.CollapsedSize(Horizontal))
            if my < y+firstH {
                return t.panelAtNode(node.Split.First, x, y, w, firstH, mx, my)
            }
            if my >= y+firstH+borderH {
                return t.panelAtNode(node.Split.Second, x, y+firstH+borderH, w, secondH, mx, my)
            }
            return nil
        }
    }

    if node.Flex != nil {
        return t.panelAtFlex(node.Flex, x, y, w, h, mx, my)
    }

    return nil
}

func (t *Tab) panelAtFlex(flex *FlexConfig, x, y, w, h int, mx, my int) *panelHit {
    if len(flex.Items) == 0 {
        return nil
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1

    switch flex.Direction {
    case Horizontal:
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        cx := x
        for i, item := range flex.Items {
            if i > 0 {
                cx += borderSize
            }
            if mx >= cx && mx < cx+sizes[i] {
                return t.panelAtNode(item.Node, cx, y, sizes[i], h, mx, my)
            }
            cx += sizes[i]
        }
    case Vertical:
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        cy := y
        for i, item := range flex.Items {
            if i > 0 {
                cy += borderSize
            }
            if my >= cy && my < cy+sizes[i] {
                return t.panelAtNode(item.Node, x, cy, w, sizes[i], mx, my)
            }
            cy += sizes[i]
        }
    }
    return nil
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
    if node.IsLeaf() && node.Panel == panel {
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
            if item.Node.Panel == panel {
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

func (t *Tab) updateSplitDrag(mx, my, cw, ch int) {
    for _, bh := range t.lastBorders {
        if bh.Split == t.dragging {
            switch bh.Direction {
            case Vertical:
                if cw > 0 {
                    frac := float64(mx) / float64(cw)
                    t.dragging.Fraction = clampFraction(frac)
                }
            case Horizontal:
                if ch > 0 {
                    frac := float64(my) / float64(ch)
                    t.dragging.Fraction = clampFraction(frac)
                }
            }
            break
        }
    }
}

func (t *Tab) updateFlexDrag(mx, my, cw, ch int) {
    if t.flexDragIdx < 0 || t.flexDragIdx >= len(t.flexDragging.Items)-1 {
        return
    }

    switch t.flexDragging.Direction {
    case Horizontal:
        if cw <= 0 {
            return
        }
        // Total available width for flex items
        borderSize := 1
        numBorders := len(t.flexDragging.Items) - 1
        availW := cw - numBorders*borderSize
        if availW <= 0 {
            return
        }

        // Mouse position relative to flex start
        // Find cumulative size up to dragIdx
        sizes := computeFlexSizes(availW, t.flexDragging.Items)
        cum := 0
        for i := 0; i <= t.flexDragIdx; i++ {
            cum += sizes[i]
        }
        cum += t.flexDragIdx * borderSize

        // New relative position
        rel := mx - cum + sizes[t.flexDragIdx]
        if rel < MinPanelSize {
            rel = MinPanelSize
        }
        if rel > availW-MinPanelSize {
            rel = availW - MinPanelSize
        }

        // Adjust grow weights proportionally
        leftGrow := t.flexDragging.Items[t.flexDragIdx].Grow
        rightGrow := t.flexDragging.Items[t.flexDragIdx+1].Grow
        if leftGrow+rightGrow == 0 {
            leftGrow = 1
            rightGrow = 1
        }
        ratio := float64(rel) / float64(availW)
        totalGrow := leftGrow + rightGrow
        t.flexDragging.Items[t.flexDragIdx].Grow = int(ratio * float64(totalGrow))
        if t.flexDragging.Items[t.flexDragIdx].Grow < 1 {
            t.flexDragging.Items[t.flexDragIdx].Grow = 1
        }
        t.flexDragging.Items[t.flexDragIdx+1].Grow = totalGrow - t.flexDragging.Items[t.flexDragIdx].Grow

    case Vertical:
        if ch <= 0 {
            return
        }
        borderSize := 1
        numBorders := len(t.flexDragging.Items) - 1
        availH := ch - numBorders*borderSize
        if availH <= 0 {
            return
        }

        sizes := computeFlexSizes(availH, t.flexDragging.Items)
        cum := 0
        for i := 0; i <= t.flexDragIdx; i++ {
            cum += sizes[i]
        }
        cum += t.flexDragIdx * borderSize

        rel := my - cum + sizes[t.flexDragIdx]
        if rel < MinPanelSize {
            rel = MinPanelSize
        }
        if rel > availH-MinPanelSize {
            rel = availH - MinPanelSize
        }

        leftGrow := t.flexDragging.Items[t.flexDragIdx].Grow
        rightGrow := t.flexDragging.Items[t.flexDragIdx+1].Grow
        if leftGrow+rightGrow == 0 {
            leftGrow = 1
            rightGrow = 1
        }
        ratio := float64(rel) / float64(availH)
        totalGrow := leftGrow + rightGrow
        t.flexDragging.Items[t.flexDragIdx].Grow = int(ratio * float64(totalGrow))
        if t.flexDragging.Items[t.flexDragIdx].Grow < 1 {
            t.flexDragging.Items[t.flexDragIdx].Grow = 1
        }
        t.flexDragging.Items[t.flexDragIdx+1].Grow = totalGrow - t.flexDragging.Items[t.flexDragIdx].Grow
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
