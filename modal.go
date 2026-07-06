package warp

import (
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

// ShowModalMsg tells a panel to show a modal dialog.
type ShowModalMsg struct {
	Title   string
	Content string
	Buttons []ModalButton
	OnClose func()
	Width   int
}

// CloseModalMsg tells a panel to close the current modal.
type CloseModalMsg struct{}

// Modal is a dialog rendered on top of a panel's content.
// Any panel can embed and render a Modal in its View().
type Modal struct {
	Title   string        // title (displayed in top border area)
	Content string        // rendered content (ANSI string)
	Buttons []ModalButton // buttons at the bottom
	OnClose func()        // called when modal is closed via ✕ or Esc
	Width   int           // box width (0 = auto: 3/5 of screen, clamped 30-50)

	// Rendered box position/size (set during Overlay)
	startX, startY int
	boxWidth       int
	boxHeight      int

	// Drag state
	dragging bool
	dragX    int
	dragY    int
	offsetX  int
	offsetY  int

	// dimsSet is true after ensureDimensions has run
	dimsSet bool

	// last known total dimensions (set by EnsureDimensions)
	totalW int
	totalH int
}

// ModalButton describes a button in a modal dialog.
type ModalButton struct {
	Label  string
	Action func()
}

// NewModal creates a Modal from its parts.
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal {
	return &Modal{
		Title:   title,
		Content: content,
		Buttons: buttons,
		OnClose: onClose,
	}
}

// EnsureDimensions computes and stores box dimensions if not yet set.
// Must be called before HandleMouse if Overlay hasn't been called yet.
func (m *Modal) EnsureDimensions(totalW, totalH int) {
	if m.dimsSet {
		return
	}
	m.dimsSet = true

	boxWidth := m.Width
	if boxWidth <= 0 {
		boxWidth = totalW * 3 / 5
	}
	if boxWidth < 30 {
		boxWidth = 30
	}
	if boxWidth > 50 {
		boxWidth = 50
	}
	if boxWidth > totalW {
		boxWidth = totalW
	}
	m.boxWidth = boxWidth

	m.totalW = totalW
	m.totalH = totalH

	// Compute boxHeight without lipgloss (same logic as Overlay)
	// RoundedBorder + Padding(1,2) with 3 content lines => 7 lines
	m.boxHeight = 7

	// Center + drag offset
	centerX := (totalW - boxWidth) / 2
	centerY := (totalH - m.boxHeight) / 2
	if centerX < 0 {
		centerX = 0
	}
	if centerY < 0 {
		centerY = 0
	}
	m.startX = centerX + m.offsetX
	m.startY = centerY + m.offsetY
}

// Overlay renders the modal on top of existing content lines.
// Returns the overlaid lines.
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string {
	if totalW <= 0 || len(lines) == 0 {
		return lines
	}

	m.EnsureDimensions(totalW, totalH)
	boxWidth := m.boxWidth
	startX := m.startX
	startY := m.startY

	innerWidth := boxWidth - 6 // borders (2) + horizontal padding (4)

	// Build content lines
	titleLine := m.Title + strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(m.Title)-1)) + "✕"

	contentLine := m.Content
	contentWidth := lipgloss.Width(contentLine)
	if contentWidth > innerWidth {
		contentLine = ansi.Truncate(contentLine, innerWidth-1, "") + "…"
	}
	contentLine += strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(contentLine)))

	var btnParts []string
	for _, btn := range m.Buttons {
		btnParts = append(btnParts, "["+btn.Label+"]")
	}
	btnLine := strings.Join(btnParts, "  ")
	btnLine += strings.Repeat(" ", max(0, innerWidth-lipgloss.Width(btnLine)))

	box := modalBorderStyle.Width(boxWidth - 2).Render(titleLine + "\n" + contentLine + "\n" + btnLine)
	boxLines := strings.Split(box, "\n")
	if len(boxLines) == 0 {
		return lines
	}

	// Dim background
	for i := range lines {
		lines[i] = dimStyle.Copy().Background(lipgloss.Color(gbDark0)).Render(stripANSI(lines[i]))
	}

	// Overlay box, preserving dimmed content left and right
	for i, bl := range boxLines {
		if startY+i >= len(lines) {
			break
		}
		original := lines[startY+i]

		leftPart := ansi.Truncate(original, startX, "")
		rightStart := visualBytePos(original, startX+boxWidth)
		rightPart := ""
		if rightStart < len(original) {
			rightPart = original[rightStart:]
			if ansiRe.ReplaceAllString(rightPart, "") == "" {
				rightPart = ""
			}
		}

		leftWidth := ansi.StringWidth(leftPart)
		if leftWidth < startX {
			leftPart += strings.Repeat(" ", startX-leftWidth)
		}

		lines[startY+i] = leftPart + bl + rightPart
	}

	return lines
}

