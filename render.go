package warp

import (
    "strings"

    "github.com/charmbracelet/x/ansi"
)

// renderNode renders a node tree into lines of the given dimensions.
func renderNode(node *Node, w, h int) []string {
    if node == nil {
        return makeEmptyLines(w, h)
    }
    if node.IsLeaf() {
        content := node.Panel.View(w, h)
        return padContent(content, w, h)
    }

    if node.Split != nil {
        switch node.Split.Direction {
        case Vertical:
            return renderVerticalSplit(node.Split, w, h)
        case Horizontal:
            return renderHorizontalSplit(node.Split, w, h)
        }
    }

    if node.Flex != nil {
        return renderFlex(node.Flex, w, h)
    }

    return makeEmptyLines(w, h)
}

func renderVerticalSplit(split *SplitConfig, w, h int) []string {
    borderW := 1
    availW := w - borderW
    firstW, secondW := computeSplitSizes(availW, split.Fraction, split.First.IsCollapsed(), split.Second.IsCollapsed(), split.First.CollapsedSize(Vertical), split.Second.CollapsedSize(Vertical))

    firstLines := renderNode(split.First, firstW, h)
    secondLines := renderNode(split.Second, secondW, h)

    // When one side is collapsed, omit the border so the collapsed panel
    // sits flush against the expanded panel.
    noBorder := split.First.IsCollapsed() || split.Second.IsCollapsed()

    borderChar := borderStyle.Render("│")
    if split.Dragging {
        borderChar = borderDragStyle.Render("│")
    }
    // Isolate the border from panel styles on either side.
    borderChar = ansi.ResetStyle + borderChar + ansi.ResetStyle

    // Collapse symbol ("<") shown at CollapseRow instead of "│".
    // Only shown when OnCollapse is set (explicitly configured).
    collapseChar := ""
    if split.OnCollapse != nil && split.CollapseRow >= 0 && !noBorder {
        collapseChar = ansi.ResetStyle + collapseStyle.Render("<") + ansi.ResetStyle
    }

    result := make([]string, h)
    for y := 0; y < h; y++ {
        left := ""
        right := ""
        if y < len(firstLines) {
            left = firstLines[y]
        }
        if y < len(secondLines) {
            right = secondLines[y]
        }
        if noBorder {
            result[y] = left + right
        } else if collapseChar != "" && y == split.CollapseRow {
            result[y] = left + collapseChar + right
        } else {
            result[y] = left + borderChar + right
        }
    }
    return result
}

func renderHorizontalSplit(split *SplitConfig, w, h int) []string {
    borderH := 1
    availH := h - borderH
    firstH, secondH := computeSplitSizes(availH, split.Fraction, split.First.IsCollapsed(), split.Second.IsCollapsed(), split.First.CollapsedSize(Horizontal), split.Second.CollapsedSize(Horizontal))

    firstLines := renderNode(split.First, w, firstH)
    secondLines := renderNode(split.Second, w, secondH)

    // When one side is collapsed, omit the border.
    noBorder := split.First.IsCollapsed() || split.Second.IsCollapsed()

    borderLine := borderStyle.Render(strings.Repeat("─", w))
    if split.Dragging {
        borderLine = borderDragStyle.Render(strings.Repeat("─", w))
    }
    // Isolate the border from panel styles above/below.
    borderLine = ansi.ResetStyle + borderLine + ansi.ResetStyle

    result := make([]string, 0, h)
    result = append(result, firstLines...)
    if !noBorder {
        result = append(result, borderLine)
    }
    result = append(result, secondLines...)
    return result
}

// renderFlex renders a flex layout with weighted children.
func renderFlex(flex *FlexConfig, w, h int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1

    switch flex.Direction {
    case Horizontal: // Row
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        return renderFlexRow(flex, w, h, sizes)
    case Vertical: // Column
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        return renderFlexColumn(flex, w, h, sizes)
    }

    return makeEmptyLines(w, h)
}

