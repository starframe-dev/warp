package warp

import (
	"math"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/parser"
	"github.com/rivo/uniseg"
)

type renderBlankKey struct {
	width  int
	height int
}

type renderHorizontalBorderKey struct {
	width    int
	dragging bool
}

type renderContext struct {
	blankLines        map[renderBlankKey][]string
	verticalBorders   map[bool]string
	horizontalBorders map[renderHorizontalBorderKey]string
	collapseBorder    string
}

func renderNode(node *Node, w, h int) []string {
	return renderLayout(newLayout(node, layoutRect{w: max(0, w), h: max(0, h)}))
}

func renderLayout(layout *layoutNode) []string {
	var context renderContext
	return context.renderLayout(layout)
}

func (context *renderContext) renderLayout(layout *layoutNode) []string {
	if layout == nil {
		return nil
	}
	bounds := layout.bounds
	if bounds.h <= 0 {
		return nil
	}
	if layout.node == nil {
		return context.renderBlankLines(bounds.w, bounds.h)
	}
	if layout.node.IsLeaf() {
		if bounds.w <= 0 || layout.node.Panel == nil {
			return context.renderBlankLines(bounds.w, bounds.h)
		}
		return padContent(layout.node.Panel.View(bounds.w, bounds.h), bounds.w, bounds.h)
	}
	if layout.node.Split != nil {
		return context.renderSplitLayout(layout)
	}
	if layout.node.Flex != nil {
		return context.renderFlexLayout(layout)
	}
	return context.renderBlankLines(bounds.w, bounds.h)
}

func (context *renderContext) renderSplitLayout(layout *layoutNode) []string {
	if len(layout.children) < 2 {
		return context.renderBlankLines(layout.bounds.w, layout.bounds.h)
	}
	split := layout.node.Split
	first := context.renderLayout(layout.children[0])
	second := context.renderLayout(layout.children[1])
	borderVisible := len(layout.borders) > 0
	bounds := layout.bounds

	switch split.Direction {
	case Vertical:
		border := context.renderVerticalBorder(split.Dragging)
		collapseBorder := ""
		if split.OnCollapse != nil && split.CollapseRow >= 0 {
			collapseBorder = context.renderCollapseBorder()
		}
		lines := make([]string, bounds.h)
		for y := range lines {
			firstLine := lineAt(first, y)
			secondLine := lineAt(second, y)
			if !borderVisible {
				lines[y] = firstLine + secondLine
				continue
			}
			middle := border
			if collapseBorder != "" && y == split.CollapseRow {
				middle = collapseBorder
			}
			lines[y] = firstLine + middle + secondLine
		}
		return lines
	case Horizontal:
		lines := make([]string, 0, bounds.h)
		lines = append(lines, first...)
		if borderVisible {
			lines = append(lines, context.renderHorizontalBorder(bounds.w, split.Dragging))
		}
		lines = append(lines, second...)
		return lines
	default:
		return context.renderBlankLines(bounds.w, bounds.h)
	}
}

func (context *renderContext) renderFlexLayout(layout *layoutNode) []string {
	flex := layout.node.Flex
	if flex == nil || len(layout.children) == 0 {
		return context.renderBlankLines(layout.bounds.w, layout.bounds.h)
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
			children[i] = context.renderLayout(child)
		}
		lines := make([]string, layout.bounds.h)
		border := context.renderVerticalBorder(flex.Dragging)
		for y := range lines {
			rowBytes := 0
			for i, child := range children {
				if i > 0 && visible[i-1] {
					rowBytes += len(border)
				}
				rowBytes += len(lineAt(child, y))
			}
			var row strings.Builder
			row.Grow(rowBytes)
			for i, child := range children {
				if i > 0 && visible[i-1] {
					row.WriteString(border)
				}
				row.WriteString(lineAt(child, y))
			}
			lines[y] = row.String()
		}
		return lines
	case Vertical:
		lines := make([]string, 0, layout.bounds.h)
		for i, child := range layout.children {
			if i > 0 && visible[i-1] {
				lines = append(lines, context.renderHorizontalBorder(layout.bounds.w, flex.Dragging))
			}
			lines = append(lines, context.renderLayout(child)...)
		}
		return lines
	default:
		return context.renderBlankLines(layout.bounds.w, layout.bounds.h)
	}
}

