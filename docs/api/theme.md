---
title: Theme
description: Runtime semantic color palette for Warp controls and overlays.
---

# Theme

Warp ships with a Gruvbox Dark palette. `SetTheme` replaces the package-level colors and rebuilds the styles used by tabs, borders, floats, dropdowns, popovers, modals, inputs, and collapsibles.

## ThemeColors

```go
type ThemeColors struct {
    Background          string
    Surface             string
    Raised              string
    Border              string
    BorderMuted         string
    Text                string
    TextMuted           string
    TextStrong          string
    Accent              string
    AccentMuted         string
    Error               string
    Success             string
    Warning             string
    SelectionBackground string
    SelectionForeground string
}
```

Fields contain color strings accepted by `lipgloss.Color`, commonly hex values such as `"#282828"`.

`Text` and `AccentMuted` are reserved semantic fields and are currently not mapped to a package-level color. The other fields map to Warp's internal palette; `SelectionBackground` and `SelectionForeground` control the active-tab and selected-popover colors.

## SetTheme

```go
func SetTheme(colors ThemeColors)
```

Call `SetTheme` before rendering to apply a palette. The function does not validate color strings or return an error. Calling it again replaces the previous palette.

## Example

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

Theme state is process-local and the last call wins.
