# `dropdown.go` Specification

## Overview

`dropdown.go` implements a single-select dropdown menu component for the `warp` terminal UI library. It is built on top of [Bubble Tea](https://github.com/charmbracelet/bubbletea) (`tea`) and uses [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling. The component renders either a collapsed button or an expanded list of selectable items, and supports both mouse and keyboard interaction.

## Package

```go
package warp
```

## Imports

```go
import (
    "strings"
    tea "github.com/charmbracelet/bubbletea"
)
```

- `strings` — used to join rendered lines of the open menu.
- `github.com/charmbracelet/bubbletea` — used for `tea.Msg`, `tea.KeyMsg`, `tea.MouseMsg`, and `tea.Cmd`.

## Public Types

### `DropdownItem`

Represents a single entry in the dropdown list.

```go
type DropdownItem struct {
    Label    string
    Selected bool
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Label` | `string` | Display text for the item. |
| `Selected` | `bool` | Whether this item is currently selected. Only one item is selected at a time. |

### `DropdownMenu`

The dropdown component. It tracks its label, items, open/closed state, hovered index, and an optional selection callback.

```go
type DropdownMenu struct {
    Label    string
    Items    []DropdownItem
    Open     bool
    Hovered  int // index of hovered item, -1 if none
    OnSelect func(idx int)
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Label` | `string` | Text shown on the collapsed button and at the top of the open menu. |
| `Items` | `[]DropdownItem` | The list of items to display. |
| `Open` | `bool` | Whether the menu is currently expanded. |
| `Hovered` | `int` | Index of the keyboard- or mouse-hovered item. `-1` means no item is hovered. |
| `OnSelect` | `func(idx int)` | Optional callback invoked when an item is selected. |

## Public API

### `NewDropdownMenu`

```go
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu
```

Creates a new `DropdownMenu` with the given button label and item list. The new menu is closed (`Open == false`) and no item is hovered (`Hovered == -1`).

### `(*DropdownMenu) View`

```go
func (d *DropdownMenu) View(w, h int) string
```

Renders the dropdown. If `d.Open` is `false`, it renders the collapsed button. If `true`, it renders the expanded menu constrained to the supplied width `w` and height `h`.

**Button rendering (`renderButton`)**
- Displays `Label + " ▼"`.
- Truncates with `"…"` if the label exceeds `w`.
- Pads the label to width `w` and applies `dropdownButtonStyle`.

**Menu rendering (`renderMenu`)**
- Computes `menuH = len(d.Items) + 1` (one line for the button plus one line per item), capped at `h`.
- Line 0 is the button rendered as `Label + " ▲"` with `dropdownButtonStyle`.
- Items are rendered on subsequent lines.
- Each item line has the prefix `"✓ "` if selected, otherwise `"  "`.
- Item labels are truncated with `"…"` if they exceed `w`.
- Items are padded to width `w` and styled as follows:
  - `dropdownItemHoverStyle` if `i == d.Hovered` (hover takes precedence).
  - `dropdownItemSelectedStyle` if `item.Selected`.
  - `dropdownItemStyle` otherwise.
- If the available height is smaller than the number of items, only the lines that fit are rendered.

### `(*DropdownMenu) Update`

```go
func (d *DropdownMenu) Update(msg tea.Msg) tea.Cmd
```

Handles Bubble Tea messages. Always returns `nil` (no commands produced).

**Mouse handling (`tea.MouseMsg`)**
- Only reacts to `tea.MouseActionPress` events; other mouse actions are ignored.
- When closed, clicking on row 0 opens the menu and resets `Hovered` to `-1`.
- When open:
  - Clicking on row 0 closes the menu.
  - Clicking on row `msg.Y` where `1 <= msg.Y <= len(d.Items)` selects the item at index `msg.Y - 1`.

**Keyboard handling (`tea.KeyMsg`)**
- Keys are only processed when the menu is open.
- Supported keys:
  - `"up"` — moves `Hovered` up by one, clamped at `0`.
  - `"down"` — moves `Hovered` down by one, clamped at `len(d.Items) - 1`.
  - `"enter"` — selects the currently hovered item if `Hovered >= 0`.
  - `"esc"` — closes the menu without selecting.

### `(*DropdownMenu) Close`

```go
func (d *DropdownMenu) Close()
```

Closes the dropdown by setting `d.Open = false`.

## Important Implementation Details

### Selection semantics (`selectItem`)

```go
func (d *DropdownMenu) selectItem(idx int)
```

This private method is invoked by both mouse clicks and keyboard `enter`:
1. Clears `Selected` on all items.
2. Sets `d.Items[idx].Selected = true`.
3. Closes the menu (`d.Open = false`).
4. Invokes `d.OnSelect(idx)` if the callback is non-nil.

The component enforces a single-selection model: selecting an item deselects all others.

### Styling

The dropdown uses the following package-level styles declared in `styles.go`:

| Style | Usage |
|-------|-------|
| `dropdownButtonStyle` | Background `gbDark2`, foreground `gbLight1`. Used for the collapsed button and the open menu header. |
| `dropdownItemStyle` | Background `gbDark0`, foreground `gbLight1`. Default item style. |
| `dropdownItemHoverStyle` | Background `gbDark2`, foreground `gbYellow`. Used for the currently hovered item. |
| `dropdownItemSelectedStyle` | Background `gbDark2`, foreground `gbGreen`, bold. Used for the selected item. |

Hover styling takes precedence over selected styling when rendering an item.

### Text layout helpers

- `padRight(s string, w int) string` (defined in `tabgroup.go`) is used to pad each rendered line to the supplied width `w`.
- Labels and items are truncated to `w - 1` characters with a trailing `"…"` if they are too long.

### Coordinate assumptions

Mouse coordinates (`msg.Y`) are interpreted relative to the top-left of the dropdown's bounding box. Row 0 is always the button, and item `i` is at row `i + 1`.

### State rules

- `Hovered` is reset to `-1` when the menu opens via mouse click.
- `Hovered` does not automatically wrap; moving up at the top item or down at the bottom item has no effect.
- Keyboard navigation is only available while the menu is open.
- `Open` can be toggled programmatically by setting the field or calling `Close()`.
