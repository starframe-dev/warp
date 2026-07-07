# Popover

`popover.go` implements a context-menu-style popover that renders over an existing screen and can be driven by mouse and keyboard events.

## Overview

A `Popover` is a modal-ish menu rendered at a fixed `(X, Y)` position in screen coordinates. It composes its own bordered box over the supplied content lines, preserving the original content to the left and right of the menu so the underlying UI remains visible. It supports:

- Mouse click-to-select and hover highlighting
- Keyboard navigation (up/down/enter/esc)
- A configurable close callback
- Automatic clamping so the menu stays within the terminal bounds

## Types

### `PopoverItem`

A single menu entry.

```go
type PopoverItem struct {
    Name   string
    Action func()
}
```

- `Name` – the text shown in the menu.
- `Action` – the callback invoked when the item is selected by click or enter.

### `Popover`

```go
type Popover struct {
    Items      []PopoverItem
    X, Y       int
    Width      int
    OnClose    func()

    // internal state
    boxW, boxH int
    selected   int
}
```

- `Items` – the menu entries to display.
- `X, Y` – top-left anchor position in screen coordinates (0 corresponds to the header row). These are the same coordinates used by mouse messages.
- `Width` – requested content width of the menu, excluding the border. If `0`, the default width of 20 columns is used. It is clamped to `totalW`.
- `OnClose` – callback invoked when the popover should close (outside click, escape, or item selection).
- `boxW`, `boxH` – rendered dimensions of the bordered box, set by `Overlay` and used by input handlers for hit testing.
- `selected` – index of the currently highlighted item.

## Public API

### `(*Popover) Overlay(lines []string, totalW, totalH int) []string`

Renders the popover over the provided `lines` and returns the overlaid lines. The input slice is mutated in place.

Behavior:

1. Returns `lines` unchanged if there are no items or no rendered box lines.
2. Determines the content width, defaulting to `20` and clamping to `totalW`.
3. Renders each item, with the selected item using `popoverSelectedStyle` and others using `popoverBaseStyle`. Each line is padded/truncated to the content width.
4. Joins the content lines and wraps them in a bordered box using `lipgloss.NormalBorder` with `gbDark4` foreground and `gbDark1` background.
5. Stores the resulting box width and height in `p.boxW` and `p.boxH`.
6. Clamps the menu position so the entire box fits inside `totalW` × `len(lines)`.
7. For each overlaid row, computes the left and right portions of the original line using ANSI-aware truncation (`ansi.Truncate`) and a `visualBytePos` helper, then centers the box line between them. The left portion is padded with spaces to reach the exact visual width of `menuX`.

### `(*Popover) HandleMouse(msg tea.MouseMsg) bool`

Handles mouse events for the popover. Returns `true` if the event was consumed.

- `MouseActionPress`:
  - If the click is inside the box, it computes the item index from the Y position (accounting for the top border), invokes the item action, and then calls `OnClose`.
  - If the click is outside the box, it calls `OnClose`.
  - Always returns `true`.
- `MouseActionMotion`:
  - Updates `selected` to the item under the cursor, ignoring the top and bottom border rows.
  - Always returns `true`.
- `MouseActionRelease`:
  - Returns `true`.
- Any other action returns `false`.

The handler currently uses the unclamped `X` and `Y` for hit testing, but compares against the actual rendered box dimensions stored in `boxW`/`boxH`.

### `(*Popover) HandleKey(msg tea.KeyMsg) bool`

Handles keyboard events for the popover. Returns `true` if the event was consumed.

- `Esc` – closes the popover.
- `Enter` – invokes the currently selected item's action and closes the popover.
- `Up` – decrements the selection, clamping at `0`.
- `Down` – increments the selection, clamping at `len(Items)-1`.
- All other keys return `false`.

If `boxW` is `0`, meaning `Overlay` has not yet been called, both `HandleMouse` and `HandleKey` return `false`.

## Styles

```go
var (
    popoverBaseStyle = lipgloss.NewStyle().
        Background(lipgloss.Color(gbDark1)).
        Foreground(lipgloss.Color(gbLight1))

    popoverSelectedStyle = lipgloss.NewStyle().
        Background(lipgloss.Color(gbBlue)).
        Foreground(lipgloss.Color(gbDark0))
)
```

- `popoverBaseStyle` – normal item background (`gbDark1`) with light text (`gbLight1`).
- `popoverSelectedStyle` – highlighted item background (`gbBlue`) with dark text (`gbDark0`).

Both styles are applied with an explicit width equal to the content width, and each item is rendered with a leading space.

## Important Implementation Details

- `Overlay` mutates the input `lines` slice. Callers should pass a copy if the original content must be preserved.
- The menu box includes a two-character border on both sides, so the actual screen width used by the box is `contentW + 2` (depending on lipgloss output). The stored `boxW` is derived from the rendered string width, so it is exact.
- Position clamping is performed against `totalW` for the horizontal axis and `len(lines)` for the vertical axis. This assumes the supplied `lines` represent the full terminal height.
- The hit-test logic in `HandleMouse` assumes the item content starts at `Y + 1` and ends at `Y + boxH - 1`, which matches the single-line top and bottom borders produced by `lipgloss.NormalBorder`.
- The overlay logic uses `visualBytePos` (defined elsewhere in the package) to map a visual column to a byte offset in a string that may contain ANSI escape sequences. This preserves any styling sequences in the right portion of the original line.
- `OnClose` may be called both when the user explicitly selects an item and when the user dismisses the popover.
