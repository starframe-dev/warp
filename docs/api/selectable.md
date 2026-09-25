# Selectable

The `Selectable` type adds text-selection support to a wrapped `Panel`.
Mouse dragging and Shift+arrow keys create a selection, which is
rendered with reversed colors and can be extracted as plain text.

``` go
package warp

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

## Overview

`Selectable` wraps any `Panel` and adds:

- Mouse press/motion/release selection (left button)
- Shift+arrow keyboard selection
- Ctrl+A to select all, Esc to clear
- Copy selected text to the system clipboard via OSC 52

The selected region is rendered with the ANSI reverse-video style
(`\x1b[7m`) so the user sees exactly what is selected.

## Types

### Selectable

Public struct with the following exported fields:

| Field | Type | Description |
|----|----|----|
| Content | `Panel` | The wrapped content panel to which selection is added. |
| AnchorX, AnchorY | `int` | One selection boundary in terminal-cell coordinates. |
| CursorX, CursorY | `int` | Active end-exclusive boundary; `CursorX` may equal the viewport width. |
| HasSelection | `bool` | True when a non-empty selection exists. |
| Selecting | `bool` | True while an active mouse drag is in progress. |

The unexported fields `lastW`, `lastH`, and `lastLines` cache the last
rendered dimensions and lines so that `SelectedText` can extract text
using the exact same coordinate system that was used for rendering.

## Public API

``` go
func NewSelectable(content Panel) *Selectable
```

Constructs a new `Selectable` wrapping the given panel. All selection
state starts empty.

``` go
func (s *Selectable) SelectedText() string
```

Returns the currently selected text as a plain string. Uses the cached
rendered lines so that the coordinate system matches exactly what is
visible on screen. Falls back to rendering at a large size (9999×9999)
if `View` has not been called yet (e.g. in unit tests).

``` go
func (s *Selectable) ClearSelection()
```

Removes the current selection, resetting both `HasSelection` and
`Selecting`.

``` go
func (s *Selectable) Copy() tea.Cmd
```

Returns a `tea.Cmd` that copies the selected text to the system
clipboard via OSC 52. Intended to be called when the user presses
Ctrl+C. Returns `nil` if there is no selection.

``` go
func (s *Selectable) SelectAll(w, h int)
```

Selects all visible content within the given width and height bounds.

``` go
func (s *Selectable) View(w, h int) string
```

Renders the wrapped content with selection highlight and caches the
rendered lines for later text extraction. Selection ranges are
end-exclusive (`[start, end)`); horizontal boundaries are clamped to
`[0, w]`, so the boundary after the final visible cell remains valid. If
`Content` is nil, `View` returns exactly `h` blank lines (`""` when
`h == 0`; otherwise `h-1` newline characters).

``` go
func (s *Selectable) Update(msg tea.Msg) tea.Cmd
```

Handles mouse and keyboard messages for selection. Messages not consumed
by the selection logic are proxied to the wrapped `Panel`.

## Behavior

### Mouse selection

- **Press (left button):** Records the origin cell and starts dragging.
- **Motion:** Converts pointer cell `x` to end-exclusive boundary `x+1` when dragging forward; reverse dragging uses the boundary after the origin cell. `HasSelection` becomes true.
- **Release:** Finalizes the end-exclusive boundary. Press/release without motion does not create a selection.

### Keyboard selection

- **Shift+Up / Shift+Down / Shift+Left / Shift+Right:** Extends or
  creates a selection in the given direction. Shift+Tab is intentionally
  *not* handled and is proxied to the wrapped panel (e.g. a terminal) so
  that TUI apps inside the PTY receive it.
- **Ctrl+A:** Selects all visible content (falls back to 80×24 if
  dimensions are unknown).
- **Esc:** Clears the selection.

### Rendering

`View` renders the content and applies reverse-video style to the
selected range. `sortedBounds` normalizes the anchor and cursor into
end-exclusive bounds. Lines outside the selected range pass through
unchanged. Highlighting uses terminal-cell widths and grapheme
boundaries; selecting any part of a wide grapheme highlights the whole
grapheme.

### Text extraction

`SelectedText` walks the cached rendered lines and extracts complete
graphemes overlapping each selected terminal-cell range. ANSI escape
sequences are skipped, so the extracted text matches the visible
selection.

### ANSI / Unicode handling

`highlightRange` and `extractVisRange` track ANSI-aware terminal-cell
positions and complete grapheme clusters rather than byte or rune
indices. CSI and OSC sequences do not advance the visual position; a
selection overlapping any cell of a grapheme includes that grapheme.

## Implementation details

- **OSC 52 clipboard:** The `Copy` method encodes the selected text as
  base64 and emits the sequence `\x1b]52;c;<data>\x07`. This works in
  Bubbletea's alternate screen buffer.
- **End-exclusive bounds:** `SelectAll(w, h)` sets the final horizontal boundary to `w`, so the last cell is included. Mouse dragging and Shift+Right use the same boundary convention for ASCII and Unicode text.
- **Selection invalidation:** If the cursor and anchor coincide after a
  release or a keyboard step, the selection is cleared automatically.
- **Message proxying:** Any message not handled by the selection logic
  (unrecognized keys, Shift+Tab, resize messages, etc.) is forwarded to
  `Content.Update(msg)`, keeping the wrapped panel fully functional.

## Example usage

``` go
sel := NewSelectable(myPanel)
model.SetProgramView(sel)

// In the model update:
case tea.KeyMsg:
    if msg.String() == "ctrl+c" {
        return sel.Copy()
    }
    if msg.String() == "esc" {
        sel.ClearSelection()
    }

// To get the selected text (e.g. for paste or display):
text := sel.SelectedText()
```
