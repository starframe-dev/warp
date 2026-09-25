# Dropdown Menu (dropdown.go)

The `dropdown.go` file implements a reusable TUI dropdown menu component
built on top of the bubbletea framework. It models a classic "button +
expandable list" interaction: a collapsed button that, when activated,
expands into a vertically stacked list of items, one of which can be
selected.

## Public API

### `DropdownItem` (struct)

Represents a single entry in the dropdown.

``` typescript
type DropdownItem struct {
    Label    string
    Selected bool
}
```

### `DropdownMenu` (struct)

Holds the full state of the dropdown: its label, the list of items, the
open/close flag, the index of the hovered item, and the selection
callback.

``` typescript
type DropdownMenu struct {
    Label    string
    Items    []DropdownItem
    Open     bool
    Hovered  int // index of hovered item, -1 if none
    OnSelect func(idx int)
}
```

### `NewDropdownMenu`

Constructor that builds a closed dropdown menu.

``` typescript
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu
```

- *label* — text rendered on the collapsed button.
- *items* — slice of `DropdownItem` values to show inside the open menu.

### `DropdownMenu.View(w, h int) string`

Renders the dropdown as a multi-line string. The method switches between
the collapsed button and the open menu based on `d.Open`, and clamps the
height to the available `h` lines.

### `DropdownMenu.Update(msg tea.Msg) tea.Cmd`

Handles mouse and keyboard messages for the dropdown:

- Mouse press on the button (row *y = 0*) toggles open/close.
- Mouse motion over a visible item row updates `Hovered`; motion over the button, outside the menu, or over a clipped row resets it to `-1`. Hovering does not select an item.
- Mouse press on an item row selects that item.
- Keyboard *up*/*down* moves `Hovered` only among visible rows; *enter*
  selects only a visible hovered item; *esc* closes the menu. Until the
  open menu is rendered, the previous full-list fallback is retained.

### `DropdownMenu.Close()`

Programmatically closes the menu and resets its open state.

## Behavior

### State Machine

The dropdown has two visible states driven by `Open`:

- **Closed**: only the button row is rendered, showing `Label + " ▼"`.
- **Open**: the button row shows `Label + " ▲"` and the item list is
  rendered below it. Hover and selection state are tracked with
  `Hovered` (index, `-1` for none) and each `DropdownItem.Selected`
  flag.

### Mouse Interaction

- If the menu is closed and the user presses on row `0` (the button),
  `Open` is set to `true` and `Hovered` is reset to `-1`.
- While the menu is open, motion over a visible item row `y` sets `Hovered = y-1`; motion on the button row, outside the menu, or over a clipped row sets `Hovered = -1`. Motion never selects an item.
- If the menu is open and the user presses on row `0`, the menu closes.
- If the menu is open and the user presses on a row inside the item
  list, the corresponding item is selected (its `Selected` flag is set,
  other items are cleared, `Open` is set to `false`, and `OnSelect` is
  invoked if non-nil).

### Keyboard Interaction

- *up*/*down*: move `Hovered` within the currently visible item bounds.
- *enter*: selects the hovered item only when its index is currently
  visible; clipped items cannot be selected by keyboard.
- *esc*: closes the menu without selecting anything.

### Selection Semantics

Selecting an item (via mouse or *enter*) clears the `Selected` flag on
all items, sets it on the chosen one, closes the menu, and invokes
`OnSelect(idx)`. The `OnSelect` field is optional — if it is `nil`, no
callback is invoked.

## Implementation Notes

- Rendering is done through unexported helpers `renderButton(w)` and
  `renderMenu(w, h)`, which are called by `View`. The menu height is
  clamped to `h` so the dropdown never overflows the available vertical
  space.
- Labels are truncated with a `…` ellipsis when longer than the
  available width `w`.
- Row alignment uses a `padRight(label, w)` helper (defined elsewhere in
  the package) to right-pad labels to the column width.
- Styles are applied per-row using `dropdownButtonStyle`,
  `dropdownItemStyle`, `dropdownItemHoverStyle`, and
  `dropdownItemSelectedStyle` (defined elsewhere in the package). The
  currently hovered row uses the hover style, and a selected item uses
  the selected style.
- Mouse motion updates hover state and mouse press handles open/close or selection; other mouse actions are ignored. `Update` always returns `nil` because the dropdown does not schedule follow-up commands.

## Integration with bubbletea

The component is composed into a larger model: the host model must call
`View(w, h)` during its own `View()` (passing the region width/height)
and forward relevant `tea.Msg` values into `Update`. The host model is
responsible for placing the returned string into the final frame and for
calling `Close()` when it wants to dismiss the dropdown from outside of
the component.

## Example

``` go
menu := NewDropdownMenu("Color", []DropdownItem{
    {Label: "Red", Selected: false},
    {Label: "Green", Selected: true},
    {Label: "Blue", Selected: false},
})
menu.OnSelect = func(idx int) {
    fmt.Println("selected:", menu.Items[idx].Label)
}
```
