package warp

import (
	"encoding/base64"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

// Selectable wraps a Panel with text selection support.
// Mouse drag and Shift+arrows create a selection.
// The selected region is rendered with reversed colors.
type Selectable struct {
	Content Panel

	// Selection anchor (fixed on press) and cursor (active end).
	// Both are in cell coordinates relative to the panel.
	AnchorX, AnchorY int
	CursorX, CursorY int
	mouseStartX      int
	mouseStartY      int

	HasSelection bool
	Selecting    bool // true during active mouse drag

	// Last rendered dimensions and content lines, used to extract selected
	// text that matches exactly what is currently visible.
	lastW, lastH int
	lastLines    []string
}

// NewSelectable creates a new selectable wrapper.
func NewSelectable(content Panel) *Selectable {
	return &Selectable{Content: content}
}

// SelectedText returns the currently selected text.
func (s *Selectable) SelectedText() string {
	if !s.HasSelection || isNilPanel(s.Content) {
		return ""
	}

	// Use the same rendered lines as in View so that coordinates match exactly.
	lines := s.lastLines
	if len(lines) == 0 {
		// View hasn't been called yet (e.g. in unit tests). Fall back to a
		// large render so that selection bounds still map to content lines.
		w, h := s.lastW, s.lastH
		if w <= 0 {
			w = 80
		}
		if h <= 0 {
			h = 24
		}
		content := s.Content.View(max(0, w), max(0, h))
		lines = strings.Split(content, "\n")
	}

	sx, sy, ex, ey := s.sortedBounds()
	var parts []string
	for y := sy; y <= ey && y < len(lines); y++ {
		line := lines[y]
		lineVis := ansi.Strip(line)
		startX := 0
		endX := ansi.StringWidth(lineVis)
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex
		}
		if startX < 0 {
			startX = 0
		}
		if endX > ansi.StringWidth(lineVis) {
			endX = ansi.StringWidth(lineVis)
		}
		parts = append(parts, extractVisRange(line, startX, endX))
	}
	return strings.Join(parts, "\n")
}

// ClearSelection removes the current selection.
func (s *Selectable) ClearSelection() {
	s.HasSelection = false
	s.Selecting = false
}

// Copy returns a tea.Cmd that copies the selected text to the system
// clipboard via OSC 52. Call this when the user presses Ctrl+C.
func (s *Selectable) Copy() tea.Cmd {
	text := s.SelectedText()
	if text == "" {
		return nil
	}
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	seq := fmt.Sprintf("\x1b]52;c;%s\x07", encoded)
	return func() tea.Msg {
		// OSC 52 works even in Bubbletea's alternate screen buffer
		fmt.Print(seq)
		return nil
	}
}

// SelectAll selects all visible content.
func (s *Selectable) SelectAll(w, h int) {
	if w <= 0 || h <= 0 {
		s.ClearSelection()
		return
	}
	if s.lastW != w || s.lastH != h {
		s.lastLines = nil
	}
	s.lastW, s.lastH = w, h
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = w, h-1
	s.HasSelection = true
}

// View renders the content with selection highlight.
func (s *Selectable) View(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	if isNilPanel(s.Content) {
		s.lastW, s.lastH = w, h
		s.lastLines = nil
		return emptyView(h)
	}

	// Get content first and remember the rendered lines so SelectedText can
	// extract text using the exact same coordinate system.
	content := s.Content.View(w, h)
	s.lastW = w
	s.lastH = h
	s.lastLines = strings.Split(content, "\n")

	if s.HasSelection || s.Selecting {
		s.clampSelection(w, h)
	}

	if !s.HasSelection || w == 0 || h == 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	result := make([]string, len(lines))

	sx, sy, ex, ey := s.sortedBounds()
	for y, line := range lines {
		if y < sy || y > ey {
			result[y] = line
			continue
		}
		startX := 0
		endX := ansi.StringWidth(line)
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex
		}
		result[y] = highlightRange(line, startX, endX)
	}
	return strings.Join(result, "\n")
}

// Elements transparently forwards semantic elements from the wrapped panel.
func (s *Selectable) Elements(w, h int) []Element {
	return cloneElements(collectElements(s.Content, max(0, w), max(0, h)))
}

// Update handles mouse and keyboard for selection.
func (s *Selectable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ResizeMsg:
		if !isNilPanel(s.Content) {
			return s.Content.Update(msg)
		}
		return nil
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			switch msg.Action {
			case tea.MouseActionPress:
				s.mouseStartX = int(msg.X)
				s.mouseStartY = int(msg.Y)
				s.AnchorX = int(msg.X)
				s.AnchorY = int(msg.Y)
				s.CursorX = int(msg.X)
				s.CursorY = int(msg.Y)
				s.HasSelection = false
				s.Selecting = true
			case tea.MouseActionMotion:
				if s.Selecting {
					s.setMouseSelection(int(msg.X), int(msg.Y))
					s.HasSelection = true
				}
			case tea.MouseActionRelease:
				if s.Selecting {
					if s.HasSelection {
						s.setMouseSelection(int(msg.X), int(msg.Y))
					}
					s.Selecting = false
					if s.AnchorX == s.CursorX && s.AnchorY == s.CursorY {
						s.HasSelection = false
					}
				}
			}
		}
	case tea.KeyMsg:
		key := msg.String()
		handled := false

		// Selection keys — shift-modified arrows. Shift+Tab is intentionally
		// excluded and proxied to the wrapped panel (terminal) so TUI apps
		// inside the PTY receive it.
		if strings.HasPrefix(key, "shift+") && key != "shift+tab" {
			if !s.HasSelection && !s.Selecting {
				s.AnchorX = s.CursorX
				s.AnchorY = s.CursorY
			}
			s.Selecting = true
			s.HasSelection = true
			switch key {
			case "shift+up":
				if s.CursorY > 0 {
					s.CursorY--
				}
			case "shift+down":
				if s.lastH <= 0 {
					s.CursorY++
				} else {
					s.CursorY = min(s.CursorY+1, s.lastH-1)
				}
			case "shift+left":
				if s.CursorX > 0 {
					s.CursorX--
				}
			case "shift+right":
				if s.lastW <= 0 {
					s.CursorX++
				} else {
					s.CursorX = min(s.CursorX+1, s.lastW)
				}
			}
			s.Selecting = false
			handled = true
		}

		if !handled {
			switch key {
			case "ctrl+a":
				w, h := s.lastW, s.lastH
				if w <= 0 {
					w = 80
				}
				if h <= 0 {
					h = 24
				}
				s.SelectAll(w, h)
				handled = true
			case "esc":
				if s.HasSelection {
					s.ClearSelection()
					handled = true
				}
			}
		}

		if handled {
			return nil
		}
	}

	if !isNilPanel(s.Content) {
		return s.Content.Update(msg)
	}
	return nil
}

