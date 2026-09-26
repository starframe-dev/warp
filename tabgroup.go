package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func padRight(s string, w int) string {
	width := ansi.StringWidth(s)
	if width < w {
		return s + strings.Repeat(" ", w-width)
	}
	return s
}

// tabRegion describes a clickable area in the tab bar.
type tabRegion struct {
	idx    int
	startX int
	endX   int
	closeX int // X position of close button, -1 if none
}

// TabGroup is a Panel that renders a tab bar and switches between tabs.
// Use it as a component inside splits, flex layouts, or as the root panel.
type TabGroup struct {
	tabs      []*Tab
	activeTab int
	width     int
	height    int
	ownership *panelOwnership

	tabPosition      TabPosition
	tabRegions       []tabRegion
	newTabRegion     *tabRegion
	verticalTabWidth int
}

// NewTabGroup creates a TabGroup panel with one default tab.
func NewTabGroup(pos TabPosition) *TabGroup {
	tg := &TabGroup{tabPosition: pos}
	tg.ownership = newPanelOwnership(tg)
	tg.NewTab("main")
	return tg
}

// NewTab creates a new tab and switches to it.
func (tg *TabGroup) NewTab(name string) *Tab {
	ownership := tg.ensureOwnership()
	tab := newTab(name, tg)
	tab.ownership = ownership
	tg.tabs = append(tg.tabs, tab)
	attachPanelOwnership(tab, ownership)
	tg.switchTab(len(tg.tabs) - 1)
	return tab
}

// ActiveTab returns the currently active tab.
func (tg *TabGroup) ActiveTab() *Tab {
	if tg.activeTab < 0 || tg.activeTab >= len(tg.tabs) {
		return nil
	}
	return tg.tabs[tg.activeTab]
}

func (tg *TabGroup) closeTab(idx int) {
	if idx < 0 || idx >= len(tg.tabs) || len(tg.tabs) <= 1 {
		return
	}
	ownership := tg.ensureOwnership()
	active := tg.activeTab
	closing := tg.tabs[idx]
	candidates := collectTabPanels(closing)
	closing.setFocus(nil)
	tg.tabs = removeSliceAt(tg.tabs, idx)
	closing.clearAfterRemoval()
	switch {
	case idx < active:
		tg.activeTab = active - 1
	case idx == active && active >= len(tg.tabs):
		tg.activeTab = len(tg.tabs) - 1
	}
	if idx == active {
		if next := tg.ActiveTab(); next != nil {
			next.resumeFocus()
		}
	}
	ownership.unmountRemoved(candidates)
}

func (tg *TabGroup) switchTab(idx int) {
	if idx < 0 || idx >= len(tg.tabs) || idx == tg.activeTab {
		return
	}
	if current := tg.ActiveTab(); current != nil {
		current.suspendFocus()
	}
	tg.activeTab = idx
	if next := tg.ActiveTab(); next != nil {
		next.resumeFocus()
	}
}

// NextTab switches to the next tab.
func (tg *TabGroup) NextTab() {
	if len(tg.tabs) > 1 {
		tg.switchTab((tg.activeTab + 1) % len(tg.tabs))
	}
}

// PrevTab switches to the previous tab.
func (tg *TabGroup) PrevTab() {
	if len(tg.tabs) > 1 {
		tg.switchTab((tg.activeTab - 1 + len(tg.tabs)) % len(tg.tabs))
	}
}

func (tg *TabGroup) contentWidth(totalW int) int {
	if tg.tabPosition == TabLeft || tg.tabPosition == TabRight {
		return max(0, totalW-tg.verticalTabWidth)
	}
	return max(0, totalW)
}

func (tg *TabGroup) contentHeight(totalH int) int {
	if tg.tabPosition == TabTop || tg.tabPosition == TabBottom {
		return max(0, totalH-1)
	}
	return max(0, totalH)
}

func (tg *TabGroup) contentOffset() (int, int) {
	switch tg.tabPosition {
	case TabTop:
		return 0, 1
	case TabBottom:
		return 0, 0
	case TabLeft:
		return tg.verticalTabWidth, 0
	case TabRight:
		return 0, 0
	case TabNone:
		return 0, 0
	}
	return 0, 0
}

// View renders the tab bar + active tab content.
func (tg *TabGroup) View(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	tg.width = w
	tg.height = h

	tab := tg.ActiveTab()
	if tab == nil {
		return emptyView(h)
	}

	var verticalBar string
	if tg.tabPosition == TabLeft || tg.tabPosition == TabRight {
		verticalBar = tg.renderTabBar(w)
	}
	cw := tg.contentWidth(w)
	ch := tg.contentHeight(h)

	switch tg.tabPosition {
	case TabTop:
		tabBar := tg.renderTabBar(w)
		content := tab.renderContent(cw, ch)
		return lipgloss.JoinVertical(lipgloss.Left, tabBar, content)

	case TabBottom:
		content := tab.renderContent(cw, ch)
		tabBar := tg.renderTabBar(w)
		return lipgloss.JoinVertical(lipgloss.Left, content, tabBar)

	case TabLeft:
		content := tab.renderContent(cw, ch)
		return lipgloss.JoinHorizontal(lipgloss.Top, verticalBar, content)

	case TabRight:
		content := tab.renderContent(cw, ch)
		return lipgloss.JoinHorizontal(lipgloss.Top, content, verticalBar)

	case TabNone:
		return tab.renderContent(cw, ch)
	}

	return ""
}

