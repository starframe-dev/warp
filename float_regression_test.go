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

func TestFloatReclampsAutomaticallyOnViewportResize(t *testing.T) {
	tab := NewTab("resize")
	tab.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	tab.Float(&geometryTestPanel{name: "float"}, 90, 30, 30, 10)
	fp := tab.floats[0]
	if fp.X != 90 || fp.Y != 30 {
		t.Fatalf("large viewport position=(%d,%d), want (90,30)", fp.X, fp.Y)
	}

	tab.Update(ResizeMsg{Width: 60, Height: 20})
	if fp.X != 30 || fp.Y != 10 || fp.Width != 30 || fp.Height != 10 {
		t.Fatalf("float after resize = %+v, want (30,10) size 30x10", fp)
	}
	view := tab.View(60, 20)
	assertFloatWithinViewport(t, fp, 60, 20)
	if !strings.Contains(StripANSI(view), "╭") {
		t.Fatal("resized float is not visible in the rendered viewport")
	}
}

func TestFloatRestoresPreferredSizeAfterViewportShrinkAndExpand(t *testing.T) {
	tab := NewTab("preferred-size")
	tab.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	tab.Float(&geometryTestPanel{name: "float"}, 90, 30, 30, 10)
	fp := tab.floats[0]

	tab.Update(tea.WindowSizeMsg{Width: 5, Height: 2})
	if fp.Width != 5 || fp.Height != 2 || fp.preferredWidth != 30 || fp.preferredHeight != 10 {
		t.Fatalf("float at 5x2 = %+v, preferred=%dx%d", fp, fp.preferredWidth, fp.preferredHeight)
	}
	assertFloatWithinViewport(t, fp, 5, 2)
	tab.View(5, 2)

	tab.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	if fp.Width != 30 || fp.Height != 10 {
		t.Fatalf("float after expand = %dx%d, want 30x10", fp.Width, fp.Height)
	}
	assertFloatWithinViewport(t, fp, 80, 24)
	if !strings.Contains(StripANSI(tab.View(80, 24)), "╭") {
		t.Fatal("restored float is not visible")
	}
}

func TestManualFloatResizeUpdatesPreferredSize(t *testing.T) {
	tab := NewTab("manual-preferred-size")
	tab.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	tab.Float(&geometryTestPanel{name: "float"}, 10, 5, 30, 10)
	fp := tab.floats[0]

	fp.handleMouseWithin(
		tea.MouseMsg{Action: tea.MouseActionPress, Button: tea.MouseButtonLeft},
		fp.X+fp.Width-1, fp.Y+fp.Height-1, 80, 24,
	)
	if !fp.resizing || fp.resizeEdge != "se" {
		t.Fatalf("resize did not start at southeast corner: edge=%q", fp.resizeEdge)
	}
	fp.handleMouseWithin(
		tea.MouseMsg{Action: tea.MouseActionMotion, Button: tea.MouseButtonLeft},
		fp.dragStartX+12, fp.dragStartY+2, 80, 24,
	)
	fp.handleMouseWithin(
		tea.MouseMsg{Action: tea.MouseActionRelease, Button: tea.MouseButtonLeft},
		fp.X+fp.Width-1, fp.Y+fp.Height-1, 80, 24,
	)
	if fp.Width != 42 || fp.Height != 12 || fp.preferredWidth != 42 || fp.preferredHeight != 12 {
		t.Fatalf("manual resize = %dx%d, preferred=%dx%d; want 42x12", fp.Width, fp.Height, fp.preferredWidth, fp.preferredHeight)
	}

	tab.Update(ResizeMsg{Width: 5, Height: 2})
	if fp.Width != 5 || fp.Height != 2 || fp.preferredWidth != 42 || fp.preferredHeight != 12 {
		t.Fatalf("float after shrink = %+v, preferred=%dx%d", fp, fp.preferredWidth, fp.preferredHeight)
	}
	tab.Update(ResizeMsg{Width: 80, Height: 24})
	if fp.Width != 42 || fp.Height != 12 {
		t.Fatalf("float after expand = %dx%d, want manual preferred size 42x12", fp.Width, fp.Height)
	}
	assertFloatWithinViewport(t, fp, 80, 24)
}

func TestFloatCreatedBeforeFirstWindowSizeIsReclamped(t *testing.T) {
	tab := NewTab("before-size")
	tab.Float(&geometryTestPanel{name: "float"}, 100, 30, 30, 10)
	fp := tab.floats[0]
	if fp.X != 100 || fp.Y != 30 {
		t.Fatalf("pre-size position=(%d,%d), want (100,30)", fp.X, fp.Y)
	}

	tab.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	assertFloatWithinViewport(t, fp, 60, 20)
	if fp.X != 30 || fp.Y != 10 {
		t.Fatalf("float after first WindowSizeMsg position=(%d,%d), want (30,10)", fp.X, fp.Y)
	}
}

func TestFloatFitsTinyViewports(t *testing.T) {
	for _, size := range [][2]int{{1, 1}, {2, 2}, {5, 2}} {
		tab := NewTab("tiny")
		tab.Update(tea.WindowSizeMsg{Width: size[0], Height: size[1]})
		tab.Float(&geometryTestPanel{name: "float"}, 100, 100, 20, 10)
		fp := tab.floats[0]
		view := tab.View(size[0], size[1])

		assertFloatWithinViewport(t, fp, size[0], size[1])
		lines := strings.Split(view, "\n")
		if len(lines) != size[1] {
			t.Fatalf("viewport %v rendered %d lines, want %d", size, len(lines), size[1])
		}
		for y, line := range lines {
			if got := ansi.StringWidth(line); got != size[0] {
				t.Errorf("viewport %v line %d width=%d, want %d", size, y, got, size[0])
			}
		}
		if !strings.Contains(StripANSI(view), "╭") {
			t.Errorf("float is unreachable in viewport %v: %q", size, StripANSI(view))
		}
	}
}

func assertFloatWithinViewport(t *testing.T, fp *FloatPane, width, height int) {
	t.Helper()
	if fp.X < 0 || fp.Y < 0 || fp.Width < 0 || fp.Height < 0 || fp.X+fp.Width > width || fp.Y+fp.Height > height {
		t.Fatalf("float %+v exceeds viewport %dx%d", fp, width, height)
	}
	if fp.Width == 0 || fp.Height == 0 {
		t.Fatalf("float %+v is not reachable in viewport %dx%d", fp, width, height)
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
