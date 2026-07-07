# Specification: `selectable.go`

## Overview

`selectable.go` implements `Selectable`, a `Panel` wrapper that adds mouse and keyboard text selection to any `warp.Panel`. It renders the selected region with reversed colors (ANSI highlight) and can copy the selected text to the system clipboard using the OSC 52 escape sequence.

## Type

### `Selectable`

```go
type Selectable struct {
    Content Panel

    AnchorX, AnchorY int
    CursorX, CursorY int

    HasSelection bool
    Selecting    bool

    lastW, lastH int
    lastLines    []string
}
```

A `Selectable` wraps a `Panel` and tracks a selection as an anchor point and a cursor point, both in cell coordinates relative to the panel. `Anchor` is the fixed end of the selection; `Cursor` is the active end that moves with the user.

`HasSelection` indicates whether a non-empty selection exists. `Selecting` is true while the user is actively dragging the mouse. The `lastW`, `lastH`, and `lastLines` fields cache the most recently rendered content so that `SelectedText` extracts exactly the same text that is currently visible.

## Public API

### `NewSelectable(content Panel) *Selectable`

Creates a new `Selectable` wrapping the provided `Panel`. All coordinates and state start at zero with no selection.

### `SelectedText() string`

Returns the text covered by the current selection.

Behavior:

- Returns `""` if there is no selection or if the wrapped panel is nil.
- Uses the rendered lines from the last `View` call (`lastLines`).
- If `View` has not yet been called, falls back to rendering the content with a large width/height (`9999` cells) so that selection bounds can still map to content lines.
- Converts selection coordinates to sorted bounds via `sortedBounds()`.
- For each line within the selection, strips ANSI sequences, clamps the start/end column to the visible line width, and extracts the runes in the selected visual range.
- Joins selected parts with `"\n"`.

### `ClearSelection()`

Removes the current selection by setting `HasSelection` and `Selecting` to `false`. Does not reset the anchor/cursor coordinates.

### `Copy() tea.Cmd`

Returns a Bubble Tea command that copies the selected text to the system clipboard using the OSC 52 sequence.

Behavior:

- Returns `nil` if there is no selected text.
- Base64-encodes the selected text.
- Emits the sequence `\x1b]52;c;<base64>\x07`.
- The command prints the sequence directly to standard output; this works even inside Bubble Tea’s alternate screen buffer.

### `SelectAll(w, h int)`

Selects the entire visible area.

Behavior:

- Sets the anchor to `(0, 0)` and the cursor to `(w-1, h-1)`.
- Sets `HasSelection` to `true`.
- Does not validate `w` or `h`; callers should supply positive dimensions.

### `View(w, h int) string`

Renders the wrapped panel with the selection highlighted.

Behavior:

- If `Content` is nil, returns `h` newlines (blank panel).
- Renders the wrapped panel with `Content.View(w, h)` and stores the dimensions and split lines in `lastW`, `lastH`, and `lastLines`.
- If a selection exists or is in progress, clamps the cursor to the panel bounds `[0, w)` × `[0, h)`. The anchor is also clamped to bounds, but only so that reverse-direction drag remains valid. If after clamping the cursor equals the anchor and the user is not actively dragging, clears `HasSelection`.
- If there is no selection, returns the rendered content unchanged.
- Otherwise, for each line:
  - Lines outside the selected vertical range are left unchanged.
  - Lines inside the range are highlighted by `highlightRange`, which wraps the visual range `[startX, endX)` with the ANSI reverse-video style (`\x1b[7m`) and resets the style afterward (`\x1b[0m`).
- Returns the joined lines.

### `Update(msg tea.Msg) tea.Cmd`

Handles input messages for selection and proxies unhandled messages to the wrapped panel.

Behavior:

- `ResizeMsg`: forwarded to the wrapped panel if present.
- `tea.MouseMsg` with left button:
  - `Press`: sets anchor and cursor to the mouse cell, clears `HasSelection`, and sets `Selecting` to true.
  - `Motion`: if `Selecting` is true, updates the cursor to the mouse cell and sets `HasSelection` to true.
  - `Release`: if `Selecting` is true, updates the cursor, sets `Selecting` to false, and clears `HasSelection` if the selection collapsed to a single cell.
- `tea.KeyMsg`:
  - `shift+up` / `shift+down` / `shift+left` / `shift+right`: moves the cursor and sets `HasSelection` to true. After the first movement, the anchor is set to the cursor position (the original cursor position becomes the anchor). `shift+tab` is intentionally **not** handled and is forwarded to the wrapped panel so that terminal applications inside a PTY receive it.
  - `ctrl+a`: selects the entire visible area using the last rendered dimensions (falling back to `80×24` if `View` has not been called yet).
  - `esc`: clears the current selection if one exists.
- Any unhandled key or other message is forwarded to the wrapped panel via `Content.Update`.

## Important Implementation Details

- **Coordinate system:** All selection coordinates are in cells, not bytes or runes. The implementation uses `lipgloss.Width` and `utf8.DecodeRuneInString` to account for Unicode and ANSI sequences.
- **ANSI handling:** `highlightRange` and `extractVisRange` both walk the raw string one rune at a time, skipping ANSI escape sequences so that visual positions are computed from printable characters only.
- **Style constants:**
  - `selectionStyleANSI = "\x1b[7m"` (reverse video)
  - `resetStyle = "\x1b[0m"`
- **Clipboard:** `Copy` uses OSC 52, which does not require shell access and works over many terminal emulators and remote sessions.
- **State consistency:** `View` clamps the cursor and anchor to the panel bounds, but the anchor is clamped only to keep it within the visible area during reverse-direction drag; the original press position is not preserved outside the panel.
- **Collapsed selection:** If the anchor and cursor are the same cell and the user is not dragging, the selection is considered empty and `HasSelection` is set to false.
- **Last rendered cache:** `SelectedText` relies on `lastLines` matching the current view. If the panel content changes without `View` being called, the extracted text may be stale or empty.
- **Dependencies:** Uses `github.com/charmbracelet/bubbletea` for messages and commands, `github.com/charmbracelet/lipgloss` for width measurement, and `encoding/base64`, `fmt`, `strings`, and `unicode/utf8` from the standard library.
