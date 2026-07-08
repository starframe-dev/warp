---
title: Tab
description: Tab — manages a tree of splits, flex, floats
---

# Tab

Manages a tree of panels with splits, flex layouts, and floating panels.

## Root Panel

```go
func (t *Tab) RootPanel() Panel
func (t *Tab) SetRootPanel(panel Panel)
```

## Splits

```go
func (t *Tab) SplitVertical(parent Panel, fraction float64, panel Panel)
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, panel Panel)
```

Splits the parent panel. `fraction` is clamped to [0.1, 0.9].

## Flex

```go
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec)
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec)
```

```go
type FlexItemSpec struct {
    Panel Panel
    Grow  int
}
```

## Floats

```go
func (t *Tab) Float(panel Panel, x, y, w, h int)
func (t *Tab) CloseFloat(fp *FloatPane)
```

Floats overlay on top of rendered content. Draggable, resizable, closable.

## Focus

```go
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

Warp does **not** bind Tab/Shift+Tab. Developer decides which keys trigger focus.

## Collapsible

```go
func (t *Tab) ToggleCollapsible(panel Panel)
```

Toggles a collapsible panel by its content panel reference.

## Context Menu

```go
func (t *Tab) ShowContextMenu(items []PopoverItem, x, y int)
```

Shows a popover at the given position.
