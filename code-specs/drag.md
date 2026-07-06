# drag.go Specification

## Overview

`drag.go` is a minimal documentation/placeholder file in the `warp` package that
records where the drag-and-drop logic for split borders is actually implemented.
It contains no executable code, public API, or types of its own.

## Behavior

Drag-and-drop interactions for split borders are **not** implemented in
`drag.go`. Instead, they are handled in `tab.go` and rendered in `render.go`:

- `tab.go` implements the input handling and state management for dragging a
  border between split panes.
- `render.go` computes the screen positions of borders via `findBorders()`.

## Public API

`drag.go` does not declare any functions, methods, variables, constants, or
types.

## Related Implementation Details

The relevant functions in `tab.go` are:

- `handleMouse` — responds to `MouseActionPress`, `MouseActionMotion`, and
  `MouseActionRelease` events that occur on a split border.
- `updateDrag` — computes a new `Fraction` value based on the current mouse
  position while dragging.
- `hitBorder` — determines whether the mouse cursor is positioned over a
  draggable border.

Border geometry is computed in `render.go` by:

- `findBorders()` — returns the positions of borders so they can be hit-tested
  and rendered.

## Notes

- No tests or additional logic belong in this file under the current design.
- Any changes to drag behavior should be made in `tab.go` and `render.go`, not
  here.
