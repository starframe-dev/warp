package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// updateTestPanel is like testPanel but records messages passed to Update.
type updateTestPanel struct {
	testPanel
	msgs []tea.Msg
}

func (p *updateTestPanel) Update(msg tea.Msg) tea.Cmd {
	p.msgs = append(p.msgs, msg)
	return nil
}

func TestNewCollapsible(t *testing.T) {
	c := NewCollapsible("Panel Title", &testPanel{name: "inner"})
	if c == nil {
		t.Fatal("NewCollapsible returned nil")
	}
	if c.Title != "Panel Title" {
		t.Errorf("Title = %q, want %q", c.Title, "Panel Title")
	}
	if c.Collapsed != false {
		t.Errorf("Collapsed = %v, want false", c.Collapsed)
	}
	if c.Content == nil {
		t.Error("Content is nil")
	}
}

func TestCollapsibleViewExpanded(t *testing.T) {
	inner := &testPanel{name: "expanded content"}
	c := NewCollapsible("Title", inner)

	got := c.View(20, 1)
	if !strings.Contains(got, "expanded content") {
		t.Errorf("expanded view missing content: %q", got)
	}
	if ansi.StringWidth(got) != 20 {
		t.Errorf("expanded view width = %d, want 20", ansi.StringWidth(got))
	}
}

func TestCollapsibleViewCollapsed(t *testing.T) {
	inner := &testPanel{name: "hidden"}
	c := NewCollapsible("Collapsed Title", inner)
	c.Collapsed = true

	got := c.View(30, 1)
	if strings.Contains(got, "hidden") {
		t.Errorf("collapsed view should not render inner content, got %q", got)
	}
	if !strings.Contains(got, "Collapsed Title") {
		t.Errorf("collapsed view missing title: %q", got)
	}
	if ansi.StringWidth(got) == 0 {
		t.Errorf("collapsed view returned empty width: %q", got)
	}
}

func TestCollapsibleViewNilContent(t *testing.T) {
	c := &Collapsible{Title: "Empty", Content: nil}

	if got := c.View(10, 3); got != "" {
		t.Errorf("nil content view = %q, want empty", got)
	}
	if cmd := c.Update(nil); cmd != nil {
		t.Errorf("nil content update cmd = %v, want nil", cmd)
	}
}

func TestCollapsibleUpdate(t *testing.T) {
	inner := &updateTestPanel{testPanel: testPanel{name: "x"}}
	c := NewCollapsible("Title", inner)

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

func TestCollapsibleToggle(t *testing.T) {
	c := NewCollapsible("Title", &testPanel{name: "inner"})
	if c.Collapsed != false {
		t.Fatalf("initial Collapsed = %v, want false", c.Collapsed)
	}

	c.Toggle()
	if c.Collapsed != true {
		t.Errorf("after Toggle Collapsed = %v, want true", c.Collapsed)
	}

	c.Toggle()
	if c.Collapsed != false {
		t.Errorf("after second Toggle Collapsed = %v, want false", c.Collapsed)
	}
}

func TestCollapsibleRenderCollapsed(t *testing.T) {
	c := NewCollapsible("Title", &testPanel{name: "inner"})
	c.Collapsed = true

	got := c.View(20, 1)
	if got == "" {
		t.Fatal("collapsed view returned empty string")
	}
	if ansi.StringWidth(got) == 0 {
		t.Errorf("collapsed view returned empty width: %q", got)
	}
	if !strings.Contains(got, "▶") {
		t.Errorf("collapsed view missing collapsed indicator: %q", got)
	}
	if !strings.Contains(got, "Title") {
		t.Errorf("collapsed view missing title: %q", got)
	}

	// Zero width returns empty string.
	if got := c.View(0, 1); got != "" {
		t.Errorf("zero-width collapsed view = %q, want empty", got)
	}
}

func TestCollapsibleRenderCollapsedTitleTruncation(t *testing.T) {
	longTitle := strings.Repeat("A", 100)
	c := NewCollapsible(longTitle, &testPanel{name: "inner"})
	c.Collapsed = true

	got := c.View(20, 1)
	if strings.Contains(got, longTitle) {
		t.Errorf("long title should be truncated, got %q", got)
	}
	if ansi.StringWidth(got) == 0 {
		t.Errorf("truncated view returned empty width: %q", got)
	}
	if !strings.Contains(got, "...") {
		t.Errorf("truncated title missing ellipsis: %q", got)
	}
}

func TestCollapsibleRenderCollapsedIndicator(t *testing.T) {
	c := NewCollapsible("T", &testPanel{name: "inner"})
	c.Collapsed = true

	got := c.View(10, 1)
	if !strings.Contains(got, "▶") {
		t.Errorf("collapsed indicator missing: %q", got)
	}
	if strings.Contains(got, "▼") {
		t.Errorf("collapsed indicator should not be down: %q", got)
	}
}
