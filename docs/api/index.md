---
title: Warp API Reference
description: Public types, components, interfaces, and helper functions in Warp.
---

# API Reference

Warp is a Go TUI layout engine for Bubbletea. Build a screen from a root [`Panel`](./panel), then compose tabs, splits, flex layouts, floats, and interactive components.

## Core API

| Type | Purpose |
|------|---------|
| [`Warp`](./warp) | Bubbletea model, root panel, optional element-tree HTTP endpoint |
| [`Panel`](./panel) | Interface implemented by every layout panel |
| [`TabGroup`](./tabgroup) | Tab bar and active-tab management |
| [`Tab`](./tab) | Layout tree, floats, focus, resizing, and collapse operations |
| [`Node`](./split) | Layout-tree node |
| [`SplitConfig`](./split) / [`FlexConfig`](./split) | Split and flex configuration |
| [`FloatPane`](./float) | Floating panel state |
| [`Element`](./element) / [`Bounds`](./element) | Semantic UI tree and cell coordinates |
| [`ThemeColors`](./theme) | Runtime theme palette |

## Components

| Component | Purpose |
|-----------|---------|
| [`Collapsible`](./collapsible) | Expandable panel section |
| [`Scrollable`](./scrollable) | Viewport with keyboard and mouse scrolling |
| [`DropdownMenu`](./dropdown) | Button with an expandable item list |
| [`Selectable`](./selectable) | Mouse and keyboard text selection |
| [`Input`](./input) | Single-line editable input |
| [`Modal`](./modal) | Draggable dialog overlay |
| [`Popover`](./popover) | Context-menu overlay |

## Layout and utilities

- [`Split`](./split) — directions, resize messages, collapse state, and flex data types
- [`Focus`](./focus) — explicit focus interfaces
- [`Styles`](./styles) — public border styles and default palette
- [`WordWrap`](./wrap) and `SpaceWrap` — terminal-width-aware text wrapping
- [`StripANSI`](./float) — remove CSI escape sequences from a string
- [`FindElement`](./element) — recursively locate a semantic element
- `WrapToString` — wrap text and join the resulting lines

## Contracts

- Panel sizes are measured in terminal cells, not pixels or bytes.
- `ResizeMsg` carries the allocated content size to leaf panels.
- Focus traversal is explicit: applications choose their own key bindings and call `FocusNext`, `FocusPrev`, or `FocusPanel`.
- Warp does not reserve `Tab` or `Shift+Tab` for focus traversal.
- The default palette is Gruvbox Dark; [`SetTheme`](./theme) changes the semantic palette at runtime.
