---
title: Split & Flex
description: Layout tree types for splits, flex containers, and collapsible nodes.
---

# Split & Flex

Warp stores a layout as a tree of `Node` values. A node is either a leaf `Panel`, a two-child split, or a flex container.

## Node

```go
type Node struct {
    Panel    Panel
    Split    *SplitConfig
    Flex     *FlexConfig
    Collapse *NodeCollapse
}

func (n *Node) IsLeaf() bool
func (n *Node) IsCollapsed() bool
func (n *Node) CollapsedSize(direction Direction) int
```

## SplitConfig

```go
type SplitConfig struct {
    Direction  Direction
    Fraction   float64
    First      *Node
    Second     *Node
    Dragging   bool
    CollapseRow int
    OnCollapse func() tea.Cmd
}
```

`Fraction` is the share of the first child and is clamped to `0.1..0.9`. `CollapseRow` and `OnCollapse` configure an optional collapse marker on a vertical border.

## FlexConfig and FlexItem

```go
type FlexConfig struct {
    Direction Direction
    Items     []*FlexItem
    Dragging  bool
}

type FlexItem struct {
    Node      *Node
    Grow      int
    Shrink    int
    Basis     int
    Collapsed bool
}

type FlexItemSpec struct {
    Panel Panel
    Grow  int
}
```

`Tab.FlexRow` and `Tab.FlexColumn` convert `FlexItemSpec` values into layout items. A negative `Grow` is treated as zero.

## Direction

```go
type Direction int

const (
    Vertical Direction = iota
    Horizontal
)
```

`Vertical` places children side by side. `Horizontal` stacks them top to bottom.

## Supporting types

```go
type NodeCollapse struct {
    Active bool
    Width  int
    Height int
    Saved  float64
}

type ResizeMsg struct {
    Width  int
    Height int
}
```

`ResizeMsg` carries the allocated cell size to a panel. `MinPanelSize` is the minimum panel size and equals `3`.
