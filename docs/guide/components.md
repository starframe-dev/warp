---
title: Components
description: Warp's built-in input, menu, selection, overlay, scrolling, and wrapping components.
---

# Components

Warp components implement `Panel` and compose through the same layout tree.

## Input and menus

- `NewInput(prompt)` creates a single-line field with a rune cursor and editing keys.
- `NewDropdownMenu(label, items)` creates a button and an expandable list.
- `Popover` provides a context menu that overlays existing lines.

## Content wrappers

- `NewSelectable(panel)` adds mouse and keyboard text selection; `Copy()` returns an OSC 52 command.
- `NewScrollable(panel)` adds wheel, page, and line scrolling.
- `NewCollapsible(title, panel)` adds a collapsible title section.

## Dialogs

`NewModal(title, content, buttons, onClose)` creates a draggable modal with an overlay, close button, and button actions.

## Text

`WordWrap`, `SpaceWrap`, and `WrapToString` handle fixed-width text using terminal display width.

## Theme

Use `SetTheme(ThemeColors{...})` to replace the default palette. The <a href="../api/theme.html">Theme API</a> lists semantic fields and their mappings.
