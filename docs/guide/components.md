---
title: Components
description: Built-in Warp components for input, menus, selection, overlays, context menus, and text wrapping.
---

# Components

Warp includes small components you can compose into panels.

## Input

Create an input with:

```go
warp.NewInput(prompt)
```

Input supports cursor movement, backspace, arrow keys, `Home`, `End`, and `Delete`.

## DropdownMenu

Create a dropdown menu with:

```go
warp.NewDropdownMenu(label, items)
```

The component renders a `▼` button. Users can hover items and select an item from the list.

## Selectable

Wrap a panel with selectable text support:

```go
warp.NewSelectable(panel)
```

Selectable supports mouse drag selection, `Shift` + arrows, `Ctrl+A`, and `Esc`. It can copy selected text through OSC 52 clipboard sequences.

## Modal

Create a modal dialog with:

```go
warp.NewModal(title, content, buttons, onClose)
```

A modal renders as an overlay. Users can drag it and close it.

## Popover

Create a popover directly:

```go
&warp.Popover{
	Items:   items,
	X:       x,
	Y:       y,
	OnClose: onClose,
}
```

Use popovers for context menus. Clicking outside closes the popover.

## Text wrapping

Warp provides `WordWrap` and `SpaceWrap` utilities for wrapping text inside a fixed width.
