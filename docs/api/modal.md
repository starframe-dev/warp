---
title: Modal
description: Modal dialog window
---

# Modal

A dialog window rendered as an overlay with dimmed background.

## Types

```go
type ModalButton struct {
    Label  string
    Action func()
}
```

## Constructor

```go
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

## Messages

```go
type ShowModalMsg struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
}

type CloseModalMsg struct{}
```

## Methods

```go
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool
```

## Features

- Draggable by title bar
- Close button (×)
- Button actions
- Auto-width: 3/5 of screen, clamped to [30, 50]
