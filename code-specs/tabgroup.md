# TabGroup Specification

## Overview

`TabGroup` is a `Panel` implementation that renders a tab bar and displays the active tab's content. It can be used as a root panel or embedded inside splits, flex layouts, and other containers. Tabs are positioned at the top, bottom, left, right, or can be hidden entirely (`TabNone`). Each tab owns a `Tab` instance that manages its own panel tree, focus state, and floating panes.

The tab group is responsible for:
- Rendering the tab bar and the active tab's content.
- Routing keyboard and mouse input to the active tab or tab bar.
- Tracking clickable regions in the tab bar for switching, closing, and creating tabs.
- Forwarding resize and unknown messages to all tabs so background panels continue to receive updates.

## Public API

### Types

#### `type TabGroup struct`

The panel that contains a list of tabs and renders the active tab.

Key fields (internal):
- `tabs []*Tab` — all tabs in the group.
- `activeTab int` — index of the currently visible tab.
- `width, height int` — last allocated dimensions.
- `tabPosition TabPosition` — where the tab bar is drawn.
- `tabRegions []tabRegion` — clickable regions for each tab in the current frame.
- `newTabRegion *tabRegion` — clickable region for the "+" new-tab button.
- `verticalTabWidth int` — width of the tab bar when positioned on the left or right.

#### `type TabPosition int`

Defined in `tab.go` and used by `TabGroup` to control tab bar placement:

```go
const (
    TabTop    TabPosition = iota
    TabBottom
    TabLeft
    TabRight
    TabNone
)
```

### Constructor

#### `func NewTabGroup(pos TabPosition) *TabGroup`

Creates a `TabGroup` with the requested tab bar position and a single default tab named `"main"`. The new tab is automatically active.

### Methods

#### `func (tg *TabGroup) NewTab(name string) *Tab`

Creates a new tab, appends it to the group, and switches focus to it. Returns the created `*Tab`. The tab is initialized with an empty placeholder panel.

#### `func (tg *TabGroup) ActiveTab() *Tab`

Returns the currently active tab, or `nil` if no tab is selected.

#### `func (tg *TabGroup) NextTab()`

Cycles forward to the next tab, wrapping around if necessary. Does nothing if there is only one tab.

#### `func (tg *TabGroup) PrevTab()`

Cycles backward to the previous tab, wrapping around if necessary. Does nothing if there is only one tab.

#### `func (tg *TabGroup) View(w, h int) string`

Renders the full tab group: the tab bar plus the active tab's content. The layout depends on `tabPosition`:

- `TabTop`: tab bar on the first row, content below.
- `TabBottom`: content first, tab bar on the last row.
- `TabLeft`: vertical tab bar on the left, content on the right.
- `TabRight`: content first, vertical tab bar on the right.
- `TabNone`: only the active tab content is rendered.

For top/bottom positions the tab bar consumes one row. For left/right positions the tab bar consumes `verticalTabWidth` columns, which is computed each frame based on the widest tab label.

#### `func (tg *TabGroup) Elements(w, h int) []Element`

Implements `ElementProvider`. Collects interactive elements from the active tab's panel tree and offsets their coordinates by the content area origin (accounts for the tab bar position). Also shifts child element coordinates recursively.

#### `func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd`

Handles input and lifecycle messages:

- `tea.KeyMsg`: routed to `handleKeyMsg`.
- `tea.MouseMsg`: routed to `handleMouseMsg`.
- `tea.WindowSizeMsg`: stores the new size and forwards the message to every tab.
- `ResizeMsg`: stores the new size and forwards it to every tab so child panels can receive their allocated size.
- Unknown messages (e.g. `PtyReadyMsg`, `PtyOutputMsg`): broadcast to every tab via `tab.broadcastMsg` so all panels, including background tabs, can respond.

## Behavior

### Tab Bar Rendering

`renderTabBar` dispatches to horizontal or vertical rendering based on `tabPosition`.

#### Horizontal Tab Bar (`TabTop`/`TabBottom`)

- Each tab is rendered as a label with surrounding padding.
- Active tab: `▎ <name> ×` (left bar, name, close glyph).
- Inactive tab: ` <name> ` (padded name).
- Tab names longer than 20 bytes are truncated to `...`.
- A ` + ` button is appended at the end for creating a new tab.
- Clickable regions are stored in `tabRegions` with the tab index, start column, end column, and close-button column (active tab only).
- The tab bar is padded to the full width with the tab bar background.

