package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type boundedScrollablePanel struct {
	rows          []string
	requestedSize int
}

func (p *boundedScrollablePanel) View(_, height int) string {
	p.requestedSize = height
	return strings.Join(p.rows[:min(len(p.rows), max(0, height))], "\n")
}

func (*boundedScrollablePanel) Update(tea.Msg) tea.Cmd { return nil }

func TestScrollableRequestsOnlyThroughVisibleEnd(t *testing.T) {
	rows := []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	content := &boundedScrollablePanel{rows: rows}
	scrollable := NewScrollable(content)
	scrollable.Offset = 3

	got := scrollable.View(1, 4)
	if content.requestedSize != 7 {
		t.Fatalf("Content.View requested height %d, want offset+viewport=7", content.requestedSize)
	}
	if got != "3\n4\n5\n6" {
		t.Fatalf("view=%q, want rows 3-6", got)
	}

	scrollable.Offset = 100
	got = scrollable.View(1, 4)
	if scrollable.Offset != 6 || got != "6\n7\n8\n9" {
		t.Fatalf("clamped offset=%d view=%q, want 6 and rows 6-9", scrollable.Offset, got)
	}
}

func TestScrollablePadsAndTruncatesUnicodeByCells(t *testing.T) {
	line := "\x1b[31m界🙂x\x1b[0m"
	for _, width := range []int{0, 1, 2, 3, 4, 5} {
		got := padLine(line, width)
		if actual := ansi.StringWidth(got); actual != width {
			t.Errorf("width=%d produced %d cells: %q", width, actual, got)
		}
	}
	if got := ansi.StringWidth(padLine("界", 3)); got != 3 {
		t.Fatalf("wide short line padded to %d cells, want 3", got)
	}
}

func TestScrollableClampsNegativeDimensionsAndOffset(t *testing.T) {
	content := &scrollableMockPanel{view: "0\n1\n2"}
	scrollable := NewScrollable(content)
	scrollable.Offset = -10
	if got := scrollable.View(-1, -2); got != "" {
		t.Fatalf("negative viewport returned %q", got)
	}
	if scrollable.Offset != 0 {
		t.Fatalf("negative offset clamped to %d, want 0", scrollable.Offset)
	}
	scrollable.Update(tea.KeyMsg{Type: tea.KeyDown})
	if scrollable.Offset != 1 {
		t.Fatalf("down from negative offset produced %d, want 1", scrollable.Offset)
	}
}
