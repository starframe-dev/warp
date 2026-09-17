# Element

The `element.go` file defines the representation of semantic UI elements
and the helpers used to expose, collect and look them up inside a panel.
An element is a tree node: it carries a *role*, an optional *name* and
an optional *action*, a screen *bounds* (a cell-coordinates rectangle),
and a list of child elements.

## Types

### `Element`

``` go
type Element struct {
    Role     string    `json:"role"`
    Name     string    `json:"name"`
    Action   string    `json:"action,omitempty"`
    Bounds   Bounds    `json:"bounds"`
    Children []Element `json:"children,omitempty"`
}
```

JSON encoding notes:

- `Role` and `Name` are always present.
- `Action` is omitted when empty.
- `Children` is omitted when the slice is empty (use the pointer form of
  `json` tag).

### `Bounds`

``` go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

Defines a rectangular screen region in cell coordinates: top-left corner
at `(X, Y)`, width `W`, height `H`.

### `Bounds.Center`

``` go
func (b Bounds) Center() (int, int)
```

Returns the centre cell as integer coordinates using integer division:
`(X + W/2, Y + H/2)`.

### `ElementProvider`

``` go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

Implemented by panels that can expose their UI elements. The method
takes the current panel dimensions in cells and returns the root list of
elements (children are nested inside each `Element`).

### `ElementProviderFunc`

``` go
type ElementProviderFunc func(width, height int) []Element

func (f ElementProviderFunc) Elements(width, height int) []Element
```

Adapts a plain function to the `ElementProvider` interface. Useful for
one-off providers that don't warrant a struct.

## Functions

### `collectElements`

``` go
func collectElements(panel Panel, width, height int) []Element
```

Package-private helper. Returns `nil` when `panel` is nil. Otherwise it
type-asserts the panel to `ElementProvider` and calls
`Elements(width, height)`; if the assertion fails, `nil` is returned.
Used by the warp pipeline to collect elements from a panel.

### `FindElement`

``` go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Recursively searches a list of elements for the first element whose
role, name and action match. Empty search strings are treated as
wildcards (match anything). Returns the matched `Element` and `true`
when found, zero value and `false` otherwise.

``` go
if el, ok := FindElement(panel.Elems, "", "submit", ""); ok {
    x, y := el.Bounds.Center()
    // press at (x, y)
}
```

## Behaviour Notes

- The tree is *ordered*: `FindElement` returns the *first* match in a
  depth-first pre-order traversal, so the caller is responsible for
  disambiguating by role/name/action.
- `collectElements` is the only path through which the warp runtime
  gathers elements. It treats "no provider" and "nil panel" identically
  (both yield `nil`).
- Bounds are in cell coordinates, not pixels — the warp runtime scales
  them to pixels at draw time.
