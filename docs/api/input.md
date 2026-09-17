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
| `SetCursor(pos int)` | Sets the cursor position in runes and clamps it into the valid range. |
| `Focused() bool` | Reports whether the input currently has focus. |
| `Focus()` | Gives the input focus. |
| `Blur()` | Removes focus from the input. |
| `View(w, h int) string` | Renders the input; uses a bordered box when `h >= 3`, otherwise an inline line. |
| `Update(msg tea.Msg) tea.Cmd` | Handles keyboard input (backspace, delete, left/right, home/end, text insert) while focused. |

## Behavior

### Rendering

`View` dispatches to `viewBoxed` when the available height is at least
three lines, drawing a rounded-corner box with a focused or unfocused
border style. Otherwise it falls back to `viewInline`, which renders the
prompt and value as a single styled line repeated across the height.

The visible line is built by `renderLine`, which concatenates the prompt
prefix with the value. If the combined width exceeds the available
width, the value is truncated from the left so that the cursor remains
visible. The cursor position is rendered with an inverse-video escape
sequence (`\x1b[7m ... \x1b[0m`) around the rune at the cursor position,
or a trailing inverse space when the cursor is at the end of the value.

### Cursor and editing

`Update` processes `tea.KeyMsg` only when the input is focused.
Recognized keys:

- `backspace` — deletes the rune before the cursor.
- `delete` — deletes the rune at the cursor.
- `left` / `right` — move the cursor by one rune, clamped to bounds.
- `home` / `end` — move the cursor to the start or end of the value.
- `tab` / `shift+tab` — no-op; handled by the parent for focus
  traversal.
- `enter` — no-op placeholder for a future submit signal.
- Single-character keys — insert the rune at the cursor position.

After every edit, `clampCursor` ensures the cursor stays within
`[0, len(value)]` in runes.

## Internal helpers (package-private)

- `viewBoxed` / `viewInline` — layout-specific rendering paths.
- `renderLine` — builds the prompt + value line with cursor highlight.
- `truncateTailToWidth` — keeps the visible value centered around the
  cursor by shifting the start index when the cursor is beyond the
  visible width.
- `insertAtCursor` / `deleteBeforeCursor` / `deleteAtCursor` —
  rune-level text mutations.
- `clampCursor` — invariants enforcement on cursor position.

## Styling

Package-level lipgloss styles are defined at the bottom of the file:
`inputStyle` (light foreground), `inputBorderStyle` (dark border), and
`inputFocusBorderStyle` (blue border when focused). These styles are
shared by both the boxed and inline rendering paths.
