---
title: Input
description: Text input field
---

# Input

Single-line text input with cursor and navigation.

## Constructor

```go
func NewInput(prompt string) *Input
```

## Methods

```go
func (in *Input) SetValue(v string)
func (in *Input) Value() string
func (in *Input) Focus()
func (in *Input) Blur()
func (in *Input) Focused() bool
```

## Controls

- Type — insert characters
- `Backspace` — delete before cursor
- `Delete` — delete after cursor
- `Left`/`Right` — move cursor
- `Home` — start of line
- `End` — end of line

## Interfaces

Implements both `Focusable` and `Panel`.
