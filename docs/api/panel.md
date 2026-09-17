---
title: Panel
description: The interface implemented by every panel in a Warp layout.
---

# Panel

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

`View` receives the panel's allocated size in terminal cells. `Update` handles Bubbletea messages and may return a command.

## BasePanel

```go
type BasePanel struct{}

func (BasePanel) View(width, height int) string
func (BasePanel) Update(msg tea.Msg) tea.Cmd
```

Embed `BasePanel` when a panel only needs the default empty view and no-op update behavior. Components that need focus or semantic elements add `Focusable` or `ElementProvider` separately.
