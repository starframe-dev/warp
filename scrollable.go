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

	lastWidth, lastHeight int
	hasViewport           bool
	hasLocalResize        bool
	extentCache           scrollableExtentCache
}

type scrollableExtentCache struct {
	content Panel
	width   int
	height  int
	known   bool
}

type scrollableViewport struct {
	offset int
	lines  []string
	probed bool
	known  bool
}

// NewScrollable creates a new scrollable wrapper.
func NewScrollable(content Panel) *Scrollable {
	return &Scrollable{Content: content}
}

// View renders the visible viewport of the content.
func (s *Scrollable) View(w, h int) string {
	w = max(0, w)
	h = max(0, h)
	s.rememberViewport(w, h)

	viewport := s.effectiveOffset(w, h)
	s.Offset = viewport.offset
	if isNilPanel(s.Content) {
		return emptyView(h)
	}
	if h == 0 {
		return ""
	}

	lines := viewport.lines
	lineOffset := viewport.offset
	if viewport.known {
		if renderer, ok := s.Content.(ViewportRenderer); ok {
			lines = strings.Split(renderer.ViewAt(w, h, viewport.offset), "\n")
			lineOffset = 0
		} else if !viewport.probed {
			content := s.Content.View(w, scrollableRequestHeight(viewport.offset, h))
			lines = strings.Split(content, "\n")
		}
	} else if !viewport.probed {
		content := s.Content.View(w, scrollableRequestHeight(viewport.offset, h))
		lines = strings.Split(content, "\n")
	}

	visible := make([]string, h)
	for i := range visible {
		idx := saturatingAddNonNegative(lineOffset, i)
		if idx < len(lines) {
			visible[i] = padLine(lines[idx], w)
		} else {
			visible[i] = strings.Repeat(" ", w)
		}
	}
	return strings.Join(visible, "\n")
}

// Elements returns semantic elements translated into the visible viewport.
// It uses the effective offset without changing the stored Offset.
func (s *Scrollable) Elements(w, h int) []Element {
	w = max(0, w)
	h = max(0, h)
	if isNilPanel(s.Content) || w == 0 || h == 0 {
		return nil
	}

	viewport := s.effectiveOffset(w, h)
	requestHeight := scrollableRequestHeight(viewport.offset, h)
	var sourceElements []Element
	if viewport.known {
		if provider, ok := s.Content.(ViewportElementProvider); ok {
			sourceElements = provider.ElementsAt(w, h, viewport.offset)
		} else {
			sourceElements = collectElements(s.Content, w, requestHeight)
		}
	} else {
		sourceElements = collectElements(s.Content, w, requestHeight)
	}
	elements := cloneElements(sourceElements)
	elements = clipElements(elements, Bounds{X: 0, Y: viewport.offset, W: w, H: h})
	shiftElements(elements, 0, -viewport.offset)
	return elements
}

// ContentHeight is unknown because a Scrollable viewport is not an intrinsic
// content extent for a containing Scrollable.
func (*Scrollable) ContentHeight(int) (int, bool) {
	return 0, false
}

func (s *Scrollable) effectiveOffset(w, h int) scrollableViewport {
	w = max(0, w)
	h = max(0, h)
	offset := max(0, s.Offset)
	if isNilPanel(s.Content) {
		return scrollableViewport{}
	}

	if contentHeight, known := panelContentHeight(s.Content, w); known {
		return scrollableViewport{offset: clampScrollableOffset(offset, contentHeight, h), known: true}
	}
	if contentHeight, known := s.cachedContentHeight(w); known {
		return scrollableViewport{offset: clampScrollableOffset(offset, contentHeight, h), known: true}
	}
	if h == 0 {
		return scrollableViewport{offset: offset}
	}

	probeHeight := saturatingAddNonNegative(scrollableRequestHeight(offset, h), 1)
	lines := strings.Split(s.Content.View(w, probeHeight), "\n")
	if len(lines) < probeHeight {
		contentHeight := len(lines)
		s.rememberContentHeight(w, contentHeight)
		return scrollableViewport{
			offset: clampScrollableOffset(offset, contentHeight, h),
			lines:  lines,
			probed: true,
			known:  true,
		}
	}
	return scrollableViewport{offset: offset, lines: lines, probed: true}
}

func (s *Scrollable) cachedContentHeight(width int) (int, bool) {
	if !s.extentCache.known || s.extentCache.width != width || !samePanel(s.extentCache.content, s.Content) {
		return 0, false
	}
	return s.extentCache.height, true
}

func (s *Scrollable) rememberContentHeight(width, height int) {
	s.extentCache = scrollableExtentCache{
		content: s.Content,
		width:   width,
		height:  max(0, height),
		known:   true,
	}
}

func (s *Scrollable) invalidateContentHeight() {
	s.extentCache = scrollableExtentCache{}
}

func clampScrollableOffset(offset, contentHeight, viewportHeight int) int {
	maxOffset := max(0, max(0, contentHeight)-max(0, viewportHeight))
	return min(max(0, offset), maxOffset)
}

func scrollableRequestHeight(offset, viewportHeight int) int {
	return saturatingAddNonNegative(max(0, offset), max(0, viewportHeight))
}

func saturatingAddNonNegative(a, b int) int {
	a = max(0, a)
	b = max(0, b)
	maxInt := int(^uint(0) >> 1)
	if a > maxInt-b {
		return maxInt
	}
	return a + b
}

func (s *Scrollable) rememberViewport(w, h int) {
	s.lastWidth = max(0, w)
	s.lastHeight = max(0, h)
	s.hasViewport = true
}

// Update handles scroll messages (mouse wheel, keys) and forwards all messages.
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd {
	if s.Offset < 0 {
		s.Offset = 0
	}

	viewportChanged := false
	scrollChanged := false
	switch msg := msg.(type) {
	case ResizeMsg:
		s.rememberViewport(msg.Width, msg.Height)
		s.hasLocalResize = true
		viewportChanged = true
	case tea.WindowSizeMsg:
		if !s.hasLocalResize {
			s.rememberViewport(msg.Width, msg.Height)
			viewportChanged = true
		}
	case tea.MouseMsg:
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			s.scrollBy(-3)
			scrollChanged = true
		case tea.MouseButtonWheelDown:
			s.scrollBy(3)
			scrollChanged = true
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "up":
			s.scrollBy(-1)
			scrollChanged = true
		case "down":
			s.scrollBy(1)
			scrollChanged = true
		case "pgup":
			s.scrollBy(-10)
			scrollChanged = true
		case "pgdown":
			s.scrollBy(10)
			scrollChanged = true
		}
	}

	var cmd tea.Cmd
	if !isNilPanel(s.Content) {
		s.invalidateContentHeight()
		cmd = s.Content.Update(msg)
	}
	if s.hasViewport && (viewportChanged || scrollChanged) {
		s.Offset = s.effectiveOffset(s.lastWidth, s.lastHeight).offset
	}
	return cmd
}

func (s *Scrollable) scrollBy(delta int) {
	if delta < 0 {
		amount := -delta
		if s.Offset < amount {
			s.Offset = 0
		} else {
			s.Offset -= amount
		}
		return
	}
	s.Offset = saturatingAddNonNegative(s.Offset, delta)
}

func padLine(line string, w int) string {
	return padVisualLine(line, max(0, w))
}
