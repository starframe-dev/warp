---
title: Layouts
description: Build Warp layouts with splits, flex containers, floats, nested tab groups, collapsible panels, and scrollable panels.
---

# Layouts

Warp layouts start from a panel and grow into a tree.

## Splits

Use splits when you want two regions separated by a border:

```go
warp.SplitVertical(parent, fraction, panel)
warp.SplitHorizontal(parent, fraction, panel)
```

The `fraction` controls how much space the first side receives. Users can drag split borders with the mouse.

## Flex

Use flex layouts when several children share one row or column:

```go
warp.FlexRow(parent, items)
warp.FlexColumn(parent, items)
```

Each item uses a `Grow` weight. Larger weights receive more space.

## Floats

Use floats for panels that should overlay the main layout:

```go
warp.Float(panel, x, y, w, h)
```

Floating panels can be dragged, resized, and closed. Clicking a float brings it to the front.

## Nested layouts

`TabGroup` can act as a `Panel` through `AsPanel()`. This lets you place tab groups inside splits, flex regions, floats, or another Warp instance.

You can also nest Warp inside Warp when an embedded area needs its own layout engine.

## Collapsible panels

Create a collapsible region with:

```go
warp.NewCollapsible(title, panel)
```

Toggle it with:

```go
warp.ToggleCollapsible(panel)
```

## Scrollable panels

Wrap content in a scrollable panel with:

```go
warp.NewScrollable(panel)
```

Scrollable panels support mouse wheel input plus `PgUp` and `PgDn`.
