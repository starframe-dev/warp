# Layout Engine

`layout.go` builds an internal, cell-coordinate layout tree from a panel `Node` tree and a requested screen rectangle. The resulting tree is used to locate panels, collect panel elements, identify draggable borders, and notify panels when their allocated sizes change.

The functions and tree structs declared in this file are package-private. The implementation works with public types declared elsewhere in `warp`, including `Node`, `SplitConfig`, `FlexConfig`, `Direction`, `Bounds`, `Element`, `BorderHit`, and `ResizeMsg`.

## Layout model

The input is a recursive tree of `Node` values:

- A leaf node contains a `Panel` and neither a `Split` nor a `Flex`.
- A split node has a `SplitConfig` with two children and divides its rectangle in one direction.
- A flex node has a `FlexConfig` with an ordered list of items and distributes its rectangle across them.

Every rectangle is expressed in terminal cells. Its origin is the upper-left coordinate `(x, y)`; width and height are non-negative. A border occupies one cell between adjacent visible regions and is not included in either child's assigned rectangle.

`Direction` is defined in `split.go`:

| Value | Geometry |
| --- | --- |
| `Vertical` | Side-by-side, left to right; width is divided. |
| `Horizontal` | Stacked, top to bottom; height is divided. |

## Public types used by this implementation

| Type | Role in layout |
| --- | --- |
| `Node` | Panel leaf or internal split/flex node. `IsLeaf`, `IsCollapsed`, and `CollapsedSize` determine how it is treated. |
| `SplitConfig` | Defines split direction, fraction, first and second children, and drag/collapse interaction state. |
| `FlexConfig` | Defines the flex direction and ordered `Items`. |
| `FlexItem` | Supplies a node, `Grow` weight, `Basis`, and collapsed state. `Shrink` is currently unused by flex size allocation. |
| `Bounds` | Public rectangle with `X`, `Y`, `W`, and `H` cell coordinates. |
| `BorderHit` | Identifies a split or flex border, its position and length, owning bounds, and flex item index when applicable. |
| `Element` | Panel-provided UI element with bounds; layout offsets its local bounds into screen coordinates. |
| `ResizeMsg` | Carries the leaf panel's allocated width and height. |

`MinPanelSize` is defined as 3 cells in `split.go` and is used by the split and flex sizing helpers called here.

## Internal data structures

- `layoutRect` stores an internal rectangle (`x`, `y`, `w`, `h`) and converts to public `Bounds` using `toBounds`.
- `layoutNode` stores the corresponding input node, its allocated rectangle, its laid-out child nodes, and borders owned at that level.

These types remain internal; callers typically reach layout behavior through higher-level rendering, hit-testing, and tab operations.

## Layout construction

`newLayout(node, bounds)` clamps width and height to at least zero, then constructs the layout recursively. A nil node or leaf ends recursion immediately. For internal nodes, a non-nil split takes precedence; otherwise a non-nil flex is laid out. A node with neither applicable configuration has no children.

### Split layout

`layoutSplit` allocates two children according to `SplitConfig.Direction`:

- For `Vertical`, it divides width and places the second child to the right.
- For `Horizontal`, it divides height and places the second child below.

A one-cell border is reserved only if both child nodes are not collapsed and the relevant parent axis has positive size. If either child is collapsed, no border is reserved or reported. The remaining axis length is clamped to zero before `computeSplitSizes` is called. That helper receives the configured fraction, collapsed flags, and direction-specific collapsed sizes. The two child layouts are created even when their allocated sizes are zero.

A visible split border records the `SplitConfig`, direction, screen position, length along the cross-axis, parent bounds, and `FlexIndex: -1`.

### Flex layout

`layoutFlex` returns without children when there are no items. It chooses width for a `Vertical` flex and height otherwise. The number of visible separators is counted only between adjacent non-collapsed items, and is limited by the available axis size. That count is subtracted from the axis size before `computeFlexSizes` assigns item sizes.

Items are processed in order. Each child receives the parent's cross-axis dimensions and its allocated size on the flex axis. A border is placed before an item only when it and its immediately preceding item are not collapsed and a border can fit within the axis. Vertical flex separators are horizontal borders; horizontal flex separators are vertical borders. Each flex border records its `FlexConfig`, owner bounds, position and length, and `FlexIndex` of the item preceding the separator. Nil flex items count as collapsed and produce a child layout with a nil node.

The flex sizing helper (in `render.go`) uses each item's basis, treating a non-positive basis as `MinPanelSize` and a collapsed item's basis as 1. If bases exceed available space, the bases are proportionally reduced. Otherwise remaining space is distributed among non-collapsed items according to positive `Grow` weights; when no item has a positive grow weight, it is distributed uniformly. `Shrink` is not used. Integer rounding is handled by the sizing helper.

## Queries and output traversal

### Borders

`collectLayoutBorders` gathers borders in depth-first pre-order: borders belonging to a layout node come before borders from its children. `appendLayoutBorders` tolerates nil layouts. Split and flex border records can be distinguished by whether `BorderHit.Split` or `BorderHit.Flex` is populated.

### Flex lookup

`findFlexLayout` recursively finds the first layout whose node points to the requested `FlexConfig` by identity. It returns nil if the layout is nil or no matching configuration exists.

### Panel hit-testing

`findLayoutPanel(layout, x, y)` first checks whether the coordinate lies within the current rectangle, using half-open bounds: left/top edges are included and right/bottom edges excluded. If the matching node is a leaf, it returns a `panelHit` containing that node and its allocated rectangle. Otherwise it searches children in order and returns the first hit. Nil layouts, points outside the layout, and points in unassigned gaps (such as borders) return nil.

### Elements

`elementsFromLayout` walks the tree and returns elements only for leaf nodes. For each leaf, `elementsAtLayout` calls `collectElements` with the panel and its local allocated width and height, clones the returned elements, and translates each element's bounds by the leaf's screen origin. It also shifts nested child elements. The returned geometry is therefore screen-relative while element providers receive panel-local dimensions. Nil layouts and nodes yield no elements.

### Resize notifications

`(*Tab).broadcastLayoutResize` traverses the layout and accumulates non-nil Bubble Tea `tea.Cmd` values returned by panel updates. `appendLayoutResize` sends one `ResizeMsg{Width, Height}` to each non-nil panel on a leaf, using that leaf's allocated content dimensions. Internal nodes recurse through children; nil layouts and nodes are ignored. The method does not emit resize messages for internal nodes or borders.

## Implementation notes

- Geometry and hit tests use terminal-cell integers, not pixels.
- Negative requested dimensions are normalized to zero at layout creation. Child widths and heights are likewise protected from negative sizes where they are assigned.
- Borders have their own hit geometry but consume a cell outside both adjacent child rectangles.
- Collapsed state affects split border visibility and split sizing; flex item collapse affects separator visibility and the item's flex basis.
- These helpers do not mutate the input panel tree. They create a separate layout tree that snapshots the rectangles used for subsequent queries and resize messages.
- The file imports Bubble Tea only to use `tea.Cmd` for returned panel update commands.
