# warp — `float.go` (float pane)

The `float` package file implements a *floating panel* that is rendered
on top of the main warp layout. It owns its own rectangle (`X`, `Y`,
`Width`, `Height`), renders a titled, closeable border, handles mouse
interaction (drag, resize, close, focus), and finally overlays its
content onto already-rendered screen lines — all while being ANSI-aware,
so styled text and borders never corrupt each other.

## Public API

### `FloatPane`

`type FloatPane struct { Panel Panel X, Y, Width, Height int Title string // exported state used by the owner CloseRequested bool CloseOnOutsideClick bool // unexported interaction state (drag / resize bookkeeping) ... }`

A floating panel that is composed from a regular `Panel` plus a screen
rectangle. The exported fields are:

- `Panel` — the content provider. It is called with the *inner* size
  (`Width-2`, `Height-2`) so the 1-cell border is not counted.
- `X`, `Y`, `Width`, `Height` — screen rectangle.
- `Title` — text drawn in the title bar.
- `CloseRequested` — set to `true` when the user clicks the `×` button.
  The owning tab is responsible for polling this flag after
  `handleMouse` and then calling its `CloseFloat` routine.
- `CloseOnOutsideClick` — when `true`, the owner may close the float if
  the user clicks outside its rectangle.

## Behavior

### Rendering — `render(w, h)`

Produces a fixed `Height`-long slice of lines. The first line is the
title bar, the last is the bottom border, and the rows in between are
the `Panel.View` output padded by `padContent`.

- The top border is built from `╭` + title + dashes + ` ×` + `╮`. The
  close button always reserves 4 visual columns (`╭` + `" ×"` + `╮`). If
  the title is too wide it is truncated with an ellipsis (`...`) so the
  close button stays reachable.
- The content area is
  `padContent(fp.Panel.View(fp.Width-2, fp.Height-2), fp.Width-2, fp.Height-2)`.
  Each content line is wrapped in vertical border characters
  (`│ ... │`).
- The bottom border is `╰ ─ ... ─ ╯`.

### Mouse handling — `handleMouse(msg, mx, my)`

Processes a single `tea.MouseMsg` at screen coordinates (`mx`, `my`) and
returns a `tea.Cmd` only when the event was forwarded to the inner
`Panel`.

Hit-testing rules:

- Close button — exactly the cell at `relY == 0` and
  `relX == fp.Width-2`. Sets `CloseRequested` and does *not* forward to
  the panel.
- Title bar drag — `relY == 0` and `0 < relX < fp.Width-1`. Enters
  `dragging` state and saves the origin.
- Resize edges — `hitEdge` maps any of the eight border cells to `"n"`,
  `"s"`, `"e"`, `"w"` and the four corners `"nw"`, `"ne"`, `"sw"`,
  `"se"`. A non-empty edge starts a `resizing` gesture.
- Inner click — forwarded to `fp.Panel.Update` with coordinates shifted
  by one cell so the panel sees a zero-based content area.

During an active drag/resize the bounds check is intentionally skipped
so the pointer may leave the float rectangle. Otherwise the handler
bails out when the pointer is outside `[X, X+Width) × [Y, Y+Height)`.

### Geometry invariants

| Constant             | Meaning                |
|----------------------|------------------------|
| `floatMinWidth = 10` | Minimum usable width.  |
| `floatMinHeight = 3` | Minimum usable height. |
| `floatTitleH = 1`    | Number of title rows.  |

`applyResize` clamps `Width ≥ 10`, `Height ≥ 3`, `X ≥ 0`, `Y ≥ 0`. The
`applyResize` switch handles all eight edge strings in one place so the
same arithmetic works for every edge.

## ANSI safety

Warp renders styled text with lipgloss and Bubble Tea, so raw byte
offsets do not map to visual columns. Two helpers make the float safe to
overlay:

### `StripANSI(s string) string`

Removes every `\x1b[` (CSI) sequence from a string by skipping the
parameter bytes and the final byte. The result is plain text and can
safely be measured with `lipgloss.Width` or `ansi.StringWidth`.

### `overlayFloat(lines []string, fp *FloatPane, totalW, totalH int)`

Draws `fp.render(totalW, totalH)` on top of the existing `lines` without
disturbing the rest of the screen. For every row it:

1.  Locates the visual column `fp.X` by scanning the original line
    byte-by-byte, copying complete ANSI sequences as opaque units and
    counting visual cells one at a time.
2.  Truncates the styled float line so that it never extends past
    `totalW`. If truncation is needed a `\x1b[0m` reset is appended so
    the truncated line cannot leak color into the suffix.
3.  Skips the original bytes that the float visually covers.
4.  Appends the remaining original suffix verbatim.

The function is idempotent: a float that starts at `(0,0)` with the full
terminal size simply replaces the entire `lines` slice while leaving
every byte of the suffix untouched.

## Style tokens used by the float

The float renders through four `lipgloss` styles defined in `styles.go`.
They are intentionally composed of exactly two colors from the theme so
the border, title, close and background are easy to re-skin through
[theme.go](#theme).

- `floatBorderStyle` — border characters (`╭│─╮╰╯`).
- `floatTitleStyle` — title text; bold, background and foreground from
  the theme.
- `floatCloseStyle` — the `×` glyph; bold foreground.
- `floatBgStyle` — opaque background for the close-cell and corner
  cells.

## Example

``` go
type FloatPane struct {
    Panel  Panel
    X, Y, Width, Height int
    Title string

    CloseRequested      bool
    CloseOnOutsideClick bool
}

// Typical owner loop:
func (t *Tab) handleMouse(msg tea.MouseMsg) tea.Cmd {
    fp := t.Float
    if fp == nil {
        return nil
    }
    cmd := fp.handleMouse(msg, msg.X, msg.Y)
    if fp.CloseRequested {
        t.CloseFloat()
    } else if fp.CloseOnOutsideClick && !fp.contains(msg.X, msg.Y) {
        t.CloseFloat()
    }
    return cmd
}
```

## Notes & constraints

- The float is *not* a `Panel`. It is a screen rectangle that *contains*
  a panel; the float itself owns no keys other than the ones consumed by
  `handleMouse`.
- The close button occupies exactly one cell ( `fp.Width-2`, `0`).
  Anything to its right is part of the `╮` corner and is not a hit
  target.
- Drag/resize bookkeeping uses the *original*
  `(origX, origY, origW, origH)` snapshot, so a long drag never
  accumulates rounding error.
- `render` is the only function that produces styled bytes;
  `overlayFloat` and the mouse path only consume it.
