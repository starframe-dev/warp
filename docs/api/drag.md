# drag.go

This file exists as a package-level documentation marker for the `warp`
package. It intentionally contains no code, no exported identifiers, and
no implementation. Its only purpose is to hold the package documentation
comment that describes where the drag-and-drop behaviour for split
borders actually lives.

## Purpose

The file documents that drag-and-drop interactions for split borders are
not implemented in this file. They are handled in `tab.go` and rendered
in `render.go`:

- `tab.go` implements the input handling and state management for
  dragging a border between split panes. The relevant functions are
  `handleMouse`, `updateDrag`, and `hitBorder`.
- `render.go` computes the screen positions of borders via
  `findBorders`.

## Public API

This file does not declare any exported functions, types, variables, or
constants. The only public symbol present is the package name `warp`
itself.

## Implementation Details

The file body is a single doc comment starting with a blank line. It
contains no executable code, no imports, and no side effects. It is safe
to treat the file as pure documentation.

## Related Files

| File | Responsibility |
|----|----|
| `tab.go` | Input handling and state management for dragging a border between split panes. |
| `render.go` | Computes the on-screen positions of borders. |
