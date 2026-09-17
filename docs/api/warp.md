# Warp

`Warp` is the root model of a Bubbletea TUI application. It wraps an
arbitrary root `Panel` (defaulting to a `TabGroup` with a single tab)
and forwards every incoming message straight to that panel without
interception.

In addition to being a Bubbletea model (`Init`, `Update`, `View`),
`Warp` exposes:

- Tab convenience methods that delegate to the root `TabGroup` when it
  is one.
- Optional HTTP serving of the live element tree for external
  inspection.
- A `Panel` adapter (`AsPanel`) so a `Warp` can be embedded inside
  another panel or another `Warp`.

## Public API

### Construction

` w := warp.New() `

`New` returns a `*Warp` whose root is a fresh `TabGroup` positioned at
`TabTop` with a single default tab. The model is safe to use immediately
after creation.

### Root panel access

` func (w *Warp) SetRoot(panel Panel) func (w *Warp) Root() Panel `

`SetRoot` replaces the root panel — install splits, flex containers,
nested tab groups, or any other `Panel` implementation. `Root` returns
the currently installed panel.

### Size accessors

` func (w *Warp) Width() int func (w *Warp) Height() int `

Return the last known terminal size. These values are updated from
`tea.WindowSizeMsg` during `Update` and are what `View` and the HTTP
endpoint use when rendering.

### Tab delegation

` func (w *Warp) NewTab(name string) *Tab func (w *Warp) ActiveTab() *Tab func (w *Warp) SetTabPosition(pos TabPosition) func (w *Warp) NextTab() func (w *Warp) PrevTab() `

These are pure convenience delegates: if the root panel is a `*TabGroup`
they forward to it; otherwise they are no-ops (return `nil` where a
value is expected). They exist so callers do not need to downcast the
root.

### Embedding as a panel

` func (w *Warp) AsPanel() Panel `

Returns a `warpPanel` adapter implementing the `Panel` interface (i.e.
`View(width, height)` and `Update(msg)`). This is how a `Warp` can be
nested inside another panel or another `Warp`.

### Lifecycle

` func (w *Warp) Run() error `

Starts the Bubbletea program (alt-screen + cell mouse motion) and blocks
until the program exits. The returned error is the program's exit error.

### HTTP serving

` func (w *Warp) ServeHTTP(addr string) error func (w *Warp) CloseHTTP() error func (w *Warp) HTTPAddr() string `

`ServeHTTP` starts a single-connection HTTP server on `addr` (or the
`WARP_HTTP_PORT` env var, or `:0` for a random port if empty) exposing
two endpoints:

| Path | Description |
|----|----|
| `/elements` | JSON array of the currently rendered element tree at the current (or 80x24 fallback) size |
| `/healthz` | Plain-text `ok` health probe |

The server is idempotent to start (second call while running is a no-op)
and idempotent to close. `HTTPAddr` returns the bound address, or empty
string when not serving.

## Bubbletea Model

` func (w *Warp) Init() tea.Cmd func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd) func (w *Warp) View() string `

`Init` returns `nil`. `Update` intercepts only `tea.WindowSizeMsg` to
record the window dimensions, then forwards every message (including the
size message) to `w.root.Update`. If the root panel is `nil`, `Update`
returns no command.

`View` renders `w.root.View(width, height)`. If the size is not yet
known (0x0) it renders the placeholder string `Loading...`; if the root
is `nil` it renders the empty string.

## Nesting

Because `AsPanel` exposes the `Panel` interface, a `Warp` can be placed
inside another panel (or another `Warp`). The adapter's `View` writes
the parent-supplied dimensions back into the inner `Warp` and then
renders its root, while `Update` simply forwards the message and returns
any resulting command.

``` go
parent := warp.New()
inner  := warp.New()
parent.SetRoot(inner.AsPanel()) // inner Warp embedded in parent
```

## Implementation notes

- `httpMu` serialises access to the HTTP server fields and is also
  re-entrantly held while `handleElements` snapshots `width`, `height`,
  and `root`. The `/elements` handler therefore renders the tree outside
  the lock while holding it only to publish the JSON.
- The HTTP server is intentionally minimal: it binds one `net.Listener`,
  serves two mux handlers, and never re-binds. Calling `ServeHTTP` twice
  (before the first closes) returns `nil` and keeps the original
  listener.
- The element-tree response always serialises as a JSON array; if the
  root is `nil` it is an empty array (not `null`).
- `parsePort` is an internal helper used to expose the bound port; it is
  intentionally unexported.
