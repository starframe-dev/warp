---
title: Layouts
description: Build Warp layouts with splits, flex containers, floats, nested tab groups, collapsible panels, and scrolling.
---

# Layouts

Start with a `Tab` and use its methods to replace panels in the layout tree.

## Splits

Use a split for two regions separated by a draggable border:

```go
tab.SplitVertical(parent, 0.5, rightPanel)
tab.SplitHorizontal(parent, 0.5, bottomPanel)
```

The fraction controls the share of the first child. Warp clamps it to `0.1..0.9`.

## Flex

Use flex when several children share a row or column:

```go
tab.FlexRow(parent, []warp.FlexItemSpec{
    {Panel: leftPanel, Grow: 1},
    {Panel: rightPanel, Grow: 2},
})
tab.FlexColumn(parent, items)
```

The `Grow` values divide the available space among flex items. A negative value becomes zero.

## Floats

Place a panel above the normal layout:

```go
tab.Float(panel, 10, 4, 30, 8)
```

Float panes support dragging, edge resizing, a close button, and z-order changes when clicked.

## Nested layouts

`TabGroup` already implements `Panel`, so insert it directly into a split or flex layout:

```go
tg := warp.NewTabGroup(warp.TabLeft)
tg.NewTab("inspector")
tab.FlexRow(tab.RootPanel(), []warp.FlexItemSpec{
    {Panel: tg, Grow: 1},
    {Panel: content, Grow: 2},
})
```

For a complete nested Warp, use `inner.AsPanel()`.

## Collapsible panels

```go
collapsed := warp.NewCollapsible("Details", panel)
tab.SetRootPanel(collapsed)
tab.ToggleCollapsible(collapsed)
```

## Scrollable panels

```go
scroll := warp.NewScrollable(panel)
tab.SetRootPanel(scroll)
```

Scrolling supports the mouse wheel, `PgUp`, `PgDn`, and line navigation keys.
