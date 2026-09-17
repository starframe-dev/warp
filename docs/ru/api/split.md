---
title: Split и Flex
description: Типы дерева компоновки для split, flex и сворачиваемых узлов.
---

# Split и Flex

Warp хранит компоновку как дерево `Node`. Узел содержит листовую `Panel`, разделение на два дочерних узла или flex-контейнер.

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
    Direction   Direction
    Fraction    float64
    First       *Node
    Second      *Node
    Dragging    bool
    CollapseRow int
    OnCollapse  func() tea.Cmd
}
```

`Fraction` задаёт долю первого дочернего узла и ограничивается диапазоном `0.1..0.9`. `CollapseRow` и `OnCollapse` включают необязательный маркер сворачивания на вертикальной границе.

## FlexConfig и FlexItem

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

`Tab.FlexRow` и `Tab.FlexColumn` преобразуют `FlexItemSpec` в элементы компоновки. Отрицательный `Grow` трактуется как ноль.

## Direction

```go
type Direction int

const (
    Vertical Direction = iota
    Horizontal
)
```

`Vertical` располагает дочерние панели рядом. `Horizontal` располагает их сверху вниз.

## Дополнительные типы

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

`ResizeMsg` передаёт панели выделенный размер в терминальных ячейках. Минимальный размер панели `MinPanelSize` равен `3`.
