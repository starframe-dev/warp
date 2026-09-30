package integration_test

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	warp "github.com/starframe-dev/warp"
)

type scrollTestPanel struct {
	lines       []string
	known       bool
	pad         bool
	viewCalls   int
	viewWidths  []int
	viewHeights []int
	updates     []tea.Msg
	elements    []warp.Element
}

func (p *scrollTestPanel) View(width, height int) string {
	p.viewCalls++
	p.viewWidths = append(p.viewWidths, width)
	p.viewHeights = append(p.viewHeights, height)
	n := len(p.lines)
	if p.pad && height > n {
		n = height
	}
	out := make([]string, n)
	for i := range out {
		if i < len(p.lines) {
			out[i] = p.lines[i]
		}
	}
	return strings.Join(out, "\n")
}
func (p *scrollTestPanel) Update(msg tea.Msg) tea.Cmd       { p.updates = append(p.updates, msg); return nil }
func (p *scrollTestPanel) ContentHeight(int) (int, bool)    { return len(p.lines), p.known }
func (p *scrollTestPanel) Elements(_, _ int) []warp.Element { return p.elements }

type scrollViewportPanel struct {
	*scrollTestPanel
	atCalls int
}

func (p *scrollViewportPanel) ViewAt(_, height, offset int) string {
	p.atCalls++
	out := make([]string, height)
	for i := range out {
		if offset+i < len(p.lines) {
			out[i] = p.lines[offset+i]
		}
	}
	return strings.Join(out, "\n")
}
func (p *scrollViewportPanel) ElementsAt(_, _, offset int) []warp.Element {
	return []warp.Element{{Role: "item", Name: fmt.Sprint(offset), Bounds: warp.Bounds{Y: offset, W: 2, H: 1}}}
}

func TestScrollableKnownHeightViewportRenderingAndElements(t *testing.T) {
	content := &scrollViewportPanel{scrollTestPanel: &scrollTestPanel{lines: []string{"zero", "one", "two", "three", "four"}, known: true,
		elements: []warp.Element{{Role: "item", Name: "zero", Bounds: warp.Bounds{W: 2, H: 1}}}}}
	s := warp.NewScrollable(content)
	s.Offset = 2
	view := s.View(5, 2)
	if view != "two  \nthree" {
		t.Fatalf("View()=%q", view)
	}
	if content.atCalls != 1 || content.viewCalls != 0 {
		t.Fatalf("viewport fast path: ViewAt=%d View=%d", content.atCalls, content.viewCalls)
	}
	if s.Offset != 2 {
		t.Fatalf("offset=%d, want 2", s.Offset)
	}
	elements := s.Elements(5, 2)
	if len(elements) != 1 || elements[0].Name != "2" || elements[0].Bounds != (warp.Bounds{Y: 0, W: 2, H: 1}) {
		t.Fatalf("viewport elements=%+v", elements)
	}
	if _, known := s.ContentHeight(5); known {
		t.Fatal("Scrollable must not expose intrinsic height")
	}
}

func TestScrollableNormalizesOverscrollAndProbesNaturalContent(t *testing.T) {
	known := &scrollTestPanel{lines: []string{"a", "b", "c", "d"}, known: true}
	s := warp.NewScrollable(known)
	s.Offset = 100
	if got := s.View(3, 2); got != "c  \nd  " {
		t.Fatalf("overscroll View()=%q", got)
	}
	if s.Offset != 2 {
		t.Fatalf("normalized offset=%d, want 2", s.Offset)
	}

	fallback := &scrollTestPanel{lines: []string{"a", "b", "c", "d", "e"}}
	probed := warp.NewScrollable(fallback)
	probed.Offset = 3
	got := probed.View(4, 2)
	if got != "d   \ne   " {
		t.Fatalf("probed View()=%q", got)
	}
	if fallback.viewCalls != 1 || fallback.viewHeights[0] != 6 {
		t.Fatalf("probe calls/heights=%d/%v", fallback.viewCalls, fallback.viewHeights)
	}

	padded := &scrollTestPanel{lines: []string{"x", "y"}, pad: true}
	unknown := warp.NewScrollable(padded)
	unknown.Offset = 5
	_ = unknown.Elements(8, 2) // A padded fallback deliberately cannot infer its true extent.
	if unknown.Offset != 5 {
		t.Fatalf("unknown extent clamped offset to %d", unknown.Offset)
	}
}

