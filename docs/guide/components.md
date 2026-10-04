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
- `NewScrollable(panel)` adds wheel, page, and line scrolling. Its offset is clamped to the content end when the wrapped panel provides a known intrinsic height.
- `NewCollapsible(title, panel)` adds a collapsible title section.

## Dialogs

`NewModal(title, content, buttons, onClose)` creates a draggable modal with an overlay, close button, and button actions.

## Text

`WordWrap`, `SpaceWrap`, and `WrapToString` handle fixed-width text using terminal display width.

Custom scrollable content can implement the optional `ContentHeightProvider` to report its intrinsic height without rendering. When that height is known, `ViewportRenderer` and `ViewportElementProvider` can render or inspect only the visible region; ordinary `Panel` implementations remain compatible without these interfaces.

## Theme

Use `SetTheme(ThemeColors{...})` to replace the default palette. Configure it before the Bubble Tea program starts; theme changes are not synchronized with rendering. The [Theme API](../api/theme) lists semantic fields and their mappings.