#### Vertical Tab Bar (`TabLeft`/`TabRight`)

- Each tab is rendered on its own line.
- Active tab: `▎ <name> ×`.
- Inactive tab: ` <name> `.
- Tab names longer than 15 bytes are truncated.
- A ` + ` button appears on the line after the last tab.
- All tab labels are padded to the width of the widest label so the bar is rectangular.
- `verticalTabWidth` is set to the final padded width.
- Clickable regions use the row index for `startX`/`endX` and the last column for the close button.

### Mouse Handling

When a mouse message arrives:

1. `isOnTabBar` checks whether the event coordinates fall within the tab bar region.
2. If on the tab bar, `handleTabBarClick` handles left presses:
   - Clicking a tab region switches to that tab.
   - Clicking the `×` close glyph on the active tab closes it.
   - Clicking the `+` region creates a new tab named `"tab"`.
   - Vertical tabs use the row coordinate to index `tabRegions`; the new-tab row uses `idx == -1`.
3. If not on the tab bar, the mouse message is translated by the content offset and forwarded to `tab.handleMouse` with the active tab's content dimensions.

### Keyboard Handling

`handleKeyMsg` intercepts the following global shortcuts before forwarding other keys to the active tab:

| Shortcut | Action |
|----------|--------|
| `ctrl+c` | Quit the application (`tea.Quit`). |
| `ctrl+tab` | Switch to the next tab. |
| `ctrl+shift+tab` | Switch to the previous tab. |
| `ctrl+w` | Close the active tab (ignored if only one tab remains). |
| `ctrl+t` | Create a new tab named `"tab"`. |

Any other key message is forwarded to `tab.handleKeys` on the active tab, which in turn forwards it to the focused panel.

### Tab Closing

`closeTab` removes a tab by index. The last tab cannot be closed. After removal, if the active index is out of bounds it is clamped to the new last tab. `ctrl+w` closes the currently active tab.

### Tab Switching

`switchTab` validates the index and sets `activeTab`. `NextTab` and `PrevTab` wrap modulo the number of tabs.

## Important Implementation Details

- `TabGroup` implements the `Panel` interface (`View`, `Update`) and `ElementProvider` (`Elements`).
- The tab bar always consumes one row for top/bottom positions and `verticalTabWidth` columns for left/right positions. Content dimensions are reduced accordingly.
- `contentOffset` returns the origin `(x, y)` of the content area relative to the tab group's top-left corner. It is used to shift mouse coordinates and element bounds.
- `padRight` is a small helper that pads a string to a given visual width using `lipgloss.Width`.
- `tabRegion` is an internal struct that records the index, start/end columns, and close-button column of a tab in the rendered bar. It is rebuilt every frame during rendering.
- `newTabRegion` stores the region of the `+` button in horizontal layouts; in vertical layouts the new-tab row is represented by a `tabRegion` with `idx == -1`.
- Tab label truncation uses byte-length checks (`len(name) > 20` / `len(name) > 15`), not rune count, which can cut multi-byte runes in edge cases.
- The active tab's close glyph is always shown and clickable; inactive tabs have no close button in the current implementation.
- `Elements` recursively shifts not only top-level element bounds but also all `Children` via `shiftElements`.
- `Update` broadcasts unrecognized messages to all tabs, not just the active one, so background panels such as terminal emulators continue to receive output.

## Dependencies

- `github.com/charmbracelet/bubbletea` for `tea.Msg`, `tea.KeyMsg`, `tea.MouseMsg`, `tea.WindowSizeMsg`, and command types.
- `github.com/charmbracelet/lipgloss` for styles and layout helpers (`lipgloss.Width`, `lipgloss.JoinVertical`, `lipgloss.JoinHorizontal`).
- `tab.go` for `Tab`, `TabPosition`, and related tab behavior.
- `styles.go` for `tabBarStyle`, `activeTabStyle`, `inactiveTabStyle`, and `newTabStyle`.
- `element.go` for `Element`, `collectElements`, and `shiftElements`.

## Notes

- `TabGroup` does not currently expose a way to programmatically close a specific tab by index from outside the package; `closeTab` is unexported.
- Tab names are not validated or deduplicated; callers can create multiple tabs with the same name.
- The new tab created via the `+` button or `ctrl+t` is always named `"tab"`. Callers can rename a tab by modifying the returned `*Tab` if the field is exported, or by creating a tab with `NewTab(name)` directly.
