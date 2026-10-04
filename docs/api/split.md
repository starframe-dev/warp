# split.go

The `split.go` file in package `warp` defines data structures for a hierarchical panel layout tree and provides tree-query and mutation helpers. It does not implement layout calculation, rendering, or hit testing.

## Public API

### `Direction`

```go
type Direction int

const (
    Vertical Direction = iota
    Horizontal
)
```

`Direction` identifies an orientation. `Vertical` represents side-by-side (left/right) arrangement, and `Horizontal` represents top/bottom arrangement.

### `ResizeMsg`

```go
type ResizeMsg struct {
    Width  int
    Height int
}
```

A resize message carrying a panel's allocated content size in cells, excluding borders and padding. The type itself does not send messages or define command behavior.

### `MinPanelSize`

```go
const MinPanelSize = 3
```

Minimum size in cells when enough space is available. This file declares the constant; it does not enforce sizing.

### `SplitConfig`

```go
type SplitConfig struct {
    Direction   Direction
    Fraction    float64
    First       *Node
    Second      *Node
    Dragging    bool
    CollapseRow int
    OnCollapse  func() tea.Cmd
}
```

An internal node description with two child pointers. `Fraction` is documented in code as the share of the first child in the range 0.0–1.0; this file does not clamp or otherwise apply it. `Dragging` records drag-and-drop state. `CollapseRow` is the zero-indexed row where the border shows `"<"`; `OnCollapse` is called when that marker is clicked and returns a `tea.Cmd` to trigger a re-render.

### `NodeCollapse`

```go
type NodeCollapse struct {
    Active bool
    Width  int
    Height int
    Saved  float64
}
```

Stores collapse state and dimensions: `Width` is used for vertical layouts, `Height` for horizontal layouts, and `Saved` stores a fraction to restore on expansion. This type alone does not perform collapse or expansion.

### `Node`

```go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}
```

A tree node can hold a panel and/or split or flex configuration; the struct does not enforce that exactly one of these is set. `Collapse` stores optional collapse state.

### Methods on `Node`

| Signature | Description |
|---|---|
| `func (n *Node) IsLeaf() bool` | True when `n` is non-nil, its `Panel` is not nil (including typed-nil detection via `isNilPanel`), and both `Split` and `Flex` are nil. |
| `func (n *Node) IsCollapsed() bool` | True when `n` is non-nil, `Collapse` is non-nil, and `Collapse.Active` is true. |
| `func (n *Node) CollapsedSize(d Direction) int` | Returns 0 for a nil node, missing/inactive collapse state, or otherwise the active collapsed width for `Vertical` or height for other directions, falling back to 1 when the selected dimension is not positive. |

### `FlexItem` and `FlexConfig`

```go
type FlexItem struct {
    Node      *Node
    Grow      int
    Shrink    int
    Basis     int
    Collapsed bool
}

type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}
```

These structs store flex-layout configuration. `Shrink` is marked unused for now; this file does not implement weighted size distribution or use the `Collapsed` flag.

## Tree helpers

The following unexported methods operate on the tree:

- `findNode` returns the leaf containing the given panel, or nil if the receiver or panel is nil or no matching leaf is found. It traverses split children in `First`, then `Second` order, and flex items in slice order; nil flex items are skipped.
- `findSplitParent` returns the split node with a direct leaf child matching the given panel, or searches recursively and returns nil if none is found. It also traverses flex items in slice order and skips nil items.
- `replaceNode` replaces a child pointer equal to `old` with `new` and returns true, or recursively searches descendants and returns false if not found. Nil receivers return false. Nil flex items are skipped.
- `collectLeafNodes` returns leaf nodes in traversal order: split `First` then `Second`, followed by flex items in slice order. A nil receiver returns nil; nil flex items are skipped.

`IsLeaf` defines a leaf as a node with a non-nil panel and no split or flex configuration. The helpers do not validate the overall tree shape or prevent cycles.
