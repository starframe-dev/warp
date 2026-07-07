# Specification: `tab.go`

## Overview
`tab.go` implements the `Tab` type for the `warp` package. A `Tab` is a container that owns a tree of panels (split, flex, or floating), manages focus, processes mouse and keyboard input, and renders its contents to a string. It is itself a `Panel` and can be embedded inside another `Panel`.

## File Location
`/Users/a/Space/Projects/Starframe/warp/tab.go`

## Package
`package warp`

## Public API

### Types

```go
type TabPosition int
```
Enum-like type that defines where a tab bar is rendered.

Constants:
- `TabTop`
- `TabBottom`
- `TabLeft`
- `TabRight`
- `TabNone`

```go
type Tab struct { ... }
```
The main container type. Holds a panel tree, floating panes, focus state, and internal drag state.

```go
type FlexItemSpec struct {
    Panel Panel
    Grow  int
}
```
Describes a panel and its flex-grow weight for `FlexRow` / `FlexColumn`.

### Constructors

```go
func newTab(name string, parent *TabGroup) *Tab
```
Package-private constructor that creates a `Tab` associated with a `TabGroup`. The root is initialized with an `emptyPanel` placeholder.

```go
func NewTab(name string) *Tab
```
Public constructor that creates a standalone `Tab` with no parent `TabGroup`. Useful for embedding a warp layout inside a `Panel`.

### Root Panel Access

```go
func (t *Tab) RootPanel() Panel
```
Returns the root panel. Useful as the parent argument for the first `Split` or `Flex` call.

```go
func (t *Tab) SetRootPanel(panel Panel)
```
Replaces the root panel with the given panel.

### Layout Operations

```go
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel)
```
Replaces `parent` with a vertical split (left/right). `fraction` is the share for the left panel, clamped to `[0.1, 0.9]`.

```go
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel)
```
Replaces `parent` with a horizontal split (top/bottom). `fraction` is the share for the top panel, clamped to `[0.1, 0.9]`.

```go
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec)
```
Replaces `parent` with a horizontal flex layout. `Grow` values below `0` are treated as `0`.

```go
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec)
```
Replaces `parent` with a vertical flex layout. `Grow` values below `0` are treated as `0`.

### Floating Panes

```go
func (t *Tab) Float(panel Panel, x, y, width, height int)
```
Adds a new floating pane on top of the layout at the specified position and size.

```go
func (t *Tab) CloseFloat(fp *FloatPane)
```
Removes the given floating pane from the tab.

### Focus Management

```go
func (t *Tab) Focus() Panel
```
Returns the currently focused panel.

```go
func (t *Tab) SetFocus(panel Panel) tea.Cmd
```
Sets the focused panel. Returns `nil`.

```go
func (t *Tab) FocusNext()
```
Moves focus to the next focusable panel in the layout tree.

```go
func (t *Tab) FocusPrev()
```
Moves focus to the previous focusable panel in the layout tree.

```go
func (t *Tab) FocusFirst()
```
Moves focus to the first focusable panel.

```go
func (t *Tab) FocusPanel(panel Panel)
```
Sets focus to a specific panel if it implements `Focusable`.

> **Note:** Warp does not automatically bind `Tab` / `Shift+Tab` for focus traversal. Call `FocusNext()` / `FocusPrev()` explicitly if desired.

### Panel Interface

```go
func (t *Tab) View(width, height int) string
```
Renders the tab content, including the panel tree and floating panes, joined by newlines.

```go
func (t *Tab) Update(msg tea.Msg) tea.Cmd
```
Handles `tea.Msg` updates:
- `tea.WindowSizeMsg`: Stores dimensions, broadcasts `ResizeMsg` to leaves, and also broadcasts the original `WindowSizeMsg` for backward compatibility.
- `ResizeMsg`: Stores dimensions and broadcasts `ResizeMsg` to leaves.
- Other messages: Broadcasts to all leaf panels and floats.

### Mouse Handling

```go
func (t *Tab) HandleMouse(msg tea.MouseMsg) tea.Cmd
```
Public entry point for mouse handling. Processes the message with no offset using the tab's current content dimensions. Used by embedded tabs such as `Container`'s inner tab.

Behavior:
- Floating panes are checked first in top-to-bottom z-order.
- Left-click inside a float brings it to the top and focuses its panel.
- Clicking outside an auto-close float dismisses it.
- Left-click on a border starts split or flex dragging.
- Left-click on a `Collapsible` title bar toggles its collapsed state.
- Drag motion updates split fractions or flex grow weights, with live `ResizeMsg` broadcasts.
- Release stops dragging and broadcasts a final resize.
- Mouse events are forwarded to the leaf panel under the cursor using coordinates relative to that panel.

