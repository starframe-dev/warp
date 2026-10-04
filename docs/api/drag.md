# drag.go

`drag.go` declares the `warp` package and contains a comment describing where split-border drag-and-drop behavior is implemented. It does not implement that behavior or declare any exported identifiers.

## Purpose

The comment states that drag-and-drop interactions for split borders are not implemented in this file. They are handled in `tab.go` and rendered in `render.go`:

- `tab.go` implements input handling and state management for dragging a border between split panes. The relevant functions are `handleMouse`, `updateDrag`, and `hitBorder`.
- `render.go` computes the screen positions of borders via `findBorders`.

## Implementation Details

The file contains the `package warp` clause followed by a Go line comment. It has no imports, executable statements, or exported identifiers.

## Related Files

| File | Responsibility |
|----|----|
| `tab.go` | Input handling and state management for dragging a border between split panes. |
| `render.go` | Computes the on-screen positions of borders. |
