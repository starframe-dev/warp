# Warp

**Go TUI Layout Engine** — tabs, splits, flexbox, floating panels, modals, popover, and more.

Built on [Bubbletea](https://github.com/charmbracelet/bubbletea) and [Lipgloss](https://github.com/charmbracelet/lipgloss).

## Features

- **Tabs** — `TabGroup` as a `Panel`, nestable anywhere
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
- **Element tree** — semantic UI tree for E2E testing
- **Gruvbox Dark** theme

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

- [docs](https://starframe-dev.github.io/warp/docs/) — full documentation (English + Russian)
- [specs](specs/) — project specifications
- [Plans](https://github.com/starframe-dev/warp/issues) — roadmap and plans

## License

MIT