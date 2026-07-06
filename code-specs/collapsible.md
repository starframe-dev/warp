# Specification: `collapsible.go`

## Overview

`collapsible.go` implements `Collapsible`, a `warp` panel decorator that can be toggled between an expanded state (which renders its inner `Panel` normally) and a collapsed state (which renders only a single-line title bar). It is part of the `warp` package and uses the Charm `lipgloss` styles defined in `styles.go`.

## Public Types

### `Collapsible`

```go
type Collapsible struct {
    Title     string
    Collapsed bool
    Content   Panel
}
```

- `Title` — text displayed in the collapsed title bar.
- `Collapsed` — current state; `true` shows the title bar, `false` renders `Content`.
- `Content` — the inner `Panel` to render when expanded. May be `nil`.

`Collapsible` satisfies the `Panel` interface (`View` and `Update`).

## Public API

### `NewCollapsible(title string, content Panel) *Collapsible`

Constructor. Creates a `Collapsible` with the given title and inner content panel. `Collapsed` defaults to `false`.

### `(c *Collapsible) View(w, h int) string`

Renders the panel.

- If `Collapsed` is `true`, returns the single-line collapsed title bar via `renderCollapsed(w)`.
- If `Collapsed` is `false` and `Content` is non-`nil`, delegates to `c.Content.View(w, h)`.
- If `Collapsed` is `false` and `Content` is `nil`, returns an empty string.

### `(c *Collapsible) Update(msg tea.Msg) tea.Cmd`

Forwards Bubbletea messages to the inner content panel. If `Content` is `nil`, returns `nil`.

### `(c *Collapsible) Toggle()`

Inverts the `Collapsed` state (`true` ↔ `false`).

## Important Implementation Details

### Collapsed Title Bar Rendering (`renderCollapsed`)

- Returns empty string if width `w <= 0`.
- Uses a `▶` indicator. The indicator is currently set to `▶` in the collapsed branch; the code also contains logic for `▼` when not collapsed, but that branch is unreachable because `renderCollapsed` is only called when `Collapsed` is `true`.
- Reserves `5` rune cells for the indicator, spaces, and box corners: `indicator + " " + title` plus the left corner `┌` and right corner `┐`.
- Truncates the title with `...` if it exceeds `w - reserve - 3`. If the available title width is `0`, the title is omitted.
- Pads the remaining width with `─` characters so the bar fills exactly width `w`.
- Applies `collapsibleStyle` to the left portion and `collapsibleBorderStyle` to the right border/padding portion.

### Style Dependencies

`renderCollapsed` depends on two package-level styles defined in `styles.go`:

- `collapsibleStyle` — foreground `gbLight1`, background `gbDark1`.
- `collapsibleBorderStyle` — foreground `gbDark4`.

### Dependencies

- Standard library: `strings`.
- External: `github.com/charmbracelet/bubbletea` (`tea.Msg`).
- Package: `Panel` interface defined in `panel.go`.

## Behavior Summary

- `Collapsible` is a lightweight wrapper around a `Panel` that provides a collapse/expand mechanic.
- In the expanded state, it behaves transparently: input messages and rendering pass through to the inner panel.
- In the collapsed state, only the title bar is rendered; the inner panel receives no messages and produces no content.
- State changes are explicit via `Toggle()`; there is no built-in keyboard or mouse handling in this file.