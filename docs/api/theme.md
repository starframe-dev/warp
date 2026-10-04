# theme.go

## Overview

`theme.go` exposes `ThemeColors` and `SetTheme` for changing colors used by Warp's package-level UI styles. `SetTheme` assigns the supplied values to internal palette variables and rebuilds the styles for tabs, borders, float panes, collapsibles, dropdowns, popovers, modals, dim text, and inputs.

The default palette is initialized in `styles.go`. Calling `SetTheme` replaces the relevant package-level colors and styles; it does not persist the theme.

## Public API

### ThemeColors

`ThemeColors` is a struct of 15 string fields accepted by `SetTheme`:

| Field | Used for |
|---|---|
| `Background` | Tab bar, float pane, dropdown item, and dim-text backgrounds |
| `Surface` | Float title, panel border, collapsible body, popover, and modal backgrounds |
| `Raised` | Dropdown button, hover, and selected-item backgrounds |
| `Border` | Modal and collapsible borders, and input border |
| `BorderMuted` | Hovered split border |
| `Text` | Not currently used by `SetTheme` |
| `TextMuted` | Inactive tab labels, float border, and dim text |
| `TextStrong` | Float title, collapsible, dropdown, popover, modal, and input foregrounds |
| `Accent` | Focused input border |
| `AccentMuted` | Not currently used by `SetTheme` |
| `Error` | Close-tab and float-close foregrounds |
| `Success` | New-tab and selected-dropdown-item foregrounds |
| `Warning` | Dragged split border and dropdown hover foreground |
| `SelectionBackground` | Active tab and selected-popover backgrounds |
| `SelectionForeground` | Active tab and selected-popover foregrounds |

Values are passed to `lipgloss.Color` as supplied; `SetTheme` does not validate or normalize them. Although fields are plain strings and may contain color values supported by Lipgloss, the implementation does not require a particular format.

### SetTheme

```go
func SetTheme(colors ThemeColors)
```

`SetTheme` has no return value. It updates internal palette and semantic color variables, then reconstructs the package-level Lipgloss styles derived from them. Calling it again replaces the same values, so the most recent call determines the package-level theme.

The `Text` and `AccentMuted` fields are currently not read by this function and therefore do not affect the rendered styles.

## Usage

```go
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
```

## Notes

- The default palette is Gruvbox Dark, declared in `styles.go`.
- `SetTheme` performs no validation, returns no error, and has no persistence or configuration-file behavior.
- The function reassigns package-level state. Applications should avoid concurrent calls to `SetTheme` while other goroutines access the affected styles unless they provide their own synchronization.