// Elements implements ElementProvider. It returns the active tab's elements
// offset by the tab bar position, if any.
func (tg *TabGroup) Elements(w, h int) []Element {
	w = max(0, w)
	h = max(0, h)
	tab := tg.ActiveTab()
	if tab == nil {
		return nil
	}

	cw, ch := w, h
	offX, offY := 0, 0
	switch tg.tabPosition {
	case TabTop:
		ch = max(0, h-1)
		offY = 1
	case TabBottom:
		ch = max(0, h-1)
	case TabLeft:
		offX = tg.verticalTabBarWidth(w)
		cw = max(0, w-offX)
	case TabRight:
		cw = max(0, w-tg.verticalTabBarWidth(w))
	}

	elems := collectElements(tab, cw, ch)
	for i := range elems {
		elems[i].Bounds.X += offX
		elems[i].Bounds.Y += offY
		shiftElements(elems[i].Children, offX, offY)
	}
	return elems
}

func shiftElements(elems []Element, dx, dy int) {
	for i := range elems {
		elems[i].Bounds.X += dx
		elems[i].Bounds.Y += dy
		shiftElements(elems[i].Children, dx, dy)
	}
}

// Update handles keys, mouse, and window resize for the tab group.
func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return tg.handleKeyMsg(msg)
	case tea.MouseMsg:
		return tg.handleMouseMsg(msg)
	case tea.WindowSizeMsg:
		tg.width = msg.Width
		tg.height = msg.Height
		var cmds []tea.Cmd
		for _, tab := range tg.tabs {
			if cmd := tab.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return tea.Batch(cmds...)
	case ResizeMsg:
		// Forward ResizeMsg to tabs so they can broadcast to child panels
		// with correct allocated sizes (via broadcastResize).
		tg.width = msg.Width
		tg.height = msg.Height
		var cmds []tea.Cmd
		for _, tab := range tg.tabs {
			if cmd := tab.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
		return tea.Batch(cmds...)
	}
	// Broadcast unknown messages (PtyReadyMsg, PtyOutputMsg, etc.)
	// to all panels so emulators can receive them.
	var cmds []tea.Cmd
	for _, tab := range tg.tabs {
		if tab != nil {
			tab.appendBroadcastMsg(&cmds, msg)
		}
	}
	return tea.Batch(cmds...)
}

func (tg *TabGroup) handleKeyMsg(msg tea.KeyMsg) tea.Cmd {
	if tab := tg.ActiveTab(); tab != nil {
		if receiver, ok := tab.focused.(RawKeyReceiver); ok && receiver.WantsRawKeys() {
			return tab.handleKeys(msg)
		}
	}

	keyStr := msg.String()
	switch keyStr {
	case "ctrl+c":
		return tea.Quit
	case "ctrl+tab":
		tg.NextTab()
		return nil
	case "ctrl+shift+tab":
		tg.PrevTab()
		return nil
	case "ctrl+w":
		tg.closeTab(tg.activeTab)
		return nil
	case "ctrl+t":
		tg.NewTab("tab")
		return nil
	}

	if tab := tg.ActiveTab(); tab != nil {
		return tab.handleKeys(msg)
	}
	return nil
}

func (tg *TabGroup) handleMouseMsg(msg tea.MouseMsg) tea.Cmd {
	if tg.isOnTabBar(msg.X, msg.Y) {
		return tg.handleTabBarClick(msg)
	}

	ox, oy := tg.contentOffset()
	cw := tg.contentWidth(tg.width)
	ch := tg.contentHeight(tg.height)
	if tab := tg.ActiveTab(); tab != nil {
		return tab.handleMouse(msg, ox, oy, cw, ch)
	}
	return nil
}

func (tg *TabGroup) isOnTabBar(x, y int) bool {
	switch tg.tabPosition {
	case TabTop:
		return y == 0
	case TabBottom:
		return y == tg.height-1
	case TabLeft:
		return x < tg.verticalTabWidth
	case TabRight:
		return x >= tg.width-tg.verticalTabWidth
	case TabNone:
		return false
	}
	return false
}

