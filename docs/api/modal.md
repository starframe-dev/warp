# Modal

The `Modal` type implements a draggable dialog box rendered on top of a panel's content. It has a title and close glyph, a content area, and a row of buttons. A panel embeds a modal, calls `Overlay` when rendering, and routes mouse events to `HandleMouse`.

## Package

Package `warp`, file `modal.go`. Imports:

- `github.com/charmbracelet/bubbletea` (as `tea`) — mouse event types.
- `github.com/charmbracelet/lipgloss` — box rendering, borders, padding, and colors.
- `github.com/charmbracelet/x/ansi` — ANSI string width, truncation, and stripping.
- `github.com/charmbracelet/x/ansi/parser` — ANSI state machine used by `visualBytePos`.
- `github.com/rivo/uniseg` — Unicode grapheme-cluster width computation used by `visualBytePos`.

## Types

### ShowModalMsg

```go
type ShowModalMsg struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
}
```

A Bubble Tea message type that can carry the parameters for showing a modal. The panel handles the message and creates or stores the modal; this type does not perform that handling itself.

### CloseModalMsg

```go
type CloseModalMsg struct{}
```

A Bubble Tea message type that a panel can use to request closing its current modal. The message itself does not close a modal.

### Modal

```go
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

    totalW, totalH int
}
```

A dialog rendered on top of a panel's content. The unexported fields hold the last computed box geometry, drag state, offsets, and viewport dimensions. They are managed by the modal methods.

| Field | Description |
|----|----|
| `Title` | Title text. Framework control characters/sequences are sanitized before rendering. |
| `Content` | Content string; ANSI formatting is retained where valid, while control characters are normalized. |
| `Buttons` | Buttons rendered at the bottom of the box; each carries an optional `Action`. |
| `OnClose` | Callback invoked when the close glyph is clicked. This implementation does not handle Esc or external close paths. |
| `Width` | Requested box width. A non-positive value selects 3/5 of the viewport width; the resulting width is clamped to 30–50 columns and then to the viewport width. |

### ModalButton

```go
type ModalButton struct {
    Label  string
    Action func()
}
```

A button in a modal. Its sanitized label is rendered as `[Label]`; its `Action` is invoked when its visible bracketed range is clicked, if the callback is non-nil.

## Constructor

```go
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

Creates and returns a modal with the supplied title, content, buttons, and close callback. `Width` remains zero, so automatic width selection is used unless the caller sets it.

## Methods

### EnsureDimensions

```go
func (m *Modal) EnsureDimensions(totalW, totalH int)
```

Clamps the supplied viewport dimensions to non-negative values, computes the box width and fixed height (7 rows), and updates the modal's position and cached viewport dimensions. Width is `Width` when positive or `totalW * 3 / 5` otherwise; it is clamped to 30–50 columns and then to `totalW`. The modal is centered, with its stored drag offset applied, and the result is clamped to the viewport. Offsets are updated to reflect the clamped position. Each call recalculates dimensions and position; it is not a one-time initialization. A nil receiver is ignored.

Call this before `HandleMouse` if `Overlay` has not run for the current viewport dimensions.

### Overlay

```go
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string
```

Renders the modal on top of the supplied lines and returns the overlaid lines.

1. Returns the input unchanged when `totalW <= 0` or `lines` is empty.
2. Calls `EnsureDimensions(totalW, totalH)` and uses the resulting geometry.
3. Builds three content lines within `innerWidth = max(0, boxWidth - 6)`: a sanitized title padded before the `✕` glyph, normalized content truncated to the available visual width (with `…` if it initially exceeds that width), and sanitized bracketed button labels joined by two spaces. The content and button lines are padded to the inner width.
4. Renders the lines in a Lip Gloss rounded-border box using `modalBorderStyle` (background `gbDark1`, foreground `gbLight1`, border `gbBlue`, padding `(1, 2)`).
5. Dims every supplied line after stripping its ANSI sequences, using `dimStyle` (gray foreground and `gbDark0` background).
6. Places box lines over the dimmed lines, preserving available content to the left and right. The right-side byte offset is calculated by `visualBytePosAfter`; the left side is truncated to the visual starting column and padded if needed. Resulting lines are padded to `totalW`.

### HandleMouse

```go
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool
```

Processes a mouse event and reports whether it was consumed. The box layout is:

```text
startY + 0  top border
startY + 1  top padding (draggable strip)
startY + 2  title and close glyph
startY + 3  content
startY + 4  buttons row
startY + 5  bottom padding
startY + 6  bottom border
```

Returns `false` for a nil receiver or when `boxHeight == 0`.

- **Left-button press:** A click exactly at `(startX + boxWidth - 4, startY + 2)` invokes `OnClose` if non-nil and consumes the event even if the callback is nil. A click in a visible button's bracket range on row `startY + 4` invokes its non-nil `Action` and consumes the event. Truncated or hidden buttons have no hit region. A press anywhere within the box on row `startY + 1` starts dragging and records the pointer position.
- **Motion:** While dragging, applies the pointer delta to the position and clamps the box to the cached viewport bounds. The drag offset is recalculated from the resulting position.
- **Release:** Ends an active drag and consumes the event; otherwise it is not consumed.

Mouse Y coordinates must use the same line-relative coordinate system as the lines passed to `Overlay` (0 is the first content line). The caller is responsible for translating screen coordinates if needed.

### Geometry accessors

The following methods return the most recently computed geometry, after `Overlay` or `EnsureDimensions` has run:

| Method | Returns |
|----|----|
| `StartX()` | `startX`, the left X coordinate of the box. |
| `StartY()` | `startY`, the top Y coordinate of the box. |
| `BoxWidth()` | `boxWidth`, the box width. |
| `BoxHeight()` | `boxHeight`, fixed at 7 rows. |

## Internal helpers

- `buildButtonLine() string` — builds the sanitized `[Label]` entries joined by two spaces.
- `findBracketPair(s string, pos int) (int, int)` — finds the next bracket pair at or after `pos`; returns `(-1, -1)` if none is found. Used for button hit-testing.
- `visualBytePos(s string, targetW int) int` — walks an ANSI-aware stream and returns a byte offset based on visual width, or `len(s)` if the target width is not reached. It is defined in this file but is not used by `Overlay`.
- `stripANSI(s string) string` — delegates to `ansi.Strip`.
- `max(a, b int) int` — returns the greater integer.

Rendering also uses shared terminal-text helpers defined elsewhere in the package to normalize controls, sanitize framework labels, truncate fragments, and pad visual lines.

## Styles

Two package-level `lipgloss.Style` values are used by the modal:

- `modalBorderStyle` — background `gbDark1`, foreground `gbLight1`, rounded border in `gbBlue`, padding `(1, 2)`.
- `dimStyle` — gray foreground on `gbDark0` background.

## Integration notes

- The modal has no independent rendering surface. A panel composes `Overlay`'s result into its `View()` output.
- `ShowModalMsg` and `CloseModalMsg` are message types only; a parent panel must implement the corresponding state changes.
- `HandleMouse` handles clicking the close glyph but does not implement keyboard input such as Esc.
- Dragging starts only on the top padding strip (`startY + 1`), not on the title row.
