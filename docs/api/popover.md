# Popover

`Popover` is a context-menu overlay in the `warp` package. It renders over existing content lines without inserting rows and supports mouse and keyboard interaction.

## Public API

```go
type PopoverItem struct {
    Name   string
    Action func()
}

type Popover struct {
    Items   []PopoverItem
    X, Y    int
    Width   int
    OnClose func()
}

func (p *Popover) Overlay(lines []string, totalW, totalH int) []string
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool
func (p *Popover) HandleKey(msg tea.KeyMsg) bool
```

`Popover` also contains unexported fields used to track rendered dimensions, the clamped position, and the selection. `tea.MouseMsg` and `tea.KeyMsg` are from Bubble Tea.

## Fields

- `Items` — menu entries. Each item's `Name` is sanitized before rendering; its `Action` is optional.
- `X`, `Y` — requested screen position. Row 0 corresponds to the header row.
- `Width` — desired content width. A value less than or equal to zero uses 20; the width is capped at `totalW`.
- `OnClose` — optional callback invoked when a supported close interaction occurs.

## Behavior

### `Overlay`

```go
func (p *Popover) Overlay(lines []string, totalW, totalH int) []string
```

If `Items` is empty, returns `lines` unchanged. Otherwise it renders one styled content line per item, wraps them in a normal-border box, and overlays that box on the input lines while preserving visible content to its left and right.

The box's horizontal position is clamped using `totalW`, and its vertical position is clamped using `len(lines)`. The resulting box dimensions and position are stored for subsequent event handling. `totalH` is accepted but is not used by the implementation. Overlaying does not add lines; box lines beyond the input slice are skipped.

### `HandleMouse`

```go
func (p *Popover) HandleMouse(msg tea.MouseMsg) bool
```

Returns `false` until a non-empty menu has been rendered (the stored box width is zero); otherwise:

- A press on an item row invokes its non-nil `Action`, then invokes `OnClose` if set.
- A press elsewhere, including on the box border, invokes `OnClose` if set. Presses inside the box are consumed even when they are not on an item row.
- Motion updates the selection when the pointer is over an item row. Motion is consumed even when outside the item rows.
- Release is consumed.

Hit-testing uses the clamped position saved by the last `Overlay` call. If a saved coordinate is negative, the corresponding requested `X` or `Y` is used instead.

### `HandleKey`

```go
func (p *Popover) HandleKey(msg tea.KeyMsg) bool
```

Returns `false` until a non-empty menu has been rendered. Otherwise:

- `Esc` invokes `OnClose` if set.
- `Enter` invokes the selected item's non-nil `Action`, then invokes `OnClose` if set.
- `Up` and `Down` change the selection, clamped to the available item indices.
- Other keys return `false`.

Handled keys return `true`. The initial selection is index 0.

## Styling

Items use base or selected Lip Gloss styles. The menu box uses a normal border, with colors taken from the package's shared `gbDark4` and `gbDark1` constants.
