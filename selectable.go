package warp

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	if !s.HasSelection || s.Content == nil {
		return ""
	}

	// Use the same rendered lines as in View so that coordinates match exactly.
	lines := s.lastLines
	if len(lines) == 0 {
		// View hasn't been called yet (e.g. in unit tests). Fall back to a
		// large render so that selection bounds still map to content lines.
		w, h := s.lastW, s.lastH
		if w <= 0 {
			w = 9999
		}
		if h <= 0 {
			h = 9999
		}
		content := s.Content.View(w, h)
		lines = strings.Split(content, "\n")
	}

	sx, sy, ex, ey := s.sortedBounds()
	var parts []string
	for y := sy; y <= ey && y < len(lines); y++ {
		line := lines[y]
		lineVis := StripANSI(line)
		startX := 0
		endX := len(lineVis)
		if y == sy {
			startX = sx
		}
		if y == ey {
			endX = ex
		}
		if startX < 0 {
			startX = 0
		}
		if endX > len(lineVis) {
			endX = len(lineVis)
		}
		if startX < endX {
			parts = append(parts, extractVisRange(line, startX, endX))
		}
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
	s.AnchorX, s.AnchorY = 0, 0
	s.CursorX, s.CursorY = w-1, h-1
	s.HasSelection = true
}

// View renders the content with selection highlight.
func (s *Selectable) View(w, h int) string {
	if s.Content == nil {
		return strings.Repeat("\n", h)
	}

	// Get content first and remember the rendered lines so SelectedText can
	// extract text using the exact same coordinate system.
	content := s.Content.View(w, h)
	s.lastW = w
	s.lastH = h
	s.lastLines = strings.Split(content, "\n")

	// Clamp selection to panel bounds. Only clamp the cursor end; the anchor
	// must stay where the user originally pressed so reverse-direction drag
	// continues to work correctly.
	if s.HasSelection || s.Selecting {
		if s.AnchorX < 0 {
			s.AnchorX = 0
		}
		if s.AnchorX >= w {
			s.AnchorX = w - 1
		}
		if s.AnchorY < 0 {
			s.AnchorY = 0
		}
		if s.AnchorY >= h {
			s.AnchorY = h - 1
		}
		if s.CursorX < 0 {
			s.CursorX = 0
		}
		if s.CursorX >= w {
			s.CursorX = w - 1
		}
		if s.CursorY < 0 {
			s.CursorY = 0
		}
		if s.CursorY >= h {
			s.CursorY = h - 1
		}
		if s.CursorX == s.AnchorX && s.CursorY == s.AnchorY {
			if !s.Selecting {
				s.HasSelection = false
			}
		}
	}

	if !s.HasSelection {
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
		endX := lipgloss.Width(StripANSI(line))
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

// Update handles mouse and keyboard for selection.
func (s *Selectable) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case ResizeMsg:
		if s.Content != nil {
			return s.Content.Update(msg)
		}
		return nil
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonLeft:
			switch msg.Action {
			case tea.MouseActionPress:
				s.AnchorX = msg.X
				s.AnchorY = msg.Y
				s.CursorX = msg.X
				s.CursorY = msg.Y
				s.HasSelection = false
				s.Selecting = true
			case tea.MouseActionMotion:
				if s.Selecting {
					s.CursorX = msg.X
					s.CursorY = msg.Y
					s.HasSelection = true
				}
			case tea.MouseActionRelease:
				if s.Selecting {
					s.CursorX = msg.X
					s.CursorY = msg.Y
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
			switch key {
			case "shift+up":
				if s.CursorY > 0 {
					s.CursorY--
					s.HasSelection = true
					handled = true
				}
			case "shift+down":
				s.CursorY++
				s.HasSelection = true
				handled = true
			case "shift+left":
				if s.CursorX > 0 {
					s.CursorX--
					s.HasSelection = true
					handled = true
				}
			case "shift+right":
				s.CursorX++
				s.HasSelection = true
				handled = true
			}
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
		} else {
			// Shift+arrow extends selection
			if !s.HasSelection && !s.Selecting {
				s.AnchorX = s.CursorX
				s.AnchorY = s.CursorY
				s.HasSelection = true
			}
			s.Selecting = true
			switch key {
			case "shift+left":
				s.CursorX--
				if s.CursorX < 0 {
					s.CursorX = 0
				}
				handled = true
			case "shift+right":
				s.CursorX++
				handled = true
			case "shift+up":
				s.CursorY--
				if s.CursorY < 0 {
					s.CursorY = 0
				}
				handled = true
			case "shift+down":
				s.CursorY++
				handled = true
			}
			s.Selecting = false
		}

		if handled {
			return nil
		}
	}

	if s.Content != nil {
		return s.Content.Update(msg)
	}
	return nil
}

// sortedBounds returns selection bounds with start <= end.
func (s *Selectable) sortedBounds() (sx, sy, ex, ey int) {
	sx, sy = s.AnchorX, s.AnchorY
	ex, ey = s.CursorX, s.CursorY
	if sy > ey || (sy == ey && sx > ex) {
		sx, ex = ex, sx
		sy, ey = ey, sy
	}
	return
}

// highlightRange applies selection highlight to the visual range [startX, endX).
func highlightRange(line string, startX, endX int) string {
	if startX >= endX {
		return line
	}

	// Walk through the line tracking visual rune position so that Unicode and
	// ANSI sequences are handled correctly. The fast byte-index path is skipped
	// because terminal lines may contain multi-byte UTF-8 runes.
	var result strings.Builder
	result.Grow(len(line) + 20)

	visPos := 0
	i := 0
	inSelection := false

	for i < len(line) {
		if line[i] == '\x1b' {
			// Copy ANSI sequence
			start := i
			if i+1 < len(line) && line[i+1] == '[' {
				i += 2
				for i < len(line) && line[i] < 0x40 {
					i++
				}
				if i < len(line) {
					i++
				}
			} else {
				i++
			}
			esc := line[start:i]
			// If we're inside selection, close it before the escape,
			// then reopen after
			if inSelection {
				result.WriteString(resetStyle)
				result.WriteString(esc)
				result.WriteString(selectionStyleANSI)
			} else {
				result.WriteString(esc)
			}
			continue
		}

		r, size := utf8.DecodeRuneInString(line[i:])
		wasInSelection := inSelection
		inSelection = visPos >= startX && visPos < endX

		if !wasInSelection && inSelection {
			result.WriteString(selectionStyleANSI)
		}
		if wasInSelection && !inSelection {
			result.WriteString(resetStyle)
		}

		result.WriteRune(r)
		i += size
		visPos++
	}

	if inSelection {
		result.WriteString(resetStyle)
	}

	return result.String()
}

// extractVisRange extracts text from visual range [startX, endX).
func extractVisRange(line string, startX, endX int) string {
	var result strings.Builder
	visPos := 0
	for i := 0; i < len(line); {
		if line[i] == '\x1b' {
			if i+1 < len(line) && line[i+1] == '[' {
				i += 2
				for i < len(line) && line[i] < 0x40 {
					i++
				}
				if i < len(line) {
					i++
				}
			} else {
				i++
			}
			continue
		}
		if visPos >= startX && visPos < endX {
			r, size := utf8.DecodeRuneInString(line[i:])
			result.WriteRune(r)
			i += size
		} else if visPos >= endX {
			break
		} else {
			_, size := utf8.DecodeRuneInString(line[i:])
			i += size
		}
		visPos++
	}
	return result.String()
}

var (
	selectionStyleANSI = "\x1b[7m"
	resetStyle         = "\x1b[0m"
)
