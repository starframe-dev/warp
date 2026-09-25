package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// DropdownItem is a single item in a dropdown menu.
type DropdownItem struct {
	Label    string
	Selected bool
}

// DropdownMenu is a panel that shows a button and an expandable list of items.
type DropdownMenu struct {
	Label    string
	Items    []DropdownItem
	Open     bool
	Hovered  int // index of hovered item, -1 if none
	OnSelect func(idx int)

	visibleItemCount int
	menuLayoutKnown  bool
}

// NewDropdownMenu creates a dropdown menu.
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu {
	return &DropdownMenu{
		Label:   label,
		Items:   items,
		Hovered: -1,
	}
}

// View renders the dropdown button or the open menu.
func (d *DropdownMenu) View(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	if !d.Open {
		return d.renderButton(w)
	}
	return d.renderMenu(w, h)
}

func (d *DropdownMenu) renderButton(w int) string {
	label := d.Label + " ▼"
	if ansi.StringWidth(label) > w {
		label = ansi.Truncate(label, w, "…")
	}
	return dropdownButtonStyle.Render(padRight(label, w))
}

func (d *DropdownMenu) renderMenu(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	menuH := min(len(d.Items)+1, h)
	d.visibleItemCount = max(0, menuH-1)
	d.menuLayoutKnown = true
	d.normalizeHovered()
	if menuH == 0 {
		return ""
	}

	lines := make([]string, menuH)
	button := d.Label + " ▲"
	if ansi.StringWidth(button) > w {
		button = ansi.Truncate(button, w, "…")
	}
	lines[0] = dropdownButtonStyle.Render(padRight(button, w))

	for i := 0; i < len(d.Items) && i+1 < menuH; i++ {
		item := d.Items[i]
		prefix := "  "
		if item.Selected {
			prefix = "✓ "
		}
		label := prefix + item.Label
		if ansi.StringWidth(label) > w {
			label = ansi.Truncate(label, w, "…")
		}
		style := dropdownItemStyle
		if i == d.Hovered {
			style = dropdownItemHoverStyle
		}
		if item.Selected {
			style = dropdownItemSelectedStyle
		}
		lines[i+1] = style.Render(padRight(label, w))
	}
	return strings.Join(lines, "\n")
}

// Update handles mouse and keyboard for the dropdown.
func (d *DropdownMenu) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionMotion {
			if d.Open {
				idx := msg.Y - 1
				if idx >= 0 && idx < d.hoverableItemCount() {
					d.Hovered = idx
				} else {
					d.Hovered = -1
				}
			}
			return nil
		}
		if msg.Action != tea.MouseActionPress {
			return nil
		}
		if !d.Open {
			// Click on button opens menu
			if msg.Y == 0 {
				d.Open = true
				d.Hovered = -1
				d.menuLayoutKnown = false
			}
			return nil
		}
		// Click inside menu
		if msg.Y == 0 {
			// Click on button closes menu
			d.Open = false
			return nil
		}
		idx := msg.Y - 1
		if idx >= 0 && idx < d.hoverableItemCount() {
			d.selectItem(idx)
		}
	case tea.KeyMsg:
		if !d.Open {
			return nil
		}
		d.normalizeHovered()
		switch msg.String() {
		case "up":
			if d.Hovered > 0 {
				d.Hovered--
			}
		case "down":
			if d.Hovered < d.hoverableItemCount()-1 {
				d.Hovered++
			}
		case "enter":
			if d.Hovered >= 0 && d.Hovered < d.hoverableItemCount() {
				d.selectItem(d.Hovered)
			}
		case "esc":
			d.Open = false
		}
	}
	return nil
}

func (d *DropdownMenu) hoverableItemCount() int {
	if d.menuLayoutKnown {
		return min(len(d.Items), d.visibleItemCount)
	}
	// Before the first open-menu render, the visible item count is unknown.
	return len(d.Items)
}

func (d *DropdownMenu) normalizeHovered() {
	if d.Hovered < -1 || d.Hovered >= d.hoverableItemCount() {
		d.Hovered = -1
	}
}

func (d *DropdownMenu) selectItem(idx int) {
	for i := range d.Items {
		d.Items[i].Selected = false
	}
	d.Items[idx].Selected = true
	d.Open = false
	if d.OnSelect != nil {
		d.OnSelect(idx)
	}
}

// Close closes the dropdown menu.
func (d *DropdownMenu) Close() {
	d.Open = false
}
