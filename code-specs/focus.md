# Specification: focus.go

## Overview

`focus.go` defines the focus-management system for the `warp` UI panel tree. It provides interfaces, helpers, and traversal logic for determining which panels can receive keyboard focus, enumerating them in visual order, and moving focus forward or backward through the tree.

This file is part of the `warp` package and does not expose package-level state; all focus mutation is performed by callers that supply the current and target focusables.

## Public API

### `Focusable` interface

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

A `Panel` that can receive keyboard focus. Implementations must provide:

- `Focus()` — request focus for the panel.
- `Blur()` — release focus from the panel.
- `Focused() bool` — report whether the panel currently has focus.

The embedded `Panel` interface means a `Focusable` is also a valid tree leaf or node panel.

### `RawKeyReceiver` interface

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

A `Panel` that wants to receive all keyboard input without interception (for example, a terminal emulator running inside a PTY). Callers can type-assert a `Panel` to this interface to decide whether to route raw keystrokes directly to the panel.

## Internal Functions

These functions are not exported, but they are the primary implementation surface for focus traversal. Public focus management elsewhere in the package is expected to delegate to these helpers.

### `isFocusable(panel Panel) (Focusable, bool)`

Type-asserts a `Panel` to `Focusable`. Returns `nil, false` if `panel` is `nil`.

### `collectFocusables(node *Node) []Focusable`

Walks the panel tree starting at `node` and returns all focusable panels in visual order. It traverses:

- `node.Split.First`, then `node.Split.Second` for split containers.
- `node.Flex.Items` in order for flex containers.

If the node is a leaf, the function checks the panel itself.

If `node` is `nil`, returns `nil`.

### `focusIndex(list []Focusable, current Panel) int`

Returns the index of `current` in `list`, or `-1` if `current` is `nil` or not found.

### `focusNext(list []Focusable, current Panel) Focusable`

Returns the focusable after `current` in `list`, wrapping to the beginning. Returns `nil` if `list` is empty.

### `focusPrev(list []Focusable, current Panel) Focusable`

Returns the focusable before `current` in `list`, wrapping to the end. Returns `nil` if `list` is empty.

### `applyFocus(current, next Focusable)`

Calls `Blur()` on `current` if it is non-`nil` and not the same panel as `next`, then calls `Focus()` on `next` if it is non-`nil`.

## Important Implementation Details

- Focus traversal is **visual order**: splits are visited first-child then second-child, and flex items are visited in slice order.
- Traversal is **recursive** and allocates a new slice for each call; large trees may create many temporary slices.
- Identity is checked by Go interface comparison (`f == current`). This relies on the underlying panel implementation being comparable in the sense used by the caller (typically the same concrete value).
- Focus wrapping is cyclic: moving past the last focusable returns the first, and moving before the first returns the last.
- `applyFocus` does not check whether `next` is already focused; callers should ensure `next` is the intended destination.
- The `RawKeyReceiver` interface is declared here alongside focus logic because it is closely related to keyboard input routing, but it does not participate in focus traversal.

## Relationships

- `Focusable` embeds `Panel` (defined elsewhere in the package).
- `collectFocusables` depends on `Node.IsLeaf()`, `Node.Split`, and `Node.Flex` (defined elsewhere in the package).
- Callers that want to move focus should collect focusables, compute the next target with `focusNext`/`focusPrev`, and then apply it with `applyFocus`.
