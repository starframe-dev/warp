package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

// mockPanel is a simple panel for testing layout/rendering code.
type mockPanel struct {
	name string
}

func (p *mockPanel) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := make([]string, height)
	for i := range lines {
		if i == 0 {
			line := p.name
			if len(line) > width {
				line = line[:width]
			}
			line += strings.Repeat(" ", width-len(line))
			lines[i] = line
		} else {
			lines[i] = strings.Repeat(" ", width)
		}
	}
	return strings.Join(lines, "\n")
}

func (p *mockPanel) Update(msg tea.Msg) tea.Cmd { return nil }

func TestMakeEmptyLines(t *testing.T) {
	if got := makeEmptyLines(0, 5); got != nil {
		t.Errorf("makeEmptyLines(0,5) = %v, want nil", got)
	}
	if got := makeEmptyLines(5, 0); got != nil {
		t.Errorf("makeEmptyLines(5,0) = %v, want nil", got)
	}
	if got := makeEmptyLines(-1, -1); got != nil {
		t.Errorf("makeEmptyLines(-1,-1) = %v, want nil", got)
	}

	got := makeEmptyLines(4, 3)
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	want := strings.Repeat(" ", 4)
	for i, line := range got {
		if line != want {
			t.Errorf("line %d = %q, want %q", i, line, want)
		}
	}
}

func TestPadContent(t *testing.T) {
	if got := padContent("hello", 0, 3); got != nil {
		t.Errorf("padContent with zero width = %v, want nil", got)
	}
	if got := padContent("hello", 3, 0); got != nil {
		t.Errorf("padContent with zero height = %v, want nil", got)
	}

	got := padContent("hello", 3, 1)
	if len(got) != 1 {
		t.Fatalf("expected 1 line, got %d", len(got))
	}
	if ansi.StringWidth(got[0]) != 3 {
		t.Errorf("line width = %d, want 3", ansi.StringWidth(got[0]))
	}
	if !strings.HasSuffix(got[0], ansi.ResetStyle) {
		t.Errorf("expected reset suffix, got %q", got[0])
	}

	got = padContent("a\nb", 5, 3)
	if len(got) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(got))
	}
	for i, line := range got {
		if ansi.StringWidth(line) != 5 {
			t.Errorf("line %d width = %d, want 5", i, ansi.StringWidth(line))
		}
	}

	styled := "\x1b[31mhello\x1b[0m"
	got = padContent(styled, 3, 1)
	if ansi.StringWidth(got[0]) != 3 {
		t.Errorf("styled width = %d, want 3", ansi.StringWidth(got[0]))
	}
}

func TestComputeSplitSizes(t *testing.T) {
	first, second := computeSplitSizes(100, 0.3, false, false, 0, 0)
	if first != 30 || second != 70 {
		t.Errorf("split sizes = (%d,%d), want (30,70)", first, second)
	}

	first, second = computeSplitSizes(10, 0.1, false, false, 0, 0)
	if first != MinPanelSize || second != 10-MinPanelSize {
		t.Errorf("min clamp = (%d,%d), want (%d,%d)", first, second, MinPanelSize, 10-MinPanelSize)
	}

	first, second = computeSplitSizes(100, 0.0, true, false, 10, 0)
	if first != 10 || second != 90 {
		t.Errorf("first collapsed = (%d,%d), want (10,90)", first, second)
	}

	first, second = computeSplitSizes(100, 0.0, false, true, 0, 20)
	if second != 20 || first != 80 {
		t.Errorf("second collapsed = (%d,%d), want (80,20)", first, second)
	}

	first, second = computeSplitSizes(10, 0.0, true, true, 1, 1)
	if first != 1 || second != 9 {
		t.Errorf("both collapsed = (%d,%d), want (1,9)", first, second)
	}
}

func TestComputeFlexSizes(t *testing.T) {
	if got := computeFlexSizes(10, nil); got != nil {
		t.Errorf("empty items = %v, want nil", got)
	}

	collapsed := []*FlexItem{
		{Node: &Node{Panel: &mockPanel{name: "a"}}, Collapsed: true},
		{Node: &Node{Panel: &mockPanel{name: "b"}}, Collapsed: true},
	}
	sizes := computeFlexSizes(10, collapsed)
	if len(sizes) != 2 || sizes[0] != 1 || sizes[1] != 1 {
		t.Errorf("collapsed sizes = %v, want [1 1]", sizes)
	}

	items := []*FlexItem{
		{Node: &Node{Panel: &mockPanel{name: "a"}}, Basis: 2, Grow: 0},
		{Node: &Node{Panel: &mockPanel{name: "b"}}, Basis: 2, Grow: 0},
	}
	sizes = computeFlexSizes(8, items)
	if len(sizes) != 2 || sizes[0] != 4 || sizes[1] != 4 {
		t.Errorf("equal distribution = %v, want [4 4]", sizes)
	}

	items = []*FlexItem{
		{Node: &Node{Panel: &mockPanel{name: "a"}}, Basis: 2, Grow: 1},
		{Node: &Node{Panel: &mockPanel{name: "b"}}, Basis: 2, Grow: 2},
	}
	sizes = computeFlexSizes(17, items)
	if len(sizes) != 2 || sizes[0] != 6 || sizes[1] != 11 {
		t.Errorf("grow distribution = %v, want [6 11]", sizes)
	}

	items = []*FlexItem{
		{Node: &Node{Panel: &mockPanel{name: "a"}}, Basis: 10, Grow: 0},
		{Node: &Node{Panel: &mockPanel{name: "b"}}, Basis: 10, Grow: 0},
	}
	sizes = computeFlexSizes(5, items)
	if len(sizes) != 2 || sizes[0] != 10 || sizes[1] != 10 {
		t.Errorf("avail < basis = %v, want [10 10]", sizes)
	}
}

