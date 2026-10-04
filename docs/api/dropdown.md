# Dropdown Menu (`dropdown.go`)

The `dropdown.go` file implements a reusable TUI dropdown menu component built on the Bubble Tea framework. It models a button and an expandable list of items, one of which can be selected.

## Public API

### `DropdownItem`

Represents a single entry in the dropdown.

```go
type DropdownItem struct {
    Label    string
    Selected bool
}
```

### `DropdownMenu`

Holds the dropdown state and its optional selection callback.

```go
type DropdownMenu struct {
    Label    string
    Items    []DropdownItem
    Open     bool
    Hovered  int // index of hovered item, -1 if none
    OnSelect func(idx int)
}
```

The implementation also stores private layout state used to limit interaction to visible menu items after the open menu has been rendered.

### `NewDropdownMenu`

```go
func NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu
```

Creates a closed menu with the supplied label and items. `Hovered` starts at `-1`.

### `DropdownMenu.View(w, h int) string`

Renders the collapsed button or open menu according to `Open`. Negative width and height are treated as zero. The open menu contains at most `h` rows, including its button row; the collapsed button renders as one row.

### `DropdownMenu.Elements(w, h int) []Element`

Returns semantic elements for the button and, when open, the visible options. Returns `nil` if `w` or `h` is zero or negative. Otherwise, it exposes the button as a `combobox` with action `toggle` and bounds `(0, 0, w, 1)`. When open, it also exposes at most `h-1` options as `option` elements with action `select`; clipped options are omitted.

### `DropdownMenu.Update(msg tea.Msg) tea.Cmd`

Handles mouse and keyboard messages. It returns `nil` for all messages.

- A mouse press on row `y = 0` opens a closed menu or closes an open one. Mouse handling checks the row (`Y`), not the horizontal coordinate.
- While open, mouse motion sets `Hovered` to `y-1` for a visible item row and resets it to `-1` for the button row or any non-visible row. Motion does not select an item.
- A mouse press on a visible item row selects that item.
- While open, keyboard `up` and `down` move `Hovered` among visible items; `enter` selects a visible hovered item; `esc` closes the menu.
- Before the first open-menu render, the visible item count is unknown, so keyboard and mouse selection use the full item list as a fallback.

### `DropdownMenu.Close()`

Closes the menu by setting `Open` to `false`. It does not change the hover or selection state.

## Behavior

### Rendering and state

- When closed, the button shows `Label + " ▼"`; when open, it shows `Label + " ▲"` above the items.
- Labels are sanitized and truncated with an ellipsis when wider than the available width. Rows are padded to the requested width.
- The selected item is rendered with the selected style, which takes precedence over the hover style when an item is both selected and hovered. Other hovered items use the hover style.
- An open-menu render records how many item rows fit, and subsequent interaction is limited to those visible rows.

### Selection

Selecting an item clears `Selected` on all items, sets it on the chosen item, closes the menu, and invokes `OnSelect(idx)` if the callback is non-nil. Pressing `esc` or calling `Close()` does not select an item.

## Integration with Bubble Tea

The host model should call `View(w, h)` while rendering its own view and forward relevant `tea.Msg` values to `Update`. The host places the returned string in the final frame and may call `Close()` to dismiss the dropdown.

## Example

```go
menu := NewDropdownMenu("Color", []DropdownItem{
    {Label: "Red", Selected: false},
    {Label: "Green", Selected: true},
    {Label: "Blue", Selected: false},
})
menu.OnSelect = func(idx int) {
    fmt.Println("selected:", menu.Items[idx].Label)
}
```
