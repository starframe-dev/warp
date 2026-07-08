---
title: Split и Flex
description: API дерева Node, split-раскладки и flex-раскладки Warp.
---

# Split и Flex

Warp использует `Node` для описания дерева раскладки.

## Node

`Node` представляет элемент дерева layout: панель, split или flex-контейнер.

## SplitConfig

```go
type SplitConfig struct {
    Direction Direction
    Fraction  float64
    First     *Node
    Second    *Node
    Dragging  bool
}
```

`SplitConfig` описывает деление области на две части.

- `Direction` — направление деления.
- `Fraction` — доля первой области.
- `First` и `Second` — дочерние узлы.
- `Dragging` — состояние перетаскивания разделителя.

## FlexConfig

```go
type FlexConfig struct {
    Direction Direction
    Items     []FlexItem
}
```

`FlexConfig` описывает flex-раскладку с набором элементов.

## FlexItemSpec

```go
type FlexItemSpec struct {
    Panel Panel
    Grow  int
}
```

`FlexItemSpec` задаёт панель и её коэффициент роста.

## Direction

```go
VerticalSplit
HorizontalSplit
```

`VerticalSplit` делит область по вертикали. `HorizontalSplit` делит область по горизонтали.
