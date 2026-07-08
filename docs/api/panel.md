---
title: Panel
description: Panel interface
---

# Panel

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

All UI components implement `Panel`. `View` renders the component, `Update` handles messages.
