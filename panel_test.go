package warp

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type testContentPanel struct {
	BasePanel
	height int
	known  bool
	width  int
}

func (p *testContentPanel) ContentHeight(width int) (int, bool) {
	p.width = width
	return p.height, p.known
}

type testViewportRenderer struct {
	BasePanel
	result string
	width  int
	height int
	offset int
}

func (p *testViewportRenderer) ViewAt(width, height, offset int) string {
	p.width, p.height, p.offset = width, height, offset
	return p.result
}

type testUnmounter struct {
	BasePanel
	unmounted bool
}

func (p *testUnmounter) Unmount() { p.unmounted = true }

func TestBasePanelView(t *testing.T) {
	bp := BasePanel{}
	if got := bp.View(80, 24); got != "" {
		t.Errorf("BasePanel.View(80, 24) = %q; want empty string", got)
	}
	if got := bp.View(0, 0); got != "" {
		t.Errorf("BasePanel.View(0, 0) = %q; want empty string", got)
	}
}

func TestBasePanelUpdate(t *testing.T) {
	bp := BasePanel{}
	msgs := []tea.Msg{
		tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("a")},
		tea.KeyMsg{Type: tea.KeyCtrlC},
		tea.WindowSizeMsg{Width: 100, Height: 40},
		tea.MouseMsg{X: 5, Y: 7},
	}
	for i, msg := range msgs {
		if got := bp.Update(msg); got != nil {
			t.Errorf("BasePanel.Update(msgs[%d]) = %v; want nil", i, got)
		}
	}
}

func TestPanelInterface(t *testing.T) {
	var p Panel = BasePanel{}
	if got := p.View(60, 30); got != "" {
		t.Errorf("Panel.View(60, 30) = %q; want empty string", got)
	}
	if cmd := p.Update(tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
		t.Errorf("Panel.Update(tea.KeyEnter) = %v; want nil", cmd)
	}
}

func TestContentHeightProviderAndPanelContentHeight(t *testing.T) {
	provider := &testContentPanel{height: 12, known: true}
	var p Panel = provider
	got, known := panelContentHeight(p, -4)
	if !known || got != 12 {
		t.Fatalf("panelContentHeight() = (%d, %t); want (12, true)", got, known)
	}
	if provider.width != 0 {
		t.Errorf("ContentHeight received width %d; want normalized width 0", provider.width)
	}

	provider.height = -3
	got, known = panelContentHeight(p, 20)
	if !known || got != 0 {
		t.Errorf("negative content height = (%d, %t); want (0, true)", got, known)
	}
	if provider.width != 20 {
		t.Errorf("ContentHeight received width %d; want 20", provider.width)
	}

	provider.known = false
	if got, known = panelContentHeight(p, 20); got != 0 || known {
		t.Errorf("unknown content height = (%d, %t); want (0, false)", got, known)
	}
	if got, known = panelContentHeight(BasePanel{}, 20); got != 0 || known {
		t.Errorf("panel without provider = (%d, %t); want (0, false)", got, known)
	}
	var nilPanel *testContentPanel
	if got, known = panelContentHeight(nilPanel, 20); got != 0 || known {
		t.Errorf("nil panel = (%d, %t); want (0, false)", got, known)
	}
}

func TestViewportRenderer(t *testing.T) {
	renderer := &testViewportRenderer{result: "rows"}
	var p ViewportRenderer = renderer
	if got := p.ViewAt(80, 5, 7); got != "rows" {
		t.Fatalf("ViewAt() = %q; want %q", got, "rows")
	}
	if renderer.width != 80 || renderer.height != 5 || renderer.offset != 7 {
		t.Errorf("ViewAt arguments = (%d, %d, %d); want (80, 5, 7)", renderer.width, renderer.height, renderer.offset)
	}
}

func TestUnmounter(t *testing.T) {
	unmounter := &testUnmounter{}
	var p Unmounter = unmounter
	p.Unmount()
	if !unmounter.unmounted {
		t.Error("Unmount() did not mark the panel as unmounted")
	}
}
