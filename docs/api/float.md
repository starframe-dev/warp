# warp — `float.go` (float pane)

`FloatPane` is rendered on top of a tab's main layout. It owns a screen
rectangle (`X`, `Y`, `Width`, `Height`), draws a titled, closeable border,
and supports mouse dragging and resizing. The owning `Tab` manages focus,
z-order, outside-click closing, and overlay rendering; ANSI-aware clipping
keeps styled text and borders within the viewport.

## Public API

### `FloatPane`

```go
type FloatPane struct {
    Panel Panel
    X, Y, Width, Height int
    Title string

    preferredWidth int
    preferredHeight int

    CloseRequested bool
    CloseOnOutsideClick bool
    // Drag and resize bookkeeping is unexported.
}
```

A floating panel that is composed from a regular `Panel` plus a screen
rectangle. The exported fields are:

- `Panel` — the content provider. It is called with the *inner* size
  (`Width-2`, `Height-2`) so the 1-cell border is not counted.
- `X`, `Y`, `Width`, `Height` — screen rectangle.
- `Title` — text drawn in the title bar.
- `CloseRequested` — set to `true` when the user clicks the `×` button;
  the owning `Tab` observes this and removes the float.
- `CloseOnOutsideClick` — when `true`, the owner may close the float if
  the user clicks outside its rectangle.

## Behavior

### Preferred size and viewport

`Tab.Float(panel, x, y, width, height)` creates a float and stores the requested dimensions as its preferred size, raised to the minimum dimensions (10×3). `Width` and `Height` are the current visible dimensions; the preferred dimensions are stored separately.

On `tea.WindowSizeMsg`, `ResizeMsg`, and before rendering, the tab clamps the visible rectangle to the viewport without changing the preferred dimensions. When the viewport grows, the float restores its preferred width and height and reclamps `X` and `Y` so the rectangle fits. A mouse resize updates the preferred size to the user's selected dimensions, so that size is restored after subsequent viewport changes. A viewport smaller than 10×3 may temporarily show a smaller float; automatic clamping preserves the preferred size.

A float created before its tab receives a viewport keeps its preferred dimensions. The first resize or render clamps the visible rectangle to the known viewport.

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

`applyResizeWithin` enforces the minimum dimensions when the viewport allows them; in a smaller viewport it clamps the visible size to the available cells. The eight edge strings share one resize calculation. A manual resize records its resulting dimensions as the new preferred size.

## ANSI safety

Warp renders styled text with lipgloss and Bubble Tea, so raw byte
offsets do not map to visual columns. Two helpers make the float safe to
overlay:

### `StripANSI(s string) string`

Removes ANSI terminal control sequences from a string. The result is
plain text and can safely be measured with `lipgloss.Width` or
`ansi.StringWidth`.

### `overlayFloat(lines []string, fp *FloatPane, totalW, totalH int)`

Draws `fp.render(totalW, totalH)` on top of the existing `lines` without
disturbing the rest of the screen. For every row it:

1.  Locates visual column `fp.X` using ANSI- and grapheme-aware width
    measurement; wide clusters are not split.
2.  Truncates the styled float line to the visible viewport width and
    adds an ANSI reset before the uncovered suffix so styles cannot leak.
3.  Skips the original bytes that the float visually covers.
4.  Appends the remaining original suffix verbatim.


## Style tokens used by the float

The float renders through four `lipgloss` styles defined in `styles.go`.
They use the package theme colors and can be changed through
[theme.go](./theme.md).

- `floatBorderStyle` — border characters (`╭│─╮╰╯`).
- `floatTitleStyle` — title text; bold, background and foreground from
  the theme.
- `floatCloseStyle` — the `×` glyph; bold foreground.
- `floatBgStyle` — opaque background for the close-cell and corner
  cells.

## Example

```go
// detailsPanel implements warp.Panel.
tab := warp.NewTab("details")
tab.Float(detailsPanel, 10, 5, 30, 10)
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
  accumulates rounding error. Automatic viewport clamping changes only
  visible dimensions; manual resizing updates the preferred dimensions.
- `render` is the only function that produces styled bytes;
  `overlayFloat` and the mouse path only consume it.
