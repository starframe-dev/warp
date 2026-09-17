# split.go

The `split` module (package `warp`) defines the structural types for a
hierarchical panel layout tree. Nodes are either leaf panels or internal
split/flex containers. It carries the data model of the layout but not
its rendering or hit-testing logic.

## Public API

### `Direction`

``` go
type Direction int

const (
    Vertical   Direction = iota // side by side (left/right)
    Horizontal                  // stacked (top/bottom)
)
```

Integer type (`Direction`) selecting the split orientation of a
`SplitConfig` or `FlexConfig`. `Vertical` lays children side by side,
`Horizontal` stacks them top to bottom. It is stored directly as a
`Direction` value in the config structs, not as a pointer.

### `ResizeMsg`

``` go
type ResizeMsg struct {
    Width  int
    Height int
}
```

Message type the layout engine sends to a panel when its allocated
rectangle changes. `Width` and `Height` are measured in cells, without
borders or padding. It is a bubbletea command (`tea.Cmd`), so it
participates in the per-frame message loop of the running app.

### `SplitConfig`

``` go
type SplitConfig struct {
    Direction   Direction
    Fraction    float64   // share of First (0.0..1.0)
    First       *Node
    Second      *Node
    Dragging   bool       // true during drag-and-drop
    CollapseRow  int     // border row showing "<"
    OnCollapse   func() tea.Cmd
}
```

An internal node that divides its area between two children. `Fraction`
is the relative share of the *first* child, clamped to \[0.0, 1.0\].
`CollapseRow` is the 0-indexed row where the border renders a collapse
handle (“\|” or “\<”); clicking it invokes `OnCollapse`, which returns a
`tea.Cmd` to re-render.

### `NodeCollapse`

``` go
type NodeCollapse struct {
    Active  bool
    Width   int  // fixed width when collapsed (vertical layouts)
    Height  int  // fixed height when collapsed (horizontal layouts)
    Saved   float64 // fraction to restore on expand
}
```

Holds the collapsed state of a node. When `Active` is true the node
renders at its fixed `Width`/`Height`; `Saved` preserves the
pre-collapse fraction for later restoration.

### `Node`

``` go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}
```

Generic panel-tree node. Exactly one of `Panel`, `Split`, `Flex` is set:
a leaf has a non-nil `Panel` and nil internal config; an internal node
has one of `Split`/`Flex` and a nil `Panel`. `Collapse` is present on
any node that supports collapse/expand.

### Methods on `Node`

| Signature | Description |
|----|----|
| `func (n *Node) IsLeaf() bool` | Returns `true` when `n.Panel` is non-nil (terminal leaf node). |
| `func (n *Node) IsCollapsed() bool` | Returns `true` when the node is currently in a collapsed state (`Collapse.Active == true`). |
| `func (n *Node) CollapsedSize(d Direction) int` | Returns the collapsed size along direction `d`, or 0 if the node is not collapsed. |

Both `IsLeaf` and `IsCollapsed` are cheap field checks; they are the
primary entry points for tree walks. `CollapsedSize` returns the stored
dimension (or 1 as a minimum) so that the layout engine can query a
uniform size regardless of whether the node is currently collapsed.

### `FlexItem` / `FlexConfig`

``` go
type FlexItem struct {
    Node      *Node
    Grow      int  // flex-grow weight
    Shrink    int  // flex-shrink (currently unused)
    Basis     int  // flex-basis (min size); 0 = auto
    Collapsed bool
}

type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}
```

Weighted row/column layout. Each item wraps a `Node` and carries its
grow/shrink/basis weights plus a local `Collapsed` flag. `FlexConfig`
lays its children out in one row or column with the weights distributed
proportionally. `Shrink` is declared but intentionally unused (the
layout does not implement shrinking in the current release).

## Behavior

The tree is a *binary* or *multi-child* tree depending on which internal
node is in use: a `SplitConfig` always has exactly two children; a
`FlexConfig` has an ordered list of `FlexItem`s (each wrapping a
`Node`). A `Node` is either a leaf (`Panel != nil`) or an internal
container (`Split != nil` or `Flex != nil`), never both at once.

Direction-dependent sizing: `SplitConfig` and `FlexConfig` use a single
`Direction` to pick the split axis; collapsed dimensions are queried
through `CollapsedSize(d)` so the caller can request either axis
uniformly.

Drag state is tracked per-config via `Dragging bool`; setting it to
`true` pauses layout re-computation until the next resize message
arrives.

## Implementation details

- `findNode`, `findSplitParent`, `replaceNode` and `collectLeafNodes`
  are recursive walkers defined as methods on `*Node`; they handle both
  `SplitConfig` and `FlexConfig` children.
- `findNode` returns the leaf node that matches the given `Panel`;
  `findSplitParent` returns the parent `Split` node whose two children
  are the direct parents of the target panel; `replaceNode` rewrites a
  pointer in place and returns a success flag; `collectLeafNodes`
  returns every leaf `*Node` in in-order (split left-to-right / flex
  top-to-bottom) traversal order.
- All walkers guard against a nil receiver so they can be called on a
  nil `*Node` safely.
