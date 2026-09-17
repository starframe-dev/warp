---
title: Architecture
description: Warp's panel tree, layout nodes, floats, focus, themes, and semantic elements.
---

# Architecture

Warp is a Bubbletea-based TUI layout engine. It owns a root `Panel`; `warp.New()` starts with a `TabGroup` containing one tab.

## Panel tree

Every value inserted into a layout implements:

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

A `Tab` stores panels in a `Node` tree:

- a leaf holds a `Panel`;
- `SplitConfig` divides an area between two children;
- `FlexConfig` distributes an area among several children;
- `NodeCollapse` stores collapsed state.

The `Tab` methods mutate this tree. `TabGroup` is itself a `Panel`, so a complete tab group can appear inside another tree.

## Rendering and events

Warp renders the tree recursively, pads each child to its allocated cell rectangle, and draws float panes above the result. Mouse events visit the topmost float first, then split borders, then the active panel. A click inside a float does not reach panels below it.

`WindowSizeMsg` and other Bubbletea messages are forwarded to the appropriate panels. Warp does not install application-level `Tab` or `Shift+Tab` bindings.

## Focus

A focusable panel implements `Focusable`. `Tab.FocusNext`, `FocusPrev`, `FocusFirst`, and `FocusPanel` provide explicit traversal. The application chooses the key bindings.

## Theme

The default palette is Gruvbox Dark. `SetTheme` rebuilds the package styles at runtime from semantic colors. See [Theme](../api/theme).

## Element tree

A panel may implement `ElementProvider` to expose semantic controls with `Role`, `Name`, `Action`, and `Bounds`. Warp serves this tree through its optional `/elements` HTTP endpoint, which keeps E2E tests independent from terminal text layout.
