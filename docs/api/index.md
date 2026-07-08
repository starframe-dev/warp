---
title: Warp — API Reference
description: Complete API reference for the Warp Go TUI layout engine
---

# API Reference

## Core Types

| Type | Description |
|------|-------------|
| `Warp` | Root Bubbletea model, wraps a Panel |
| `TabGroup` | Panel with tab bar and content area |
| `Tab` | Manages a tree of splits, flex, floats |
| `Panel` | Interface: `View(w, h)` and `Update(msg)` |
| `Node` | Tree node: leaf Panel, SplitConfig, or FlexConfig |
| `SplitConfig` | Vertical/horizontal split with fraction |
| `FlexConfig` | Row/column flex with grow weights |
| `FlexItemSpec` | Panel + Grow weight for flex layouts |
| `FloatPane` | Floating panel overlay |
| `TabPosition` | TabTop, TabBottom, TabLeft, TabRight, TabNone |

## Components

| Component | Constructor | Description |
|-----------|-------------|-------------|
| Collapsible | `NewCollapsible(title, panel)` | Expand/collapse section |
| Scrollable | `NewScrollable(panel)` | Scrollable viewport |
| DropdownMenu | `NewDropdownMenu(label, items)` | Dropdown list |
| Selectable | `NewSelectable(panel)` | Text selection |
| Input | `NewInput(prompt)` | Text input field |
| Modal | `NewModal(title, content, buttons, onClose)` | Dialog window |
| Popover | `&warp.Popover{Items, X, Y, OnClose}` | Context menu |

## Interfaces

| Interface | Methods |
|-----------|---------|
| `Focusable` | `Focus()`, `Blur()`, `Focused() bool` |
| `RawKeyReceiver` | `HandleRawKey(msg tea.KeyMsg) tea.Cmd` |
| `ElementProvider` | `Elements(w, h) []Element` |

## Helpers

| Function | Description |
|----------|-------------|
| `WordWrap(text, width)` | Word-wrap text |
| `SpaceWrap(text, width)` | Space-wrap text |
| `StripANSI(s)` | Remove ANSI codes |
| `FindElement(elems, role, name, action)` | Find in element tree |
