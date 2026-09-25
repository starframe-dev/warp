# Tab

The `Tab` type in `tab.go` is a self-contained warp layout: it owns a
root panel tree (`Node`), an optional set of floating panes, and
keyboard focus. It implements the `Panel` interface (`View`/`Update`) so
a tab itself can be embedded as a panel in a larger layout (e.g. inside
a `Container`).

A tab is created standalone via `NewTab` (no parent group) or internally
via `newTab(name, parent)` when it belongs to a `TabGroup`. In both
cases the root starts with an empty placeholder panel so the tab is
always renderable.

## Public API

| Method | Purpose |
|----|----|
| `NewTab(name string) *Tab` | Create a standalone tab with an empty root. |
| `RootPanel() Panel` | Return the root panel (safe to pass as the first split parent). |
| `SetRootPanel(panel Panel)` | Replace the root panel. |
| `SplitVertical(parent Panel, fraction float64, newPanel Panel)` | Split a panel into left/right (fraction is the left share). |
| `SplitHorizontal(parent Panel, fraction float64, newPanel Panel)` | Split a panel into top/bottom (fraction is the top share). |
| `FlexRow(parent Panel, items []FlexItemSpec)` | Replace a panel with a horizontal flex row. |
| `FlexColumn(parent Panel, items []FlexItemSpec)` | Replace a panel with a vertical flex column. |
| `Float(panel Panel, x, y, width, height int)` | Add a floating pane on top of the layout. |
| `CloseFloat(fp *FloatPane)` | Remove a floating pane. |
| `Focus() Panel` | Return the currently focused panel. |
| `SetFocus(panel Panel) tea.Cmd` | Set the focused panel. |
| `FocusNext() / FocusPrev()` | Move focus to the next/previous focusable panel (wraps around). |
| `FocusFirst()` | Move focus to the first focusable panel. |
| `FocusPanel(panel Panel)` | Focus a specific panel if it is `Focusable`. |
| `SetSplitFraction(panel Panel, fraction float64) bool` | Set the fraction of the split containing `panel`. |
| `GetSplitFraction(panel Panel) (float64, bool)` | Read the fraction of the split containing `panel`. |
| `Collapse(panel Panel, size int) bool` | Shrink `panel` to a fixed size, saving its split fraction for restore. |
| `Expand(panel Panel) bool` | Restore a collapsed panel to its saved fraction. |
| `ToggleSplitCollapse(parent Panel)` | Toggle the collapse state of the split containing `parent`. |
| `SetSplitCollapse(parent, collapseRow, onCollapse)` | Attach a collapse symbol + callback to a split border. |
| `ToggleCollapsible(panel Panel)` | Toggle a collapsible panel's collapsed state. |
| `BroadcastResize() tea.Msg` | Broadcast a `ResizeMsg` to every leaf with its computed size. |
| `HandleMouse(msg tea.MouseMsg) tea.Cmd` | Public mouse entry point for embedded tabs. |
| `View(width, height int) string` | Render the tab content (implements `Panel`). |
| `Update(msg tea.Msg) tea.Cmd` | Dispatch messages through the panel tree (implements `Panel`). |
| `Elements(w, h int) []Element` | Collect interactive elements from the tree. |

## Types

### TabPosition

Integer type selecting where the tab bar is drawn: `TabTop`,
`TabBottom`, `TabLeft`, `TabRight`, or `TabNone`.

### FlexItemSpec

Describes a single item to add to a flex layout: `Panel` (the panel) and
`Grow` (its flex-grow weight, clamped to ≥ 0).

### emptyPanel

Internal placeholder panel used as the initial root so a tab can always
be queried as an element provider before real panels are attached.

## Behavior & Implementation Details

- **Splitting:** `SplitVertical`/`SplitHorizontal` wrap the existing
  panel and the new panel into a `SplitConfig` with clamped fractions
  (`clampFraction` keeps values in \[0.1, 0.9\]).
- **Flex layouts:** `FlexRow`/`FlexColumn` replace the node's content
  with a `FlexConfig`. Negative grow weights are clamped to 0; an empty
  item slice is ignored.
- **Message flow:** `Update` handles `tea.WindowSizeMsg` (records size,
  broadcasts `ResizeMsg` plus the original window msg), `ResizeMsg`
  (broadcast resize only), and forwards all other messages (e.g.
  `PtyOutputMsg`) to every leaf.
- **Rendering:** `View` renders the node tree, refreshes border hit data
  (`findBorders`), then overlays floating panes.
- **Mouse:** `handleMouse` (exposed as `HandleMouse`) processes float
  panes first (z-order, bring-to-top, close-on-outside-click), then
  border dragging (split or flex), collapse symbols, and finally
  forwards events to the panel under the cursor in relative coordinates.
  Dragging live-broadcasts resizes so panels update while dragging.
- **Keyboard:** `handleKeys` forwards only to the focused panel; warp
  never intercepts Tab/Shift+Tab automatically — use the explicit
  `FocusNext`/`FocusPrev` methods.
- **Focus cleanup:** `CloseFloat` blurs and clears the focus target when
  it removes the focused float. An outside click may then focus the panel
  under the click.
- **Collapse/Expand:** `Collapse` saves the split fraction and shrinks
  the panel to a fixed size; `Expand` restores the saved fraction. Flex
  items and `Collapsible` panels are assigned the requested state
  explicitly, making repeated Collapse/Expand calls idempotent. A nil
  `SetSplitCollapse` callback does not disable the internal state toggle.
- **Element collection:** `Elements` recursively walks splits and flex
  layouts, computing per-branch sizes (accounting for 1-cell borders)
  and offsetting element bounds so coordinates are relative to the tab
  content area.

## Usage

``` go
tab := warp.NewTab("main")
tab.SplitVertical(tab.RootPanel(), 0.6, rightPanel)
tab.SplitHorizontal(leftPanel, 0.7, bottomPanel)
tab.FlexRow(centerPanel, []warp.FlexItemSpec{{Panel: a, Grow: 1}, {Panel: b, Grow: 2}})
tab.Float(overlay, 0, 0, 40, 10)

// focus traversal
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()

// split geometry
tab.SetSplitFraction(leftPanel, 0.7)
frac, ok := tab.GetSplitFraction(leftPanel)

// collapse / expand
tab.Collapse(leftPanel, 2)
tab.Expand(leftPanel)

// broadcast size to leaves
cmd := tab.BroadcastResize()
```