func lineAt(lines []string, index int) string {
	if index < 0 || index >= len(lines) {
		return ""
	}
	return lines[index]
}

func (context *renderContext) renderBlankLines(w, h int) []string {
	if h <= 0 {
		return nil
	}
	key := renderBlankKey{width: max(0, w), height: h}
	if context.blankLines != nil {
		if lines, ok := context.blankLines[key]; ok {
			return lines
		}
	} else {
		context.blankLines = make(map[renderBlankKey][]string)
	}
	lines := renderBlankLines(key.width, key.height)
	context.blankLines[key] = lines
	return lines
}

func (context *renderContext) renderVerticalBorder(dragging bool) string {
	if context.verticalBorders != nil {
		if border, ok := context.verticalBorders[dragging]; ok {
			return border
		}
	} else {
		context.verticalBorders = make(map[bool]string, 2)
	}
	border := renderVerticalBorder(dragging)
	context.verticalBorders[dragging] = border
	return border
}

func (context *renderContext) renderHorizontalBorder(width int, dragging bool) string {
	key := renderHorizontalBorderKey{width: width, dragging: dragging}
	if context.horizontalBorders != nil {
		if border, ok := context.horizontalBorders[key]; ok {
			return border
		}
	} else {
		context.horizontalBorders = make(map[renderHorizontalBorderKey]string)
	}
	border := renderHorizontalBorder(width, dragging)
	context.horizontalBorders[key] = border
	return border
}

func (context *renderContext) renderCollapseBorder() string {
	if context.collapseBorder == "" {
		context.collapseBorder = ansi.ResetStyle + collapseStyle.Render("<") + ansi.ResetStyle
	}
	return context.collapseBorder
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
	baseTotal := float64(0)
	growTotal := float64(0)
	eligibleCount := 0

	for _, item := range items {
		baseTotal += flexItemBasis(item)
		if flexItemCollapsed(item) {
			continue
		}
		eligibleCount++
		if item.Grow > 0 {
			growTotal += float64(item.Grow)
		}
	}

	if avail == 0 {
		return sizes
	}
	if baseTotal > float64(avail) {
		distributeFlexBases(avail, sizes, items, baseTotal)
		return sizes
	}

	used := 0
	for i, item := range items {
		sizes[i] = int(flexItemBasis(item))
		used += sizes[i]
	}
	remaining := avail - used
	if remaining <= 0 || eligibleCount == 0 {
		return sizes
	}

	distributeFlexGrow(remaining, sizes, items, eligibleCount, growTotal)
	return sizes
}

func flexItemBasis(item *FlexItem) float64 {
	if flexItemCollapsed(item) {
		return 1
	}
	basis := item.Basis
	if basis <= 0 {
		basis = MinPanelSize
	}
	return float64(basis)
}

