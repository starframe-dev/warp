# TabGroup

`TabGroup` is a Bubble Tea panel that renders a tab bar and switches
between child tabs. It can be embedded inside splits, flex layouts, or
used as the root panel of a TUI application. Tab placement is
configurable via `TabPosition` (top, bottom, left, right, or none).

## Types

### TabGroup

`TabGroup` keeps an ordered list of `*Tab` panels, tracks the active tab
index, and stores its tab position, ownership, and layout and hit-test
state.

## Public API

| Symbol | Kind | Description |
|----|----|----|
| `NewTabGroup(pos TabPosition) *TabGroup` | function | Creates a `TabGroup` with a single default tab named "main". |
| `ActiveTab() *Tab` | method | Returns the currently active tab, or `nil` if no valid tab is selected. |
| `NewTab(name string) *Tab` | method | Creates a new tab, switches to it, and returns the new tab. |
| `NextTab()` | method | Switches to the next tab (wrapping). |
| `PrevTab()` | method | Switches to the previous tab (wrapping). |
| `View(w, h int) string` | method | Renders the tab bar plus active tab content. |
| `Elements(w, h int) []Element` | method | Returns active-tab elements offset by the tab bar position, along with semantic tab-bar elements. |
| `Update(msg tea.Msg) tea.Cmd` | method | Handles key, mouse, resize, and broadcast messages. |

## Behavior

### Tab switching

`NextTab` and `PrevTab` wrap around the tab list using modulo
arithmetic. They are no-ops when only one tab exists. Internal
`switchTab(idx)` ignores invalid indices and no-ops when the requested
tab is already active.

### Tab creation and closing

`NewTab` appends a new tab and immediately activates it. `closeTab`
removes a tab by index; if the active tab is closed, the tab at that
index is activated, or the last remaining tab when the removed tab was
last. Closing an invalid index or the only tab is a no-op.

### Focus lifecycle

Switching tabs calls `Blur` on the old tab's focused `Focusable` while
retaining its focus target, then calls `Focus` on the new active tab's
retained target. Thus an inactive tab does not report a focused child,
but its focus target is restored when reactivated. Closing a tab blurs
and clears its focused panel; closing the active tab restores focus in
the newly active tab. After removal, panels implementing `Unmounter` are
unmounted only if no remaining tab or float in the same ownership hierarchy
references that instance. A standalone `TabGroup` owns its hierarchy; when
embedded in a `Warp`, the outer `Warp` owns it.

### Rendering

`View` composes the tab bar and active tab content using
`lipgloss.JoinVertical` / `lipgloss.JoinHorizontal` depending on the
`TabPosition`. The tab bar is rendered as either a horizontal row of tab
labels (for top/bottom) or a vertical column (for left/right). Labels
are truncated by terminal display width to at most 20 columns
(horizontal) or 15 columns (vertical), using an ellipsis where specified
by `ansi.Truncate`. The active tab label shows a `▎` prefix and `×` close
glyph. If `ActiveTab()` is nil, `View` returns `emptyView(h)` (with the
requested height normalized to at least zero).

### Semantic tab chrome

`Elements` exposes framework-owned tab-bar controls in addition to active-tab
content. Tab labels use role `tab` with action `activate-tab`; the active close
cell is a child `button` with action `close-tab`; the `+` cell is a `button`
with action `new-tab`. Bounds are computed from tab-bar label geometry and do
not require a prior `View` call.

### Mouse interaction

Tab bar hit regions (`tabRegion`) are populated while rendering the tab bar.
Clicking a tab switches to it, clicking its close glyph closes it, and clicking
the `+` region creates a new tab. Hit-testing logic differs for horizontal vs.
vertical bar layouts.

### Message dispatch

`Update` handles `tea.KeyMsg`, `tea.MouseMsg`, `tea.WindowSizeMsg`, and
`ResizeMsg`. Tab-group shortcuts are ctrl+c, ctrl+tab, ctrl+shift+tab,
ctrl+w, and ctrl+t; if the active tab's focused panel requests raw keys,
key messages are forwarded to it before shortcut handling. Other key messages
are forwarded to the active tab. Window-size and resize messages are forwarded
to every tab, and other messages are broadcast to all tabs so child panels
(e.g. PTY emulators) can receive them.

## Implementation details

- `padRight` pads a string with spaces to a target terminal display width, using
  `ansi.StringWidth`.
- `tabRegion` stores per-tab click bounds; `idx == -1` marks the "new
  tab" region.
- Horizontal bar regions are computed left-to-right with a trailing `+`
  region; vertical bar regions stack top-to-bottom with the `+` region
  last.
- Content geometry adjusts for the tab bar's position so child panels are
  laid out correctly.
- `shiftElements` recursively offsets element bounds so hit-testing
  remains consistent after the tab bar shifts child coordinates.

## Example

```go
tg := warp.NewTabGroup(warp.TabTop)
tg.NewTab("editor")
tg.NewTab("logs")

// Switch to the next tab
tg.NextTab()
```
