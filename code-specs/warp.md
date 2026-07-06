# warp.go Specification

## Overview

`warp.go` defines the top-level `Warp` model, which is the root Bubble Tea model for the Starframe Warp UI framework. `Warp` wraps a root `Panel` and exposes lifecycle methods, tab conveniences, an optional HTTP introspection server, and a `Panel` adapter so that a `Warp` can be nested inside another `Warp`.

## Behavior

- `Warp` implements the `tea.Model` interface (`Init`, `Update`, `View`, `Run`).
- It forwards every Bubble Tea message to its root `Panel` without interception, with the exception of `tea.WindowSizeMsg`, which it stores as `width` and `height` before forwarding.
- The default root panel is a top-positioned `TabGroup` created with one default tab.
- The root panel can be replaced with any custom `Panel` (e.g., splits, flex layouts, nested tab groups) via `SetRoot`.
- `Warp` can optionally serve an HTTP endpoint that exposes the current element tree as JSON, useful for debugging and external tooling.
- The `AsPanel` adapter allows a `Warp` instance to be embedded as a panel inside another `Warp`, enabling nested applications.

## Public API

### Types

#### `Warp`

```go
type Warp struct {
    root   Panel
    width  int
    height int

    httpServer *http.Server
    httpAddr   string
    httpMu     sync.Mutex
}
```

The root Bubble Tea model. Holds the root panel, terminal dimensions, and optional HTTP server state.

- `root` — the current root `Panel`.
- `width` / `height` — last known terminal dimensions from `tea.WindowSizeMsg`.
- `httpServer` / `httpAddr` / `httpMu` — state for the optional HTTP introspection server, guarded by a mutex.

### Constructors

#### `New() *Warp`

Creates a new `Warp` whose root is a `TabGroup` positioned at the top (`TabTop`).

### Root Panel Management

#### `SetRoot(panel Panel)`

Replaces the current root panel with a custom layout. This is the intended way to install splits, flex layouts, nested tab groups, or any other panel implementation.

#### `Root() Panel`

Returns the current root panel.

### Dimension Access

#### `Width() int`

Returns the last known terminal width.

#### `Height() int`

Returns the last known terminal height.

### TabGroup Convenience Delegates

These methods delegate to the root panel when it is a `*TabGroup`. If the root is not a `TabGroup`, they are no-ops (or return `nil`).

| Method | Behavior |
|--------|----------|
| `NewTab(name string) *Tab` | Creates a new tab in the root `TabGroup`, returning it, or `nil` if the root is not a `TabGroup`. |
| `ActiveTab() *Tab` | Returns the currently active tab, or `nil` if the root is not a `TabGroup`. |
| `SetTabPosition(pos TabPosition)` | Sets the tab bar position on the root `TabGroup` (e.g., `TabTop`, `TabBottom`, `TabLeft`, `TabRight`). |
| `NextTab()` | Activates the next tab in the root `TabGroup`. |
| `PrevTab()` | Activates the previous tab in the root `TabGroup`. |

### Bubble Tea Model Methods

#### `Init() tea.Cmd`

Initialization command. Currently returns `nil` because `Warp` itself needs no startup command.

#### `Update(msg tea.Msg) (tea.Model, tea.Cmd)`

Handles incoming Bubble Tea messages. On `tea.WindowSizeMsg` it records the new terminal size. All messages are forwarded to the root panel's `Update` method. Returns `w` and the command produced by the root panel, or `nil` if there is no root panel.

#### `View() string`

Renders the root panel by calling `root.View(w.width, w.height)`. Returns an empty string if no root panel is set.

#### `Run() error`

Starts the Bubble Tea program with alternate screen and mouse cell-motion enabled. Returns any error from `p.Run()`.

### Nesting

#### `AsPanel() Panel`

Returns a `Panel` adapter that wraps this `Warp`, enabling a `Warp` to be embedded inside another `Warp` as a panel.

### HTTP Introspection Server

#### `ServeHTTP(addr string) error`

Starts an HTTP server in a goroutine that exposes the current element tree.

- If the server is already running, returns `nil` immediately.
- If `addr` is empty, it falls back to the `WARP_HTTP_PORT` environment variable, or `:0` if that is unset.
- Registers the following handlers:
  - `GET /elements` — returns the current element tree as JSON.
  - `GET /healthz` — returns `200 OK` with body `ok`.
- The actual listening address is stored in `httpAddr` and can be retrieved via `HTTPAddr()`.
- Returns an error if TCP listening fails.

#### `CloseHTTP() error`

Stops the running HTTP server with `context.Background()` and clears the server state. If no server is running, returns `nil`.

#### `HTTPAddr() string`

Returns the current HTTP listening address (e.g., `127.0.0.1:12345`), or an empty string if no server is running.

## Internal Types

### `warpPanel`

```go
type warpPanel struct {
    warp *Warp
}
```

A private adapter that lets a `Warp` satisfy the `Panel` interface so it can be nested inside another `Warp`.

#### `View(width, height int) string`

Sets the warp's internal dimensions and returns `warp.View()`.

#### `Update(msg tea.Msg) tea.Cmd`

Forwards the message to `warp.Update` and returns the resulting command.

## Important Implementation Details

- **No message interception:** `Warp` does not interpret messages itself except to store the terminal size. Custom key bindings, focus management, or quit behavior must be handled inside the root panel.
- **Tab delegates are root-type dependent:** The tab helper methods only work when the root panel is a `*TabGroup`. Using `SetRoot` with a non-TabGroup panel means these helpers become no-ops.
- **HTTP server state is mutex-guarded:** `httpMu` protects `httpServer`, `httpAddr`, and the dimensions/root snapshots used by the `/elements` handler. This makes the server safe to start/stop from concurrent callers.
- **HTTP defaults:** When no address is provided, the server binds to a port from `WARP_HTTP_PORT` or to an OS-assigned port (`:0`).
- **Nested dimensions:** The `warpPanel` adapter writes incoming dimensions into `warp.width` and `warp.height` before rendering, so the inner `Warp` always uses the space allocated by the outer panel.
- **Unexported helper:** `parsePort(addr string) string` is defined but unused in the current file. It splits a host:port string and returns the port portion, or an empty string on error.
- **Element collection:** The `/elements` handler uses `collectElements(root, width, height)` (defined elsewhere) to produce a JSON representation of the visible UI tree. It defaults to `80x24` if no terminal size has been reported yet.
- **CORS:** The `/elements` response sets `Access-Control-Allow-Origin: *`, allowing browsers and external tools to consume the JSON tree without restriction.

## Dependencies

- `context` — HTTP server shutdown.
- `encoding/json` — JSON encoding of the element tree.
- `fmt` — Error wrapping for HTTP listen failures.
- `net` / `net/http` — HTTP server and listener.
- `os` — Reading `WARP_HTTP_PORT`.
- `sync` — Mutex for HTTP server state.
- `github.com/charmbracelet/bubbletea` — Bubble Tea framework types.
