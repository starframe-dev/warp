package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// Collapsible is a panel that can be collapsed to a single-line title bar.
type Collapsible struct {
	Title     string
	Collapsed bool
	Content   Panel
}

// NewCollapsible creates a new collapsible panel.
func NewCollapsible(title string, content Panel) *Collapsible {
	return &Collapsible{
		Title:   title,
		Content: content,
	}
}

// View renders a title row and, when expanded, the inner content below it.
func (c *Collapsible) View(w, h int) string {
	if h <= 0 {
		return ""
	}
	title := c.renderTitle(w)
	if c.Collapsed {
		return title
	}

	contentHeight := h - 1
	content := ""
	if !isNilPanel(c.Content) {
		content = c.Content.View(w, contentHeight)
	}
	lines := padContent(content, w, contentHeight)
	if len(lines) == 0 && contentHeight > 0 {
		return title + strings.Repeat("\n", contentHeight)
	}
	return strings.Join(append([]string{title}, lines...), "\n")
}

// Elements returns the persistent title control plus visible content elements.
func (c *Collapsible) Elements(w, h int) []Element {
	w = max(0, w)
	h = max(0, h)
	if w == 0 || h == 0 {
		return nil
	}

	elements := []Element{{
		Role:   "button",
		Name:   sanitizeFrameworkLabel(c.Title),
		Action: "toggle-collapse",
		Bounds: Bounds{W: w, H: 1},
	}}
	if c.Collapsed || isNilPanel(c.Content) || h <= 1 {
		return elements
	}

	contentHeight := h - 1
	contentElements := cloneElements(collectElements(c.Content, w, contentHeight))
	contentElements = clipElements(contentElements, Bounds{W: w, H: contentHeight})
	shiftElements(contentElements, 0, 1)
	return append(elements, contentElements...)
}

// ContentHeight adds the persistent title row to a known content height.
func (c *Collapsible) ContentHeight(width int) (int, bool) {
	if c.Collapsed {
		return 1, true
	}
	contentHeight, known := panelContentHeight(c.Content, width)
	if !known {
		return 0, false
	}
	maxInt := int(^uint(0) >> 1)
	if contentHeight == maxInt {
		return maxInt, true
	}
	return contentHeight + 1, true
}

// Update forwards visible content events and adjusts coordinates for the title row.
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ResizeMsg:
		if isNilPanel(c.Content) {
			return nil
		}
		contentHeight := msg.Height
		if c.Collapsed {
			contentHeight = 0
		} else {
			contentHeight = max(0, contentHeight-1)
		}
		msg.Height = contentHeight
		return c.Content.Update(msg)
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft && msg.Y == 0 {
			c.Toggle()
			return nil
		}
		if isNilPanel(c.Content) || c.Collapsed || msg.Y <= 0 {
			return nil
		}
		msg.Y--
		return c.Content.Update(msg)
	default:
		if isNilPanel(c.Content) {
			return nil
		}
		return c.Content.Update(msg)
	}
}

// Toggle switches between collapsed and expanded states.
func (c *Collapsible) Toggle() {
	c.Collapsed = !c.Collapsed
}

// renderTitle renders the title row in either state.
func (c *Collapsible) renderTitle(w int) string {
	if w <= 0 {
		return ""
	}

	indicator := "▶"
	if !c.Collapsed {
		indicator = "▼"
	}

	titleWidth := max(0, w-4) // indicator, spacing and corner glyphs
	title := ansi.Truncate(sanitizeFrameworkLabel(c.Title), titleWidth, "...")
	padding := max(0, w-4-ansi.StringWidth(title))
	line := collapsibleStyle.Render("┌"+indicator+" "+title) +
		collapsibleBorderStyle.Render(strings.Repeat("─", padding)+"┐")
	return padVisualLine(line, w)
}
