package warp

import tea "github.com/charmbracelet/bubbletea"

type layoutRect struct {
	x int
	y int
	w int
	h int
}

type layoutNode struct {
	node     *Node
	bounds   layoutRect
	children []*layoutNode
	borders  []BorderHit
}

func newLayout(node *Node, bounds layoutRect) *layoutNode {
	bounds.w = max(0, bounds.w)
	bounds.h = max(0, bounds.h)
	layout := &layoutNode{node: node, bounds: bounds}
	if node == nil || node.IsLeaf() {
		return layout
	}

	if split := node.Split; split != nil {
		layoutSplit(layout, split)
		return layout
	}
	if flex := node.Flex; flex != nil {
		layoutFlex(layout, flex)
	}
	return layout
}

func layoutSplit(layout *layoutNode, split *SplitConfig) {
	bounds := layout.bounds
	firstCollapsed := nodeCollapsed(split.First)
	secondCollapsed := nodeCollapsed(split.Second)

	switch split.Direction {
	case Vertical:
		borderVisible := !firstCollapsed && !secondCollapsed && bounds.w > 0
		borderSize := 0
		if borderVisible {
			borderSize = 1
		}
		avail := max(0, bounds.w-borderSize)
		firstW, secondW := computeSplitSizes(
			avail,
			split.Fraction,
			firstCollapsed,
			secondCollapsed,
			nodeCollapsedSize(split.First, Vertical),
			nodeCollapsedSize(split.Second, Vertical),
		)
		firstBounds := layoutRect{x: bounds.x, y: bounds.y, w: firstW, h: bounds.h}
		secondX := bounds.x + firstW + borderSize
		secondBounds := layoutRect{x: secondX, y: bounds.y, w: secondW, h: bounds.h}
		layout.children = []*layoutNode{
			newLayout(split.First, firstBounds),
			newLayout(split.Second, secondBounds),
		}
		if borderVisible {
			layout.borders = append(layout.borders, BorderHit{
				Split:     split,
				Direction: Vertical,
				X:         bounds.x + firstW,
				Y:         bounds.y,
				Length:    bounds.h,
				Bounds:    bounds.toBounds(),
				FlexIndex: -1,
			})
		}
	case Horizontal:
		borderVisible := !firstCollapsed && !secondCollapsed && bounds.h > 0
		borderSize := 0
		if borderVisible {
			borderSize = 1
		}
		avail := max(0, bounds.h-borderSize)
		firstH, secondH := computeSplitSizes(
			avail,
			split.Fraction,
			firstCollapsed,
			secondCollapsed,
			nodeCollapsedSize(split.First, Horizontal),
			nodeCollapsedSize(split.Second, Horizontal),
		)
		firstBounds := layoutRect{x: bounds.x, y: bounds.y, w: bounds.w, h: firstH}
		secondY := bounds.y + firstH + borderSize
		secondBounds := layoutRect{x: bounds.x, y: secondY, w: bounds.w, h: secondH}
		layout.children = []*layoutNode{
			newLayout(split.First, firstBounds),
			newLayout(split.Second, secondBounds),
		}
		if borderVisible {
			layout.borders = append(layout.borders, BorderHit{
				Split:     split,
				Direction: Horizontal,
				X:         bounds.x,
				Y:         bounds.y + firstH,
				Length:    bounds.w,
				Bounds:    bounds.toBounds(),
				FlexIndex: -1,
			})
		}
	}
}

func layoutFlex(layout *layoutNode, flex *FlexConfig) {
	count := len(flex.Items)
	if count == 0 {
		return
	}

	bounds := layout.bounds
	vertical := flex.Direction == Vertical
	axisSize := bounds.w
	if vertical {
		axisSize = bounds.h
	}
	visible := make([]bool, max(0, count-1))
	visibleCount := 0
	for i := range visible {
		if !flexItemCollapsed(flex.Items[i]) && !flexItemCollapsed(flex.Items[i+1]) && visibleCount < axisSize {
			visible[i] = true
			visibleCount++
		}
	}

	avail := max(0, axisSize-visibleCount)
	sizes := computeFlexSizes(avail, flex.Items)
	cursor := bounds.x
	if vertical {
		cursor = bounds.y
	}
	for i, item := range flex.Items {
		if i > 0 && visible[i-1] {
			if vertical {
				layout.borders = append(layout.borders, BorderHit{
					Flex:      flex,
					Direction: Horizontal,
					X:         bounds.x,
					Y:         cursor,
					Length:    bounds.w,
					Bounds:    bounds.toBounds(),
					FlexIndex: i - 1,
				})
			} else {
				layout.borders = append(layout.borders, BorderHit{
					Flex:      flex,
					Direction: Vertical,
					X:         cursor,
					Y:         bounds.y,
					Length:    bounds.h,
					Bounds:    bounds.toBounds(),
					FlexIndex: i - 1,
				})
			}
			cursor++
		}

		childSize := 0
		if i < len(sizes) {
			childSize = max(0, sizes[i])
		}
		childBounds := bounds
		if vertical {
			childBounds.y = cursor
			childBounds.h = childSize
		} else {
			childBounds.x = cursor
			childBounds.w = childSize
		}
		layout.children = append(layout.children, newLayout(flexItemNode(item), childBounds))
		cursor += childSize
	}
}

