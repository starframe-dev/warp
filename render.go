package warp

import (
	"math"
	"strings"

	"github.com/charmbracelet/x/ansi"
)

func renderNode(node *Node, w, h int) []string {
	return renderLayout(newLayout(node, layoutRect{w: max(0, w), h: max(0, h)}))
}

func renderLayout(layout *layoutNode) []string {
	if layout == nil {
		return nil
	}
	bounds := layout.bounds
	if bounds.h <= 0 {
		return nil
	}
	if layout.node == nil {
		return renderBlankLines(bounds.w, bounds.h)
	}
	if layout.node.IsLeaf() {
		if bounds.w <= 0 || layout.node.Panel == nil {
			return renderBlankLines(bounds.w, bounds.h)
		}
		return padContent(layout.node.Panel.View(bounds.w, bounds.h), bounds.w, bounds.h)
	}
	if layout.node.Split != nil {
		return renderSplitLayout(layout)
	}
	if layout.node.Flex != nil {
		return renderFlexLayout(layout)
	}
	return renderBlankLines(bounds.w, bounds.h)
}

func renderSplitLayout(layout *layoutNode) []string {
	if len(layout.children) < 2 {
		return renderBlankLines(layout.bounds.w, layout.bounds.h)
	}
	split := layout.node.Split
	first := renderLayout(layout.children[0])
	second := renderLayout(layout.children[1])
	borderVisible := len(layout.borders) > 0
	bounds := layout.bounds

	switch split.Direction {
	case Vertical:
		border := renderVerticalBorder(split.Dragging)
		collapseBorder := ""
		if split.OnCollapse != nil && split.CollapseRow >= 0 {
			collapseBorder = ansi.ResetStyle + collapseStyle.Render("<") + ansi.ResetStyle
		}
		lines := make([]string, bounds.h)
		for y := range lines {
			var row strings.Builder
			row.WriteString(lineAt(first, y))
			if borderVisible {
				if collapseBorder != "" && y == split.CollapseRow {
					row.WriteString(collapseBorder)
				} else {
					row.WriteString(border)
				}
			}
			row.WriteString(lineAt(second, y))
			lines[y] = padVisualLine(row.String(), bounds.w)
		}
		return lines
	case Horizontal:
		lines := make([]string, 0, bounds.h)
		lines = append(lines, first...)
		if borderVisible {
			lines = append(lines, renderHorizontalBorder(bounds.w, split.Dragging))
		}
		lines = append(lines, second...)
		return padLayoutLines(lines, bounds.w, bounds.h)
	default:
		return renderBlankLines(bounds.w, bounds.h)
	}
}

