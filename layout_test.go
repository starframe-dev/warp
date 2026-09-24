package warp

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type geometryTestPanel struct {
	BasePanel
	name       string
	lastResize ResizeMsg
}

func (p *geometryTestPanel) View(width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := make([]string, height)
	lines[0] = ansi.Truncate(p.name, width, "") + strings.Repeat(" ", max(0, width-ansi.StringWidth(p.name)))
	for i := 1; i < height; i++ {
		lines[i] = strings.Repeat(" ", width)
	}
	return strings.Join(lines, "\n")
}

func (p *geometryTestPanel) Update(msg tea.Msg) tea.Cmd {
	if resize, ok := msg.(ResizeMsg); ok {
		p.lastResize = resize
	}
	return nil
}

func (p *geometryTestPanel) Elements(width, height int) []Element {
	return []Element{{
		Role:   "panel",
		Name:   p.name,
		Bounds: Bounds{W: width, H: height},
	}}
}

func TestComputeSplitSizesNeverExceedsAvailable(t *testing.T) {
	for avail := 0; avail <= 8; avail++ {
		for _, collapsed := range [][2]bool{{false, false}, {true, false}, {false, true}, {true, true}} {
			first, second := computeSplitSizes(avail, 0.5, collapsed[0], collapsed[1], 7, 5)
			if first < 0 || second < 0 || first+second > avail {
				t.Fatalf("avail=%d collapsed=%v: got (%d,%d)", avail, collapsed, first, second)
			}
		}
	}
}

func TestComputeFlexSizesTinyManyChildren(t *testing.T) {
	items := make([]*FlexItem, 12)
	for i := range items {
		items[i] = &FlexItem{
			Node:      &Node{Panel: &geometryTestPanel{name: string(rune('a' + i))}},
			Basis:     i + 1,
			Grow:      i + 1,
			Collapsed: i%3 == 0,
		}
	}
	for _, avail := range []int{0, 1, 2, 3, 4, 5} {
		sizes := computeFlexSizes(avail, items)
		if len(sizes) != len(items) {
			t.Fatalf("avail=%d: got %d sizes, want %d", avail, len(sizes), len(items))
		}
		total := 0
		for _, size := range sizes {
			if size < 0 {
				t.Fatalf("avail=%d: negative child size %d", avail, size)
			}
			total += size
		}
		if total > avail {
			t.Fatalf("avail=%d: child sizes sum to %d", avail, total)
		}
	}
}

func TestLayoutGeometryFitsTinyContainers(t *testing.T) {
	leaves := make([]*FlexItem, 7)
	for i := range leaves {
		panel := &geometryTestPanel{name: string(rune('a' + i))}
		node := &Node{Panel: panel}
		if i == 2 {
			node.Collapse = &NodeCollapse{Active: true, Width: 8, Height: 8}
		}
		leaves[i] = &FlexItem{Node: node, Grow: i + 1, Collapsed: i == 4}
	}
	tree := &Node{Split: &SplitConfig{
		Direction: Vertical,
		Fraction:  0.5,
		First: &Node{Flex: &FlexConfig{
			Direction: Horizontal,
			Items:     leaves,
		}},
		Second: &Node{Split: &SplitConfig{
			Direction: Horizontal,
			Fraction:  0.5,
			First:     &Node{Panel: &geometryTestPanel{name: "x"}},
			Second:    &Node{Panel: &geometryTestPanel{name: "y"}},
		}},
	}}

	for width := 0; width <= 10; width++ {
		for height := 0; height <= 6; height++ {
			layout := newLayout(tree, layoutRect{w: width, h: height})
			assertLayoutFits(t, layout)
			lines := renderLayout(layout)
			if height == 0 {
				if len(lines) != 0 {
					t.Fatalf("size %dx%d rendered %d lines", width, height, len(lines))
				}
				continue
			}
			if len(lines) != height {
				t.Fatalf("size %dx%d rendered %d lines, want %d", width, height, len(lines), height)
			}
			for y, line := range lines {
				if ansi.StringWidth(line) != width {
					t.Fatalf("size %dx%d line %d has width %d", width, height, y, ansi.StringWidth(line))
				}
			}
		}
	}
}

func assertLayoutFits(t *testing.T, layout *layoutNode) {
	t.Helper()
	if layout == nil {
		return
	}
	bounds := layout.bounds
	if bounds.w < 0 || bounds.h < 0 {
		t.Fatalf("negative layout bounds: %+v", bounds)
	}
	axisSize := 0
	if layout.node != nil && layout.node.Split != nil {
		if layout.node.Split.Direction == Vertical {
			axisSize = bounds.w
		} else if layout.node.Split.Direction == Horizontal {
			axisSize = bounds.h
		}
	} else if layout.node != nil && layout.node.Flex != nil {
		if layout.node.Flex.Direction == Horizontal {
			axisSize = bounds.w
		} else if layout.node.Flex.Direction == Vertical {
			axisSize = bounds.h
		}
	}
	if axisSize > 0 || len(layout.borders) > 0 {
		used := len(layout.borders)
		for _, child := range layout.children {
			if layout.node.Split != nil && layout.node.Split.Direction == Vertical ||
				layout.node.Flex != nil && layout.node.Flex.Direction == Horizontal {
				used += child.bounds.w
			} else {
				used += child.bounds.h
			}
		}
		if used > axisSize {
			t.Fatalf("layout uses %d cells on axis of %d: %+v", used, axisSize, layout)
		}
	}
	for _, child := range layout.children {
		if child.bounds.x < bounds.x || child.bounds.y < bounds.y ||
			child.bounds.x+child.bounds.w > bounds.x+bounds.w ||
			child.bounds.y+child.bounds.h > bounds.y+bounds.h {
			t.Fatalf("child bounds %+v exceed parent %+v", child.bounds, bounds)
		}
		assertLayoutFits(t, child)
	}
}
