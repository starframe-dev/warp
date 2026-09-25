package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type collapsibleTestPanel struct {
	name       string
	viewWidth  int
	viewHeight int
}

func (p *collapsibleTestPanel) View(w, h int) string {
	p.viewWidth, p.viewHeight = w, h
	return p.name
}

func (p *collapsibleTestPanel) Update(msg tea.Msg) tea.Cmd {
	return nil
}

// capturingPanel records the messages forwarded to Update.
type capturingPanel struct {
	collapsibleTestPanel
	msgs []tea.Msg
}

func (p *capturingPanel) Update(msg tea.Msg) tea.Cmd {
	p.msgs = append(p.msgs, msg)
	return nil
}

func TestNewCollapsible(t *testing.T) {
	inner := &collapsibleTestPanel{name: "inner"}
	c := NewCollapsible("My Panel", inner)

	if c == nil {
		t.Fatal("NewCollapsible returned nil")
	}
	if c.Title != "My Panel" {
		t.Errorf("Title = %q, want %q", c.Title, "My Panel")
	}
	if c.Collapsed {
		t.Error("NewCollapsible Collapsed should be false")
	}
	if c.Content == nil {
		t.Error("NewCollapsible Content is nil")
	}
}

func TestCollapsibleViewExpandedWithContent(t *testing.T) {
	inner := &collapsibleTestPanel{name: "visible content"}
	c := NewCollapsible("T", inner)

	lines := strings.Split(ansi.Strip(c.View(20, 5)), "\n")
	if len(lines) != 5 {
		t.Fatalf("expanded view has %d rows, want 5", len(lines))
	}
	if !strings.Contains(lines[0], "▼") || !strings.Contains(lines[0], "T") {
		t.Fatalf("expanded title row=%q, want title and ▼", lines[0])
	}
	if strings.TrimSpace(lines[1]) != "visible content" {
		t.Fatalf("first content row=%q, want visible content", lines[1])
	}
	if inner.viewWidth != 20 || inner.viewHeight != 4 {
		t.Fatalf("content View size=%dx%d, want 20x4", inner.viewWidth, inner.viewHeight)
	}
}

func TestCollapsibleViewNilContent(t *testing.T) {
	c := &Collapsible{Title: "empty", Content: nil}
	lines := strings.Split(ansi.Strip(c.View(12, 4)), "\n")
	if len(lines) != 4 {
		t.Fatalf("nil-content view has %d rows, want 4", len(lines))
	}
	if !strings.Contains(lines[0], "empty") || !strings.Contains(lines[0], "▼") {
		t.Fatalf("nil-content title row=%q", lines[0])
	}
}

func TestCollapsibleViewCollapsed(t *testing.T) {
	inner := &collapsibleTestPanel{name: "hidden"}
	c := NewCollapsible("Collapsed Title", inner)
	c.Collapsed = true

	got := c.View(40, 1)
	if strings.Contains(got, "hidden") {
		t.Errorf("collapsed view leaked inner content: %q", got)
	}
	if !strings.Contains(got, "Collapsed Title") {
		t.Errorf("collapsed view missing title: %q", got)
	}
}

func TestCollapsibleViewZeroAndNegativeWidth(t *testing.T) {
	c := NewCollapsible("T", &collapsibleTestPanel{name: "x"})
	c.Collapsed = true

	if got := c.View(0, 1); got != "" {
		t.Errorf("zero-width collapsed view = %q, want empty", got)
	}
	if got := c.View(-1, 1); got != "" {
		t.Errorf("negative-width collapsed view = %q, want empty", got)
	}
}

func TestCollapsibleRenderCollapsedTinyWidthEmptyTitle(t *testing.T) {
	// A tiny width forces maxLen = 0 -> empty title, but bar is still rendered.
	c := NewCollapsible("A Very Long Title", &collapsibleTestPanel{name: "x"})
	c.Collapsed = true

	if got := c.View(1, 1); got == "" {
		t.Errorf("tiny-width collapsed view = empty, want non-empty bar")
	}
}

func TestCollapsibleRenderCollapsedTruncation(t *testing.T) {
	longTitle := strings.Repeat("A", 100)
	c := NewCollapsible(longTitle, &collapsibleTestPanel{name: "x"})
	c.Collapsed = true

	got := c.View(20, 1)
	if strings.Contains(got, longTitle) {
		t.Errorf("long title should be truncated, got %q", got)
	}
	if !strings.Contains(got, "...") {
		t.Errorf("truncated title missing ellipsis: %q", got)
	}
}

