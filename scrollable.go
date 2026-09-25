package warp

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Scrollable wraps a Panel with scroll support.
// When content exceeds the allocated height, the user can scroll via mouse wheel.
type Scrollable struct {
	Content Panel
	Offset  int // scroll offset in lines
}

// NewScrollable creates a new scrollable wrapper.
func NewScrollable(content Panel) *Scrollable {
	return &Scrollable{Content: content}
}

// View renders the visible viewport of the content.
func (s *Scrollable) View(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	if s.Offset < 0 {
		s.Offset = 0
	}
	if isNilPanel(s.Content) {
		return emptyView(h)
	}
	if h == 0 {
		return ""
	}

	// Render only through the end of the requested viewport, not an arbitrary height.
	requestHeight := s.requestedContentHeight(h)
	fullContent := s.Content.View(w, requestHeight)
	lines := strings.Split(fullContent, "\n")

	maxOffset := max(0, len(lines)-h)
	if s.Offset > maxOffset {
		s.Offset = maxOffset
	}

	visible := make([]string, h)
	for i := range visible {
		idx := s.Offset + i
		if idx < len(lines) {
			visible[i] = padLine(lines[idx], w)
		} else {
			visible[i] = strings.Repeat(" ", w)
		}
	}
	return strings.Join(visible, "\n")
}

// Elements returns semantic elements translated into the visible viewport.
func (s *Scrollable) Elements(w, h int) []Element {
	w = max(0, w)
	h = max(0, h)
	if s.Offset < 0 {
		s.Offset = 0
	}
	if isNilPanel(s.Content) || w == 0 || h == 0 {
		return nil
	}

	requestHeight := s.requestedContentHeight(h)
	visibleHeight := requestHeight - s.Offset
	elements := collectElements(s.Content, w, requestHeight)
	elements = clipElements(elements, Bounds{X: 0, Y: s.Offset, W: w, H: visibleHeight})
	shiftElements(elements, 0, -s.Offset)
	return elements
}

func (s *Scrollable) requestedContentHeight(viewportHeight int) int {
	requestHeight := s.Offset + viewportHeight
	if requestHeight < s.Offset {
		return int(^uint(0) >> 1)
	}
	return requestHeight
}

// Update handles scroll messages (mouse wheel, keys).
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd {
	if s.Offset < 0 {
		s.Offset = 0
	}
	scrollBy := func(delta int) {
		if delta < 0 {
			if s.Offset < -delta {
				s.Offset = 0
			} else {
				s.Offset += delta
			}
			return
		}
		maxInt := int(^uint(0) >> 1)
		if s.Offset > maxInt-delta {
			s.Offset = maxInt
		} else {
			s.Offset += delta
		}
	}

	switch msg := msg.(type) {
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			scrollBy(-3)
		case tea.MouseButtonWheelDown:
			scrollBy(3)
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			scrollBy(-1)
		case "down":
			scrollBy(1)
		case "pgup":
			scrollBy(-10)
		case "pgdown":
			scrollBy(10)
		}
	}

	if !isNilPanel(s.Content) {
		return s.Content.Update(msg)
	}
	return nil
}

func padLine(line string, w int) string {
	return padVisualLine(line, max(0, w))
}
