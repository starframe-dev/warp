# TabGroup

`TabGroup` is a Bubble Tea panel that renders a tab bar and switches
between child tabs. It can be embedded inside splits, flex layouts, or
used as the root panel of a TUI application. Tab placement is
configurable via `TabPosition` (top, bottom, left, right, or none).

## Types

### TabGroup

``` go
type TabGroup struct {
    tabs      []*Tab
    activeTab int
    width     int
    height    int

    tabPosition      TabPosition
    tabRegions       []tabRegion
    newTabRegion     *tabRegion
    verticalTabWidth int
}
```

`TabGroup` keeps an ordered list of `*Tab` panels, tracks the active tab
index, and caches per-frame geometry (width, height, tab hit regions)
used for mouse hit-testing.

## Public API

| Symbol | Kind | Description |
|----|----|----|
| `NewTabGroup(pos TabPosition) *TabGroup` | function | Creates a `TabGroup` with a single default tab named "main". |
| `ActiveTab() *Tab` | method | Returns the currently active tab, or `nil` if no valid tab is selected. |
| `NewTab(name string) *Tab` | method | Creates a new tab, switches to it, and returns the new tab. |
| `NextTab()` | method | Switches to the next tab (wrapping). |
| `PrevTab()` | method | Switches to the previous tab (wrapping). |
| `View(w, h int) string` | method | Renders the tab bar plus active tab content. |
| `Elements(w, h int) []Element` | method | Returns active tab elements offset by the tab bar position. |
| `Update(msg tea.Msg) tea.Cmd` | method | Handles key, mouse, resize, and broadcast messages. |

## Behavior

### Tab switching

`NextTab` and `PrevTab` wrap around the tab list using modulo
arithmetic. They are no-ops when only one tab exists. Internal
`switchTab(idx)` ignores invalid indices and no-ops when the requested
tab is already active.

### Tab creation and closing

`NewTab` appends a new tab and immediately activates it. `closeTab`
removes a tab by index; if the active tab is closed, the next tab at that
index is activated, or the last remaining tab when the removed tab was
last. Closing the only tab is a no-op.

### Focus lifecycle

Switching tabs calls `Blur` on the old tab's focused `Focusable` while
retaining its focus target, then calls `Focus` on the new active tab's
retained target. Thus an inactive tab does not report a focused child,
but its focus target is restored when reactivated. Closing a tab blurs
and clears its focused panel; closing the active tab restores focus in
the newly active tab.

### Rendering

`View` composes the tab bar and active tab content using
`lipgloss.JoinVertical` / `lipgloss.JoinHorizontal` depending on the
`TabPosition`. The tab bar is rendered as either a horizontal row of tab
labels (for top/bottom) or a vertical column (for left/right). Tab
labels are truncated to 20 chars (horizontal) or 15 chars (vertical)
with an ellipsis. The active tab label shows a `▎` prefix and `×` close
glyph. If `ActiveTab()` is nil, `View` returns exactly `h` blank lines
(`""` when `h == 0`; otherwise `h-1` newline characters).

### Mouse interaction

Tab bar hit regions (`tabRegion`) are recomputed each frame during
rendering. Clicking a tab switches to it, clicking the close glyph
closes the tab, and clicking the `+ /` region creates a new tab.
Hit-testing logic differs for horizontal vs. vertical bar layouts.

### Message dispatch

`Update` intercepts `tea.KeyMsg` (ctrl+c, ctrl+tab, ctrl+shift+tab,
ctrl+w, ctrl+t), `tea.MouseMsg`, `tea.WindowSizeMsg`, and `ResizeMsg`.
All other messages are broadcast to every tab so that child panels (e.g.
PTY emulators) receive them.

## Implementation details

- `padRight` pads a string with spaces to a target width, using
  lipgloss's rune-aware width calculation.
- `tabRegion` stores per-tab click bounds; `idx == -1` marks the "new
  tab" region.
- Horizontal bar: regions are computed left-to-right with a trailing `+`
  region; vertical bar: regions stack top-to-bottom with the `+` region
  last.
- Content geometry (`contentWidth`, `contentHeight`, `contentOffset`)
  adjusts for the tab bar's position so child panels are laid out
  correctly.
- `shiftElements` recursively offsets element bounds so hit-testing
  remains consistent after the tab bar shifts child coordinates.

## Example

``` go
tg := warp.NewTabGroup(warp.TabTop)
tg.NewTab("editor")
tg.NewTab("logs")

// Switch to the next tab
tg.NextTab()
```
