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
		switch layout.node.Split.Direction {
		case Vertical:
			axisSize = bounds.w
		case Horizontal:
			axisSize = bounds.h
		}
	} else if layout.node != nil && layout.node.Flex != nil {
		switch layout.node.Flex.Direction {
		case Horizontal:
			axisSize = bounds.w
		case Vertical:
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

func TestLayoutHelpersAndTraversal(t *testing.T) {
	left := &geometryTestPanel{name: "left"}
	right := &geometryTestPanel{name: "right"}
	split := &SplitConfig{Direction: Vertical, Fraction: .5, First: &Node{Panel: left}, Second: &Node{Panel: right}}
	root := &Node{Split: split}
	layout := newLayout(root, layoutRect{x: 2, y: 3, w: 11, h: 4})
	if len(layout.children) != 2 || len(layout.borders) != 1 {
		t.Fatalf("split layout children/borders = %d/%d", len(layout.children), len(layout.borders))
	}
	if got := layout.borders[0]; got.X != 7 || got.Y != 3 || got.Length != 4 || got.Bounds != (Bounds{X: 2, Y: 3, W: 11, H: 4}) {
		t.Fatalf("unexpected split border: %+v", got)
	}
	if got := findLayoutPanel(layout, 2, 3); got == nil || got.Node.Panel != left || got.W != 5 {
		t.Fatalf("unexpected panel hit: %+v", got)
	}
	if findLayoutPanel(layout, 13, 3) != nil || findLayoutPanel(nil, 0, 0) != nil {
		t.Fatal("out-of-bounds or nil layout hit a panel")
	}
	borders := collectLayoutBorders(layout)
	if len(borders) != 1 || findFlexLayout(layout, &FlexConfig{}) != nil {
		t.Fatalf("unexpected traversal results: borders=%d", len(borders))
	}
	if got := elementsFromLayout(layout); len(got) != 2 || got[0].Bounds.X != 2 || got[1].Bounds.X <= got[0].Bounds.X {
		t.Fatalf("unexpected translated elements: %+v", got)
	}
	if elementsFromLayout(nil) != nil || elementsFromLayout(&layoutNode{}) != nil {
		t.Fatal("nil layouts should not produce elements")
	}
	if appendLayoutBorders(nil, nil) != nil || findFlexLayout(nil, nil) != nil || appendElementsFromLayout(nil, nil) != nil {
		t.Fatal("nil traversal should preserve an empty result")
	}
	if got := newLayout(nil, layoutRect{w: -2, h: -3}); got.bounds.w != 0 || got.bounds.h != 0 {
		t.Fatalf("negative bounds not clamped: %+v", got.bounds)
	}
	leaf := newLayout(&Node{Panel: left}, layoutRect{x: 1, y: 2, w: 3, h: 4})
	if len(leaf.children) != 0 || leaf.bounds != (layoutRect{x: 1, y: 2, w: 3, h: 4}) {
		t.Fatalf("unexpected leaf layout: %+v", leaf)
	}
	var nested []Element
	nested = appendElementsFromLayout(nested, layout)
	if len(nested) != 2 {
		t.Fatalf("appendElementsFromLayout returned %d elements", len(nested))
	}
}

func TestLayoutFlexTraversalAndResize(t *testing.T) {
	first := &geometryTestPanel{name: "first"}
	second := &geometryTestPanel{name: "second"}
	flex := &FlexConfig{Direction: Horizontal, Items: []*FlexItem{{Node: &Node{Panel: first}, Grow: 1}, nil, {Node: &Node{Panel: second}, Grow: 1}}}
	root := &Node{Flex: flex}
	layout := newLayout(root, layoutRect{x: 4, y: 5, w: 9, h: 3})
	if found := findFlexLayout(layout, flex); found != layout {
		t.Fatal("failed to locate flex layout")
	}
	if len(layout.children) != 3 || len(collectLayoutBorders(layout)) != 0 {
		t.Fatalf("unexpected flex children/borders: %d/%d", len(layout.children), len(layout.borders))
	}
	if flexItemNode(nil) != nil || !flexItemCollapsed(nil) || flexItemCollapsed(&FlexItem{}) || nodeCollapsed(nil) || nodeCollapsedSize(nil, Vertical) != 0 {
		t.Fatal("nil/collapse helpers returned unexpected values")
	}
	var tab Tab
	commands := tab.broadcastLayoutResize(layout)
	if len(commands) != 0 || first.lastResize.Width != 4 || first.lastResize.Height != 3 || second.lastResize.Width != 4 || second.lastResize.Height != 3 {
		t.Fatalf("resize broadcast failed: commands=%d first=%+v second=%+v", len(commands), first.lastResize, second.lastResize)
	}
	if got := tab.broadcastLayoutResize(nil); len(got) != 0 {
		t.Fatalf("nil resize layout returned %d commands", len(got))
	}
	if elems := elementsAtLayout(layout.children[0]); len(elems) != 1 || elems[0].Bounds.X != 4 || elems[0].Bounds.Y != 5 {
		t.Fatalf("unexpected elementsAtLayout output: %+v", elems)
	}

	vertical := newLayout(&Node{Flex: &FlexConfig{Direction: Vertical, Items: []*FlexItem{{Node: &Node{Panel: first}}, {Node: &Node{Panel: second}}}}}, layoutRect{w: 5, h: 6})
	if len(vertical.children) != 2 || len(vertical.borders) != 1 || vertical.borders[0].Direction != Horizontal {
		t.Fatalf("unexpected vertical flex layout: %+v", vertical)
	}
	unknown := newLayout(&Node{Split: &SplitConfig{Direction: Direction(99)}}, layoutRect{w: 5, h: 5})
	if len(unknown.children) != 0 {
		t.Fatal("unknown split direction should not create children")
	}
	if got := newLayout(&Node{Flex: &FlexConfig{}}, layoutRect{}); len(got.children) != 0 {
		t.Fatal("empty flex should not create children")
	}
}
