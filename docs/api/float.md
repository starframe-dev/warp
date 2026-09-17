---
title: Float
description: Floating panel state and rendering behavior.
---

# Float

A `FloatPane` describes a panel rendered above a tab's normal layout. Use `Tab.Float` to create one and `Tab.CloseFloat` to remove it.

## FloatPane

```go
type FloatPane struct {
    Panel  Panel
    X, Y   int
    Width  int
    Height int
    Title  string
    CloseRequested    bool
    CloseOnOutsideClick bool
}
```

`CloseRequested` becomes true after the close button is pressed. The owning `Tab` removes the pane. `CloseOnOutsideClick` enables closing when the user presses outside the pane.

There is no public `NewFloatPane` constructor or public overlay method. Rendering and mouse dispatch are managed by `Tab`.

## Creating a float

```go
tab.Float(panel, 10, 4, 30, 8)
```

Floats support dragging, edge resizing, a title bar, and a close button. A click inside a float brings it to the front and does not propagate to panels below it.

## ANSI helper

```go
func StripANSI(s string) string
```

Removes ANSI escape sequences. Warp uses visual cell widths when composing overlays.
