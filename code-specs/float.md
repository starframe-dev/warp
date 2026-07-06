# FloatPane Specification

`float.go` implements a draggable, resizable, floating panel rendered on top of the main layout in the `warp` package. It provides ANSI-aware compositing, mouse-driven interaction, and a close button.

## Overview

A `FloatPane` is a rectangular window drawn above the primary screen content. It supports:

- Positioning via `X`, `Y` coordinates.
- Sizing via `Width`, `Height`.
- Dragging by the title bar.
- Resizing from all edges and corners.
- A close button in the top-right corner.
- Optional `CloseOnOutsideClick` behavior.
- Compositing over existing rendered lines while preserving ANSI escape sequences.

## Public API

### Type: `FloatPane`

```go
type FloatPane struct {
    Panel Panel
    X, Y  int
    Width int
    Height int
    Title string

    CloseRequested bool
    CloseOnOutsideClick bool
}
```

Fields:

- `Panel` — the underlying `Panel` whose content is rendered inside the float. The `Panel` interface must provide `View(width, height) string` and `Update(msg tea.Msg) tea.Cmd`.
- `X`, `Y` — screen coordinates of the top-left corner of the float.
- `Width`, `Height` — outer dimensions of the float including borders.
- `Title` — text displayed in the top border.
- `CloseRequested` — set to `true` when the user clicks the close button (`×`). The owning `Tab` must check this after `handleMouse` and call `CloseFloat`.
- `CloseOnOutsideClick` — when `true`, the caller should close the float if a mouse press occurs outside the float's bounds.

### Function: `StripANSI`

```go
func StripANSI(s string) string
```

Removes ANSI CSI escape sequences (`\x1b[...`) from the input string. Used internally during visual-width calculations and overlay truncation. It is exported and may be used by other components.

### Function: `overlayFloat`

```go
func overlayFloat(lines []string, fp *FloatPane, totalW, totalH int)
```

Composites the rendered float pane onto the provided `lines` slice. The slice is mutated in place. Handles ANSI sequences in both the original content and the float's border/title styles so that visual positions are tracked correctly. Floats are clipped against `totalW`/`totalH` and truncated to avoid extending beyond the right edge.

## Constants

```go
const (
    floatMinWidth  = 10
    floatMinHeight = 3
    floatTitleH    = 1
)
```

- `floatMinWidth` — minimum outer width of a float pane during resizing.
- `floatMinHeight` — minimum outer height of a float pane during resizing.
- `floatTitleH` — height reserved for the title bar (1 row).

## Rendering

### `FloatPane.render(w, h int) []string`

Renders the float into a slice of `Height` strings. The layout is:

1. **Top border** — composed of:
   - A rounded corner `╭`.
   - The title styled with `floatTitleStyle`.
   - Horizontal dashes `─` styled with `floatBorderStyle` to fill the top border.
   - A leading space and the close button `×` styled with `floatCloseStyle`.
   - A rounded corner `╮`.
   - The title is truncated with `...` if it exceeds the available width. The close button and corners reserve 4 visual columns.

2. **Content rows** — each row is `│` + padded content + `│`. The content area is `(Width - 2) × (Height - 2)`.

3. **Bottom border** — `╰` + `─` repeated `Width - 2` times + `╯`.

Background styling is applied via `floatBgStyle` (corners and spaces), border styling via `floatBorderStyle`, title text via `floatTitleStyle`, and the close button via `floatCloseStyle`.

## Mouse Interaction

### `FloatPane.handleMouse(msg tea.MouseMsg, mx, my int) tea.Cmd`

Processes mouse events for the float. `mx` and `my` are relative to the content area (not absolute screen coordinates). The function returns a `tea.Cmd` if the event is forwarded to the underlying panel.

Behavior:

- **Bounds check:** If not currently dragging or resizing, events outside the float's bounds are ignored (return `nil`).
- **Close button:** Clicking the `×` character at `relX == fp.Width-2` and `relY == 0` sets `CloseRequested = true`.
- **Title bar drag:** Pressing on the top border (excluding corners) starts a drag. The current position is recorded as `dragStartX`/`dragStartY`, and the original position is saved in `origX`/`origY`.
- **Edge resize:** Pressing on any border or corner starts a resize. The edge is determined by `hitEdge` and stored as `resizeEdge` (`"n"`, `"s"`, `"e"`, `"w"`, `"ne"`, `"nw"`, `"se"`, `"sw"`). Original position and size are saved.
- **Content forwarding:** A press inside the content area is forwarded to the underlying `Panel` via `Panel.Update` with `X`/`Y` adjusted to be content-relative.
- **Motion:** During dragging, `X`/`Y` are updated by the mouse delta and clamped to `>= 0`. During resizing, `applyResize` is called with the mouse delta.
- **Release:** Ends dragging and resizing, clearing state flags.

### Edge Hit Detection

`hitEdge` returns the resize edge/corner based on the relative coordinates. Corners take precedence over edges.

### Resize Logic

`applyResize` applies the mouse delta to the saved original geometry based on the active edge:

- `"n"`, `"nw"`, `"ne"` move the top edge up and reduce height.
- `"s"`, `"sw"`, `"se"` extend the bottom edge and increase height.
- `"w"`, `"nw"`, `"sw"` move the left edge left and reduce width.
- `"e"`, `"ne"`, `"se"` extend the right edge and increase width.

After resizing, `Width`, `Height`, `X`, and `Y` are clamped to minimums and non-negative values.

## Important Implementation Details

- **ANSI awareness:** `overlayFloat` and `StripANSI` work on visual columns, not byte positions. Each function iterates over runes and skips ANSI CSI sequences (`\x1b[...`) when counting width or copying text.
- **Style reset on truncation:** When a float line is truncated to fit the screen width, `\x1b[0m` is appended to prevent incomplete ANSI sequences from bleeding into the suffix.
- **Style reset after float:** After writing the float line into the output buffer, `\x1b[0m` is written so the original content suffix does not inherit the float's styles.
- **Clipping:** Floats are skipped entirely if `fp.X >= totalW` or `fp.Y >= totalH`. Rows outside the `lines` slice are also skipped.
- **Content padding:** `padContent` (not shown in this file) is assumed to pad/trim the underlying panel's view to fit the content area exactly.
- **Close behavior:** The float does not close itself; it only signals intent via `CloseRequested`. The owner is responsible for removing the float from its collection.

## Dependencies

- `strings` — for string building and repeating characters.
- `unicode/utf8` — for rune-aware byte decoding during ANSI-aware iteration.
- `github.com/charmbracelet/bubbletea` — for `tea.MouseMsg`, `tea.MouseButton`, `tea.MouseAction`, and `tea.Cmd`.
- `github.com/charmbracelet/lipgloss` — for `lipgloss.Width` and styling helpers (`floatBgStyle`, `floatBorderStyle`, `floatTitleStyle`, `floatCloseStyle`).
