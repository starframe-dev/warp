# Specification: `split.go`

File: `warp/split.go`  
Package: `warp`  
Language: Go

## Overview

`split.go` defines the core tree data structures used by the `warp` package to model a hierarchical panel layout. It supports binary splits (fixed fraction of available space), flex layouts (weighted distribution), and collapsing nodes to a minimal size. The file also provides tree navigation helpers used to locate, replace, and enumerate leaf nodes.

## Types

### `Direction`

```go
type Direction int
```

Orientation constant for a split or flex layout.

| Constant | Value | Meaning |
|---|---|---|
| `Vertical` | `0` | Splits/layout runs left-to-right (children side by side). |
| `Horizontal` | `1` | Splits/layout runs top-to-bottom (children stacked). |

### `ResizeMsg`

```go
type ResizeMsg struct {
    Width  int
    Height int
}
```

Message sent by the layout engine to a panel when its allocated content rectangle changes. Dimensions are in cells, excluding borders and padding.

### `MinPanelSize`

```go
const MinPanelSize = 3
```

Minimum allowed size in characters for any panel.

### `SplitConfig`

```go
type SplitConfig struct {
    Direction Direction
    Fraction  float64 // 0.0–1.0, share of First child
    First     *Node
    Second    *Node
    Dragging  bool
}
```

Internal binary-split node. The area is divided between `First` and `Second` according to `Fraction`. `Dragging` is set to `true` during drag-and-drop operations (layout-specific behavior, not interpreted by this file).

### `NodeCollapse`

```go
type NodeCollapse struct {
    Active bool
    Width  int
    Height int
    Saved  float64
}
```

Holds the collapsed state of a node. When `Active` is `true`, the node renders with a fixed size: `Width` for vertical layouts and `Height` for horizontal layouts. `Saved` stores the previous layout fraction so it can be restored on expand.

### `Node`

```go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}
```

Tree node that is either a terminal panel (`Panel != nil`) or an internal layout node (`Split != nil` or `Flex != nil`).

#### Methods

- `func (n *Node) IsLeaf() bool`  
  Returns `true` if the node is a terminal panel.

- `func (n *Node) IsCollapsed() bool`  
  Returns `true` if `n.Collapse` is non-nil and `Active` is `true`.

- `func (n *Node) CollapsedSize(d Direction) int`  
  Returns the collapsed size for the given direction, or `0` if the node is not collapsed. If the configured size is not positive, returns `1` as a fallback.

### `FlexItem`

```go
type FlexItem struct {
    Node      *Node
    Grow      int
    Shrink    int
    Basis     int
    Collapsed bool
}
```

Single child inside a flex layout. `Grow` is the weight used to distribute extra space. `Shrink` is reserved but currently unused. `Basis` is the minimum size, where `0` means auto. `Collapsed` reduces the item to one line/character when `true`.

### `FlexConfig`

```go
type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}
```

Flex container that lays out children in a row or column according to `Direction`. Items are distributed based on their `Grow` weights, subject to their `Basis` minimums. `Dragging` is set during drag-and-drop.

## Public API (methods on `Node`)

| Method | Signature | Behavior |
|---|---|---|
| `IsLeaf` | `func (n *Node) IsLeaf() bool` | True when `Panel != nil`. |
| `IsCollapsed` | `func (n *Node) IsCollapsed() bool` | True when collapse is active. |
| `CollapsedSize` | `func (n *Node) CollapsedSize(d Direction) int` | Returns the effective collapsed size for a direction, or `0` if not collapsed. |
| `findNode` | `func (n *Node) findNode(panel Panel) *Node` | Recursively searches the tree for the node containing the given `Panel`. Returns `nil` if not found. |
| `replaceNode` | `func (n *Node) replaceNode(old, new *Node) bool` | Replaces `old` with `new` in the tree. Returns `true` if the replacement occurred. |
| `collectLeafNodes` | `func (n *Node) collectLeafNodes() []*Node` | Returns all leaf nodes in order (left-to-right or top-to-bottom). |

## Important Implementation Details

- A node is considered a leaf only by `n.Panel != nil`. The other fields (`Split`, `Flex`, `Collapse`) do not affect leaf status.
- `CollapsedSize` falls back to `1` when the configured width/height is zero or negative, ensuring collapsed nodes always consume at least one cell in the appropriate direction.
- `findNode` and `replaceNode` traverse `Split` first, then `Flex`, and finally return `nil`/`false` if the target is not found.
- `collectLeafNodes` preserves the natural order of children: `First` then `Second` for splits, and item order for flex layouts.
- `Shrink` in `FlexItem` is documented but not used by the layout logic in this file.
- `Dragging` fields on `SplitConfig` and `FlexConfig` are boolean markers; their exact semantics are defined by the surrounding layout/rendering code.
