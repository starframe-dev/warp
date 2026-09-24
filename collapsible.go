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

// View renders the collapsible panel.
// When collapsed, returns a single-line title bar.
func (c *Collapsible) View(w, h int) string {
	if c.Collapsed {
		return c.renderCollapsed(w)
	}
	if c.Content != nil {
		return c.Content.View(w, h)
	}
	return ""
}

// Update forwards messages to the inner content panel.
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd {
	if c.Content != nil {
		return c.Content.Update(msg)
	}
	return nil
}

// Toggle switches between collapsed and expanded states.
func (c *Collapsible) Toggle() {
	c.Collapsed = !c.Collapsed
}

// renderCollapsed renders the title bar for a collapsed panel.
func (c *Collapsible) renderCollapsed(w int) string {
	if w <= 0 {
		return ""
	}

	indicator := "▶"
	if !c.Collapsed {
		indicator = "▼"
	}

	titleWidth := max(0, w-4) // indicator, spacing and corner glyphs
	title := ansi.Truncate(c.Title, titleWidth, "...")
	padding := max(0, w-4-ansi.StringWidth(title))
	line := collapsibleStyle.Render("┌"+indicator+" "+title) +
		collapsibleBorderStyle.Render(strings.Repeat("─", padding)+"┐")
	return padVisualLine(line, w)
}
