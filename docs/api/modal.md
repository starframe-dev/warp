# Modal

The `Modal` type implements a draggable modal dialog box that renders on
top of a panel's content. It is composed of a titled box (rounded
border, background color), a content area, a row of buttons, a close (✕)
button, and optional drag behavior via mouse events. The modal is
designed to be embedded by any panel in a Bubble Tea TUI app (the `warp`
project).

## Package

Package `warp`, file `modal.go`. Imports:

- `github.com/charmbracelet/bubbletea` (as `tea`) — TUI message/event
  types.
- `github.com/charmbracelet/lipgloss` — box rendering, borders, padding,
  colors.
- `github.com/charmbracelet/x/ansi` — ANSI string width/truncation.
- `github.com/charmbracelet/x/ansi/parser` — ANSI state machine used by
  `visualBytePos`.
- `github.com/rivo/uniseg` — Unicode segmentation for grapheme-cluster
  width computation.

## Types

### ShowModalMsg

``` go
type ShowModalMsg struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
}
```

A Bubble Tea message instructing a panel to display a modal dialog. The
panel is responsible for converting this message into a concrete `Modal`
and storing it.

### CloseModalMsg

``` go
type CloseModalMsg struct{}
```

A Bubble Tea message instructing a panel to close the currently
displayed modal.

### Modal

``` go
type Modal struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int

    startX, startY int
    boxWidth, boxHeight int

    dragging bool
    dragX, dragY int
    offsetX, offsetY int

    dimsSet bool
    totalW, totalH int
}
```

A dialog rendered on top of a panel's content. Any panel can embed and
render a `Modal` in its `View()`. The unexported fields (box
position/size, drag state, cached dimensions) are managed by the methods
described below and should not be mutated directly.

| Field | Description |
|----|----|
| `Title` | Displayed in the top border area of the box. |
| `Content` | ANSI-rendered content string (may contain ANSI escape sequences). |
| `Buttons` | Buttons rendered at the bottom of the box; each carries its own `Action`. |
| `OnClose` | Called when the modal is closed via the ✕ button, Esc, or any external path. |
| `Width` | Desired box width; 0 means auto (3/5 of screen width, clamped to 30–50 columns). |

### ModalButton

``` go
type ModalButton struct {
    Label  string
    Action func()
}
```

A single button inside a modal. The label is rendered as `[Label]` and
its `Action` runs when the button is clicked.

## Constructors