func renderFlexLayout(layout *layoutNode) []string {
	flex := layout.node.Flex
	if flex == nil || len(layout.children) == 0 {
		return renderBlankLines(layout.bounds.w, layout.bounds.h)
	}

	visible := make([]bool, max(0, len(layout.children)-1))
	for _, border := range layout.borders {
		if border.FlexIndex >= 0 && border.FlexIndex < len(visible) {
			visible[border.FlexIndex] = true
		}
	}

	switch flex.Direction {
	case Horizontal:
		children := make([][]string, len(layout.children))
		for i, child := range layout.children {
			children[i] = renderLayout(child)
		}
		lines := make([]string, layout.bounds.h)
		border := renderVerticalBorder(flex.Dragging)
		for y := range lines {
			var row strings.Builder
			for i, child := range children {
				if i > 0 && visible[i-1] {
					row.WriteString(border)
				}
				row.WriteString(lineAt(child, y))
			}
			lines[y] = padVisualLine(row.String(), layout.bounds.w)
		}
		return lines
	case Vertical:
		lines := make([]string, 0, layout.bounds.h)
		for i, child := range layout.children {
			if i > 0 && visible[i-1] {
				lines = append(lines, renderHorizontalBorder(layout.bounds.w, flex.Dragging))
			}
			lines = append(lines, renderLayout(child)...)
		}
		return padLayoutLines(lines, layout.bounds.w, layout.bounds.h)
	default:
		return renderBlankLines(layout.bounds.w, layout.bounds.h)
	}
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func renderBlankLines(w, h int) []string {
	if h <= 0 {
		return nil
	}
	lines := make([]string, h)
	if w > 0 {
		blank := strings.Repeat(" ", w)
		for i := range lines {
			lines[i] = blank
		}
	}
	return lines
}

func padLayoutLines(lines []string, w, h int) []string {
	if h <= 0 {
		return nil
	}
	result := make([]string, h)
	for i := 0; i < h; i++ {
		if i < len(lines) {
			result[i] = padVisualLine(lines[i], w)
		} else if w > 0 {
			result[i] = strings.Repeat(" ", w)
		}
	}
	return result
}

func renderVerticalBorder(dragging bool) string {
	style := borderStyle
	if dragging {
		style = borderDragStyle
	}
	return ansi.ResetStyle + style.Render("│") + ansi.ResetStyle
}

func renderHorizontalBorder(width int, dragging bool) string {
	if width <= 0 {
		return ""
	}
	style := borderStyle
	if dragging {
		style = borderDragStyle
	}
	line := strings.Repeat("─", width)
	return ansi.ResetStyle + style.Render(line) + ansi.ResetStyle
}

func renderVerticalSplit(split *SplitConfig, w, h int) []string {
	return renderNode(&Node{Split: split}, w, h)
}

func renderHorizontalSplit(split *SplitConfig, w, h int) []string {
	return renderNode(&Node{Split: split}, w, h)
}

func renderFlex(flex *FlexConfig, w, h int) []string {
	return renderNode(&Node{Flex: flex}, w, h)
}

func renderFlexRow(flex *FlexConfig, w, h int, _ []int) []string {
	return renderFlex(flex, w, h)
}

func renderFlexColumn(flex *FlexConfig, w, h int, _ []int) []string {
	return renderFlex(flex, w, h)
}

func computeFlexSizes(avail int, items []*FlexItem) []int {
	if len(items) == 0 {
		return nil
	}
	avail = max(0, avail)
	sizes := make([]int, len(items))
	bases := make([]float64, len(items))
	baseTotal := float64(0)
	growTotal := float64(0)
	eligible := make([]int, 0, len(items))
	growWeights := make([]float64, len(items))

	for i, item := range items {
		if flexItemCollapsed(item) {
			bases[i] = 1
			baseTotal++
			continue
		}
		basis := item.Basis
		if basis <= 0 {
			basis = MinPanelSize
		}
		bases[i] = float64(basis)
		baseTotal += bases[i]
		eligible = append(eligible, i)
		if item.Grow > 0 {
			growWeights[i] = float64(item.Grow)
			growTotal += growWeights[i]
		}
	}

	if avail == 0 {
		return sizes
	}
	if baseTotal > float64(avail) {
		weights := make([]float64, len(items))
		copy(weights, bases)
		distributeSizes(avail, sizes, weights, allIndices(len(items)))
		return sizes
	}

	used := 0
	for i, basis := range bases {
		sizes[i] = int(basis)
		used += sizes[i]
	}
	remaining := avail - used
	if remaining <= 0 || len(eligible) == 0 {
		return sizes
	}

	weights := growWeights
	if growTotal <= 0 {
		weights = make([]float64, len(items))
		for _, i := range eligible {
			weights[i] = 1
		}
	}
	distributeSizes(remaining, sizes, weights, eligible)
	return sizes
}

func allIndices(n int) []int {
	indices := make([]int, n)
	for i := range indices {
		indices[i] = i
	}
	return indices
}

func distributeSizes(amount int, sizes []int, weights []float64, indices []int) {
	if amount <= 0 || len(indices) == 0 {
		return
	}
	totalWeight := float64(0)
	for _, i := range indices {
		totalWeight += weights[i]
	}
	if totalWeight <= 0 || math.IsInf(totalWeight, 0) || math.IsNaN(totalWeight) {
		return
	}

	remaining := amount
	for position, index := range indices {
		share := remaining
		if position < len(indices)-1 {
			share = int(math.Floor(float64(amount) * weights[index] / totalWeight))
			if share > remaining {
				share = remaining
			}
			if share < 0 {
				share = 0
			}
		}
		sizes[index] += share
		remaining -= share
	}
}

func computeSplitSizes(avail int, fraction float64, firstCollapsed, secondCollapsed bool, firstSize, secondSize int) (first, second int) {
	avail = max(0, avail)
	collapsedSize := func(size int) int {
		if size < 1 {
			return 1
		}
		return size
	}

	if firstCollapsed {
		first = min(collapsedSize(firstSize), avail)
		if secondCollapsed {
			return first, avail - first
		}
		if avail >= MinPanelSize+1 && first > avail-MinPanelSize {
			first = avail - MinPanelSize
		}
		return first, avail - first
	}
	if secondCollapsed {
		second = min(collapsedSize(secondSize), avail)
		if avail >= MinPanelSize+1 && second > avail-MinPanelSize {
			second = avail - MinPanelSize
		}
		return avail - second, second
	}

	if math.IsNaN(fraction) {
		fraction = 0.5
	}
	if fraction < 0 {
		fraction = 0
	} else if fraction > 1 {
		fraction = 1
	}
	first = int(float64(avail) * fraction)
	if avail >= 2*MinPanelSize {
		first = min(max(first, MinPanelSize), avail-MinPanelSize)
	} else {
		first = min(max(first, 0), avail)
	}
	second = avail - first
	return first, second
}

func padVisualLine(line string, width int) string {
	if width <= 0 {
		return ""
	}
	line = ansi.Truncate(line, width, "")
	lineWidth := ansi.StringWidth(line)
	if lineWidth < width {
		line += strings.Repeat(" ", width-lineWidth)
	}
	return line
}

func padContent(content string, w, h int) []string {
	if w <= 0 || h <= 0 {
		return makeEmptyLines(w, h)
	}
	lines := strings.Split(content, "\n")
	result := make([]string, h)
	for y := 0; y < h; y++ {
		line := ""
		if y < len(lines) {
			line = lines[y]
		}
		result[y] = padVisualLine(line, w) + ansi.ResetStyle
	}
	return result
}

func makeEmptyLines(w, h int) []string {
	if w <= 0 || h <= 0 {
		return nil
	}
	lines := make([]string, h)
	empty := strings.Repeat(" ", w)
	for i := range lines {
		lines[i] = empty
	}
	return lines
}

// BorderHit describes a draggable border and the layout rectangle that owns it.
type BorderHit struct {
	Split     *SplitConfig
	Flex      *FlexConfig
	Direction Direction
	X, Y      int
	Length    int
	Bounds    Bounds
	FlexIndex int
}

func findBorders(node *Node, x, y, w, h int) []BorderHit {
	layout := newLayout(node, layoutRect{x: x, y: y, w: max(0, w), h: max(0, h)})
	return collectLayoutBorders(layout)
}

func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit {
	layout := newLayout(&Node{Flex: flex}, layoutRect{x: x, y: y, w: max(0, w), h: max(0, h)})
	return collectLayoutBorders(layout)
}