func flexItemNode(item *FlexItem) *Node {
	if item == nil {
		return nil
	}
	return item.Node
}

func flexItemCollapsed(item *FlexItem) bool {
	return item == nil || item.Collapsed
}

func nodeCollapsed(node *Node) bool {
	return node != nil && node.IsCollapsed()
}

func nodeCollapsedSize(node *Node, direction Direction) int {
	if node == nil {
		return 0
	}
	return node.CollapsedSize(direction)
}

func (r layoutRect) toBounds() Bounds {
	return Bounds{X: r.x, Y: r.y, W: r.w, H: r.h}
}

func collectLayoutBorders(layout *layoutNode) []BorderHit {
	var borders []BorderHit
	return appendLayoutBorders(borders, layout)
}

func appendLayoutBorders(borders []BorderHit, layout *layoutNode) []BorderHit {
	if layout == nil {
		return borders
	}
	borders = append(borders, layout.borders...)
	for _, child := range layout.children {
		borders = appendLayoutBorders(borders, child)
	}
	return borders
}

func findFlexLayout(layout *layoutNode, flex *FlexConfig) *layoutNode {
	if layout == nil {
		return nil
	}
	if layout.node != nil && layout.node.Flex == flex {
		return layout
	}
	for _, child := range layout.children {
		if found := findFlexLayout(child, flex); found != nil {
			return found
		}
	}
	return nil
}

func findLayoutPanel(layout *layoutNode, x, y int) *panelHit {
	if layout == nil || x < layout.bounds.x || y < layout.bounds.y ||
		x >= layout.bounds.x+layout.bounds.w || y >= layout.bounds.y+layout.bounds.h {
		return nil
	}
	if layout.node != nil && layout.node.IsLeaf() {
		return &panelHit{
			Node: layout.node,
			X:    layout.bounds.x,
			Y:    layout.bounds.y,
			W:    layout.bounds.w,
			H:    layout.bounds.h,
		}
	}
	for _, child := range layout.children {
		if hit := findLayoutPanel(child, x, y); hit != nil {
			return hit
		}
	}
	return nil
}

func elementsFromLayout(layout *layoutNode) []Element {
	if layout == nil || layout.node == nil {
		return nil
	}
	if layout.node.IsLeaf() {
		return elementsAtLayout(layout)
	}
	var elements []Element
	return appendElementsFromLayout(elements, layout)
}

func appendElementsFromLayout(elements []Element, layout *layoutNode) []Element {
	if layout == nil || layout.node == nil {
		return elements
	}
	if layout.node.IsLeaf() {
		return append(elements, elementsAtLayout(layout)...)
	}
	for _, child := range layout.children {
		elements = appendElementsFromLayout(elements, child)
	}
	return elements
}

func elementsAtLayout(layout *layoutNode) []Element {
	elements := collectElements(layout.node.Panel, layout.bounds.w, layout.bounds.h)
	for i := range elements {
		elements[i].Bounds.X += layout.bounds.x
		elements[i].Bounds.Y += layout.bounds.y
		shiftElements(elements[i].Children, layout.bounds.x, layout.bounds.y)
	}
	return elements
}

func (t *Tab) broadcastLayoutResize(layout *layoutNode) []tea.Cmd {
	var commands []tea.Cmd
	t.appendLayoutResize(&commands, layout)
	return commands
}

func (t *Tab) appendLayoutResize(commands *[]tea.Cmd, layout *layoutNode) {
	if layout == nil || layout.node == nil {
		return
	}
	if layout.node.IsLeaf() {
		if layout.node.Panel != nil {
			if cmd := layout.node.Panel.Update(ResizeMsg{Width: layout.bounds.w, Height: layout.bounds.h}); cmd != nil {
				*commands = append(*commands, cmd)
			}
		}
		return
	}
	for _, child := range layout.children {
		t.appendLayoutResize(commands, child)
	}
}