func TestRenderNode(t *testing.T) {
	got := renderNode(nil, 5, 2)
	if len(got) != 2 || ansi.StringWidth(got[0]) != 5 {
		t.Errorf("nil node = %v", got)
	}

	p := &mockPanel{name: "AB"}
	got = renderNode(&Node{Panel: p}, 5, 2)
	if len(got) != 2 || !strings.Contains(got[0], "AB") || ansi.StringWidth(got[0]) != 5 {
		t.Errorf("leaf node = %v", got)
	}

	splitNode := &Node{Split: &SplitConfig{
		Direction: Vertical,
		Fraction:    0.5,
		First:       &Node{Panel: &mockPanel{name: "L"}},
		Second:      &Node{Panel: &mockPanel{name: "R"}},
	}}
	got = renderNode(splitNode, 21, 3)
	if len(got) != 3 {
		t.Errorf("split node lines = %d, want 3", len(got))
	}
	for i, line := range got {
		if ansi.StringWidth(line) != 21 {
			t.Errorf("split line %d width = %d, want 21", i, ansi.StringWidth(line))
		}
	}

	flexNode := &Node{Flex: &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}}
	got = renderNode(flexNode, 11, 3)
	if len(got) != 3 {
		t.Errorf("flex node lines = %d, want 3", len(got))
	}
	for i, line := range got {
		if ansi.StringWidth(line) != 11 {
			t.Errorf("flex line %d width = %d, want 11", i, ansi.StringWidth(line))
		}
	}

	empty := renderNode(&Node{}, 5, 2)
	if len(empty) != 2 || ansi.StringWidth(empty[0]) != 5 {
		t.Errorf("empty node = %v", empty)
	}
}

func TestRenderVerticalSplit(t *testing.T) {
	split := &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First:     &Node{Panel: &mockPanel{name: "L"}},
		Second:    &Node{Panel: &mockPanel{name: "R"}},
	}
	lines := renderVerticalSplit(split, 21, 3)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if ansi.StringWidth(line) != 21 {
			t.Errorf("line %d width = %d, want 21", i, ansi.StringWidth(line))
		}
	}
	if !strings.Contains(StripANSI(lines[0]), "│") {
		t.Errorf("expected vertical border, got %q", lines[0])
	}

	firstCollapsed := &Node{
		Panel:    &mockPanel{name: "L"},
		Collapse: &NodeCollapse{Active: true, Width: 1},
	}
	split2 := &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First:     firstCollapsed,
		Second:    &Node{Panel: &mockPanel{name: "R"}},
	}
	lines2 := renderVerticalSplit(split2, 21, 3)
	if strings.Contains(StripANSI(lines2[0]), "│") {
		t.Errorf("collapsed split should not render a border, got %q", lines2[0])
	}

	split.Dragging = true
	lines3 := renderVerticalSplit(split, 21, 3)
	if !strings.Contains(StripANSI(lines3[0]), "│") {
		t.Errorf("expected dragging border, got %q", lines3[0])
	}
}

