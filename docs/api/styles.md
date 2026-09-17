# styles.go

The `styles.go` file is the central palette and lipgloss style source of
truth for the `warp` TUI. It defines the Gruvbox Dark color set once and
derives every UI style from it: tab bar, split border, floating pane,
collapsible, dropdown and other elements. There is no runtime behavior
in this file; it only holds *constants* (colors) and *pre-built styles*
that other files consume.

## Module behavior

The module has no functions that mutate state. All exported values are
immutable `lipgloss.Style` values and unexported color variables. This
keeps the file trivially local: any UI component can render without
knowing how other elements are styled.

The color palette follows the [Gruvbox
Dark](https://github.com/morhetz/gruvbox) scheme:

- Background scale `gbDark0..gbDark4` and neutral `gbGray`
- Light text `gbLight1`
- Accent set: `gbRed`, `gbGreen`, `gbYellow`, `gbBlue`

## Public API

All public exports are read-only style values. They are safe to share
across goroutines because `lipgloss.Style` is an immutable struct.

``` go
func BorderStyle() lipgloss.Style
func BorderDragStyle() lipgloss.Style
func BorderHoverStyle() lipgloss.Style
```

| Function | Returns | Purpose |
|----|----|----|
| `BorderStyle` | `lipgloss.Style` | Foreground of the regular split border between panes. |
| `BorderDragStyle` | `lipgloss.Style` | Foreground used while a split border is being dragged. |
| `BorderHoverStyle` | `lipgloss.Style` | Foreground shown when the mouse hovers over a split border. |

The three functions are tiny wrappers rather than values because
`lipgloss.Style` is a value type and the package wants to avoid
exporting the unexported `borderStyle`, `borderDragStyle` and
`borderHoverStyle` identifiers directly. Callers should treat the return
value as a constant and never mutate it.

## Internal palette

The unexported color variables group into three conceptual buckets:

1.  **Base palette** — `gbDark0` through `gbLight1`, `gbGray` plus
    accent colors.
2.  **Split border palette** — `borderColor`, `borderDragColor`,
    `borderHoverColor` and `collapseStyle`.
3.  **Element palettes** — tab bar, floating pane, dropdown and
    collapsible colors, each paired with the lipgloss style that
    consumes them.

## Style groups

### Tab bar

The tab bar reuses the darkest background (`tabBarBg = gbDark0`) and
uses a bold, light foreground for the active tab. Inactive tabs use a
muted gray foreground. New-tab affordance is green, close-tab affordance
is red.

### Split border

`borderStyle` and `collapseStyle` both render with `gbDark1` foreground.
Dragging switches to `gbYellow`, hovering to `gbDark3`.

### Float pane

Floating panes have their own chrome: border in `gbGray`, title bar with
`gbDark1` background and `gbLight1` bold text, close button in red, body
background `gbDark0`.

### Collapse

Collapsed panes use `gbLight1` on `gbDark1`, and the border around a
collapsed pane uses `gbDark4`.

### Dropdown

Dropdown buttons use `gbDark2` background with `gbLight1` text. Items
sit on `gbDark0` background, hover to `gbDark2` background with yellow
text, and selected state turns green with bold text.

## Usage example

``` go
// Render a split border in its normal state.
lipgloss.NewRenderer().Render(wrapText("pane", warp.BorderStyle()))

// While the user drags the border.
lipgloss.NewRenderer().Render(wrapText("pane", warp.BorderDragStyle()))
```

## Implementation details

- All styles are constructed once at package init time; they are not
  rebuilt on use.
- Colors are hard-coded hex values rather than a runtime theme system;
  switching themes requires editing this file.
- There is no theme interface or DI; the file is a single source of
  truth that other files in the package import implicitly via
  package-level names.
- Styles are values, not pointers; lipgloss treats them as immutable, so
  returning them by value is safe.
- The file has no dependencies on the rest of the package except the
  `lipgloss` library itself; it can be reasoned about locally.
