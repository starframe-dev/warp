# Collapsible

`Collapsible` is a bubbletea panel wrapper in the `warp` package. It
wraps any `Panel` and adds collapse/expand behavior. When expanded, it
delegates all rendering and message handling to the inner `Content`
panel. When collapsed, it renders a single-line title bar built from an
indicator glyph, a (possibly truncated) title, and a decorative
horizontal rule.

## Public API

### Type

``` go
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel
}
```

| Field | Type | Description |
|----|----|----|
| `Title` | `string` | Display title shown in the collapsed title bar. Truncated with `...` when it does not fit the available width. |
| `Collapsed` | `bool` | Whether the panel is currently collapsed to the title bar. Public so callers can read or set state directly. |
| `Content` | `Panel` | The inner content panel rendered when expanded. May be `nil`. |

### NewCollapsible

``` go
func NewCollapsible(title string, content Panel) *Collapsible
```

Creates a new `Collapsible` with the given `title` and `content`. The
panel starts in the expanded state (`Collapsed == false`). Returns a
pointer; callers should hold it and drive state via [Toggle](#toggle).

### View

``` go
func (c *Collapsible) View(w, h int) string
```

Renders the panel at width `w` and height `h`. Behavior:

- If `Collapsed` is `true`, renders the single-line collapsed title bar
  (see [renderCollapsed](#renderCollapsed) below).
- If `Collapsed` is `false` and `Content != nil`, returns
  `c.Content.View(w, h)`.
- If `Content == nil`, returns `""`.

In collapsed mode the `h` parameter is ignored; the bar is one line
regardless of height.

### Update

``` go
func (c *Collapsible) Update(msg tea.Msg) tea.Cmd
```

Forwards a bubbletea message to the inner `Content` panel and returns
whatever command that handler produces. Returns `nil` if `Content` is
`nil`.

`Collapsed` state is *not* driven by messages — the caller owns toggling
via [Toggle](#toggle).

### Toggle

``` go
func (c *Collapsible) Toggle()
```

Flips `Collapsed`. This is the only state-mutating method on the type.
Repeated calls toggle back and forth; a `nil` `Content` does not prevent
toggling.

## renderCollapsed

``` go
func (c *Collapsible) renderCollapsed(w int) string
```

Unexported helper that builds the single-line collapsed title bar.
Layout, left to right:

- A left cell rendered with `collapsibleStyle` containing `┌`, the
  indicator glyph (`▶` when collapsed, `▼` when expanded), a space, and
  the (possibly truncated) `Title`.
- A right cell rendered with `collapsibleBorderStyle` containing a run
  of `─` characters followed by the `┐` top-right corner. The run length
  is `padding = w − len(title) − 5`, clamped to zero.

Truncation reserves 5 columns (indicator, space, and corner characters)
from the available width. If the remaining title length would be
negative, the title is emitted empty and the bar still renders. A
non-positive `w` returns an empty string.

## Null Content Semantics

A `Collapsible` with `Content == nil` behaves as a pure state holder: it
still toggles and renders a collapsed bar, but when expanded `View`
returns `""` and `Update` swallows every message. This is intentional —
the wrapper must be safe to embed in any tree without an inner panel
being constructed yet.

## Styling

The collapsed bar is assembled from two package-level lipgloss styles
defined in `styles.go` and re-bound by `theme.go`:

- `collapsibleStyle` — foreground `gbLight1`, background `gbDark1`.
  Applies to the left cell containing the corner, indicator, and title.
- `collapsibleBorderStyle` — foreground `gbDark4`. Applies to the right
  cell containing the `─` run and the `┐` corner.

## Example Usage

``` go
col := NewCollapsible("Settings", settingsPanel)

// Collapse
col.Toggle()

// Render the collapsed title bar at width 80
out := col.View(80, 1)

// Forward a message to the inner panel
cmd := col.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
```

## Dependencies

`Collapsible` depends on `github.com/charmbracelet/bubbletea` for the
`tea.Msg` and `tea.Cmd` types, the package-level `Panel` interface, and
the `collapsibleStyle` / `collapsibleBorderStyle` lipgloss values from
`styles.go`.
