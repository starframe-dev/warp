---
title: Tab
description: Tab API for layout, floats, focus, split sizing, and collapse.
---

# Tab

A `Tab` owns one panel tree, its floating panes, and its current focus.

## Construction and root

```go
func NewTab(name string) *Tab
func (t *Tab) RootPanel() Panel
func (t *Tab) SetRootPanel(panel Panel)
```

`NewTab` creates a standalone tab. `TabGroup.NewTab` creates a tab attached to a group.

## Layout

```go
func (t *Tab) SplitVertical(parent Panel, fraction float64, newPanel Panel)
func (t *Tab) SplitHorizontal(parent Panel, fraction float64, newPanel Panel)
func (t *Tab) FlexRow(parent Panel, items []FlexItemSpec)
func (t *Tab) FlexColumn(parent Panel, items []FlexItemSpec)
```

Splits replace `parent` with a two-child node. The split fraction is clamped to `0.1..0.9`. Flex methods ignore an empty item list.

## Split sizing and collapse

```go
func (t *Tab) SetSplitFraction(panel Panel, fraction float64) bool
func (t *Tab) GetSplitFraction(panel Panel) (float64, bool)
func (t *Tab) Collapse(panel Panel, size int) bool
func (t *Tab) Expand(panel Panel) bool
func (t *Tab) SetSplitCollapse(parent Panel, collapseRow int, onCollapse func() tea.Cmd)
func (t *Tab) ToggleSplitCollapse(parent Panel)
```

These methods return `false` when no matching node exists. `SetSplitCollapse` installs a callback for the collapse marker.

## Floats

```go
func (t *Tab) Float(panel Panel, x, y, width, height int)
func (t *Tab) CloseFloat(fp *FloatPane)
```

`Float` creates a floating pane above the layout. A float can be dragged, resized, and closed through mouse handling.

## Focus

```go
func (t *Tab) Focus() Panel
func (t *Tab) SetFocus(panel Panel) tea.Cmd
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

Warp does not bind `Tab` or `Shift+Tab`; the application chooses its key bindings.

## Panel lifecycle

```go
func (t *Tab) Update(msg tea.Msg) tea.Cmd
func (t *Tab) View(width, height int) string
func (t *Tab) HandleMouse(msg tea.MouseMsg) tea.Cmd
func (t *Tab) Elements(width, height int) []Element
func (t *Tab) BroadcastResize() tea.Msg
```

`Elements` exposes the tab's semantic UI tree. `BroadcastResize` creates a resize message for the current layout.

Popover menus are standalone `Popover` panels; `Tab` has no context-menu factory.
