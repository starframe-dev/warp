# theme.go

## Overview

`theme.go` provides the public runtime API for changing the color theme
of the Warp terminal UI package. It defines `ThemeColors`, a flat struct
of semantic color names, and `SetTheme`, which maps those colors onto
every package-level lipgloss style and color variable used by overlays
and controls (tabs, borders, float panes, dropdowns, popovers, modals,
inputs, collapsibles, dim text, and more).

Warp ships with a default Gruvbox Dark palette (declared in
`styles.go`). The variables in `styles.go` and the other widget files
(`input.go`, `modal.go`, `popover.go`, `dropdown.go`, etc.) are
initialized with that palette, and `SetTheme` rebuilds them in place at
runtime so every existing widget adopts the new colors immediately.

## Public API

### ThemeColors

`type ThemeColors struct { Background string Surface string Raised string Border string BorderMuted string Text string TextMuted string TextStrong string Accent string AccentMuted string Error string Success string Warning string SelectionBackground string SelectionForeground string } `

`ThemeColors` holds semantic color names as plain strings (CSS color
strings, e.g. `"#282828"`). It is the input contract for `SetTheme`.
There are 15 fields, grouped by purpose:

| Field | Purpose |
|----|----|
| Background | Root window / tab-bar background |
| Surface | Panel, float, modal, and collapsible body |
| Raised | Raised elements (dropdown buttons, dropdown hover, selected dropdown items) |
| Border | Primary border color for panels and modals |
| BorderMuted | Muted border for input borders, dropdown borders |
| Text | Default text color (used where the palette maps to it) |
| TextMuted | Muted / secondary text (inactive tab labels, dim text, float border) |
| TextStrong | Strong / primary text (modal titles, input text, dropdown labels) |
| Accent | Primary accent (focused input border) |
| AccentMuted | Secondary accent |
| Error | Error / close-tab / modal-close foreground |
| Success | Success / "new tab" foreground |
| Warning | Warning / drag-border / dropdown hover foreground |
| SelectionBackground | Active tab background |
| SelectionForeground | Active tab foreground / popover selected foreground |

### SetTheme

`func SetTheme(colors ThemeColors) `

Rebuilds all package-level color and style variables from the supplied
`ThemeColors`. The function performs no error checking and does not
return a value. Any widget rendered after the call — including
already-constructed widgets whose rendering happens lazily at draw time
— picks up the new values.

Internally `SetTheme`:

1.  Converts each hex color string in `ThemeColors` into a
    `lipgloss.Color` and assigns it to the corresponding Gruvbox-palette
    variable (`gbDark0`, `gbDark1`, `gbDark2`, `gbDark3`, `gbDark4`,
    `gbGray`, `gbLight1`, `gbRed`, `gbGreen`, `gbYellow`, `gbBlue`).
2.  Derives all semantic color variables from those palette variables:
    `tabBarBg`, `activeTabBg`, `activeTabFg`, `inactiveTabFg`,
    `newTabFg`, `closeTabFg`, `borderColor`, `borderDragColor`,
    `borderHoverColor`, `floatBorderColor`, `floatTitleBg`,
    `floatTitleFg`, `floatBg`, `floatCloseFg`.
3.  Rebuilds every lipgloss `Style` value that depends on those color
    variables, covering the tab bar, split borders, float panes,
    collapsibles, dropdowns, popovers, modals, dim text, and input
    widgets.

## Usage Example

``` go
package main

import (
    "fmt"

    "github.com/starframe-dev/warp"
)

func main() {
    // Apply a custom theme. Warp's default is Gruvbox Dark;
    // any 15-color palette can be supplied.
    warp.SetTheme(warp.ThemeColors{
        Background:          "#1e1e1e",
        Surface:             "#252525",
        Raised:              "#333333",
        Border:              "#555555",
        BorderMuted:         "#666666",
        Text:                "#cccccc",
        TextMuted:           "#999999",
        TextStrong:          "#eeeeee",
        Accent:              "#4f8aff",
        AccentMuted:         "#7aa0d4",
        Error:               "#e55",
        Success:             "#4c4",
        Warning:             "#ee8811",
        SelectionBackground: "#336699",
        SelectionForeground: "#ffffff",
    })
}
```

## Implementation Details

- **Default palette:** The default palette is Gruvbox Dark (see
  `styles.go`). The variables that `SetTheme` rewrites are declared with
  `var` (not `const`) so they can be reassigned at runtime.
- **Style rebuild, not just color swap:** `SetTheme` does not merely
  swap color constants. It re-constructs every derived lipgloss `Style`
  from the updated color variables, so any style that encodes
  background, foreground, bold, padding, or border-style all pick up the
  new values in a single call.
- **No persistence:** The theme is process-local. There is no file I/O,
  environment variable, or global config involved.
- **Idempotent:** Calling `SetTheme` more than once simply rewrites the
  same set of variables; the last call wins.

## Notes

- `SetTheme` has no error return. Invalid or empty color strings are
  passed through to lipgloss as-is; behavior depends on lipgloss's
  parsing.
- `SetTheme` must be called *before* widgets are rendered if you want
  the new theme to be visible. Widgets that cache rendered strings
  before the call will not update.
- `ThemeColors` is a flat value type — it has no methods and no
  interface. It exists purely as the input record for `SetTheme`.
