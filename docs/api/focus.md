# focus.go — Keyboard Focus Management

This module implements keyboard focus management for the `warp` terminal
UI layout engine. It defines the `Focusable` capability a panel must opt
into, and provides tree traversal helpers that compute the ordered list
of focusable panels and the navigation primitives (next / previous) used
to move the caret through that list.

## Public API

### `Focusable` interface

``` go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

A panel that can receive keyboard focus must embed a `Panel` and
additionally implement three methods: `Focus` (called when the panel
gains focus), `Blur` (called when it loses focus), and `Focused` which
reports whether the panel currently holds focus.

| Method | Signature | Description |
|----|----|----|
| `Focus` | `Focus()` | Invoked when focus is transferred to this panel. |
| `Blur` | `Blur()` | Invoked when focus leaves this panel. |
| `Focused` | `Focused() bool` | Returns `true` if the panel is the current focus owner. |

### `RawKeyReceiver` interface

``` go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

A panel that wants to receive *all* keyboard input without interception
(for example, a terminal emulator running inside a PTY). The runtime
consults `WantsRawKeys` to decide whether to skip its normal key
dispatch and forward raw key events directly to the receiver.

## Internal API (unexported)

The following helpers are unexported and are part of the package's
internal focus-resolution pipeline.

### `isFocusable`

``` go
func isFocusable(panel Panel) (Focusable, bool)
```

Performs a safe type assertion. Returns the `Focusable` value (or `nil`)
and a boolean flag indicating whether the supplied `Panel` actually
implements the interface.

### `collectFocusables`

``` go
func collectFocusables(node *Node) []Focusable
```

Walks the layout tree rooted at `node` and returns every `Focusable`
panel in *visual order* (first-split children before second-split; flex
items in their declared order). The result is used as the
focus-navigation list for the next / previous key handlers.

### `focusIndex`

``` go
func focusIndex(list []Focusable, current Panel) int
```

Locates the position of `current` inside `list`. Returns `-1` when the
panel is not present in the list.

### `focusNext` / `focusPrev`

``` go
func focusNext(list []Focusable, current Panel) Focusable
func focusPrev(list []Focusable, current Panel) Focusable
```

Wrap-around navigation primitives. `focusNext` moves to the following
panel in `list`, wrapping from the last element back to the first.
`focusPrev` moves in the opposite direction. Both return `nil` if the
list is empty.

### `applyFocus`

``` go
func applyFocus(current, next Focusable)
```

Executes the side-effects of a focus transition: calls `Blur` on the
outgoing panel (if it is a distinct non-nil value) and `Focus` on the
incoming panel.

## Behavioral Notes

- Focus traversal respects the *visual order* of the layout tree: for
  split nodes, the first child is always visited before the second; for
  flex nodes, items are visited in their declared slice order.
- Wrap-around is intentional: pressing "next" at the last panel wraps to
  the first, and "previous" at the first wraps to the last.
- `applyFocus` suppresses the `Blur` call when the outgoing panel is
  `nil` or is the same object as the incoming panel, preventing
  redundant focus churn.
- The module contains no global state and no hidden side effects beyond
  the documented `Focus` / `Blur` callbacks; all data flows through the
  explicit function parameters above.

## Usage Example

``` go

// Pseudocode for a key handler driving Tab navigation.
list := collectFocusables(root)
next  := focusNext(list, current)
applyFocus(current, next)

```
