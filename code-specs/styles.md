# Specification: `styles.go`

## Overview

`styles.go` defines the visual style system for the `warp` package. It centralizes the Gruvbox Dark color palette and pre-built `lipgloss.Style` values used to render the tab bar, split borders, floating panes, collapsible sections, and dropdown widgets. The file is pure configuration: it contains no behavior or rendering logic beyond constructing styles.

## Color Palette

The file implements the Gruvbox Dark color scheme via package-level color variables of type `lipgloss.Color`. These are used as the raw colors for all styles.

### Base colors

| Name      | Hex     | Notes                           |
|-----------|---------|---------------------------------|
| `gbDark0` | `#282828` | Deepest background / float pane background |
| `gbDark1` | `#3c3836` | Border, float title background  |
| `gbDark2` | `#504945` | Active tab background, dropdown surfaces |
| `gbDark3` | `#665c54` | Hover border color              |
| `gbDark4` | `#7c6f64` | Collapsible border color        |
| `gbGray`  | `#928374` | Inactive tab text, float border |
| `gbLight1`| `#ebdbb2` | Primary light text              |
| `gbRed`   | `#fb4934` | Close tab, float close button   |
| `gbGreen` | `#b8bb26` | New tab button, selected item   |
| `gbYellow`| `#fabd2f` | Drag border, dropdown hover     |
| `gbBlue`  | `#83a598` | Reserved in palette, not used by a style currently |

## Public API

The file exports three style accessor functions. All other style variables are package-private.

```go
func BorderStyle() lipgloss.Style
func BorderDragStyle() lipgloss.Style
func BorderHoverStyle() lipgloss.Style
```

### `BorderStyle()`

Returns the style used for normal split borders.

- Foreground: `gbDark1` (`#3c3836`)
- Other properties: none (no background, no padding/margin)

### `BorderDragStyle()`

Returns the style used while a split border is being dragged.

- Foreground: `gbYellow` (`#fabd2f`)

### `BorderHoverStyle()`

Returns the style used when the mouse pointer hovers over a split border.

- Foreground: `gbDark3` (`#665c54`)

## Internal Style Groups

The following pre-built `lipgloss.Style` variables are declared in the package. They are intended to be used by other components in the `warp` package.

### Tab bar styles

| Variable          | Background | Foreground | Other      |
|-------------------|------------|------------|------------|
| `tabBarStyle`     | `tabBarBg` (`gbDark0`) | — | — |
| `activeTabStyle`  | `activeTabBg` (`gbDark2`) | `activeTabFg` (`gbLight1`) | `Bold(true)` |
| `inactiveTabStyle`| `tabBarBg` (`gbDark0`) | `inactiveTabFg` (`gbGray`) | — |
| `newTabStyle`     | — | `newTabFg` (`gbGreen`) | — |
| `closeTabStyle`   | — | `closeTabFg` (`gbRed`) | — |

### Split border styles

| Variable            | Foreground                     |
|---------------------|--------------------------------|
| `borderStyle`       | `borderColor` (`gbDark1`)      |
| `borderHoverStyle`  | `borderHoverColor` (`gbDark3`) |
| `borderDragStyle`   | `borderDragColor` (`gbYellow`) |

### Floating pane styles

| Variable            | Background | Foreground | Other      |
|---------------------|------------|------------|------------|
| `floatBorderStyle`  | — | `floatBorderColor` (`gbGray`) | — |
| `floatTitleStyle`   | `floatTitleBg` (`gbDark1`) | `floatTitleFg` (`gbLight1`) | `Bold(true)` |
| `floatCloseStyle`   | — | `floatCloseFg` (`gbRed`) | `Bold(true)` |
| `floatBgStyle`      | `floatBg` (`gbDark0`) | — | — |

### Collapsible styles

| Variable                | Background | Foreground |
|-------------------------|------------|------------|
| `collapsibleStyle`      | `gbDark1`  | `gbLight1` |
| `collapsibleBorderStyle`| —          | `gbDark4`  |

### Dropdown styles

| Variable                    | Background | Foreground | Other      |
|-----------------------------|------------|------------|------------|
| `dropdownButtonStyle`       | `gbDark2`  | `gbLight1` | — |
| `dropdownItemStyle`         | `gbDark0`  | `gbLight1` | — |
| `dropdownItemHoverStyle`    | `gbDark2`  | `gbYellow` | — |
| `dropdownItemSelectedStyle` | `gbDark2`  | `gbGreen`  | `Bold(true)` |

## Important Implementation Details

- **Palette constants are variables, not constants.** They are declared as `var` blocks so they can be reassigned or overridden by other packages/files, although the file treats them as a fixed theme.
- **Style variables are also mutable.** All `lipgloss.Style` values are package-level variables. They are eagerly initialized at package initialization time via `lipgloss.NewStyle()`.
- **No exported types.** The file exports no new Go types; the only exported API is the three accessor functions returning `lipgloss.Style`.
- **Dependency on `lipgloss`.** The file imports `github.com/charmbracelet/lipgloss` and relies entirely on its style API.
- **Style reuse.** Styles are reused across related UI components (e.g., active tabs and dropdown selected items share `gbDark2` background). Color reuse is intentional to keep the theme consistent.
- **No runtime behavior.** The file does not define any methods, state, or conditional logic. It only declares colors and styles.
