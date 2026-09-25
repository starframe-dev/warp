package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

// FloatPane is a floating panel rendered on top of the main layout.
type FloatPane struct {
	Panel  Panel
	X, Y   int
	Width  int
	Height int
	Title  string

	preferredWidth  int
	preferredHeight int

	dragging   bool
	resizing   bool
	resizeEdge string
	dragStartX int
	dragStartY int
	origX      int
	origY      int
	origW      int
	origH      int

	// CloseRequested is set when the user clicks the × button.
	// The owning Tab checks this after handleMouse and calls CloseFloat.
	CloseRequested bool

	// CloseOnOutsideClick closes the float when the user clicks outside it.
	CloseOnOutsideClick bool
}

const (
	floatMinWidth  = 10
	floatMinHeight = 3
	floatTitleH    = 1
)

// render renders the float pane into lines.
func (fp *FloatPane) render(_, _ int) []string {
	if fp == nil || fp.Width <= 0 || fp.Height <= 0 {
		return nil
	}

	lines := make([]string, fp.Height)
	title := ansi.Truncate(fp.Title, max(0, fp.Width-4), "...")
	dashesW := max(0, fp.Width-ansi.StringWidth(title)-4)
	topBorder := floatBgStyle.Render("╭") + floatTitleStyle.Render(title) +
		floatBorderStyle.Render(strings.Repeat("─", dashesW)) +
		floatBgStyle.Render(" ") + floatCloseStyle.Render("×") + floatBgStyle.Render("╮")
	lines[0] = padVisualLine(topBorder, fp.Width)
	if fp.Height == 1 {
		return lines
	}

	contentW := max(0, fp.Width-2)
	contentH := max(0, fp.Height-2)
	content := ""
	if !isNilPanel(fp.Panel) {
		content = fp.Panel.View(contentW, contentH)
	}
	contentLines := padContent(content, contentW, contentH)
	for i, line := range contentLines {
		lines[i+1] = padVisualLine(
			floatBorderStyle.Render("│")+floatBgStyle.Render(line)+floatBorderStyle.Render("│"),
			fp.Width,
		)
	}

	if fp.Height > 1 {
		bottom := floatBgStyle.Render("╰") +
			floatBorderStyle.Render(strings.Repeat("─", max(0, fp.Width-2))) +
			floatBgStyle.Render("╯")
		lines[fp.Height-1] = padVisualLine(bottom, fp.Width)
	}
	return lines
}

// handleMouse processes mouse events for this float pane.
// mx, my are relative to the content area (not absolute screen).
// Returns tea.Cmd if the event was handled.
func (fp *FloatPane) handleMouse(msg tea.MouseMsg, mx, my int) tea.Cmd {
	return fp.handleMouseWithin(msg, mx, my, 0, 0)
}

func (fp *FloatPane) handleMouseWithin(msg tea.MouseMsg, mx, my, totalW, totalH int) tea.Cmd {
	if fp == nil || fp.Width <= 0 || fp.Height <= 0 {
		return nil
	}
	if !fp.dragging && !fp.resizing {
		if mx < fp.X || mx >= fp.X+fp.Width || my < fp.Y || my >= fp.Y+fp.Height {
			return nil
		}
	}

	relX := mx - fp.X
	relY := my - fp.Y
	switch msg.Button {
	case tea.MouseButtonLeft:
		switch msg.Action {
		case tea.MouseActionPress:
			if relY == 0 && relX == fp.Width-2 {
				fp.CloseRequested = true
				return nil
			}
			if relY == 0 && relX > 0 && relX < fp.Width-1 {
				fp.dragging = true
				fp.dragStartX = mx
				fp.dragStartY = my
				fp.origX, fp.origY = fp.X, fp.Y
				return nil
			}
			if edge := fp.hitEdge(relX, relY); edge != "" {
				fp.resizing = true
				fp.resizeEdge = edge
				fp.dragStartX = mx
				fp.dragStartY = my
				fp.origX, fp.origY = fp.X, fp.Y
				fp.origW, fp.origH = fp.Width, fp.Height
				return nil
			}
			if !isNilPanel(fp.Panel) && relY > 0 && relY < fp.Height-1 {
				innerMsg := tea.MouseMsg{
					Action: msg.Action,
					Button: msg.Button,
					X:      relX - 1,
					Y:      relY - 1,
				}
				return fp.Panel.Update(innerMsg)
			}

		case tea.MouseActionMotion:
			if fp.dragging {
				fp.X = fp.origX + mx - fp.dragStartX
				fp.Y = fp.origY + my - fp.dragStartY
				fp.clampPosition(totalW, totalH)
			}
			if fp.resizing {
				fp.applyResizeWithin(mx-fp.dragStartX, my-fp.dragStartY, totalW, totalH)
			}

		case tea.MouseActionRelease:
			fp.dragging = false
			fp.resizing = false
			fp.resizeEdge = ""
		}
	}
	return nil
}

