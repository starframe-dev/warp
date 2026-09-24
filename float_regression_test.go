package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

func TestFloatPaneRenderValidatesDimensions(t *testing.T) {
	for _, size := range [][2]int{{-1, 5}, {5, -1}, {0, 5}, {5, 0}, {1, 1}, {2, 2}, {3, 3}} {
		fp := &FloatPane{Panel: &geometryTestPanel{name: "中"}, Width: size[0], Height: size[1]}
		lines := fp.render(20, 10)
		if size[0] <= 0 || size[1] <= 0 {
			if len(lines) != 0 {
				t.Fatalf("size %v returned %d lines, want none", size, len(lines))
			}
			continue
		}
		if len(lines) != size[1] {
			t.Fatalf("size %v returned %d lines, want %d", size, len(lines), size[1])
		}
		for i, line := range lines {
			if got := ansi.StringWidth(line); got != size[0] {
				t.Errorf("size %v line %d width=%d, want %d", size, i, got, size[0])
			}
		}
	}

	var nilFloat *FloatPane
	if got := nilFloat.render(20, 10); got != nil {
		t.Fatalf("nil FloatPane render = %v, want nil", got)
	}
}

func TestTabFloatRejectsNilAndClampsPositiveDimensions(t *testing.T) {
	tab := NewTab("float-validation")
	tab.Float(nil, 0, 0, 20, 5)
	if len(tab.floats) != 0 {
		t.Fatalf("nil panel created %d floats", len(tab.floats))
	}

	panel := &geometryTestPanel{name: "float"}
	tab.Float(panel, -3, -4, 1, 1)
	if len(tab.floats) != 1 {
		t.Fatalf("valid small float count=%d, want 1", len(tab.floats))
	}
	fp := tab.floats[0]
	if fp.X != 0 || fp.Y != 0 || fp.Width != floatMinWidth || fp.Height != floatMinHeight {
		t.Fatalf("small float = %+v, want nonnegative position and minimum size", fp)
	}
}

func TestFloatDragAndResizeStayWithinViewport(t *testing.T) {
	fp := &FloatPane{Panel: &geometryTestPanel{name: "float"}, X: 5, Y: 5, Width: 20, Height: 10}
	fp.handleMouseWithin(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, 10, 5, 40, 15)
	fp.handleMouseWithin(tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft}, 100, 100, 40, 15)
	if fp.X != 20 || fp.Y != 5 {
		t.Fatalf("drag position=(%d,%d), want (20,5)", fp.X, fp.Y)
	}

	fp = &FloatPane{Panel: &geometryTestPanel{name: "float"}, X: 5, Y: 3, Width: 20, Height: 10}
	fp.handleMouseWithin(tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft}, 24, 8, 30, 15)
	if !fp.resizing || fp.resizeEdge != "e" {
		t.Fatalf("resize did not start on east edge: resizing=%v edge=%q", fp.resizing, fp.resizeEdge)
	}
	fp.handleMouseWithin(tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft}, 100, 8, 30, 15)
	if fp.X != 5 || fp.Width != 25 || fp.X+fp.Width != 30 {
		t.Fatalf("resized float = %+v, want right edge clamped at 30", fp)
	}
}

func TestOverlayFloatHandlesNilAndUnicodeCellClipping(t *testing.T) {
	base := padVisualLine("a界🙂bc", 12)
	lines := []string{base}
	overlayFloat(lines, nil, 12, 1)
	overlayFloat(lines, &FloatPane{Width: 4, Height: 1}, 12, 1)
	if lines[0] != base {
		t.Fatalf("invalid float modified background: %q", lines[0])
	}

	fp := &FloatPane{
		Panel:  &geometryTestPanel{name: "界"},
		X:      8,
		Y:      0,
		Width:  10,
		Height: 1,
		Title:  "界🙂",
	}
	overlayFloat(lines, fp, 12, 1)
	if got := ansi.StringWidth(lines[0]); got != 12 {
		t.Fatalf("overlaid line width=%d, want 12 (%q)", got, lines[0])
	}
	if !strings.Contains(StripANSI(lines[0]), "╭") {
		t.Fatalf("clipped float border missing: %q", StripANSI(lines[0]))
	}
}
