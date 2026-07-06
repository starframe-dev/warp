# Specification: `render.go`

## Overview

`render.go` is responsible for rendering the `warp` layout tree into fixed-size rectangular strings of terminal cells. It supports nested split panes, flex layouts, and individual leaf panels. The module handles border placement, collapse semantics, ANSI style isolation, and hit-testing of draggable borders.

The rendering entry point is `renderNode`. It takes a layout `Node`, a width, and a height, and returns a slice of `h` lines, each of which is exactly `w` display cells wide.

## Package

- **Package:** `warp`
- **File:** `render.go`
- **Dependencies:**
  - `strings` (standard library)
  - `github.com/charmbracelet/x/ansi` for visual-width truncation, padding, and ANSI reset isolation

## Public API

`render.go` exposes only one public type. The actual rendering functions are intentionally unexported; they are invoked by other parts of the `warp` package.

### Type: `BorderHit`

```go
type BorderHit struct {
    Split     *SplitConfig
    Flex      *FlexConfig
    Direction Direction
    X, Y      int // Start position of the border
    Length    int // Length of the border in cells
}
```

`BorderHit` describes a draggable border within the rendered layout. It is produced by `findBorders` and `findFlexBorders`.

- `Split` is set when the border belongs to a binary split; `Flex` is set when it belongs to a flex container. Only one of the two is expected to be non-nil for a given hit.
- `Direction` indicates whether the border is `Vertical` (column divider) or `Horizontal` (row divider).
- `X` and `Y` are the zero-based cell coordinates of the start of the border.
- `Length` is the number of cells along the border axis (`h` for vertical dividers, `w` for horizontal dividers).

## Behavior

### Rendering pipeline

1. `renderNode(node, w, h)` dispatches based on the node kind:
   - `nil` → empty rectangle of size `w×h`
   - Leaf (`node.IsLeaf()`) → the panel's view is rendered to `w×h` and padded/truncated
   - Binary split (`node.Split != nil`) → delegated to `renderVerticalSplit` or `renderHorizontalSplit`
   - Flex layout (`node.Flex != nil`) → delegated to `renderFlex`

2. Returned lines are always exactly `w` display cells wide and exactly `h` lines tall, unless `w` or `h` are non-positive, in which case `nil` or empty lines are returned.

3. Every panel line is terminated with an ANSI reset sequence to prevent styles from bleeding into adjacent panels or borders.

### Leaf panel rendering (`padContent`)

- The panel's `View(w, h)` string is split on `\n`.
- Each line is visually truncated to `w` using `ansi.Truncate`.
- Short lines are right-padded with spaces based on `ansi.StringWidth`.
- Excess lines beyond `h` are ignored; missing lines are blank.
- A final `ansi.ResetStyle` is appended to every line.
- If `w <= 0` or `h <= 0`, an empty-lines slice is returned.

### Binary splits

A split reserves a single-cell border between its two children.

- `renderVerticalSplit` places the divider as a vertical line (`│`) between left and right children.
- `renderHorizontalSplit` places the divider as a horizontal line (`─` repeated `w` times) between top and bottom children.
- Border rendering uses `borderStyle` normally, and `borderDragStyle` when `split.Dragging` is true.
- Each border cell is wrapped with `ansi.ResetStyle` on both sides to isolate it from panel styles.
- If either child is collapsed, the border is omitted and the children sit flush against each other.

### Flex layouts

A flex container lays out `N` children in a row or column, with `N-1` one-cell borders between them.

- `renderFlex` computes per-child sizes using `computeFlexSizes`, then delegates to `renderFlexRow` or `renderFlexColumn`.
- Child sizing is based on `Basis` (minimum size) and `Grow` (extra-space weight). Collapsed items receive only a basis of `1` and no grow share.
- If no non-collapsed item has a grow value, extra space is divided equally among non-collapsed items.
- Leftover pixels from integer division are given to the last non-collapsed item.
- Borders are omitted between adjacent collapsed items.
- Flex borders use the same style/dragging logic as split borders.

### Collapsed panels

Collapsed children receive their fixed collapsed size (or a minimum of `1`) and the remaining space is allocated to the expanded child. The allocator also ensures the expanded child does not fall below `MinPanelSize`. Both split and flex layouts respect the `Collapsed` flag on nodes/flex items.

### Split sizing (`computeSplitSizes`)

Given available space (total dimension minus one border cell), a fraction, and optional collapse states:

- If the first child is collapsed, it is sized to its collapsed size (at least `1`); the second child takes the rest, clamped to at least `MinPanelSize`.
- If the second child is collapsed, the symmetric rule applies.
- If neither is collapsed, the first child gets `int(float64(avail) * fraction)`, clamped to `MinPanelSize`, and the second child takes the remainder, also clamped to `MinPanelSize`.

### Empty-line generation (`makeEmptyLines`)

Produces `h` lines of `w` spaces. If `w <= 0` or `h <= 0`, returns `nil`.

## Border discovery (hit testing)

`findBorders` and `findFlexBorders` recursively walk the layout tree and record the on-screen position, length, and owning config of every visible border. These hits are used by the package's drag/resize logic.

- For a vertical split, the border is at `x + firstW` and spans `h` cells.
- For a horizontal split, the border is at `y + firstH` and spans `w` cells.
- For flex rows, borders are placed after each item (except the last) at the accumulated width.
- For flex columns, borders are placed after each item (except the last) at the accumulated height.
- Borders adjacent to collapsed panels are omitted from the hit list, matching the render output.
- The functions recurse into every child with the correct sub-rectangle coordinates.

## Important implementation details

- Visual width is used throughout, not byte length, to keep Unicode and ANSI-styled content aligned with the rest of the Bubble Tea/lipgloss rendering stack.
- `ansi.Truncate` and `ansi.StringWidth` are preferred over `lipgloss.Width` to avoid mismatches with true-color SGR sequences.
- All border characters are wrapped with `ansi.ResetStyle` to guarantee that background/foreground styles from a neighboring panel do not extend into the border.
- The render output is deterministic and depends only on the tree state, width, and height.
- Rendering is read-only with respect to the layout tree; it does not mutate `SplitConfig`, `FlexConfig`, or `Node`.
