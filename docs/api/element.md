# Element

`element.go` defines semantic UI elements, their screen bounds, and helpers for providing, clipping, and searching element trees.

An element contains a role, a name, an optional action, bounds in terminal-cell coordinates, and optional child elements.

## Types

### `Element`

```go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

`Role`, `Name`, and `Bounds` are always included in JSON. `Action` is omitted when empty, and `Children` is omitted when it has zero elements.

### `Bounds`

```go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

Defines a rectangle in terminal-cell coordinates, with its top-left corner at `(X, Y)` and dimensions `W` by `H`.

### `Bounds.Center`

```go
func (b Bounds) Center() (int, int)
```

Returns `(X + W/2, Y + H/2)`, using integer division.

### `ElementProvider`

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

Panels can implement this interface to provide root elements for the given panel dimensions, in cells. Child elements are nested in each element's `Children` slice.

### `ViewportElementProvider`

```go
type ViewportElementProvider interface {
    ElementsAt(width, height, offset int) []Element
}
```

A provider can implement this interface to return semantic elements intersecting a requested content viewport. The returned bounds remain relative to the full content origin. The viewport dimensions and content offset are supplied in cells.

### `ElementProviderFunc`

```go
type ElementProviderFunc func(width, height int) []Element

func (f ElementProviderFunc) Elements(width, height int) []Element
```

Adapts a function to `ElementProvider` by forwarding the dimensions and returning the function's result.

### `SemanticStableMsg`

```go
type SemanticStableMsg interface {
    SemanticStateUnchanged() bool
}
```

A message type may implement this interface to promise that processing a particular message does not change the semantic element tree. When the HTTP inspector is enabled, a message returning `true` allows Warp to retain the previously published immutable snapshot after `Update` and defer rebuilding elements until the next `View`. This is an explicit correctness promise; messages that do not implement the interface or return `false` use the conservative update behavior.

## Functions

### `collectElements`

```go
func collectElements(panel Panel, width, height int) []Element
```

Package-private helper. Returns `nil` if the panel is nil or does not implement `ElementProvider`; otherwise calls `Elements(width, height)` and returns its result.

### `FindElement`

```go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Searches recursively in depth-first pre-order and returns the first element matching all non-empty criteria. Empty role, name, or action arguments are wildcards. Returns the zero value and `false` if no match is found or the search reaches its safety limit.

```go
if el, ok := FindElement(panel.Elems, "", "submit", ""); ok {
    x, y := el.Bounds.Center()
    // Use (x, y) as the element's center cell.
}
```

## Behavior notes

- Tree traversal and cloning use limits of 128 levels and 100,000 nodes. Cloning truncates trees beyond these limits; `FindElement` stops searching when a limit is reached.
- Warp clones provider-owned semantic data before translating or clipping it, so providers should not rely on Warp mutating the slices they return.
- Bounds use terminal-cell coordinates, not pixels.
