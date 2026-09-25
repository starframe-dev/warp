# popover

Package `popover` implements a context-menu popover overlay for a
Bubbletea TUI. The popover renders on top of existing content lines
(without scrolling the underlying view) and supports mouse and keyboard
interaction: hover highlighting, click-to-activate, Enter/arrow/Esc
handling, and automatic closing.

## Public API

``` typescript
type PopoverItem struct
type Popover struct
func (p *Popover) Overlay(lines []string, totalW, totalH int) []string
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool
func (p *Popover) HandleKey(msg tea.KeyMsg) bool
```

## Types

### PopoverItem

``` typescript
type PopoverItem struct {
    Name   string
    Action func()
}
```

Describes a single entry in the menu. `Name` is the label to render;
`Action` is optional and is invoked on click or Enter only when non-nil.
The item still closes the popover when its action is nil.

### Popover

``` typescript
type Popover struct {
    Items   []PopoverItem
    X, Y    int
    Width   int
    OnClose func()
    boxW, boxH int
    clampedX, clampedY int
    selected int
}
```

Fields:

- `Items` — the menu entries to render and activate.
- `X, Y` — requested screen position (row 0 corresponds to the header
  row).
- `Width` — desired content width; `0` means auto (default 20).
- `OnClose` — callback invoked when the popover closes.

Remaining fields (`boxW/boxH`, `clampedX/clampedY`, `selected`) are
internal state maintained by the methods below.

## Behavior

### Overlay

``` typescript
func (p *Popover) Overlay(lines []string, totalW, totalH int) []string
```

Renders the menu box on top of the given content lines and returns the
overlaid lines. If `Items` is empty, the input is returned unchanged.

The method:

1.  Builds content lines (one per item), applying selected/base styles
    and width.
2.  Wraps the menu in a lipgloss normal-border box.
3.  Computes the box dimensions and stores them in `boxW`/`boxH`.
4.  Clamps the placement to stay within the terminal bounds and stores
    `clampedX`/`clampedY`.
5.  Splices each box line into the original lines, preserving content to
    the left and right of the box (using `ansi.Truncate` and a visual
    byte-position helper).

### HandleMouse

``` typescript
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool
```

Consumes mouse events for the popover. Returns `true` when the event is
handled and the caller should suppress default handling.

- Press on an item row inside the box: calls its `Action` if non-nil and closes the popover. `OnClose`, when set, is called even when the action is nil.
- Press outside: closes the popover.
- Motion: updates hover highlighting by tracking the item under the
  cursor.
- Release: always consumed.

Uses the clamped position from the last `Overlay` call for hit-testing,
falling back to `X/Y` if unset.

### HandleKey

``` typescript
func (p *Popover) HandleKey(msg tea.KeyMsg) bool
```

Consumes keyboard events:

- `Esc` — closes the popover.
- `Enter` — calls the selected item's `Action` if non-nil, then closes.
- `Up` / `Down` — move the selection, clamped to the item list.

Any other key returns `false` (not consumed).

## Styling

Two lipgloss styles are used:

- `popoverBaseStyle` — normal item background/foreground.
- `popoverSelectedStyle` — highlighted selection state.

Box border, background, and foreground colors are configured in
`Overlay` using the module's shared color constants (`gbDark4`,
`gbDark1`).

## Implementation notes

The popover composes the box with `lipgloss`, and splices it into
existing lines without shifting them, preserving content left/right.
Hit-testing uses the clamped position stored during `Overlay`; the mouse
coordinate space is assumed to match the screen rows/columns.