func (fp *FloatPane) clampPosition(totalW, totalH int) {
	if fp == nil {
		return
	}
	fp.ensurePreferredSize()
	fp.X = max(0, fp.X)
	fp.Y = max(0, fp.Y)
	if totalW > 0 {
		fp.Width = min(fp.preferredWidth, totalW)
		fp.X = min(fp.X, max(0, totalW-fp.Width))
	} else {
		fp.Width = fp.preferredWidth
	}
	if totalH > 0 {
		fp.Height = min(fp.preferredHeight, totalH)
		fp.Y = min(fp.Y, max(0, totalH-fp.Height))
	} else {
		fp.Height = fp.preferredHeight
	}
}

func (fp *FloatPane) ensurePreferredSize() {
	if fp.preferredWidth <= 0 {
		fp.preferredWidth = max(floatMinWidth, fp.Width)
	}
	if fp.preferredHeight <= 0 {
		fp.preferredHeight = max(floatMinHeight, fp.Height)
	}
}

func (fp *FloatPane) hitEdge(x, y int) string {
	onTop := y == 0
	onBottom := y == fp.Height-1
	onLeft := x == 0
	onRight := x == fp.Width-1

	if onTop && onLeft {
		return "nw"
	}
	if onTop && onRight {
		return "ne"
	}
	if onBottom && onLeft {
		return "sw"
	}
	if onBottom && onRight {
		return "se"
	}
	if onTop {
		return "n"
	}
	if onBottom {
		return "s"
	}
	if onLeft {
		return "w"
	}
	if onRight {
		return "e"
	}
	return ""
}

func (fp *FloatPane) applyResize(dx, dy int) {
	fp.applyResizeWithin(dx, dy, 0, 0)
}

func (fp *FloatPane) applyResizeWithin(dx, dy, totalW, totalH int) {
	edge := fp.resizeEdge
	west := strings.Contains(edge, "w")
	east := strings.Contains(edge, "e")
	north := strings.Contains(edge, "n")
	south := strings.Contains(edge, "s")

	left, right := fp.origX, fp.origX+fp.origW
	top, bottom := fp.origY, fp.origY+fp.origH
	if west {
		left += dx
	} else if east {
		right += dx
	}
	if north {
		top += dy
	} else if south {
		bottom += dy
	}

	minW, minH := floatMinWidth, floatMinHeight
	if totalW > 0 {
		minW = min(minW, totalW)
		right = clampInt(right, 0, totalW)
		if west {
			left = clampInt(left, 0, max(0, right-minW))
		} else {
			left = clampInt(left, 0, max(0, totalW-minW))
			right = clampInt(right, min(totalW, left+minW), totalW)
		}
	} else if west {
		left = max(0, min(left, right-minW))
	} else {
		left = max(0, left)
		right = max(right, left+minW)
	}

	if totalH > 0 {
		minH = min(minH, totalH)
		bottom = clampInt(bottom, 0, totalH)
		if north {
			top = clampInt(top, 0, max(0, bottom-minH))
		} else {
			top = clampInt(top, 0, max(0, totalH-minH))
			bottom = clampInt(bottom, min(totalH, top+minH), totalH)
		}
	} else if north {
		top = max(0, min(top, bottom-minH))
	} else {
		top = max(0, top)
		bottom = max(bottom, top+minH)
	}

	fp.X, fp.Y = left, top
	fp.Width, fp.Height = max(0, right-left), max(0, bottom-top)
	fp.preferredWidth, fp.preferredHeight = fp.Width, fp.Height
}

func clampInt(value, low, high int) int {
	if high < low {
		return high
	}
	return max(low, min(value, high))
}

// StripANSI removes terminal control sequences from a string.
func StripANSI(s string) string {
	return ansi.Strip(s)
}

// overlayFloat draws the float pane on top of existing content lines.
func overlayFloat(lines []string, fp *FloatPane, totalW, totalH int) {
	if fp == nil || isNilPanel(fp.Panel) || fp.Width <= 0 || fp.Height <= 0 || totalW <= 0 || totalH <= 0 {
		return
	}
	x := max(0, fp.X)
	y := max(0, fp.Y)
	if x >= totalW || y >= totalH {
		return
	}
	visibleW := min(fp.Width, totalW-x)
	if visibleW <= 0 {
		return
	}
	floatLines := fp.render(totalW, totalH)
	for fy, floatLine := range floatLines {
		screenY := y + fy
		if screenY < 0 || screenY >= totalH || screenY >= len(lines) {
			continue
		}

		floatLine = ansi.Truncate(floatLine, visibleW, "")
		floatLine = padVisualLine(floatLine, visibleW)
		original := lines[screenY]

		prefix := ansi.Truncate(original, x, "")
		prefixWidth := ansi.StringWidth(prefix)
		if prefixWidth < x {
			prefix += strings.Repeat(" ", x-prefixWidth)
		}
		suffixStart := visualBytePosAfter(original, x+visibleW)
		suffix := ""
		if suffixStart < len(original) {
			suffix = original[suffixStart:]
		}

		combined := prefix + floatLine + ansi.ResetStyle + suffix
		lines[screenY] = padVisualLine(combined, totalW)
	}
}

func visualBytePosAfter(s string, targetW int) int {
	if targetW <= 0 {
		return 0
	}
	start := visualBytePos(s, targetW)
	if start >= len(s) || ansi.StringWidth(s[:start]) >= targetW {
		return start
	}
	cluster, _, _, _ := uniseg.FirstGraphemeCluster([]byte(s[start:]), -1)
	return start + len(cluster)
}