func distributeFlexBases(amount int, sizes []int, items []*FlexItem, totalWeight float64) {
	if amount <= 0 || len(items) == 0 || totalWeight <= 0 || math.IsInf(totalWeight, 0) || math.IsNaN(totalWeight) {
		return
	}
	remaining := amount
	for index, item := range items {
		weight := flexItemBasis(item)
		share := remaining
		if index < len(items)-1 {
			share = int(math.Floor(float64(amount) * weight / totalWeight))
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

func distributeFlexGrow(amount int, sizes []int, items []*FlexItem, eligibleCount int, growTotal float64) {
	if amount <= 0 || eligibleCount == 0 {
		return
	}
	uniform := growTotal <= 0
	totalWeight := growTotal
	if uniform {
		totalWeight = float64(eligibleCount)
	}
	if totalWeight <= 0 || math.IsInf(totalWeight, 0) || math.IsNaN(totalWeight) {
		return
	}

	remaining := amount
	position := 0
	for index, item := range items {
		if flexItemCollapsed(item) {
			continue
		}
		weight := float64(max(0, item.Grow))
		if uniform {
			weight = 1
		}
		share := remaining
		if position < eligibleCount-1 {
			share = int(math.Floor(float64(amount) * weight / totalWeight))
			if share > remaining {
				share = remaining
			}
			if share < 0 {
				share = 0
			}
		}
		sizes[index] += share
		remaining -= share
		position++
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
	line = truncateTerminalFragment(line, width, "")
	lineWidth := ansi.StringWidth(line)
	if lineWidth < width {
		line += strings.Repeat(" ", width-lineWidth)
	}
	return line
}

func normalizeTerminalText(s string) string {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "�")
	}
	if strings.ContainsFunc(s, unicode.IsControl) {
		s = strings.Map(func(r rune) rune {
			switch r {
			case '\x1b', '\a':
				// Preserve the standard ESC introducer and BEL terminator used by
				// valid ANSI sequences. Incomplete sequences are removed later.
				return r
			case '\t', '\r', '\n':
				return ' '
			default:
				if unicode.IsControl(r) {
					return -1
				}
				return r
			}
		}, s)
	}
	return s
}

func sanitizeFrameworkLabel(s string) string {
	if utf8.ValidString(s) && strings.IndexByte(s, 0x1b) < 0 &&
		!strings.ContainsFunc(s, unicode.IsControl) {
		return s
	}
	return ansi.Strip(normalizeTerminalText(s))
}

func sanitizeTerminalFragment(s string) string {
	s = normalizeTerminalText(s)
	if s == "" || strings.IndexByte(s, 0x1b) < 0 {
		return s
	}
	if start := incompleteANSITailStart(s); start >= 0 {
		// An incomplete escape tail has no stable visual meaning. Drop it before
		// composing framework chrome so terminal/layout parsers cannot consume
		// padding, borders or controls that follow the user fragment.
		return s[:start]
	}
	return s
}

func truncateTerminalFragment(s string, width int, tail string) string {
	return sanitizeTerminalFragment(ansi.Truncate(normalizeTerminalText(s), width, tail))
}

func ansiEndsInGroundState(s string) bool {
	return incompleteANSITailStart(normalizeTerminalText(s)) < 0
}

func incompleteANSITailStart(s string) int {
	state := parser.GroundState
	sequenceStart := -1
	bytes := []byte(s)
	for i := 0; i < len(bytes); {
		next, _ := parser.Table.Transition(state, bytes[i])
		if next == parser.Utf8State {
			cluster, _, _, _ := uniseg.FirstGraphemeCluster(bytes[i:], -1)
			if len(cluster) == 0 {
				if sequenceStart >= 0 {
					return sequenceStart
				}
				return i
			}
			i += len(cluster)
			state = parser.GroundState
			sequenceStart = -1
			continue
		}
		if state == parser.GroundState && next != parser.GroundState {
			sequenceStart = i
		}
		state = next
		i++
		if state == parser.GroundState {
			sequenceStart = -1
		}
	}
	if state == parser.GroundState {
		return -1
	}
	if sequenceStart >= 0 {
		return sequenceStart
	}
	return len(s)
}

func padContent(content string, w, h int) []string {
	if w <= 0 || h <= 0 {
		return makeEmptyLines(w, h)
	}
	lines := strings.Split(content, "\n")
	result := make([]string, h)
	blank := strings.Repeat(" ", w) + ansi.ResetStyle
	for y := 0; y < h; y++ {
		if y >= len(lines) || lines[y] == "" {
			result[y] = blank
			continue
		}
		result[y] = padVisualLine(lines[y], w) + ansi.ResetStyle
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

func emptyView(height int) string {
	if height <= 0 {
		return ""
	}
	return strings.Repeat("\n", height-1)
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
