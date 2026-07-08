---
title: Popover
description: Context menu popover
---

# Popover

A context menu rendered at a specific position, overlaying existing content.

## Types

```go
type PopoverItem struct {
    Name   string
    Action func()
}
```

## Popover Struct

```go
type Popover struct {
    Items   []PopoverItem
    X, Y    int
    Width   int
    OnClose func()
}
```

## Methods

```go
func (p *Popover) Overlay(lines []string, totalW, totalH int) []string
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool
func (p *Popover) HandleKey(msg tea.KeyMsg) bool
```

## Controls

- Click item — execute action and close
- Click outside — close
- `Esc` — close
- `Enter` — select
- `Up`/`Down` — navigate
