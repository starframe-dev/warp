# styles.go

`styles.go` defines the Gruvbox Dark color palette and the pre-built
`lipgloss.Style` values used by the tab bar, split borders, floating panes,
collapsible elements, and dropdowns. It also exposes functions for retrieving
the normal, dragging, and hover split-border styles. Popover styles are defined
separately in `popover.go`.

The file contains package-level variables and style definitions; it does not
implement UI rendering or runtime theme switching. Colors are hard-coded
`lipgloss.Color` values.

## Color palette

The palette is based on [Gruvbox Dark](https://github.com/morhetz/gruvbox):

- `gbDark0` through `gbDark4` are progressively lighter dark tones.
- `gbGray` is a neutral gray, and `gbLight1` is a light text color.
- `gbRed`, `gbGreen`, `gbYellow`, and `gbBlue` are accent colors. `gbBlue` is
defined in this file but is not used by its styles.

## Public API

```go
func BorderStyle() lipgloss.Style
func BorderDragStyle() lipgloss.Style
func BorderHoverStyle() lipgloss.Style
```

| Function | Purpose |
| --- | --- |
| `BorderStyle` | Returns the style with the regular split-border foreground. |
| `BorderDragStyle` | Returns the style used while a split border is being dragged. |
| `BorderHoverStyle` | Returns the style used when the mouse hovers over a split border. |

Each function returns its corresponding package-level style value. The returned
`lipgloss.Style` is a value, not a pointer; callers can use it with Lip Gloss
without accessing the unexported style variables.

## Style groups

### Tab bar

The tab bar background is `gbDark0`. The active tab uses `gbDark2` with bold
`gbLight1` text; inactive tabs use `gbDark0` with `gbGray` text. The new-tab
style uses `gbGreen`, and the close-tab style uses `gbRed`.

### Split borders

The normal border uses `gbDark1`, the dragging border uses `gbYellow`, and the
hover border uses `gbDark3`. The internal `collapseStyle` uses `gbDark1` as its
foreground.

### Floating panes

The floating-pane border uses `gbGray`. Its title uses a `gbDark1` background
and bold `gbLight1` text. The close style uses bold `gbRed`, and the pane
background style uses `gbDark0`.

### Collapsible elements

`collapsibleStyle` uses `gbLight1` on `gbDark1`; `collapsibleBorderStyle` uses
`gbDark4` as its foreground.

### Dropdowns

The dropdown button uses `gbDark2` with `gbLight1` text. Dropdown items use
`gbDark0` with `gbLight1` text. Hovered items use `gbDark2` with `gbYellow`
text, while selected items use `gbDark2` with bold `gbGreen` text.

## Example

```go
border := warp.BorderStyle().Render("│")
```

The example assumes the calling file imports this package as `warp` and imports
`github.com/charmbracelet/lipgloss` only if it uses Lip Gloss APIs directly.

## Implementation notes

- The styles are initialized as package-level variables when the package is
  initialized; they are not rebuilt by the three exported accessor functions.
- The colors and styles are unexported variables except for the three accessor
  functions. They are not constants or exported values.
- This file imports `github.com/charmbracelet/lipgloss` and defines package-level
  styles; other files in package `warp` can use those unexported names.
- Theme customization is not provided here; changing the hard-coded palette
  requires changing this file.
