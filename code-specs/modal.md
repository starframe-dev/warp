# Modal Specification

**File:** `warp/modal.go`  
**Package:** `warp`  
**Language:** Go

## Overview

`modal.go` provides a draggable, mouse-aware modal dialog that can be rendered on top of any panel's content. It is built on top of [Charm](https://charm.sh/) libraries (`bubbletea`, `lipgloss`, `x/ansi`) and supports ANSI-styled content, clickable buttons, a close button, and background dimming.

A modal is typically embedded in a panel and rendered in that panel's `View()` method via `Overlay`. Mouse events are forwarded from the panel's `Update`/`handleMouse` logic through `HandleMouse`.

---

## Messages

### `ShowModalMsg`

```go
type ShowModalMsg struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
}
```

A `tea.Msg` that instructs a panel to display a modal dialog.

| Field | Description |
|-------|-------------|
| `Title` | Title shown in the top border / title bar. |
| `Content` | Body text (ANSI-aware string). |
| `Buttons` | Zero or more bottom buttons. |
| `OnClose` | Optional callback invoked when the modal is closed via the `✕` button or `Esc`. |
| `Width` | Desired box width. `0` means auto-width (3/5 of screen, clamped to 30–50). |

### `CloseModalMsg`

```go
type CloseModalMsg struct{}
```

A `tea.Msg` that instructs a panel to close the current modal.

---

## Types

### `ModalButton`

```go
type ModalButton struct {
    Label  string
    Action func()
}
```

A single button rendered as `[Label]` at the bottom of the modal. When clicked, `Action` is called.

### `Modal`

```go
type Modal struct {
    Title   string
    Content string
    Buttons []ModalButton
    OnClose func()
    Width   int
    // ... internal fields
}
```

The main modal widget. It stores state for content, layout, position, and drag.

---

## Public API

### `NewModal`

```go
func NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

Creates a new `Modal` value with the given title, content, buttons, and close callback. Width is left at `0` (auto).

### `EnsureDimensions`

```go
func (m *Modal) EnsureDimensions(totalW, totalH int)
```

Computes and stores the modal's box dimensions and centered position. Safe to call multiple times; only runs once unless `m.dimsSet` resets. Must be called before `HandleMouse` if `Overlay` has not yet been called.

**Width rules:**

- If `m.Width > 0`, use it.
- Otherwise use `totalW * 3 / 5`.
- Clamp to minimum `30` and maximum `50`.
- Never exceed `totalW`.

**Height:** fixed at `7` lines for the current border + padding + 3 content lines layout.

**Position:** centered, plus any accumulated drag offsets (`offsetX`, `offsetY`).

### `Overlay`

```go
func (m *Modal) Overlay(lines []string, totalW, totalH int) []string
```

Renders the modal on top of the supplied content lines. Returns a new slice of lines with the modal composited.

Behavior:

1. Calls `EnsureDimensions` to set up position/size.
2. Builds three content lines internally: title line (with `✕`), content line, and button line.
3. Truncates content with `…` if it exceeds the inner width.
4. Dims the entire background using `dimStyle` and strips ANSI from the background first.
5. Overlays the box, preserving dimmed content left and right of the box.
6. Returns the modified line slice.

### `HandleMouse`

```go
func (m *Modal) HandleMouse(msg tea.MouseMsg) bool
```

Processes mouse events for the modal. Returns `true` if the event was consumed.

Supported interactions:

| Action | Behavior |
|--------|----------|
| Left click on `✕` | Calls `OnClose` and consumes the event. |
| Left click on a button | Calls the button's `Action` and consumes the event. |
| Left click on top padding strip | Starts dragging the modal. |
| Mouse motion while dragging | Moves the modal and clamps it inside the screen. |
| Mouse release while dragging | Stops dragging. |

Coordinates are expected to be relative to the content lines (i.e., screen Y adjusted by the caller before invocation).

### Position / Size Accessors

```go
func (m *Modal) StartX() int
func (m *Modal) StartY() int
func (m *Modal) BoxWidth() int
func (m *Modal) BoxHeight() int
```

Return the box position and size computed by `Overlay` / `EnsureDimensions`. Useful for tests or external hit-testing.

---

## Important Implementation Details

### Layout

The modal is rendered with:

```go
modalBorderStyle = lipgloss.NewStyle().
    Background(lipgloss.Color(gbDark1)).
    Foreground(lipgloss.Color(gbLight1)).
    BorderStyle(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color(gbBlue)).
    Padding(1, 2)
```

With rounded border, 1 vertical padding and 2 horizontal padding, the visible box has 7 rows for 3 content lines. The inner content width is `boxWidth - 6` (2 border columns + 4 horizontal padding columns).

### Content Lines

1. **Title line:** `Title` padded to inner width, followed by `✕`.
2. **Content line:** `Content` truncated with `…` if too long, then padded.
3. **Button line:** buttons joined as `[Label]  [Label]`, padded.

### Background Dimming

Before compositing the box, every input line is passed through `dimStyle` and `stripANSI` to produce a uniform dimmed background. The box is then spliced back in, leaving the dimmed left/right gutters intact.

### ANSI-Aware Truncation

- `ansi.Truncate` is used to truncate content to the inner width.
- `visualBytePos` walks the string using `x/ansi/parser` and `uniseg.FirstGraphemeCluster` to find the byte index corresponding to a target visual width. This preserves ANSI sequences and multi-cell characters when splitting the background line.
- `stripANSI` removes ANSI escape sequences from the dimmed background.

### Dragging

Dragging is initiated only on the top padding strip (`startY + 1`), not the title line. The modal position is stored as `startX`/`startY` and an additional `offsetX`/`offsetY` is maintained to recompute position after dimension changes. Dragging is clamped so the box cannot move outside the total viewport.

### Button Hit-Testing

The button line is reconstructed with `buildButtonLine`, and bracket pairs are found with `findBracketPair`. Click coordinates are compared against the button's visual span inside the content area (which starts at `startX + 3`, after the left border and padding).

### Close Button Hit-Testing

The `✕` is expected at visual column `startX + boxWidth - 4` on the title line (`startY + 2`). A left click there invokes `OnClose`.

---

## Dependencies

- `github.com/charmbracelet/bubbletea` — `tea.Msg`, `tea.MouseMsg`, `tea.MouseAction`, `tea.MouseButton`.
- `github.com/charmbracelet/lipgloss` — styling, borders, width calculation.
- `github.com/charmbracelet/x/ansi` — ANSI-aware truncation and width measurement.
- `github.com/charmbracelet/x/ansi/parser` — ANSI state machine for `visualBytePos`.
- `github.com/rivo/uniseg` — grapheme cluster handling for `visualBytePos`.
- Standard library: `regexp`, `strings`.

## Notes

- `Width` of `0` triggers auto-sizing; explicit widths outside the 30–50 clamp or larger than the viewport are adjusted automatically.
- The modal is stateful: position and drag offsets persist across `Overlay` calls as long as the same `*Modal` is reused.
- `HandleMouse` requires `EnsureDimensions` (or a prior `Overlay`) to have run; otherwise the box height is zero and the event is not consumed.
