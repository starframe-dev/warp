---
title: Dropdown
description: Dropdown menu component
---

# Dropdown

A button with a dropdown list of options.

## Types

```go
type DropdownItem struct {
    Label    string
    Selected bool
}
```

## Constructor

```go
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu
```

## Properties

```go
menu.Open    // bool — is menu open
menu.Hovered // int — hovered item index
menu.OnSelect func(idx int) // selection callback
```

## Panel Interface

```go
func (d *DropdownMenu) View(w, h int) string
func (d *DropdownMenu) Update(msg tea.Msg) tea.Cmd
```

## Controls

- Click button — open/close
- Click item — select
- `Up`/`Down` — navigate
- `Enter` — confirm
- `Esc` — close
