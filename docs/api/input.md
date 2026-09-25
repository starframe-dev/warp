# Input

`Input` is a single-line text input component built on top of Bubble
Tea. It manages a string value, a rune-based cursor position, an
optional prompt prefix, and keyboard editing, and renders itself either
as an inline line or as a bordered box depending on the available
height.

## Types

### Input

A single-line text input component.


    type Input struct {
        Value   string
        Cursor  int // cursor position in runes
        Prompt  string
        Width   int // desired width (0 = auto from View w)
        focused bool
    }

### Input — constructor

``` go
func NewInput(prompt string) *Input
```

Creates a new empty input with the given prompt.

## Methods

| Method | Description |
|----|----|
| `SetValue(v string)` | Replaces the input value and places the cursor at the end. |
| `SetCursor(pos int)` | Sets a rune-indexed cursor position; an index inside a grapheme cluster is normalized to its right boundary. |
| `Focused() bool` | Reports whether the input currently has focus. |
| `Focus()` | Gives the input focus. |
| `Blur()` | Removes focus from the input. |
| `View(w, h int) string` | Renders a bordered box when `w >= 3` and `h >= 3`; otherwise renders an inline line. |
| `Update(msg tea.Msg) tea.Cmd` | Handles grapheme-aware editing while focused. |

## Behavior

### Rendering

`View` dispatches to `viewBoxed` when the width is at least three cells
and the height is at least three lines, drawing a rounded-corner box with
a focused or unfocused border style. Otherwise it falls back to
`viewInline`, which renders the prompt and value as a single styled line
repeated across the height.

The visible line is built by `renderLine`. The prompt and value are
measured in terminal cells, including ANSI styling and wide characters.
When the value does not fit, `truncateInputAtCursor` chooses a visible
range around the cursor without splitting a grapheme cluster; a cluster
cut by the range is represented by spaces. The cursor is rendered with
an inverse-video escape sequence (`\x1b[7m ... \x1b[0m`) around its whole
grapheme cluster, or a trailing inverse space when the cursor is at the
end of the value.

### Cursor and editing

`Update` processes `tea.KeyMsg` only when the input is focused.
Recognized keys:

- `backspace` / `delete` — delete the complete grapheme cluster before
  or at the cursor.
- `left` / `right` — move to the previous or next grapheme boundary.
- `home` / `end` — move the cursor to the start or end of the value.
- `tab` / `shift+tab` — no-op; handled by the parent for focus
  traversal.
- `enter` — no-op placeholder for a future submit signal.
- `tea.KeyRunes` — inserts all runes carried by the event; one event may
  contain multiple runes.

`Cursor` remains a rune index for API compatibility. `clampCursor`
clamps it to `[0, utf8.RuneCountInString(value)]` and normalizes a
position inside a grapheme cluster to that cluster's right boundary. The
value is not limited by the available render width.

## Internal helpers (package-private)

- `viewBoxed` / `viewInline` — layout-specific rendering paths.
- `renderLine` — builds the prompt + value line with cursor highlight.
- `truncateInputAtCursor` — selects visible terminal cells around the
  cursor without splitting grapheme clusters.
- `previousGraphemeBoundary` / `nextGraphemeBoundary` /
  `normalizeGraphemeCursor` — find and normalize rune-indexed cluster
  boundaries.
- `insertAtCursor` / `deleteBeforeCursor` / `deleteAtCursor` — perform
  rune-slice mutations while preserving grapheme boundaries.
- `clampCursor` — invariants enforcement on cursor position.

## Styling

Package-level lipgloss styles are defined at the bottom of the file:
`inputStyle` (light foreground), `inputBorderStyle` (dark border), and
`inputFocusBorderStyle` (blue border when focused). These styles are
shared by both the boxed and inline rendering paths.
