---
title: Selectable
description: Text selection component
---

# Selectable

A wrapper panel that adds text selection to any panel.

## Constructor

```go
func NewSelectable(panel Panel) *Selectable
```

## Controls

- Mouse drag — select text
- `Shift`+arrows — extend selection
- `Ctrl+A` — select all
- `Esc` — clear selection

## Clipboard

```go
func (s *Selectable) Copy()
```

Copies selected text to clipboard via OSC 52 escape sequence.

## Panel Interface

```go
func (s *Selectable) View(w, h int) string
func (s *Selectable) Update(msg tea.Msg) tea.Cmd
```
