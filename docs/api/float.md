---
title: Float
description: Floating panel overlay
---

# Float

Floating panels rendered on top of existing content.

## FloatPane

```go
type FloatPane struct {
    Panel              Panel
    X, Y, W, H        int
    CloseOnOutsideClick bool
}
```

## Constructor

```go
func NewFloatPane(panel Panel, x, y, w, h int) *FloatPane
```

## Methods

```go
func (fp *FloatPane) Overlay(lines []string, totalW, totalH int) []string
func (fp *FloatPane) HandleMouse(msg tea.MouseMsg) bool
```

## Utility

```go
func StripANSI(s string) string
```

Removes ANSI escape codes from a string.