``` go
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

Creates a new modal from its parts and returns it as a pointer. The box
width is left at its zero value so the auto-sizing logic in
`EnsureDimensions` applies on first render.

## Methods

### EnsureDimensions

``` go
func (m *Modal) EnsureDimensions(totalW, totalH int)
```

Computes and stores the box position/size if it has not been computed
yet (`dimsSet == false`). The box width is resolved as: `Width` if
positive, otherwise `totalW * 3 / 5`, clamped to a minimum of 30 and a
maximum of 50, and never wider than `totalW`. The box height is fixed at
7 rows (rounded border + padding(1,2) + 3 content lines). The box is
centered horizontally and vertically, and any existing drag offset is
added to the centered position. Subsequent calls are no-ops.

Must be called before `HandleMouse` if `Overlay` has not yet run.

### Overlay

``` go
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string
```

Renders the modal on top of the supplied content lines and returns the
overlaid lines. Behavior:

1.  Guards: returns the input unchanged when `totalW <= 0` or `lines` is
    empty.
2.  Calls `EnsureDimensions(totalW, totalH)` and reads back the box
    position/size.
3.  Builds the three content lines:
    - *Title line* — `Title` padded with spaces, terminated with a `✕`
      close glyph.
    - *Content line* — `Content` truncated to
      `innerWidth = boxWidth - 6` columns (visual width), with an
      ellipsis when truncated, padded to the same width.
    - *Button line* — buttons rendered as `[Label]` entries joined by
      two spaces, padded to `innerWidth`.
4.  Wraps the three lines in a `lipgloss` rounded-border box using the
    shared `modalBorderStyle` (background `gbDark1`, foreground
    `gbLight1`, border color `gbBlue`, padding `(1, 2)`).
5.  Dims the full background of every line with `dimStyle` (background
    `gbDark0`, gray foreground) after stripping the original ANSI
    escapes.
6.  Paste the box lines over the dimmed content while preserving the
    content on the left and right of the box. The left and right slices
    are recovered by byte position using `visualBytePos`, which walks
    the ANSI string and stops at the first byte offset where the visual
    width reaches the target column. Left part is padded with spaces
    when its visual width is shorter than `startX`.

### HandleMouse

``` go
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool
```

Processes a mouse event for the modal and returns whether the event was
consumed. The expected layout (rows are 1-based within the box):

``` plaintext
startY + 0  top border
startY + 1  top padding (draggable strip)
startY + 2  title + ✕
startY + 3  content
startY + 4  buttons row 1
startY + 5  buttons row 2 (wraps when wide)
startY + 6  bottom border
```

Behavior:

- No-op (`false`) when `boxHeight == 0` (dimensions unknown).
- **MouseActionPress** + left button:
  - If the click is exactly on the `✕` close button (at
    `(startX + boxWidth - 4, startY + 2)`) and `OnClose` is set,
    `OnClose()` is invoked and the event is consumed.
  - Otherwise, the row of buttons is hit-tested: `buildButtonLine()`
    rebuilds the `[label]` string and `findBracketPair` walks it to
    compute the byte/visual `[start, end)` ranges of each button; adding
    `startX + 3` (1 border + 2 padding) gives the X range in screen
    coordinates. A press whose X falls inside `[btnX1, btnX2)` and whose
    Y is on either of the two button rows invokes the button's `Action`
    and consumes the event.
  - If the press is on the draggable padding strip (`Y == startY + 1`, X
    inside the box), dragging starts and the drag anchor is recorded.
- **MouseActionMotion**: while dragging, the delta is applied to both
  the accumulated offset (`offsetX`/`offsetY`) and the live position
  (`startX`/`startY`). The position is clamped to the screen using
  `totalW`/`totalH`.
- **MouseActionRelease**: ends a drag if one is in progress and consumes
  the event.

### Accessors

These accessors expose the last computed box geometry after `Overlay`
(or `EnsureDimensions`) has run:

| Method        | Returns                                   |
|---------------|-------------------------------------------|
| `StartX()`    | `startX` — left X of the box.             |
| `StartY()`    | `startY` — top Y of the box.              |
| `BoxWidth()`  | `boxWidth` — width of the box.            |
| `BoxHeight()` | `boxHeight` — height of the box (7 rows). |

## Internal helpers

- `buildButtonLine() string` — re-derives the buttons row string
  (identical to the one rendered by `Overlay`) so hit-testing uses the
  same source of truth as rendering.
- `findBracketPair(s string, pos int) (int, int)` — finds the next `[…]`
  pair at/after `pos`; returns `(-1, -1)` when not found. Used by
  `HandleMouse` to iterate buttons.
- `visualBytePos(s string, targetW int) int` — walks `s` as an
  ANSI-aware stream (via the `parser` table and `uniseg` grapheme
  clusters) and returns the byte offset where the visual width first
  reaches `targetW`; returns `len(s)` if it never does.
- `stripANSI(s string) string` — delegates to `ansi.Strip` and removes
  supported ANSI/OSC control sequences, including SGR and hyperlinks.
- `max(a, b int) int` — small int max helper (pre-Go-1.21 compatible).

## Styles

Two package-level `lipgloss.Style` values are shared with other parts of
the TUI:

- `modalBorderStyle` — background `gbDark1`, foreground `gbLight1`,
  rounded border in `gbBlue`, padding `(1, 2)`. Used by `Overlay` to
  wrap the three content lines.
- `dimStyle` — gray foreground on `gbDark0` background. Used to dim the
  panel content behind the modal.

## Integration notes

- The modal does not own its own rendering surface; a panel composes the
  result of `Overlay` into its `View()` output.
- `msg.Y` passed to `HandleMouse` is expected to be in the same
  line-relative coordinate system used by `Overlay` (0 = first content
  line of the panel; the parent has already translated screen Y to line
  Y).
- Dragging is only initiated from the top padding strip (`startY + 1`),
  not from the title line itself; the title row is still hit-testable
  for the `✕` close glyph and for buttons.
- Content truncation uses visual width: `ansi.Truncate(s, n, "")`
  followed by appending `…` when the raw content was wider than the
  inner width.
- The background is stripped with `ansi.Strip` before dimming, including
  CSI/SGR and OSC sequences, so escape bytes and hyperlink controls do
  not leak into the sliced left/right parts of each line.
