# Warp

`Warp` is the root model of a Bubble Tea TUI application. It wraps an
arbitrary root `Panel` (defaulting to a `TabGroup` with a single tab),
records incoming window dimensions, and forwards messages to that panel.

In addition to being a Bubble Tea model (`Init`, `Update`, `View`),
`Warp` exposes:

- Tab convenience methods that delegate to the root `TabGroup` when it
  is one.
- Optional HTTP serving of the latest UI-thread element snapshot for
  external inspection.
- A `Panel` adapter (`AsPanel`) so a `Warp` can be embedded inside
  another panel or another `Warp`.

## Public API

### Construction

` w := warp.New() `

`New` returns a `*Warp` whose root is a fresh `TabGroup` positioned at
`TabTop` with a single default tab. The model is safe to use immediately
after creation.

### Root panel access

```go
func (w *Warp) SetRoot(panel Panel)
func (w *Warp) Root() Panel
```

`SetRoot` replaces the root panel — install splits, flex containers,
nested tab groups, or any other `Panel` implementation. `Root` returns
the currently installed panel.

### Size accessors

```go
func (w *Warp) Width() int
func (w *Warp) Height() int
```

Return the last known terminal size. These values are updated from
`tea.WindowSizeMsg` during `Update`, or from the dimensions supplied to
`AsPanel().View` when Warp is embedded. `View` uses them to render the
root, and the UI thread uses them when building the HTTP element snapshot.

### Tab delegation

```go
func (w *Warp) NewTab(name string) *Tab
func (w *Warp) ActiveTab() *Tab
func (w *Warp) SetTabPosition(pos TabPosition)
func (w *Warp) NextTab()
func (w *Warp) PrevTab()
```

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

### Panel resource ownership

A root `Warp` owns its complete panel hierarchy, including nested tabs,
wrappers, and floats. A standalone `TabGroup` or `Tab` owns a separate
hierarchy until embedded in another owner. Panels implementing `Unmounter`
are unmounted once after they are detached and no longer reachable from that
owner. Hiding, collapsing, focus changes, and reparenting do not unmount a
panel. Pointer-backed panels use instance identity; sharing the same panel
instance across independent `Warp` roots is unsupported and remains the
caller's responsibility. Warp does not maintain a global ownership registry.

### HTTP serving

```go
func (w *Warp) ServeHTTP(addr string) error
func (w *Warp) CloseHTTP() error
func (w *Warp) HTTPAddr() string
```

`ServeHTTP` starts an HTTP server on `addr` and exposes two endpoints.
If `addr` is empty, it binds to `127.0.0.1` on `WARP_HTTP_PORT`, or on an
automatically assigned port when that variable is unset. A non-empty
address is used as supplied.

| Path | Description |
|----|----|
| `/elements` | JSON array from the latest completed UI-thread snapshot (built at the current size, or 80x24 when dimensions are unknown) |
| `/healthz` | Plain-text `ok` health probe |

The server is idempotent to start (second call while running is a no-op)
and idempotent to close. `HTTPAddr` returns the bound address, or empty
string when not serving.

## Bubbletea Model

```go
func (w *Warp) Init() tea.Cmd
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (w *Warp) View() string
```

`Init` returns `nil`. `Update` intercepts only `tea.WindowSizeMsg` to
record the window dimensions, then forwards every message (including the
size message) to `w.root.Update`. If the root panel is `nil`, `Update`
returns no command.

`View` renders `w.root.View(width, height)`. If either dimension is not
yet known (zero) it renders the placeholder string `Loading...`; if the
root is `nil` it renders the empty string. After rendering, it refreshes
the element snapshot.

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

- After `Update` and `View`, the UI thread collects elements and deep-copies their `Children` before publishing the immutable snapshot. Element collection invokes `Panel.Elements` without holding Warp's mutex.
- `/elements` reads only the most recently completed snapshot. It never traverses the live panel tree or invokes `Panel.Elements`, `Panel.View`, or `Panel.Update`; a response may be one UI operation behind, but cannot observe partially updated tree data.
- A root revision prevents a snapshot collected for an old root from being published after `SetRoot`. A nil root is serialized as an empty JSON array, not `null`.
- If dimensions are unknown while building a snapshot, the inspector uses 80×24.
- The HTTP server is idempotent to start while running. `CloseHTTP` detaches the server under the mutex, then calls `http.Server.Shutdown` without holding that mutex.
