---
title: Scrollable
description: Scrollable viewport
---

# Scrollable

A wrapper panel that adds scrolling to any panel.

## Constructor

```go
func NewScrollable(panel Panel) *Scrollable
```

## Controls

- Mouse wheel — scroll up/down
- `PgUp` / `PgDn` — page scroll
- `Up` / `Down` — line scroll

## Panel Interface

```go
func (s *Scrollable) View(w, h int) string
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```
