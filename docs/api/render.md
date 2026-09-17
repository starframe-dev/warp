# render.go — Warp rendering engine

This file contains the recursive rendering logic that converts a `Node`
tree (leaves, splits, and flex containers) into a grid of text cells. It
also implements hit-testing for draggable borders so that mouse/pointer
interaction can find which split or flex border lies under a given cell.

## Public API

None of the functions in this file are exported (all are lowercase). The
only exported type is `BorderHit`. The file is an internal
implementation module of the `warp` package; consumers interact with the
package's public render entry point (defined elsewhere) which drives
these internal helpers.

`type BorderHit struct { Split *SplitConfig Flex *FlexConfig Direction Direction X, Y int // Start position of the border Length int // Length of the border in cells }`

`BorderHit` describes a single draggable border at a concrete position
in the terminal grid. Exactly one of `Split` or `Flex` is non-nil,
identifying which config owns the border. `Direction` tells whether the
border runs vertically (splitting width) or horizontally (splitting
height).

## Core rendering functions

### renderNode

``` go
func renderNode(node *Node, w, h int) []string
```

The central dispatcher. Given a node and its allocated `w` × `h` cell
rectangle it returns exactly `h` strings, each of visual width `w`. The
behaviour is decided by which layout field is set:

- *nil node* — returns blank lines via `makeEmptyLines`.
- *Leaf* (`node.IsLeaf()`) — asks the panel's `View(w, h)` for content
  and normalises it with `padContent`.
- *Vertical / Horizontal split* — delegates to `renderVerticalSplit` /
  `renderHorizontalSplit`.
- *Flex* — delegates to `renderFlex`.

### renderVerticalSplit / renderHorizontalSplit

Both compute child sizes with `computeSplitSizes`, render each child
recursively, then merge the two halves. A one-cell border of `│`
(vertical) or `─` (horizontal) is drawn between them. Key behaviours:

- If either side is *collapsed*, the border is omitted so the collapsed
  panel sits flush against the expanded one.
- When `split.Dragging` is true, the border is styled with
  `borderDragStyle`; otherwise `borderStyle`.
- The border character is wrapped in `ansi.ResetStyle` on both sides so
  panel styles cannot bleed into the border line.
- An optional collapse indicator (`<`) is drawn at `split.CollapseRow`
  only when `OnCollapse` is set and no side is collapsed.

### renderFlex

Dispatches to `renderFlexRow` (Direction=Horizontal, children laid out
left-to-right with vertical borders) or `renderFlexColumn`
(Direction=Vertical, children top-to-bottom with horizontal borders).
Border count is `len(items) − 1`; available space is reduced accordingly
before `computeFlexSizes` distributes it.

### computeFlexSizes

Two-pass flexbox-like algorithm:

1.  Assign each item a *basis* (its `Basis` or `MinPanelSize` if not
    collapsed; 1 if collapsed).
2.  If remaining space after bases is non-positive, return bases as-is.
3.  Distribute remaining space proportionally by `Grow` weights among
    non-collapsed items; if total grow is 0, split equally among
    non-collapsed items.
4.  Assign leftover pixels to the last non-collapsed item to keep total
    width/height exact.

### padContent

``` go
func padContent(content string, w, h int) []string
```

Normalises a panel's raw string to exactly `w` columns × `h` rows.
Truncation uses `ansi.Truncate` (visual-width aware) so multi-byte UTF-8
and ANSI sequences are never split. Every returned line is right-padded
with spaces and suffixed with `ansi.ResetStyle` to prevent style leakage
between adjacent panels.

### computeSplitSizes

Resolves the exact pixel widths/heights for a two-child split: collapsed
children get their fixed `CollapsedSize`; otherwise `fraction` of
`avail` goes to the first child. Both sides are clamped to
`MinPanelSize` before the second is derived.

## Border hit-testing

### findBorders

``` go
func findBorders(node *Node, x, y, w, h int) []BorderHit
```

Recursively walks the node tree and returns the absolute positions of
every visible border. Split borders are reported only when both children
are expanded (collapsed side suppresses the border). Flex borders are
gathered via `findFlexBorders`.

### findFlexBorders

Iterates the flex items, accumulating offsets to compute each inter-item
border's `(X, Y)` and `Length`. A border between items *i* and *i+1* is
omitted if either adjacent item is collapsed.

## Key invariants

- Returned line slices always have length exactly `h` (or are empty/nil
  when the node is nil or the size is ≤ 0).
- Every line is exactly `w` visual columns wide.
- No style can leak from one panel into a neighbour because each line is
  reset with `ansi.ResetStyle`.
- Border visibility is a function of child collapse state; the render
  and hit-test passes share the same size computations, keeping them
  consistent.

## Dependencies

Internal (same package): `Node`, `Panel`, `SplitConfig`, `FlexConfig`,
`FlexItem`, `Direction`, `MinPanelSize`, border style variables, and
`CollapsedSize`. External: `github.com/charmbracelet/x/ansi` for string
width, truncation, and SGR reset.