### Split/Flex Adjustment

```go
func (t *Tab) SetSplitFraction(panel Panel, fraction float64) bool
```
Updates the fraction of the split containing `panel`. Returns `true` if the panel was found inside a split. Fraction is clamped to `[0.1, 0.9]`.

```go
func (t *Tab) GetSplitFraction(panel Panel) (float64, bool)
```
Returns the current split fraction for `panel` and `true` if found. For the second child, returns `1 - parent.Fraction`.

### Collapse/Expand

```go
func (t *Tab) Collapse(panel Panel, size int) bool
```
Shrinks `panel` to a fixed size inside its parent split. For vertical splits, `size` is width; for horizontal splits, it is height. The previous split fraction is saved so `Expand` can restore it. Returns `true` if found.

```go
func (t *Tab) Expand(panel Panel) bool
```
Restores `panel` to its saved split fraction. Returns `true` if found and was collapsed.

```go
func (t *Tab) ToggleCollapsible(panel Panel)
```
Toggles the collapsed state of a `Collapsible` panel and updates the associated `FlexItem.Collapsed` flag if the panel is inside a flex layout.

### Resize Broadcast

```go
func (t *Tab) BroadcastResize() tea.Msg
```
Sends `ResizeMsg` with each leaf panel's current content size. Called automatically by `Tab` on `WindowSizeMsg` and after a drag ends. Returns a `tea.BatchMsg` of collected commands.

### Element Query

```go
func (t *Tab) Elements(w, h int) []Element
```
Returns elements from the panel tree, recursively accounting for splits and flex layouts so coordinates are relative to the tab content area.

## Key Internal Types / Fields

| Field | Purpose |
|-------|---------|
| `name` | Tab name / label. |
| `root` | Root `*Node` of the panel tree. |
| `focused` | Currently focused `Panel`. |
| `floats` | Slice of `*FloatPane` floating above the layout. |
| `parent` | Parent `*TabGroup` (may be nil for standalone tabs). |
| `width`, `height` | Current content dimensions. |
| `dragging` | `*SplitConfig` currently being dragged. |
| `flexDragging` | `*FlexConfig` currently being dragged. |
| `flexDragIdx` | Index of the border being dragged in a flex layout. |
| `lastBorders` | Cached `[]BorderHit` used for hit-testing borders. |

## Important Implementation Details

- **Root safety:** `ensureRoot()` makes sure the root node always exists; initial root holds an `emptyPanel`.
- **Node replacement:** `SplitVertical` / `SplitHorizontal` find the leaf node for `parent`, then replace it with a `SplitConfig` containing the old panel as the first child and `newPanel` as the second.
- **Layout rendering:** `renderContent` calls `renderNode` and overlays each float with `overlayFloat`. Border hits are cached in `lastBorders` for mouse interaction.
- **Coordinate translation:** `Elements`, `panelAt`, `broadcastResize`, and `handleMouse` all recursively traverse the tree, applying offsets for splits and flex layouts. Border widths are treated as 1 cell.
- **Flex sizing:** `computeFlexSizes` distributes the available space based on `FlexItem.Grow` weights. Dragging adjusts neighboring grow weights proportionally and clamps them to a minimum of 1.
- **Fraction clamping:** `clampFraction` restricts fractions to `[0.1, 0.9]` to prevent panels from disappearing.
- **Message broadcasting:** `Update` forwards unknown messages (such as `PtyOutputMsg`, `PtyReadyMsg`) to all leaf panels and floating panes so that terminal emulators and other dynamic panels receive them.
- **Keyboard handling:** `handleKeys` forwards `tea.KeyMsg` to the currently focused panel. It does not intercept `Tab`/`Shift+Tab`.
- **Focus transition:** `setFocusedFocusable` calls `applyFocus` on the current focusable (if any) so that the previous panel can lose focus gracefully before the new panel gains focus.
- **Collapsible integration:** Collapsing or expanding a `Collapsible` panel inside a flex layout updates the `FlexItem.Collapsed` flag, which affects layout sizing.

## Dependencies
- Standard library: `fmt`, `os`, `strings`
- External: `github.com/charmbracelet/bubbletea` (tea)
- Other warp package types referenced: `Panel`, `Node`, `SplitConfig`, `FlexConfig`, `FlexItem`, `FloatPane`, `BorderHit`, `Element`, `Focusable`, `Collapsible`, `ResizeMsg`, `NodeCollapse`, `MinPanelSize`, `emptyPanel`, plus helper functions such as `collectFocusables`, `focusIndex`, `applyFocus`, `collectElements`, `shiftElements`, `renderNode`, `overlayFloat`, `findBorders`, `computeSplitSizes`, `computeFlexSizes`, `computeFlexSizes`.
