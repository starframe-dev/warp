# focus.go — Keyboard Focus Management

This module defines the focus-related panel interfaces and the internal
helpers used to collect focusable panels and move focus through them.

## Public API

### `Focusable` interface

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

A panel that can receive keyboard focus implements `Panel` and the three
focus methods. `Focus` is called when focus is assigned, `Blur` when it is
removed, and `Focused` reports whether the panel currently holds focus.

| Method | Signature | Description |
|---|---|---|
| `Focus` | `Focus()` | Called when the panel gains focus. |
| `Blur` | `Blur()` | Called when the panel loses focus. |
| `Focused` | `Focused() bool` | Reports whether the panel currently holds focus. |

### `RawKeyReceiver` interface

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

A panel implementing this interface can indicate whether it wants raw keys.
As described by the code comment, `TabGroup` forwards every key to the
focused panel before handling Warp shortcuts. In particular, Ctrl+C is not
treated as quit while the focused receiver returns `true` from
`WantsRawKeys()`.

## Internal API (unexported)

These helpers support focus resolution and navigation within the package.

### `isFocusable`

```go
func isFocusable(panel Panel) (Focusable, bool)
```

Returns the panel as a `Focusable` and `true` if it implements the interface.
Nil panels, including typed nil panel values, return `nil, false`.

### `collectFocusables`

```go
func collectFocusables(node *Node) []Focusable
```

Walks the tree rooted at `node` and returns its focusable leaf panels in
visual traversal order: split first child before second child, and flex items
in their declared order. A nil node produces an empty list.

### `focusIndex`

```go
func focusIndex(list []Focusable, current Panel) int
```

Returns the index of the first panel in `list` that matches `current`, or
`-1` if `current` is nil or no panel matches. Panel matching compares
comparable values by equality and otherwise uses deep equality.

### `focusNext` / `focusPrev`

```go
func focusNext(list []Focusable, current Panel) Focusable
func focusPrev(list []Focusable, current Panel) Focusable
```

Return the next or previous panel with wrap-around, respectively, and return
`nil` for an empty list. If `current` is absent from a non-empty list,
`focusNext` returns the first item and `focusPrev` returns the last item.

### `applyFocus`

```go
func applyFocus(current, next Focusable)
```

If the two values do not match, calls `Blur` on the current panel when it is
non-nil, then calls `Focus` on the next panel when it is non-nil. Matching is
performed by the same panel comparison used by `focusIndex`; no callbacks
are made when the values match.

### `isNilPanel` / `samePanel`

```go
func isNilPanel(panel Panel) bool
func samePanel(a, b Panel) bool
```

`isNilPanel` detects both a nil interface and typed nil values of nil-able
kinds. `samePanel` treats two nil panels as equal, compares non-nil values
with different dynamic types as unequal, uses equality for comparable types,
and falls back to `reflect.DeepEqual` for non-comparable types.

### `appendFocusables`

```go
func appendFocusables(result *[]Focusable, node *Node)
```

Appends focusable leaf panels from `node` to `result` using the same traversal
order as `collectFocusables`; nil nodes are ignored.

## Behavioral Notes

- Navigation wraps around the focusable list. If the current panel is not in
'the list, the next panel is the first item and the previous panel is the last.
- `applyFocus` avoids redundant callbacks when the current and next values
  match according to `samePanel`.
- The focus helpers do not maintain global focus state; they operate on their
  explicit arguments and invoke only the `Focus` and `Blur` callbacks.

## Usage Example

```go
// Pseudocode for a key handler driving Tab navigation.
list := collectFocusables(root)
next := focusNext(list, current)
applyFocus(current, next)
```