func (s *Selectable) clampSelection(w, h int) {
	if w <= 0 || h <= 0 {
		s.AnchorX, s.AnchorY = 0, 0
		s.CursorX, s.CursorY = 0, 0
		s.ClearSelection()
		return
	}
	s.AnchorX = min(max(0, s.AnchorX), w)
	s.CursorX = min(max(0, s.CursorX), w)
	s.AnchorY = min(max(0, s.AnchorY), h-1)
	s.CursorY = min(max(0, s.CursorY), h-1)
	if s.AnchorX == s.CursorX && s.AnchorY == s.CursorY && !s.Selecting {
		s.HasSelection = false
	}
}

func (s *Selectable) setMouseSelection(x, y int) {
	startX, startY := s.mouseStartX, s.mouseStartY
	if s.lastW > 0 {
		startX = min(max(0, startX), s.lastW-1)
		x = min(max(0, x), s.lastW-1)
	}
	if s.lastH > 0 {
		startY = min(max(0, startY), s.lastH-1)
		y = min(max(0, y), s.lastH-1)
	}

	if y > startY || (y == startY && x >= startX) {
		s.AnchorX, s.AnchorY = startX, startY
		s.CursorX, s.CursorY = x+1, y
		return
	}
	s.AnchorX, s.AnchorY = startX+1, startY
	s.CursorX, s.CursorY = x, y
}

// sortedBounds returns end-exclusive selection bounds with start <= end.
func (s *Selectable) sortedBounds() (sx, sy, ex, ey int) {
	sx, sy = s.AnchorX, s.AnchorY
	ex, ey = s.CursorX, s.CursorY
	if sy > ey || (sy == ey && sx > ex) {
		sx, ex = ex, sx
		sy, ey = ey, sy
	}
	return
}

// highlightRange applies selection highlight to the terminal-cell range [startX, endX).
func highlightRange(line string, startX, endX int) string {
	if startX >= endX {
		return line
	}
	startX = max(0, startX)
	var result strings.Builder
	result.Grow(len(line) + 20)
	cellPos := 0
	inSelection := false

	for i := 0; i < len(line); {
		if line[i] == '\x1b' {
			end := ansiSequenceEnd(line, i)
			sequence := line[i:end]
			if inSelection {
				result.WriteString(resetStyle)
				result.WriteString(sequence)
				result.WriteString(selectionStyleANSI)
			} else {
				result.WriteString(sequence)
			}
			i = end
			continue
		}

		cluster, _, width, _ := uniseg.FirstGraphemeCluster([]byte(line[i:]), -1)
		if len(cluster) == 0 {
			break
		}
		selected := cellPos < endX && cellPos+width > startX
		if selected && !inSelection {
			result.WriteString(selectionStyleANSI)
		} else if !selected && inSelection {
			result.WriteString(resetStyle)
		}
		result.Write(cluster)
		i += len(cluster)
		cellPos += width
		inSelection = selected
	}
	if inSelection {
		result.WriteString(resetStyle)
	}
	return result.String()
}

func ansiSequenceEnd(text string, start int) int {
	if start+1 >= len(text) {
		return len(text)
	}
	switch text[start+1] {
	case '[':
		end := start + 2
		for end < len(text) && text[end] < 0x40 {
			end++
		}
		if end < len(text) {
			return end + 1
		}
		return end
	case ']':
		for end := start + 2; end < len(text); end++ {
			if text[end] == '\a' {
				return end + 1
			}
			if text[end] == '\x1b' && end+1 < len(text) && text[end+1] == '\\' {
				return end + 2
			}
		}
		return len(text)
	default:
		return min(len(text), start+2)
	}
}

// extractVisRange extracts complete graphemes overlapping terminal-cell range [startX, endX).
func extractVisRange(line string, startX, endX int) string {
	if startX >= endX {
		return ""
	}
	plain := ansi.Strip(line)
	var result strings.Builder
	cellPos := 0
	graphemes := uniseg.NewGraphemes(plain)
	for graphemes.Next() {
		cluster := graphemes.Str()
		width := ansi.StringWidth(cluster)
		if cellPos < endX && cellPos+width > startX {
			result.WriteString(cluster)
		}
		cellPos += width
	}
	return result.String()
}

var (
	selectionStyleANSI = "\x1b[7m"
	resetStyle         = "\x1b[0m"
)