func TestRenderHorizontalSplit(t *testing.T) {
	split := &SplitConfig{
		Direction: Horizontal,
		Fraction:  0.5,
		First:     &Node{Panel: &mockPanel{name: "T"}},
		Second:    &Node{Panel: &mockPanel{name: "B"}},
	}
	lines := renderHorizontalSplit(split, 10, 11)
	if len(lines) != 11 {
		t.Fatalf("expected 11 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if ansi.StringWidth(line) != 10 {
			t.Errorf("line %d width = %d, want 10", i, ansi.StringWidth(line))
		}
	}
	borderCount := 0
	for _, line := range lines {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 1 {
		t.Errorf("expected exactly 1 horizontal border line, got %d", borderCount)
	}

	firstCollapsed := &Node{
		Panel:    &mockPanel{name: "T"},
		Collapse: &NodeCollapse{Active: true, Height: 1},
	}
	split2 := &SplitConfig{
		Direction: Horizontal,
		Fraction:  0.5,
		First:     firstCollapsed,
		Second:    &Node{Panel: &mockPanel{name: "B"}},
	}
	lines2 := renderHorizontalSplit(split2, 10, 11)
	borderCount = 0
	for _, line := range lines2 {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 0 {
		t.Errorf("collapsed horizontal split should not render border, got %d", borderCount)
	}

	split.Dragging = true
	lines3 := renderHorizontalSplit(split, 10, 11)
	borderCount = 0
	for _, line := range lines3 {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 1 {
		t.Errorf("expected dragging border line, got %d", borderCount)
	}
}

func TestRenderFlex(t *testing.T) {
	empty := renderFlex(&FlexConfig{Direction: Horizontal, Items: []*FlexItem{}}, 5, 2)
	if len(empty) != 2 {
		t.Errorf("empty flex = %v", empty)
	}

	row := renderFlex(&FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}, 11, 3)
	if len(row) != 3 {
		t.Errorf("row lines = %d, want 3", len(row))
	}
	for i, line := range row {
		if ansi.StringWidth(line) != 11 {
			t.Errorf("row line %d width = %d, want 11", i, ansi.StringWidth(line))
		}
	}
	if !strings.Contains(StripANSI(row[0]), "│") {
		t.Errorf("expected row border, got %q", row[0])
	}

	col := renderFlex(&FlexConfig{
		Direction: Vertical,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}, 5, 7)
	if len(col) != 7 {
		t.Errorf("column lines = %d, want 7", len(col))
	}
	for i, line := range col {
		if ansi.StringWidth(line) != 5 {
			t.Errorf("col line %d width = %d, want 5", i, ansi.StringWidth(line))
		}
	}

	invalid := renderFlex(&FlexConfig{
		Direction: Direction(99),
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
		},
	}, 5, 2)
	if len(invalid) != 2 || ansi.StringWidth(invalid[0]) != 5 {
		t.Errorf("invalid direction = %v", invalid)
	}
}

func TestRenderFlexRow(t *testing.T) {
	empty := renderFlexRow(&FlexConfig{Items: []*FlexItem{}}, 5, 2, nil)
	if len(empty) != 2 {
		t.Errorf("empty row = %v", empty)
	}

	flex := &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	lines := renderFlexRow(flex, 11, 3, []int{5, 5})
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if ansi.StringWidth(line) != 11 {
			t.Errorf("line %d width = %d, want 11", i, ansi.StringWidth(line))
		}
	}
	if !strings.Contains(StripANSI(lines[0]), "│") {
		t.Errorf("expected row border, got %q", lines[0])
	}

	flex2 := &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Collapsed: true},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	lines2 := renderFlexRow(flex2, 10, 3, []int{1, 9})
	if strings.Contains(StripANSI(lines2[0]), "│") {
		t.Errorf("collapsed adjacent items should not render border, got %q", lines2[0])
	}

	flex.Dragging = true
	lines3 := renderFlexRow(flex, 11, 3, []int{5, 5})
	if !strings.Contains(StripANSI(lines3[0]), "│") {
		t.Errorf("expected dragging row border, got %q", lines3[0])
	}
}

