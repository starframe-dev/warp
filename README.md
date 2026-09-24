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
- **Custom themes** — process-wide component colors through `SetTheme`
- **Gruvbox Dark** default theme

### Keyboard precedence

`TabGroup` handles Warp shortcuts by default (`Ctrl+C`, tab navigation, `Ctrl+W`, and `Ctrl+T`). A focused panel implementing `RawKeyReceiver` with `WantsRawKeys() == true` receives every key first, so it can use those combinations itself. Normal panels retain Warp shortcuts, including `Ctrl+C` to quit.

## Theme and HTTP inspector

`SetTheme` updates Warp's package-wide component styles. Configure the theme before starting the Bubble Tea program; changing it while rendering is not synchronized.

`Warp.ServeHTTP("")` binds the inspector to `127.0.0.1` on `WARP_HTTP_PORT`, or an automatically assigned port when the variable is unset. `HTTPAddr()` returns the actual address. `/elements` exposes the semantic UI tree and allows cross-origin requests; the endpoint has no authentication. The tree is snapshotted on the UI thread after `Update`/`View`, so HTTP returns the latest completed snapshot rather than traversing live panels. An explicitly supplied non-loopback address (for example, `:8080`) may expose UI data to the network, so secure it before use outside a trusted local environment.

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

- [docs](https://starframe-dev.github.io/warp/docs/) — curated VitePress documentation (English + Russian)
- [specs](specs/) — project specifications
- [Plans](https://github.com/starframe-dev/warp/issues) — roadmap and plans

### Documentation layout

- `docs/guide/` and `docs/api/` — English VitePress pages
- `docs/ru/guide/` and `docs/ru/api/` — Russian VitePress pages

## License

MIT. See [LICENSE](LICENSE).