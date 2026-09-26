# Scrollable

`Scrollable` is a Bubble Tea UI component that wraps any `Panel` with
scroll support. When the rendered content exceeds the allocated height,
the visible portion is limited to `h` lines and the user can scroll
through the rest using a mouse wheel or keyboard.

## Public API

``` go
type Scrollable struct {
    Content Panel
    Offset  int
}
```

| Field | Type | Description |
|----|----|----|
| `Content` | `Panel` | The inner panel whose content is rendered and scrolled. |
| `Offset` | `int` | Requested scroll offset in lines. Negative values are treated as 0; when a content extent is known, the effective offset is clamped to the last full viewport. |

### NewScrollable

``` go
func NewScrollable(content Panel) *Scrollable
```

Creates a new `Scrollable` wrapping the given `Panel`. The initial
`Offset` is zero.

### ContentHeight

```go
func (s *Scrollable) ContentHeight(width int) (height int, known bool)
```

Returns `(0, false)`. A nested `Scrollable` does not expose its inner
content as an intrinsic extent; its allocated viewport does not define a
reliable outer content height.

### View

``` go
func (s *Scrollable) View(w, h int) string
```

Renders the visible viewport using the same effective offset as
`Elements`. When the content has a known intrinsic extent and implements
`ViewportRenderer`, Scrollable calls `ViewAt(w, h, effectiveOffset)` so the
panel need not render the prefix. Otherwise, known extents request through
`effectiveOffset + h`; unknown extents probe one row beyond that end and use
the probe output. Height arithmetic saturates on integer overflow. Lines
beyond the returned content are padded to fill the viewport. With nil
content, returns exactly `h` blank lines (`""` for `h == 0`; otherwise
`h-1` newline characters).

### Elements

``` go
func (s *Scrollable) Elements(w, h int) []Element
```

Uses the same effective offset as `View`. When the content has a known
intrinsic extent and implements `ViewportElementProvider`, it requests only
that viewport. Otherwise it requests semantic elements through
`effectiveOffset + h`. Elements are clipped to the visible viewport, then
the effective offset is subtracted from their Y coordinates. A viewport
provider returns bounds relative to the full content origin. Fully offscreen
elements are omitted and partially visible bounds are clipped. Inspection
does not mutate `s.Offset`.

### Optional viewport fast path

A content panel can implement `ViewportRenderer` and/or
`ViewportElementProvider` in addition to `ContentHeightProvider`. Scrollable
uses these interfaces only when the content extent is known. `ViewAt` returns
only the requested rows; `ElementsAt` returns only intersecting elements with
bounds in full-content coordinates. Other panels retain the legacy fallback.

### Update

``` go
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```

Handles incoming Bubble Tea messages. It adjusts `s.Offset` for mouse
wheel and keyboard scroll messages, remembers viewport dimensions from
`ResizeMsg` (or provisional `WindowSizeMsg`), and normalizes the stored
offset after scrolling or a viewport change when dimensions are known. It
forwards every message—including
scroll messages—to `s.Content` if non-nil.

## Scrolling behavior

The `Update` method handles two message types:

- `tea.MouseMsg`:
  - `MouseButtonWheelUp` — decrements offset by 3 (clamped at 0)
  - `MouseButtonWheelDown` — increments offset by 3
- `tea.KeyMsg`:
  - `"up"` — decrements offset by 1 (clamped at 0)
  - `"down"` — increments offset by 1
  - `"pgup"` — decrements offset by 10 (clamped at 0)
  - `"pgdown"` — increments offset by 10

When intrinsic height is known, the effective offset is
`min(max(0, Offset), max(0, contentHeight - viewportHeight))`. `View` and
`Elements` use that same value; `Elements` leaves the stored `Offset`
unchanged.

## Implementation details

- **Extent resolution:** an explicit `ContentHeightProvider` is preferred.
  Without a known extent, Scrollable probes `View` at one row beyond the
  required visible end. Fewer returned rows discover the end; exactly the
  requested number means the extent remains unknown. A panel that pads
  every request therefore has unbounded fallback scrolling. Accurate
  bounded scrolling requires an intrinsic-height provider.
- **Line padding/truncation:** each visible line is passed through
  `padLine` / `padVisualLine`, which use ANSI-aware terminal-cell widths
  and truncate without splitting graphemes. Lines shorter than `w` are
  right-padded with spaces.
- **Message forwarding:** every message is forwarded to
  `s.Content.Update(msg)`, including messages handled for scrolling. If
  `s.Content` is nil, no command is returned.
