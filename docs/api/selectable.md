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
| AnchorX, AnchorY | `int` | Selection anchor in cell coordinates (fixed on mouse press). |
| CursorX, CursorY | `int` | Active end of the selection (follows mouse/keyboard). |
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

Renders the wrapped content with selection highlight. Caches the
rendered lines for later text extraction. Clamps selection coordinates
to panel bounds — only the cursor end is clamped, while the anchor stays
where the user originally pressed so that reverse-direction drag
continues to work correctly.

``` go
func (s *Selectable) Update(msg tea.Msg) tea.Cmd
```

Handles mouse and keyboard messages for selection. Messages not consumed
by the selection logic are proxied to the wrapped `Panel`.

## Behavior

### Mouse selection

- **Press (left button):** Sets the anchor and cursor to the press
  position; `Selecting` becomes true.
- **Motion:** Moves the cursor; `HasSelection` is set to true.
- **Release:** Finalizes the cursor position; clears selection if anchor
  and cursor coincide.

### Keyboard selection

- **Shift+Up / Shift+Down / Shift+Left / Shift+Right:** Extends or
  creates a selection in the given direction. Shift+Tab is intentionally
  *not* handled and is proxied to the wrapped panel (e.g. a terminal) so
  that TUI apps inside the PTY receive it.
- **Ctrl+A:** Selects all visible content (falls back to 80×24 if
  dimensions are unknown).
- **Esc:** Clears the selection.

### Rendering

`View` renders the content and applies the reverse-video style to the
selected range. The range is computed via `sortedBounds`, which
normalizes anchor/cursor so that start ≤ end. Lines outside the
selection range pass through unchanged. The highlight is applied
per-line using the visual (rune) column range, so it is correct even
when the content contains multi-byte UTF-8 runes.

### Text extraction

`SelectedText` walks the cached rendered lines and extracts the
visual-range text for each line in the selection. ANSI escape sequences
are skipped, and only the runes within the requested visual range are
collected. This ensures the extracted text matches exactly what the user
sees on screen.

### ANSI / Unicode handling

Both `highlightRange` and `extractVisRange` parse the line byte-by-byte,
tracking the visual rune position. ANSI escape sequences (`\x1b[...m`)
are skipped without advancing the visual position. Multi-byte UTF-8
runes are decoded with `utf8.DecodeRuneInString`. This avoids incorrect
slicing of raw bytes.

## Implementation details

- **OSC 52 clipboard:** The `Copy` method encodes the selected text as
  base64 and emits the sequence `\x1b]52;c;<data>\x07`. This works in
  Bubbletea's alternate screen buffer.
- **Reverse-direction drag:** Only the cursor end is clamped to panel
  bounds; the anchor is left as-is. This ensures that dragging in the
  reverse direction still works correctly after a resize or re-render.
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
