---
title: Selectable
description: Text selection, selected text, and OSC 52 copy support.
---

# Selectable

`Selectable` wraps a panel and tracks a text selection in terminal cell coordinates.

## Constructor and fields

```go
func NewSelectable(content Panel) *Selectable
```

The public selection fields are `AnchorX`, `AnchorY`, `CursorX`, `CursorY`, `HasSelection`, and `Selecting`.

## Selection methods

```go
func (s *Selectable) SelectAll(width, height int)
func (s *Selectable) ClearSelection()
func (s *Selectable) SelectedText() string
func (s *Selectable) Copy() tea.Cmd
```

`Copy` returns a command that emits the selected text through OSC 52. It does not write to the operating system clipboard directly.

## Controls

- Mouse drag selects a region.
- `Shift` + arrows extends the selection.
- `Ctrl+A` selects all visible content.
- `Esc` clears the selection.

## Panel

```go
func (s *Selectable) View(w, h int) string
func (s *Selectable) Update(msg tea.Msg) tea.Cmd
```