// HandleMouse processes mouse events for the modal.
// Returns true if the event was consumed.
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool {
	if m.boxHeight == 0 {
		return false
	}

	startX := m.startX
	startY := m.startY
	boxWidth := m.boxWidth

	// box layout with RoundedBorder + Padding(1,2), 3 content lines:
	//   box[0]: top border    → startY + 0
	//   box[1]: top padding   → startY + 1  (draggable strip)
	//   box[2]: title + ✕     → startY + 2
	//   box[3]: content       → startY + 3
	//   box[4]: buttons       → startY + 4
	//   box[5]: bottom pad    → startY + 5
	//   box[6]: bottom border → startY + 6
	//
	// msg.Y is lines-relative (0 = first content line, no header).
	// Tree.handleMouse adjusts screen Y → lines Y before calling HandleMouse.
	titleY := startY + 1  // draggable padding strip
	xBtnY := startY + 2   // ✕ on title line / also draggable
	btnY := startY + 4    // buttons line (first row)
	btnY2 := startY + 5   // buttons line (second row, when wrap)

	// ✕ is at innerWidth-1 within content area, content starts at startX+2 padding left
	// innerWidth = boxWidth - 6
	// ✕ X = startX + 2 + innerWidth - 1 = startX + 1 + boxWidth - 6 = startX + boxWidth - 5
	// ✕ is at the last visual column of the inner content.
	// Content starts at startX+2 (border + padding left), innerWidth = boxWidth-6.
	// ✕ is at visual index (innerWidth-1) within content = startX+2+innerWidth-1 = startX+boxWidth-5.
	// BUT the rendered box string has │ at col 0, padding at 1-2, content at 3-(boxWidth-3).
	// In the rendered 48-char box, ✕ is at string index 44 = boxWidth-4.
	xBtnX := startX + boxWidth - 4

	switch msg.Action {
	case tea.MouseActionPress:
		if msg.Button == tea.MouseButtonLeft {
			// Check ✕ close button
			if int(msg.Y) == xBtnY && int(msg.X) == xBtnX {
				if m.OnClose != nil {
					m.OnClose()
				}
				return true
			}

			// Buttons: content area starts at startX+3 (border(1) + padding(2))
			btnLine := m.buildButtonLine()
			offset := 0
			for _, btn := range m.Buttons {
				btnStart, btnEnd := findBracketPair(btnLine, offset)
				if btnStart < 0 {
					break
				}
				// content starts at startX+3 (1 border + 2 padding)
				btnX1 := startX + 3 + btnStart
				btnX2 := startX + 3 + btnEnd
				if (int(msg.Y) == btnY || int(msg.Y) == btnY2) && int(msg.X) >= btnX1 && int(msg.X) < btnX2 {
					if btn.Action != nil {
						btn.Action()
					}
					return true
				}
				offset = btnEnd
			}

			// Start dragging only on the padding strip (startY+1), not the title line.
			if int(msg.Y) == titleY && int(msg.X) >= startX && int(msg.X) < startX+boxWidth {
				m.dragging = true
				m.dragX = int(msg.X)
				m.dragY = int(msg.Y)
				return true
			}
		}

	case tea.MouseActionMotion:
		if m.dragging {
			dx := int(msg.X) - m.dragX
			dy := int(msg.Y) - m.dragY
			m.offsetX += dx
			m.offsetY += dy
			m.dragX = int(msg.X)
			m.dragY = int(msg.Y)
			m.startX += dx
			m.startY += dy

			// Clamp to prevent dragging past edges
			if m.startX < 0 {
				m.offsetX -= m.startX
				m.startX = 0
			}
			if m.startY < 0 {
				m.offsetY -= m.startY
				m.startY = 0
			}
			if m.totalW > 0 && m.startX+m.boxWidth > m.totalW {
				over := m.startX + m.boxWidth - m.totalW
				m.offsetX -= over
				m.startX = m.totalW - m.boxWidth
			}
			if m.totalH > 0 && m.startY+m.boxHeight > m.totalH {
				over := m.startY + m.boxHeight - m.totalH
				m.offsetY -= over
				m.startY = m.totalH - m.boxHeight
			}
			return true
		}

	case tea.MouseActionRelease:
		if m.dragging {
			m.dragging = false
			return true
		}
	}

	return false
}

func (m *Modal) buildButtonLine() string {
	var btnParts []string
	for _, btn := range m.Buttons {
		btnParts = append(btnParts, "["+btn.Label+"]")
	}
	return strings.Join(btnParts, "  ")
}

func findBracketPair(s string, pos int) (int, int) {
	start := strings.Index(s[pos:], "[")
	if start < 0 {
		return -1, -1
	}
	start += pos
	end := strings.Index(s[start:], "]")
	if end < 0 {
		return -1, -1
	}
	end += start + 1
	return start, end
}

// --- Styles ---

var (
	modalBorderStyle = lipgloss.NewStyle().
				Background(lipgloss.Color(gbDark1)).
				Foreground(lipgloss.Color(gbLight1)).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(gbBlue)).
				Padding(1, 2)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(gbGray)).
			Background(lipgloss.Color(gbDark0))
)

// visualBytePos returns the byte position in s where the visual width reaches targetW.
// If the visual width never reaches targetW, returns len(s).
func visualBytePos(s string, targetW int) int {
	if targetW <= 0 {
		return 0
	}

	curWidth := 0
	pstate := parser.GroundState
	b := []byte(s)
	i := 0

	for i < len(b) {
		state, action := parser.Table.Transition(pstate, b[i])
		if state == parser.Utf8State {
			cluster, _, width, _ := uniseg.FirstGraphemeCluster(b[i:], -1)
			i += len(cluster)
			if curWidth+width > targetW {
				return i - len(cluster)
			}
			curWidth += width
			pstate = parser.GroundState
			continue
		}

		switch action {
		case parser.PrintAction:
			if curWidth >= targetW {
				return i
			}
			curWidth++
			fallthrough
		default:
			i++
		}

		pstate = state

		if curWidth > targetW {
			return i - 1
		}
	}

	return len(s)
}

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// StartX returns the X position of the modal box after Overlay() was called.
func (m *Modal) StartX() int { return m.startX }

// StartY returns the Y position of the modal box after Overlay() was called.
func (m *Modal) StartY() int { return m.startY }

// BoxWidth returns the width of the modal box after Overlay() was called.
func (m *Modal) BoxWidth() int { return m.boxWidth }

// BoxHeight returns the height of the modal box after Overlay() was called.
func (m *Modal) BoxHeight() int { return m.boxHeight }