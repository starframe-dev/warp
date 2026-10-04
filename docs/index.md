---
title: Warp — Go TUI Layout Engine
description: Tabs, splits, flex, floats, modals, popover — a complete layout engine for Bubbletea
---

# Warp

**Go TUI Layout Engine** — tabs, splits, flexbox, floating panels, modals, popover, and more.

Built on [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

```go
w := warp.New()
tab := w.ActiveTab()
tab.SplitVertical(tab.RootPanel(), 0.5, leftPanel)
tab.SplitVertical(tab.RootPanel(), 0.5, rightPanel)
w.Run()
```

## Features

- **Tabs** — TabGroup as a local Panel component, nestable anywhere
- **Splits** — vertical/horizontal with draggable borders
- **Flexbox** — row/column with grow weights
- **Floats** — draggable, resizable, closable panels
- **Collapsible** — expand/collapse sections
- **Scrollable** — viewport with mouse wheel
- **Dropdown** — menu with hover and select
- **Selectable** — text selection with mouse and keyboard
- **Input** — single-line text input with cursor
- **Modal** — dialog windows with overlay
- **Popover** — context menus
- **Focus API** — explicit focus switching (developer decides keys)
- **Element tree** — semantic UI tree for E2E testing, optionally exposed through an HTTP inspector
- **Custom themes** — runtime component colors through `SetTheme` (Gruvbox Dark by default)
- **Resource lifecycle** — optional `Unmounter` hook for panels permanently removed from their owner

The HTTP inspector binds to loopback by default. If you choose an externally reachable address, configure access controls before exposing UI data.

## Quick start

```bash
go get github.com/starframe-dev/warp
```

```go
package main

import (
    "github.com/starframe-dev/warp"
)

func main() {
    w := warp.New()
    tab := w.ActiveTab()
    tab.Float(&myPanel{}, 10, 5, 20, 10)
    w.Run()
}
```

## Demo

```bash
go run ./cmd/demo/
```

## Documentation

- [English Guide](/guide/getting-started)
- [Русское руководство](/ru/guide/getting-started)
- [API Reference](/api/)
- [API на русском](/ru/api/)
