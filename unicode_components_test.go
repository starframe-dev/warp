package warp

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestDropdownAndCollapsibleUseCellWidths(t *testing.T) {
	for width := 0; width <= 12; width++ {
		dropdown := NewDropdownMenu("界🙂e\u0301", []DropdownItem{{Label: "続行🙂"}})
		if got := ansi.StringWidth(dropdown.View(width, 1)); got != width {
			t.Errorf("dropdown button width=%d produced %d cells", width, got)
		}
		dropdown.Open = true
		for y, line := range strings.Split(dropdown.View(width, 2), "\n") {
			if got := ansi.StringWidth(line); got != width {
				t.Errorf("dropdown menu width=%d line %d produced %d cells", width, y, got)
			}
		}

		collapsible := NewCollapsible("界🙂e\u0301", emptyPanel{})
		collapsible.Collapsed = true
		if got := ansi.StringWidth(collapsible.View(width, 1)); got != width {
			t.Errorf("collapsible width=%d produced %d cells", width, got)
		}
	}
}

func TestWrapUsesGraphemeCellWidths(t *testing.T) {
	for _, width := range []int{1, 2, 3, 4, 5} {
		for _, line := range wrapLine("界🙂e\u0301 abc", width) {
			if got := ansi.StringWidth(line); got > width {
				t.Errorf("wrapLine width=%d produced %d cells: %q", width, got, line)
			}
		}
	}

	lines := wrapAtSpaces("界界 abcdef", 3)
	if len(lines) < 2 {
		t.Fatalf("expected SpaceWrap to split at spaces, got %q", lines)
	}
	if got := ansi.StringWidth(lines[len(lines)-1]); got <= 3 {
		t.Fatalf("SpaceWrap should preserve overlong words, got %d cells", got)
	}
}
