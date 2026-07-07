package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// PopoverItem is a single item in a context menu.
type PopoverItem struct {
	Name   string
	Action func()
}

// Popover is a context menu rendered at a specific position.
// It overlays on top of existing content lines, preserving content left/right.
type Popover struct {
	Items   []PopoverItem
	X, Y    int       // Position in screen coordinates (0 = header row)
	Width   int       // Content width (0 = auto: 20)
	OnClose func()    // Called when popover is closed

	// rendered dimensions (set by Overlay)
	boxW, boxH int

	// selection
	selected int
}

// visualBytePos returns the byte index at the given visual column (0-indexed).
// It handles ANSI escape codes by skipping them.
func visualBytePos(s string, col int) int {
	visual := 0
	pos := 0
	for i, r := range s {
		if r == '\x1b' {
			// Skip ANSI escape sequence
			for j := i + 1; j < len(s) && s[j] != 'm'; j++ {
				i = j
			}
			if i+1 < len(s) && s[i+1] == 'm' {
				i++
			}
		}
		visual++
		if visual > col {
			pos = i + 1
			break
		}
		pos = i + 1
	}
	return pos
}

// Overlay renders the popover on top of existing content lines.
// Returns the overlaid lines.
func (p *Popover) Overlay(lines []string, totalW, totalH int) []string {
	if len(p.Items) == 0 {
		return lines
	}

	contentW := p.Width
	if contentW <= 0 {
		contentW = 20
	}
	if contentW > totalW {
		contentW = totalW
	}

	// Build menu content lines.
	var contentLines []string
	for i, item := range p.Items {
		var line string
		if i == p.selected {
			line = popoverSelectedStyle.Width(contentW).Render(" " + item.Name)
		} else {
			line = popoverBaseStyle.Width(contentW).Render(" " + item.Name)
		}
		contentLines = append(contentLines, line)
	}

	// Wrap in a border box.
	menuContent := strings.Join(contentLines, "\n")
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(gbDark4)).
		Background(lipgloss.Color(gbDark1))
	boxStr := boxStyle.Render(menuContent)
	boxLines := strings.Split(boxStr, "\n")

	if len(boxLines) == 0 {
		return lines
	}

	// Save actual box dimensions for hit-test.
	p.boxW = lipgloss.Width(boxLines[0])
	p.boxH = len(boxLines)

	boxW := p.boxW
	boxH := p.boxH

	// Position: Y is screen-relative (0 = header), lines are content rows.
	menuX := p.X
	menuY := p.Y // lines include header at 0, same as mouse Y

	// Clamp to fit within terminal.
	if menuX+boxW > totalW {
		menuX = totalW - boxW
	}
	if menuX < 0 {
		menuX = 0
	}
	if menuY < 0 {
		menuY = 0
	}
	if menuY+boxH > len(lines) {
		menuY = len(lines) - boxH
		if menuY < 0 {
			menuY = 0
		}
	}

	// Place box lines, preserving original content left/right.
	for i, boxLine := range boxLines {
		idx := menuY + i
		if idx >= len(lines) {
			break
		}

		orig := lines[idx]

		// Left part: first menuX visual columns.
		leftPart := ansi.Truncate(orig, menuX, "")
		// Right part: everything after menuX+boxW visual columns.
		rightStart := visualBytePos(orig, menuX+boxW)
		if rightStart < len(orig) {
			rightPart := orig[rightStart:]
		}

		// Pad left part to menuX visual columns.
		leftWidth := ansi.StringWidth(leftPart)
		if leftWidth < menuX {
			leftPart += strings.Repeat(" ", menuX-leftWidth)
		}

		lines[idx] = leftPart + boxLine + rightPart
	}

	return lines
}

// HandleMouse processes mouse events for the popover.
// Returns true if the event was consumed.
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool {
	if p.boxW == 0 {
		return false
	}

	menuX := p.X
	menuY := p.Y // lines include header at 0, same as mouse Y
	boxW := p.boxW
	boxH := p.boxH

	// Recompute clamped position (same as Overlay).
	// We use the stored boxW/boxH and original X/Y.
	// The clamped position is computed fresh each time for consistency.

	switch msg.Action {
	case tea.MouseActionPress:
		// Check if click is inside the menu box.
		if int(msg.X) >= menuX && int(msg.X) < menuX+boxW &&
			int(msg.Y) >= menuY && int(msg.Y) < menuY+boxH {
			// Inside menu: check which item.
			contentIdx := int(msg.Y) - menuY - 1 // -1 for top border
			if contentIdx >= 0 && contentIdx < len(p.Items) {
				p.Items[contentIdx].Action()
				if p.OnClose != nil {
					p.OnClose()
				}
			}
			return true
		}
		// Click outside menu → close.
		if p.OnClose != nil {
			p.OnClose()
		}
		return true

	case tea.MouseActionMotion:
		// Track hover for visual feedback.
		if int(msg.Y) >= menuY+1 && int(msg.Y) < menuY+boxH-1 {
			contentIdx := int(msg.Y) - menuY - 1
			if contentIdx >= 0 && contentIdx < len(p.Items) {
				p.selected = contentIdx
			}
		}
		return true

	case tea.MouseActionRelease:
		return true
	}

	return false
}

// HandleKey processes keyboard events for the popover.
// Returns true if the event was consumed.
func (p *Popover) HandleKey(msg tea.KeyMsg) bool {
	if p.boxW == 0 {
		return false
	}

	switch msg.Type {
	case tea.KeyEsc:
		if p.OnClose != nil {
			p.OnClose()
		}
		return true
	case tea.KeyEnter:
		if p.selected >= 0 && p.selected < len(p.Items) {
			p.Items[p.selected].Action()
		}
		if p.OnClose != nil {
			p.OnClose()
		}
		return true
	case tea.KeyUp:
		p.selected--
		if p.selected < 0 {
			p.selected = 0
		}
		return true
	case tea.KeyDown:
		p.selected++
		if p.selected >= len(p.Items) {
			p.selected = len(p.Items) - 1
		}
		return true
	}

	return false
}

// Popover styles.
var (
	popoverBaseStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(gbDark1)).
				Foreground(lipgloss.Color(gbLight1))

	popoverSelectedStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(gbBlue)).
				Foreground(lipgloss.Color(gbDark0))
)
