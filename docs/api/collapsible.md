---
title: Collapsible
description: Expand/collapse panel section
---

# Collapsible

A wrapper panel that can be collapsed/expanded.

## Constructor

```go
func NewCollapsible(title string, panel Panel) *Collapsible
```

## Methods

```go
func (c *Collapsible) Toggle()
```

Toggles between collapsed and expanded state.

## Properties

```go
collapsed := c.Collapsed // bool
```

## Panel Interface

```go
func (c *Collapsible) View(w, h int) string
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd
```
