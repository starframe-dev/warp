---
title: Split & Flex
description: Split and flex layout types
---

# Split & Flex

## Node

```go
type Node struct {
    Panel Panel
    Split *SplitConfig
    Flex  *FlexConfig
}
```

A node is either a leaf panel, a split, or a flex container.

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

## FlexConfig

```go
type FlexConfig struct {
    Direction Direction
    Items     []FlexItem
}
```

## FlexItemSpec

```go
type FlexItemSpec struct {
    Panel Panel
    Grow  int
}
```

## Direction

```go
type Direction int

const (
    VerticalSplit   Direction = iota
    HorizontalSplit
)
```
