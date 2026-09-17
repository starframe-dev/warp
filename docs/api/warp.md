---
title: Warp
description: Root Bubbletea model, tab delegation, nesting, and HTTP inspection.
---

# Warp

`Warp` is the root Bubbletea model. It owns a root `Panel` and forwards messages to it.

## Constructor

```go
func New() *Warp
```

Creates a `Warp` with a `TabGroup` root and one `main` tab.

## Root and tab delegation

```go
func (w *Warp) SetRoot(panel Panel)
func (w *Warp) Root() Panel

func (w *Warp) NewTab(name string) *Tab
func (w *Warp) ActiveTab() *Tab
func (w *Warp) SetTabPosition(pos TabPosition)
func (w *Warp) NextTab()
func (w *Warp) PrevTab()
```

Tab helpers act only while the root is a `*TabGroup`. After `SetRoot` replaces it with another panel, tab helpers are no-ops and `ActiveTab`/`NewTab` return `nil`.

## Bubbletea model

```go
func (w *Warp) Init() tea.Cmd
func (w *Warp) Update(msg tea.Msg) (tea.Model, tea.Cmd)
func (w *Warp) View() string
```

`View` returns `Loading...` until a non-zero `WindowSizeMsg` arrives.

## Nesting

```go
func (w *Warp) AsPanel() Panel
```

Returns an adapter that passes allocated dimensions and messages to the nested Warp.

## HTTP element API

```go
func (w *Warp) ServeHTTP(addr string) error
func (w *Warp) CloseHTTP() error
func (w *Warp) HTTPAddr() string
```

`ServeHTTP` exposes:

- `GET /elements` — JSON element tree; defaults to `80 × 24` before the first resize;
- `GET /healthz` — plain-text `ok`.

An empty address uses `WARP_HTTP_PORT`; when the variable is unset, the OS chooses a free port. Calling `ServeHTTP` again while serving has no effect.

## Dimensions and lifecycle

```go
func (w *Warp) Width() int
func (w *Warp) Height() int
func (w *Warp) Run() error
```

`Width` and `Height` return the last received dimensions. `Run` starts Bubbletea with the alternate screen and mouse cell motion enabled.
