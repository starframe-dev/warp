# render.go — Warp rendering engine

This file implements the internal rendering helpers for the `warp` package. It
turns a laid-out `Node` tree into terminal lines and defines border hit-test
entry points. Layout geometry is created by `newLayout`; this file renders that
geometry rather than allocating child rectangles itself.

## Exported type

```go
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int
    Length    int
    Bounds    Bounds
    FlexIndex int
}
```

`BorderHit` describes a draggable border and the layout rectangle that owns it.
`Split` or `Flex` identifies the associated configuration; `Direction`, `X`,
`Y`, and `Length` describe the border, while `Bounds` and `FlexIndex` carry its
owning bounds and flex-item index. The type definition does not itself enforce
that exactly one of `Split` and `Flex` is non-nil.

## Rendering

### `renderNode`

```go
func renderNode(node *Node, w, h int) []string
```

Builds a layout rectangle with non-negative width and height using `newLayout`,
then renders it. A nil layout or a layout with non-positive height produces no
lines. A blank layout produces blank lines; a leaf with a panel calls
`Panel.View(w, h)` and pads its result. A leaf with non-positive width or a nil
panel renders blank lines. Split and flex layouts are dispatched according to
their layout node configuration.

The output is composed from the children and borders as allocated by the
layout. Do not rely on every result having exactly the requested width or
height: several rendering branches return or combine child lines directly.

### Split and flex layouts

Split layouts render their first two children. For vertical splits, each row
joins the child lines, with a vertical border only when the layout has a visible
border. When configured and the collapse row is non-negative, the collapse
indicator can replace that border on its row. For horizontal splits, the
children's line slices are joined with a horizontal border when visible.

Flex layouts render all layout children. In the horizontal direction, child
lines are joined row-by-row, adding vertical borders at visible flex boundaries.
In the vertical direction, child line slices are appended with horizontal
borders at visible boundaries. Border styling reflects the configuration's
dragging state. Blank-line and border strings are cached within each render
context.

`renderVerticalSplit`, `renderHorizontalSplit`, and `renderFlex` are convenience
wrappers that create a `Node` and call `renderNode`. `renderFlexRow` and
`renderFlexColumn` currently delegate to `renderFlex`; their final `[]int`
argument is unused.

### `computeSplitSizes`

```go
func computeSplitSizes(avail int, fraction float64, firstCollapsed, secondCollapsed bool, firstSize, secondSize int) (first, second int)
```

Clamps available size to zero or greater. A collapsed child's requested size is
clamped to at least one and at most the available space; if the other child is
expanded and there is sufficient room, the expanded child retains at least
`MinPanelSize`. With neither child collapsed, NaN fractions become `0.5`, and
fractions are clamped to `[0, 1]`. The first size is the truncated product of
available size and fraction, clamped to keep both children at least
`MinPanelSize` when there is room for both; otherwise it is clamped to the
available range. The second size is the remainder.

### `computeFlexSizes`

```go
func computeFlexSizes(avail int, items []*FlexItem) []int
```

Returns nil for no items and otherwise clamps available space to zero or greater.
Each collapsed item has a basis of one; other items use `Basis`, falling back
to `MinPanelSize` when the basis is non-positive. If the sum of bases exceeds
available space, the available size is allocated proportionally to the bases,
using floors for all but the final item, which receives the remainder. If the
bases fit, their integer sizes are retained and remaining space is distributed
among non-collapsed items proportionally to positive `Grow` values. If no
eligible item has positive growth, remaining space is divided equally among
non-collapsed items. Fractional shares are floored except for the final
eligible item, which receives the remainder. No growth is distributed when
available size is zero or all items are collapsed.

## Panel content and terminal text

### `padContent`

```go
func padContent(content string, w, h int) []string
```

For positive width and height, splits panel output on newline, keeps at most
`h` rows, truncates each non-empty row to the visual width `w`, pads it to that
width, and appends `ansi.ResetStyle`. Missing or empty rows are filled with
spaces followed by a style reset. For non-positive width or height it delegates
to `makeEmptyLines`, which returns nil in those cases.

Terminal fragments are normalized to valid UTF-8. Newlines, carriage returns,
and tabs are converted to spaces; other control characters are removed while
ESC and BEL are retained for ANSI parsing. Incomplete trailing ANSI sequences
are removed before fragments are composed with framework content. Framework
labels with invalid UTF-8 or control characters are stripped of ANSI styling
after normalization.

## Border hit-testing

```go
func findBorders(node *Node, x, y, w, h int) []BorderHit
func findFlexBorders(flex *FlexConfig, x, y, w, h int) []BorderHit
```

Both entry points create a layout using the supplied origin and non-negative
width and height, then return the border hits collected from that layout.

## Other helpers

`renderBlankLines` creates `h` rows of spaces of width `w` when height is
positive; `makeEmptyLines` returns nil if either dimension is non-positive.
`emptyView` returns an empty string for non-positive height, otherwise a string
containing `height - 1` newline characters. Border renderers add ANSI reset
sequences around styled vertical or horizontal glyphs; horizontal borders are
empty for non-positive width.