func TestRenderFlexColumn(t *testing.T) {
	empty := renderFlexColumn(&FlexConfig{Items: []*FlexItem{}}, 5, 2, nil)
	if len(empty) != 2 {
		t.Errorf("empty column = %v", empty)
	}

	flex := &FlexConfig{
		Direction: Vertical,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	lines := renderFlexColumn(flex, 5, 7, []int{3, 3})
	if len(lines) != 7 {
		t.Fatalf("expected 7 lines, got %d", len(lines))
	}
	for i, line := range lines {
		if ansi.StringWidth(line) != 5 {
			t.Errorf("line %d width = %d, want 5", i, ansi.StringWidth(line))
		}
	}
	borderCount := 0
	for _, line := range lines {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 1 {
		t.Errorf("expected exactly 1 horizontal border, got %d", borderCount)
	}

	flex2 := &FlexConfig{
		Direction: Vertical,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Collapsed: true},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	lines2 := renderFlexColumn(flex2, 5, 7, []int{1, 5})
	if len(lines2) != 6 {
		t.Errorf("collapsed column lines = %d, want 6", len(lines2))
	}
	borderCount = 0
	for _, line := range lines2 {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 0 {
		t.Errorf("collapsed adjacent items should not render horizontal border, got %d", borderCount)
	}

	flex.Dragging = true
	lines3 := renderFlexColumn(flex, 5, 7, []int{3, 3})
	borderCount = 0
	for _, line := range lines3 {
		if strings.Contains(StripANSI(line), "─") {
			borderCount++
		}
	}
	if borderCount != 1 {
		t.Errorf("expected dragging horizontal border, got %d", borderCount)
	}
}

func TestFindBorders(t *testing.T) {
	if got := findBorders(nil, 0, 0, 10, 5); got != nil {
		t.Errorf("nil node = %v", got)
	}
	if got := findBorders(&Node{Panel: &mockPanel{name: "x"}}, 0, 0, 10, 5); got != nil {
		t.Errorf("leaf node = %v", got)
	}

	verticalNode := &Node{Split: &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First:     &Node{Panel: &mockPanel{name: "a"}},
		Second:    &Node{Panel: &mockPanel{name: "b"}},
	}}
	borders := findBorders(verticalNode, 0, 0, 21, 5)
	if len(borders) != 1 {
		t.Fatalf("expected 1 vertical border, got %d", len(borders))
	}
	if borders[0].Direction != Vertical || borders[0].X != 10 || borders[0].Y != 0 || borders[0].Length != 5 {
		t.Errorf("unexpected border: %+v", borders[0])
	}

	horizontalNode := &Node{Split: &SplitConfig{
		Direction: Horizontal,
		Fraction:  0.5,
		First:     &Node{Panel: &mockPanel{name: "a"}},
		Second:    &Node{Panel: &mockPanel{name: "b"}},
	}}
	borders = findBorders(horizontalNode, 0, 0, 10, 11)
	if len(borders) != 1 {
		t.Fatalf("expected 1 horizontal border, got %d", len(borders))
	}
	if borders[0].Direction != Horizontal || borders[0].X != 0 || borders[0].Y != 5 || borders[0].Length != 10 {
		t.Errorf("unexpected border: %+v", borders[0])
	}

	collapsedNode := &Node{Split: &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First: &Node{
			Panel:    &mockPanel{name: "a"},
			Collapse: &NodeCollapse{Active: true, Width: 1},
		},
		Second: &Node{Panel: &mockPanel{name: "b"}},
	}}
	borders = findBorders(collapsedNode, 0, 0, 21, 5)
	if len(borders) != 0 {
		t.Errorf("expected no border for collapsed split, got %d", len(borders))
	}

	flexNode := &Node{Flex: &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}}
	borders = findBorders(flexNode, 0, 0, 11, 5)
	if len(borders) != 1 {
		t.Fatalf("expected 1 flex border, got %d", len(borders))
	}
	if borders[0].Direction != Vertical || borders[0].X != 5 || borders[0].Y != 0 || borders[0].Length != 5 {
		t.Errorf("unexpected flex border: %+v", borders[0])
	}
}

func TestFindFlexBorders(t *testing.T) {
	if got := findFlexBorders(&FlexConfig{Items: []*FlexItem{}}, 0, 0, 10, 5); got != nil {
		t.Errorf("empty flex = %v", got)
	}

	rowFlex := &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	borders := findFlexBorders(rowFlex, 0, 0, 11, 5)
	if len(borders) != 1 {
		t.Fatalf("expected 1 row border, got %d", len(borders))
	}
	if borders[0].Direction != Vertical || borders[0].X != 5 || borders[0].Y != 0 || borders[0].Length != 5 {
		t.Errorf("unexpected row border: %+v", borders[0])
	}

	colFlex := &FlexConfig{
		Direction: Vertical,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	borders = findFlexBorders(colFlex, 0, 0, 5, 5)
	if len(borders) != 1 {
		t.Fatalf("expected 1 column border, got %d", len(borders))
	}
	if borders[0].Direction != Horizontal || borders[0].X != 0 || borders[0].Y != 3 || borders[0].Length != 5 {
		t.Errorf("unexpected column border: %+v", borders[0])
	}

	collapsedFlex := &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Panel: &mockPanel{name: "a"}}, Collapsed: true},
			{Node: &Node{Panel: &mockPanel{name: "b"}}, Grow: 1},
		},
	}
	borders = findFlexBorders(collapsedFlex, 0, 0, 10, 5)
	if len(borders) != 0 {
		t.Errorf("expected no border for collapsed flex item, got %d", len(borders))
	}

	nestedFlex := &FlexConfig{
		Direction: Horizontal,
		Items: []*FlexItem{
			{Node: &Node{Split: &SplitConfig{
				Direction: Vertical,
				Fraction:  0.5,
				First:     &Node{Panel: &mockPanel{name: "a"}},
				Second:    &Node{Panel: &mockPanel{name: "b"}},
			}}, Grow: 1},
			{Node: &Node{Panel: &mockPanel{name: "c"}}, Grow: 1},
		},
	}
	borders = findFlexBorders(nestedFlex, 0, 0, 11, 5)
	if len(borders) != 2 {
		t.Errorf("expected 2 borders from nested flex+split, got %d", len(borders))
	}
}