func TestScrollableUpdateResizeWheelKeysAndForwarding(t *testing.T) {
	content := &scrollTestPanel{lines: []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15", "16", "17", "18", "19"}, known: true}
	s := warp.NewScrollable(content)
	s.Update(warp.ResizeMsg{Width: 4, Height: 4})
	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown, Action: tea.MouseActionPress})
	if s.Offset != 3 {
		t.Fatalf("wheel offset=%d, want 3", s.Offset)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyDown})
	if s.Offset != 4 {
		t.Fatalf("down offset=%d, want 4", s.Offset)
	}
	s.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if s.Offset != 14 {
		t.Fatalf("pgdown offset=%d, want 14", s.Offset)
	}
	s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelUp, Action: tea.MouseActionPress})
	if s.Offset != 11 {
		t.Fatalf("wheel-up offset=%d, want 11", s.Offset)
	}
	s.Update(tea.WindowSizeMsg{Width: 4, Height: 2}) // local ResizeMsg takes precedence.
	if s.Offset != 11 {
		t.Fatalf("window resize unexpectedly replaced local viewport: offset=%d", s.Offset)
	}
	if len(content.updates) != 6 {
		t.Fatalf("forwarded updates=%d, want 6", len(content.updates))
	}
	s.Update(warp.ResizeMsg{Width: 4, Height: 2})
	if s.Offset != 11 {
		t.Fatalf("resize did not retain in-range offset: %d", s.Offset)
	}

	short := warp.NewScrollable(&scrollTestPanel{lines: []string{"a", "b", "c"}, known: true})
	short.Update(warp.ResizeMsg{Width: 4, Height: 3})
	short.Offset = 2
	short.Update(warp.ResizeMsg{Width: 4, Height: 2})
	if short.Offset != 1 {
		t.Fatalf("resize did not clamp offset: %d", short.Offset)
	}
}

func TestScrollableSelectableAndCollapsibleHeightPropagation(t *testing.T) {
	content := &scrollTestPanel{lines: []string{"a", "b", "c", "d"}, known: true}
	selectable := warp.NewSelectable(content)
	if h, ok := selectable.ContentHeight(8); !ok || h != 4 {
		t.Fatalf("selectable height=(%d,%v), want (4,true)", h, ok)
	}
	unknownSelectable := warp.NewSelectable(&scrollTestPanel{lines: []string{"a"}})
	if _, ok := unknownSelectable.ContentHeight(8); ok {
		t.Fatal("selectable reported unknown height as known")
	}

	collapsible := warp.NewCollapsible("Details", selectable)
	if h, ok := collapsible.ContentHeight(8); !ok || h != 5 {
		t.Fatalf("expanded collapsible height=(%d,%v), want (5,true)", h, ok)
	}
	wrapped := warp.NewScrollable(collapsible)
	wrapped.Offset = 99
	if got := wrapped.View(12, 3); got == "" {
		t.Fatal("scrollable around wrappers rendered empty output")
	}
	if wrapped.Offset != 2 {
		t.Fatalf("wrapper content offset=%d, want 2", wrapped.Offset)
	}
	collapsible.Toggle()
	if h, ok := collapsible.ContentHeight(8); !ok || h != 1 {
		t.Fatalf("collapsed height=(%d,%v), want (1,true)", h, ok)
	}
	wrapped.Offset = 9
	_ = wrapped.View(12, 1)
	if wrapped.Offset != 0 {
		t.Fatalf("collapsed scroll offset=%d, want 0", wrapped.Offset)
	}
	if _, ok := (warp.NewCollapsible("unknown", &scrollTestPanel{lines: []string{"x"}})).ContentHeight(8); ok {
		t.Fatal("expanded collapsible should preserve unknown content height")
	}
}