func computeFlexSizes(avail int, items []*FlexItem) []int {
    n := len(items)
    if n == 0 {
        return nil
    }
    sizes := make([]int, n)

    // Count collapsed items and non-collapsed grow
    collapsedCount := 0
    totalGrow := 0
    for _, item := range items {
        if item.Collapsed {
            collapsedCount++
        } else {
            totalGrow += item.Grow
        }
    }

    // First pass: allocate basis (collapsed = 1)
    totalBasis := 0
    for i, item := range items {
        basis := 1
        if !item.Collapsed {
            basis = item.Basis
            if basis <= 0 {
                basis = MinPanelSize
            }
        }
        sizes[i] = basis
        totalBasis += basis
    }

    remaining := avail - totalBasis
    if remaining <= 0 {
        return sizes
    }

    // Second pass: distribute remaining space by Grow weights (collapsed get nothing)
    if totalGrow == 0 {
        // Equal distribution among non-collapsed
        nonCollapsed := n - collapsedCount
        if nonCollapsed > 0 {
            perItem := remaining / nonCollapsed
            for i, item := range items {
                if !item.Collapsed {
                    sizes[i] += perItem
                }
            }
        }
        return sizes
    }

    distributed := 0
    for i, item := range items {
        if item.Collapsed {
            continue
        }
        extra := remaining * item.Grow / totalGrow
        sizes[i] += extra
        distributed += extra
    }
    // Distribute leftover pixels to the last non-collapsed item
    leftover := remaining - distributed
    if leftover > 0 {
        for i := n - 1; i >= 0; i-- {
            if !items[i].Collapsed {
                sizes[i] += leftover
                break
            }
        }
    }

    return sizes
}

func renderFlexRow(flex *FlexConfig, w, h int, sizes []int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderChar := borderStyle.Render("│")
    if flex.Dragging {
        borderChar = borderDragStyle.Render("│")
    }
    // Isolate the border from panel styles on either side.
    borderChar = ansi.ResetStyle + borderChar + ansi.ResetStyle

    // Render each item
    itemLines := make([][]string, len(flex.Items))
    for i, item := range flex.Items {
        itemLines[i] = renderNode(item.Node, sizes[i], h)
    }

    result := make([]string, h)
    for y := 0; y < h; y++ {
        var buf strings.Builder
        for i := range flex.Items {
            if i > 0 && !flex.Items[i-1].Collapsed && !flex.Items[i].Collapsed {
                buf.WriteString(borderChar)
            }
            line := ""
            if y < len(itemLines[i]) {
                line = itemLines[i][y]
            }
            buf.WriteString(line)
        }
        result[y] = buf.String()
    }
    return result
}

func renderFlexColumn(flex *FlexConfig, w, h int, sizes []int) []string {
    if len(flex.Items) == 0 {
        return makeEmptyLines(w, h)
    }

    borderLine := borderStyle.Render(strings.Repeat("─", w))
    if flex.Dragging {
        borderLine = borderDragStyle.Render(strings.Repeat("─", w))
    }
    // Isolate the border from panel styles above/below.
    borderLine = ansi.ResetStyle + borderLine + ansi.ResetStyle

    // Render each item
    itemLines := make([][]string, len(flex.Items))
    for i, item := range flex.Items {
        itemLines[i] = renderNode(item.Node, w, sizes[i])
    }

    result := make([]string, 0, h)
    for i := range flex.Items {
        if i > 0 && !flex.Items[i-1].Collapsed && !flex.Items[i].Collapsed {
            result = append(result, borderLine)
        }
        result = append(result, itemLines[i]...)
    }
    return result
}

// padContent ensures content has exactly w×h dimensions.
// Truncates by visual width (not bytes) to avoid breaking UTF-8 or ANSI
// sequences. The width is measured with ansi.StringWidth to match Bubble Tea's
// renderer and avoid a mismatch with lipgloss.Width on true-color SGR.
// Each returned line ends with an ANSI reset so styles from one panel do not
// leak into adjacent panels or borders.
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
        line = ansi.Truncate(line, w, "")
        lineW := ansi.StringWidth(line)
        if lineW < w {
            line += strings.Repeat(" ", w-lineW)
        }
        // Reset styles at end of line so they cannot leak into neighbors.
        result[y] = line + ansi.ResetStyle
    }
    return result
}

// computeSplitSizes calculates first/second sizes for a split.
// If a child is collapsed, it uses its fixed collapsed size.
func computeSplitSizes(avail int, fraction float64, firstCollapsed, secondCollapsed bool, firstSize, secondSize int) (first, second int) {
    if firstCollapsed {
        first = firstSize
        if first <= 0 {
            first = 1
        }
        second = avail - first
        if second < MinPanelSize {
            second = MinPanelSize
            first = avail - second
        }
        return first, second
    }
    if secondCollapsed {
        second = secondSize
        if second <= 0 {
            second = 1
        }
        first = avail - second
        if first < MinPanelSize {
            first = MinPanelSize
            second = avail - first
        }
        return first, second
    }
    first = int(float64(avail) * fraction)
    if first < MinPanelSize {
        first = MinPanelSize
    }
    second = avail - first
    if second < MinPanelSize {
        second = MinPanelSize
        first = avail - second
    }
    return first, second
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

// BorderHit describes a draggable border at a given position.
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int // Start position of the border
    Length    int // Length of the border in cells
}