func (tg *TabGroup) handleTabBarClick(msg tea.MouseMsg) tea.Cmd {
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return nil
	}

	x := msg.X
	y := msg.Y
	if tg.tabPosition == TabRight {
		x -= tg.width - tg.verticalTabWidth
	}

	switch tg.tabPosition {
	case TabLeft, TabRight:
		row := y
		if row >= 0 && row < len(tg.tabRegions) {
			r := tg.tabRegions[row]
			if r.idx == -1 {
				tg.NewTab("tab")
				return nil
			}
			if r.closeX >= 0 && x >= r.closeX && x < r.endX {
				tg.closeTab(r.idx)
				return nil
			}
			tg.switchTab(r.idx)
		}
		return nil
	default:
		for _, r := range tg.tabRegions {
			if x >= r.startX && x < r.endX {
				if r.idx == -1 {
					tg.NewTab("tab")
					return nil
				}
				if r.closeX >= 0 && x >= r.closeX && x < r.endX {
					tg.closeTab(r.idx)
					return nil
				}
				tg.switchTab(r.idx)
				return nil
			}
		}
		if tg.newTabRegion != nil && x >= tg.newTabRegion.startX && x < tg.newTabRegion.endX {
			tg.NewTab("tab")
			return nil
		}
	}
	return nil
}

func (tg *TabGroup) renderTabBar(width int) string {
	if tg.tabPosition == TabLeft || tg.tabPosition == TabRight {
		return tg.renderVerticalTabBar(width)
	}
	return tg.renderHorizontalTabBar(width)
}

func (tg *TabGroup) renderHorizontalTabBar(width int) string {
	tg.tabRegions = tg.tabRegions[:0]
	activeIdx := tg.activeTab

	var contents strings.Builder
	contents.Grow(len(tg.tabs)*24 + 3)
	col := 0
	for i, tab := range tg.tabs {
		name := ansi.Truncate(tab.name, 20, "...")
		label := " " + name + " "
		if i == activeIdx {
			label = "▎ " + name + " ×"
		}

		labelW := ansi.StringWidth(label)
		endX := col + labelW
		closeX := -1
		if i == activeIdx {
			closeX = endX - 1
		}
		tg.tabRegions = append(tg.tabRegions, tabRegion{
			idx: i, startX: col, endX: endX, closeX: closeX,
		})
		col += labelW

		style := inactiveTabStyle
		if i == activeIdx {
			style = activeTabStyle
		}
		contents.WriteString(style.Render(label))
	}

	newLabel := " + "
	newW := ansi.StringWidth(newLabel)
	if tg.newTabRegion == nil {
		tg.newTabRegion = &tabRegion{}
	}
	*tg.newTabRegion = tabRegion{startX: col, endX: col + newW}
	col += newW
	contents.WriteString(newTabStyle.Render(newLabel))

	bar := tabBarStyle.Render(contents.String())
	if padding := width - col; padding > 0 {
		var result strings.Builder
		result.Grow(len(bar) + padding + 16)
		result.WriteString(bar)
		result.WriteString(tabBarStyle.Render(strings.Repeat(" ", padding)))
		return result.String()
	}
	return bar
}

func (tg *TabGroup) verticalTabLabels() ([]string, int) {
	labels := make([]string, len(tg.tabs))
	naturalWidth := ansi.StringWidth(" + ")
	for i, tab := range tg.tabs {
		name := ansi.Truncate(tab.name, 15, "...")
		label := " " + name + " "
		if i == tg.activeTab {
			label = "▎ " + name + " ×"
		}
		labels[i] = label
		naturalWidth = max(naturalWidth, ansi.StringWidth(label))
	}
	return labels, naturalWidth
}

func (tg *TabGroup) verticalTabBarWidth(maxWidth int) int {
	_, naturalWidth := tg.verticalTabLabels()
	return min(naturalWidth, max(0, maxWidth))
}

func (tg *TabGroup) renderVerticalTabBar(maxWidth int) string {
	activeIdx := tg.activeTab
	tg.tabRegions = tg.tabRegions[:0]
	tg.newTabRegion = nil

	labels, naturalWidth := tg.verticalTabLabels()
	barWidth := min(naturalWidth, max(0, maxWidth))
	tg.verticalTabWidth = barWidth

	lines := make([]string, 0, len(tg.tabs)+1)
	for i, label := range labels {
		if ansi.StringWidth(label) > barWidth {
			label = ansi.Truncate(label, barWidth, "")
		}
		labelWidth := ansi.StringWidth(label)
		closeX := -1
		if i == activeIdx && labelWidth > 0 && strings.HasSuffix(label, "×") {
			closeX = labelWidth - 1
		}
		tg.tabRegions = append(tg.tabRegions, tabRegion{
			idx: i, startX: 0, endX: labelWidth, closeX: closeX,
		})

		style := inactiveTabStyle
		if i == activeIdx {
			style = activeTabStyle
		}
		lines = append(lines, style.Render(padRight(label, barWidth)))
	}

	newLabel := ansi.Truncate(" + ", barWidth, "")
	newWidth := ansi.StringWidth(newLabel)
	tg.tabRegions = append(tg.tabRegions, tabRegion{
		idx: -1, startX: 0, endX: newWidth, closeX: -1,
	})
	lines = append(lines, newTabStyle.Render(padRight(newLabel, barWidth)))

	if len(lines) > tg.height {
		lines = lines[:tg.height]
		tg.tabRegions = tg.tabRegions[:tg.height]
	}
	return strings.Join(lines, "\n")
}
