---
title: Warp
description: Warp — root Bubbletea model
---

# Warp

Root Bubbletea model. Wraps a root `Panel` (default `TabGroup`).

## Constructor

```go
func New() *Warp
```

Creates a Warp with a root `TabGroup` and one default tab.

## Root Management

```go
func (w *Warp) SetRoot(panel Panel)
func (w *Warp) Root() Panel
```

`SetRoot` replaces the root panel. `Root` returns the current root.

## Tab Delegation

Delegated to the root `TabGroup`:

```go
func (w *Warp) NewTab(name string) *Tab
func (w *Warp) ActiveTab() *Tab
func (w *Warp) SetTabPosition(pos TabPosition)
func (w *Warp) NextTab()
func (w *Warp) PrevTab()
```

## Nesting

```go
func (w *Warp) AsPanel() Panel
```

Returns a `Panel` adapter for nesting Warp inside splits/flex.

## Lifecycle

```go
func (w *Warp) Run() error
```

Starts the Bubbletea program.

## HTTP API

```go
func (w *Warp) ServeHTTP(addr string) error
func (w *Warp) CloseHTTP() error
func (w *Warp) HTTPAddr() string
```

Serves the element tree at `/elements` and health check at `/healthz`.

## Dimensions

```go
func (w *Warp) Width() int
func (w *Warp) Height() int
```

Last known terminal dimensions.