// findBorders recursively collects all border positions from the node tree.
func findBorders(node *Node, x, y, w, h int) []BorderHit {
    if node == nil || node.IsLeaf() {
        return nil
    }

    var borders []BorderHit

    if node.Split != nil {
        split := node.Split
        switch split.Direction {
        case Vertical:
            borderW := 1
            availW := w - borderW
            firstW, secondW := computeSplitSizes(availW, split.Fraction, split.First.IsCollapsed(), split.Second.IsCollapsed(), split.First.CollapsedSize(Vertical), split.Second.CollapsedSize(Vertical))
            borderX := x + firstW
            // Omit the border when one side is collapsed.
            if !split.First.IsCollapsed() && !split.Second.IsCollapsed() {
                borders = append(borders, BorderHit{
                    Split:     split,
                    Direction: Vertical,
                    X:         borderX,
                    Y:         y,
                    Length:    h,
                })
            }
            borders = append(borders, findBorders(split.First, x, y, firstW, h)...)
            borders = append(borders, findBorders(split.Second, borderX+borderW, y, secondW, h)...)

        case Horizontal:
            borderH := 1
            availH := h - borderH
            firstH, secondH := computeSplitSizes(availH, split.Fraction, split.First.IsCollapsed(), split.Second.IsCollapsed(), split.First.CollapsedSize(Horizontal), split.Second.CollapsedSize(Horizontal))
            borderY := y + firstH
            // Omit the border when one side is collapsed.
            if !split.First.IsCollapsed() && !split.Second.IsCollapsed() {
                borders = append(borders, BorderHit{
                    Split:     split,
                    Direction: Horizontal,
                    X:         x,
                    Y:         borderY,
                    Length:    w,
                })
            }
            borders = append(borders, findBorders(split.First, x, y, w, firstH)...)
            borders = append(borders, findBorders(split.Second, x, borderY+borderH, w, secondH)...)
        }
    }

    if node.Flex != nil {
        borders = append(borders, findFlexBorders(node.Flex, x, y, w, h)...)
    }

    return borders
}

func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit {
    if len(flex.Items) == 0 {
        return nil
    }

    borderSize := 1
    numBorders := len(flex.Items) - 1
    var borders []BorderHit

    switch flex.Direction {
    case Horizontal: // Row
        availW := w - numBorders*borderSize
        sizes := computeFlexSizes(availW, flex.Items)
        cx := x
        for i := 0; i < len(flex.Items)-1; i++ {
            cx += sizes[i]
            // Omit border when either adjacent item is collapsed.
            if !flex.Items[i].Collapsed && !flex.Items[i+1].Collapsed {
                borders = append(borders, BorderHit{
                    Flex:      flex,
                    Direction: Vertical,
                    X:         cx,
                    Y:         y,
                    Length:    h,
                })
            }
            borders = append(borders, findBorders(flex.Items[i].Node, cx-sizes[i], y, sizes[i], h)...)
            cx += borderSize
        }
        // Last item
        lastIdx := len(flex.Items) - 1
        borders = append(borders, findBorders(flex.Items[lastIdx].Node, cx-sizes[lastIdx], y, sizes[lastIdx], h)...)

    case Vertical: // Column
        availH := h - numBorders*borderSize
        sizes := computeFlexSizes(availH, flex.Items)
        cy := y
        for i := 0; i < len(flex.Items)-1; i++ {
            cy += sizes[i]
            // Omit border when either adjacent item is collapsed.
            if !flex.Items[i].Collapsed && !flex.Items[i+1].Collapsed {
                borders = append(borders, BorderHit{
                    Flex:      flex,
                    Direction: Horizontal,
                    X:         x,
                    Y:         cy,
                    Length:    w,
                })
            }
            borders = append(borders, findBorders(flex.Items[i].Node, x, cy-sizes[i], w, sizes[i])...)
            cy += borderSize
        }
        // Last item
        lastIdx := len(flex.Items) - 1
        borders = append(borders, findBorders(flex.Items[lastIdx].Node, x, cy-sizes[lastIdx], w, sizes[lastIdx])...)
    }

    return borders
}
