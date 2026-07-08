---
title: TabGroup
description: TabGroup — Panel with tab bar and content area
---

# TabGroup

A `Panel` that displays a tab bar and switches between tabs. Can be nested inside splits/flex layouts.

## Constructor

```go
func NewTabGroup(pos TabPosition) *TabGroup
```

## Tab Management

```go
func (tg *TabGroup) NewTab(name string) *Tab
func (tg *TabGroup) ActiveTab() *Tab
func (tg *TabGroup) NextTab()
func (tg *TabGroup) PrevTab()
```

`NewTab` creates a tab and switches to it. `NextTab`/`PrevTab` cycle through tabs.

## TabPosition

```go
type TabPosition int

const (
    TabTop    TabPosition = iota
    TabBottom
    TabLeft
    TabRight
    TabNone
)
```

## Panel Interface

```go
func (tg *TabGroup) View(w, h int) string
func (tg *TabGroup) Update(msg tea.Msg) tea.Cmd
func (tg *TabGroup) Elements(w, h int) []Element
```

## Hotkeys

- `Ctrl+Tab` — next tab
- `Ctrl+Shift+Tab` — previous tab
- `Ctrl+W` — close active tab
- `Ctrl+T` — new tab
