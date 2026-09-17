---
title: Input
description: Single-line text input with cursor and focus.
---

# Input

`Input` is a single-line panel with a prompt, editable value, and rune cursor.

## Constructor

```go
func NewInput(prompt string) *Input
```

## Public fields

```go
type Input struct {
    Value  string
    Cursor int
    Prompt string
    Width  int
}
```

`Cursor` is a rune index. `Width == 0` lets `View` use the available width.

## Methods

```go
func (in *Input) SetValue(value string)
func (in *Input) SetCursor(position int)
func (in *Input) Focus()
func (in *Input) Blur()
func (in *Input) Focused() bool
func (in *Input) View(width, height int) string
func (in *Input) Update(msg tea.Msg) tea.Cmd
```

`Input` implements both `Panel` and `Focusable`.

## Editing keys

- Text inserts at the cursor.
- `Backspace` removes the rune before the cursor.
- `Delete` removes the rune after the cursor.
- Arrow keys move the cursor.
- `Home` and `End` move to the beginning and end.