func TestCollapsibleRenderCollapsedIndicatorCollapsed(t *testing.T) {
	c := NewCollapsible("T", &collapsibleTestPanel{name: "x"})
	c.Collapsed = true

	got := c.View(10, 1)
	if !strings.Contains(got, "▶") {
		t.Errorf("collapsed view missing ▶ indicator: %q", got)
	}
	if strings.Contains(got, "▼") {
		t.Errorf("collapsed view should not render ▼: %q", got)
	}
}

func TestCollapsibleUpdateForwardsMessage(t *testing.T) {
	inner := &capturingPanel{collapsibleTestPanel: collapsibleTestPanel{name: "x"}}
	c := NewCollapsible("T", inner)

	msg := tea.WindowSizeMsg{Width: 42, Height: 24}
	cmd := c.Update(msg)
	if cmd != nil {
		t.Errorf("Update cmd = %v, want nil", cmd)
	}
	if len(inner.msgs) != 1 {
		t.Fatalf("expected 1 forwarded message, got %d", len(inner.msgs))
	}
	got, ok := inner.msgs[0].(tea.WindowSizeMsg)
	if !ok || got.Width != 42 || got.Height != 24 {
		t.Errorf("forwarded message mismatch: got %v, want %v", inner.msgs[0], msg)
	}
}

func TestCollapsibleUpdateAdjustsContentGeometry(t *testing.T) {
	inner := &capturingPanel{collapsibleTestPanel: collapsibleTestPanel{name: "x"}}
	collapsible := NewCollapsible("title", inner)

	collapsible.Update(ResizeMsg{Width: 12, Height: 5})
	resize, ok := inner.msgs[len(inner.msgs)-1].(ResizeMsg)
	if !ok || resize.Width != 12 || resize.Height != 4 {
		t.Fatalf("expanded content resize=%v, want 12x4", inner.msgs[len(inner.msgs)-1])
	}

	collapsible.Update(tea.MouseMsg{X: 3, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	mouse, ok := inner.msgs[len(inner.msgs)-1].(tea.MouseMsg)
	if !ok || mouse.X != 3 || mouse.Y != 0 {
		t.Fatalf("first content row mouse=%v, want Y=0", inner.msgs[len(inner.msgs)-1])
	}
	messageCount := len(inner.msgs)
	collapsible.Update(tea.MouseMsg{X: 3, Y: 0, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if len(inner.msgs) != messageCount {
		t.Fatal("title-row mouse event should not reach inner content")
	}

	collapsible.Collapsed = true
	collapsible.Update(ResizeMsg{Width: 12, Height: 5})
	resize, ok = inner.msgs[len(inner.msgs)-1].(ResizeMsg)
	if !ok || resize.Height != 0 {
		t.Fatalf("collapsed content resize=%v, want height 0", inner.msgs[len(inner.msgs)-1])
	}
	messageCount = len(inner.msgs)
	collapsible.Update(tea.MouseMsg{X: 3, Y: 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if len(inner.msgs) != messageCount {
		t.Fatal("collapsed content should not receive mouse events")
	}
}

func TestCollapsibleUpdateNilContent(t *testing.T) {
	c := &Collapsible{Title: "x", Content: nil}
	if cmd := c.Update(nil); cmd != nil {
		t.Errorf("Update with nil content cmd = %v, want nil", cmd)
	}
}

func TestCollapsibleToggle(t *testing.T) {
	c := NewCollapsible("T", &collapsibleTestPanel{name: "x"})
	if c.Collapsed {
		t.Fatal("initial Collapsed should be false")
	}

	c.Toggle()
	if !c.Collapsed {
		t.Errorf("after Toggle Collapsed = false, want true")
	}

	c.Toggle()
	if c.Collapsed {
		t.Errorf("after second Toggle Collapsed = true, want false")
	}
}

func TestCollapsibleToggleNilContent(t *testing.T) {
	c := &Collapsible{Title: "x", Content: nil}
	c.Toggle()
	if !c.Collapsed {
		t.Error("Toggle should set Collapsed")
	}
	c.Toggle()
	if c.Collapsed {
		t.Error("Toggle should restore state")
	}
}
