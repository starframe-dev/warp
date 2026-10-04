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

Keyboard input normally goes to the focused panel. Resize, framework, and custom/unknown messages may also be broadcast to unfocused panels, so panel implementations should tolerate those updates. Warp does not bind `Tab` or `Shift+Tab` for focus navigation.

## Focus

A focusable panel implements `Focusable`. `Tab.FocusNext`, `FocusPrev`, `FocusFirst`, and `FocusPanel` provide explicit traversal. The application chooses the key bindings.

## Theme

The default palette is Gruvbox Dark. `SetTheme` replaces package-wide component styles using semantic colors. Configure it before starting the Bubble Tea program; changing the theme while rendering is not synchronized. See [Theme](../api/theme).

## Panel lifecycle

A resource-owning panel may implement `Unmounter`. Warp calls `Unmount` only after the panel is permanently detached and no longer referenced in its ownership domain; switching tabs, hiding, collapsing, or moving focus does not end its lifecycle. Use pointer-backed panels for stable instance identity. Sharing one resource-owning panel instance across independent Warp roots is unsupported and remains the caller's responsibility.

## HTTP inspector

The optional inspector serves `/elements` from the latest completed UI-thread snapshot; HTTP requests do not traverse live panels. With an empty address, `ServeHTTP` binds to `127.0.0.1` using `WARP_HTTP_PORT`, or an automatically assigned port if the variable is unset. Cross-origin access is disabled by default. An explicit non-loopback address may expose UI data to the network; configure a bearer token with `ServeHTTPWithOptions` before using it outside a trusted local environment.

## Element tree

A panel may implement `ElementProvider` to expose semantic controls with `Role`, `Name`, `Action`, and `Bounds`. Warp serves this tree through its optional `/elements` HTTP endpoint, which keeps E2E tests independent from terminal text layout.
