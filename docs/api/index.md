---
title: Warp API Reference
description: Public types, components, interfaces, and helper functions in Warp.
---

# API Reference

Warp is a Go TUI layout engine for Bubbletea. Build a screen from a root <a href="./panel.html"><code>Panel</code></a>, then compose tabs, splits, flex layouts, floats, and interactive components.

## Core API

| Type | Purpose |
|------|---------|
| <a href="./warp.html"><code>Warp</code></a> | Bubbletea model, root panel, optional element-tree HTTP endpoint |
| <a href="./panel.html"><code>Panel</code></a> | Interface implemented by every layout panel |
| <a href="./tabgroup.html"><code>TabGroup</code></a> | Tab bar and active-tab management |
| <a href="./tab.html"><code>Tab</code></a> | Layout tree, floats, focus, resizing, and collapse operations |
| <a href="./split.html"><code>Node</code></a> | Layout-tree node |
| <a href="./split.html"><code>SplitConfig</code></a> / <a href="./split.html"><code>FlexConfig</code></a> | Split and flex configuration |
| <a href="./float.html"><code>FloatPane</code></a> | Floating panel state |
| <a href="./element.html"><code>Element</code></a> / <a href="./element.html"><code>Bounds</code></a> | Semantic UI tree and cell coordinates |
| <a href="./theme.html"><code>ThemeColors</code></a> | Runtime theme palette |

## Components

| Component | Purpose |
|-----------|---------|
| <a href="./collapsible.html"><code>Collapsible</code></a> | Expandable panel section |
| <a href="./scrollable.html"><code>Scrollable</code></a> | Viewport with keyboard and mouse scrolling |
| <a href="./dropdown.html"><code>DropdownMenu</code></a> | Button with an expandable item list |
| <a href="./selectable.html"><code>Selectable</code></a> | Mouse and keyboard text selection |
| <a href="./input.html"><code>Input</code></a> | Single-line editable input |
| <a href="./modal.html"><code>Modal</code></a> | Draggable dialog overlay |
| <a href="./popover.html"><code>Popover</code></a> | Context-menu overlay |

## Layout and utilities

- <a href="./split.html"><code>Split</code></a> — directions, resize messages, collapse state, and flex data types
- <a href="./focus.html"><code>Focus</code></a> — explicit focus interfaces
- <a href="./styles.html"><code>Styles</code></a> — public border styles and default palette
- <a href="./wrap.html"><code>WordWrap</code></a> and `SpaceWrap` — terminal-width-aware text wrapping
- <a href="./float.html"><code>StripANSI</code></a> — remove CSI escape sequences from a string
- <a href="./element.html"><code>FindElement</code></a> — recursively locate a semantic element
- `WrapToString` — wrap text and join the resulting lines

## Contracts

- Panel sizes are measured in terminal cells, not pixels or bytes.
- `ResizeMsg` carries the allocated content size to leaf panels.
- Focus traversal is explicit: applications choose their own key bindings and call `FocusNext`, `FocusPrev`, or `FocusPanel`.
- Warp does not reserve `Tab` or `Shift+Tab` for focus traversal.
- The default palette is Gruvbox Dark; <a href="./theme.html"><code>SetTheme</code></a> changes the semantic palette at runtime.
