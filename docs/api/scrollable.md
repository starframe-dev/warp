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
| `Offset` | `int` | Current scroll offset in lines. Negative values are clamped to 0; values above the maximum are clamped to `len(lines) - h`. |

### NewScrollable

``` go
func NewScrollable(content Panel) *Scrollable
```

Creates a new `Scrollable` wrapping the given `Panel`. The initial
`Offset` is zero.

### View

``` go
func (s *Scrollable) View(w, h int) string
```

Renders the visible viewport. The full content is rendered once at the
requested width with an effectively unlimited height, split into lines,
and then sliced to show only `h` lines starting at `s.Offset`. Lines
beyond the content are padded with spaces to fill the viewport.

### Update

``` go
func (s *Scrollable) Update(msg tea.Msg) tea.Cmd
```

Handles incoming Bubble Tea messages. It adjusts `s.Offset` for mouse
wheel and keyboard scroll messages, then forwards the message to
`s.Content` if non-nil.

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

After handling a scroll message, the offset is clamped in `View`:

``` go
maxOffset := len(lines) - h
if maxOffset < 0 { maxOffset = 0 }
if s.Offset < 0 { s.Offset = 0 }
if s.Offset > maxOffset { s.Offset = maxOffset }
```

This ensures the viewport never scrolls past the end of the content.

## Implementation details

- **Rendering strategy:** the inner `Panel` is rendered with
  `View(w, 9999)`, i.e. a very large height, so that its full content is
  produced in one pass. The result is split on `"\n"` into lines.
- **Line padding/truncation:** each visible line is passed through the
  package-level helper `padLine`, which uses `lipgloss.Width` to
  determine display width. Lines shorter than `w` are right-padded with
  spaces; lines longer than `w` are truncated at a byte boundary that
  keeps the visible width within the column budget.
- **Message forwarding:** any message that does not affect scrolling
  (i.e. non-mouse, non-key, or keys other than up/down/pgup/pgdown) is
  forwarded to `s.Content.Update(msg)`. If `s.Content` is nil, no
  command is returned.
