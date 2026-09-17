---
title: Element tree
description: Semantic UI elements and bounds for testing and HTTP inspection.
---

# Element tree

Panels can expose a semantic representation of their UI. Warp uses it for E2E tests and the `/elements` HTTP endpoint.

## Element

```go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

`Role`, `Name`, and `Action` identify the element. `Children` allows nested controls.

## Bounds

```go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}

func (b Bounds) Center() (int, int)
```

Coordinates use terminal cells. `Center` returns the cell at the rectangle center.

## ElementProvider

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}

type ElementProviderFunc func(width, height int) []Element
```

Use `ElementProviderFunc` to adapt a function without defining a new type.

## FindElement

```go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Empty filter values match any value. The function searches children recursively and returns the first match.

## HTTP

`Warp.ServeHTTP` exposes the current tree at `GET /elements` as JSON. `GET /healthz` returns `ok`.
