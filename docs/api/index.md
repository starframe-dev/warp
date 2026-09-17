---
title: Warp API Reference
description: Public types, components, interfaces, and helper functions in Warp.
---

# API Reference

Warp is a Go TUI layout engine for Bubbletea. The API is organized around a root `Warp`, composable `Panel` values, and a `Tab` layout tree.

## Core API

| Type | Purpose |
|------|---------|
| [`Warp`](./warp) | Root Bubbletea model and HTTP element server |
| [`Panel`](./panel) | Component interface |
| [`TabGroup`](./tabgroup) | Tab bar and active tab management |
| [`Tab`](./tab) | Split, flex, float, focus, and collapse operations |
| [`Node`](./split) | Layout tree node |
| [`SplitConfig`](./split) | Two-child split configuration |
| [`FlexConfig`](./split) | Row or column flex configuration |
| [`FlexItem`](./split) / `FlexItemSpec` | Flex item and its grow weight |
| [`FloatPane`](./float) | Floating panel state |
| [`Element`](./element) / [`Bounds`](./element) | Semantic UI tree values |
| [`ThemeColors`](./theme) | Semantic theme palette |

## Components

| Component | Constructor | Purpose |
|-----------|-------------|---------|
| [`Collapsible`](./collapsible) | `NewCollapsible` | Expandable section |
| [`Scrollable`](./scrollable) | `NewScrollable` | Scrollable viewport |
| [`DropdownMenu`](./dropdown) | `NewDropdownMenu` | Dropdown list |
| [`Selectable`](./selectable) | `NewSelectable` | Text selection and copy |
| [`Input`](./input) | `NewInput` | Single-line text input |
| [`Modal`](./modal) | `NewModal` / `ShowModalMsg` | Modal dialog |
| [`Popover`](./popover) | `&Popover{...}` | Context menu |

## Interfaces

- [`Focusable`](./focus) — opt-in keyboard focus
- [`RawKeyReceiver`](./focus) — terminal-style raw-key preference
- [`ElementProvider`](./element) — semantic element tree

## Helpers

- [`WordWrap`](./wrap) and `SpaceWrap` — text wrapping
- [`StripANSI`](./float) — remove ANSI escape sequences
- [`FindElement`](./element) — recursive element lookup
- `WrapToString` — wrap text and join the resulting lines

## Supporting types

`Direction`, `TabPosition`, `ResizeMsg`, `NodeCollapse`, `ModalButton`, and `DropdownItem` are documented on the pages for the components that use them.

## Generated per-file reference

code-check also generates low-level HTML documentation for each Go source file. It stays under `docs/en/` and `docs/ru/` because those paths are part of the code-check contract. The curated VitePress pages above are the recommended starting point. See the [generated source reference](../guide/generated-reference) for links to the per-file HTML pages.
