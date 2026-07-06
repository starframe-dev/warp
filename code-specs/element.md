# Specification: `element.go`

## Overview

The `element.go` file defines the core data types and utility functions for describing semantic UI elements within the `warp` package. It provides a tree-based representation of on-screen elements, their bounding regions, and interfaces/functions for producing and searching these element trees.

This file is part of the package: `warp`.

---

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

A semantic UI element with a screen region and optional child elements.

**Fields:**

- `Role`: The semantic role of the element (e.g., `"button"`, `"label"`, `"input"`).
- `Name`: A human-readable or test identifier name for the element.
- `Action`: An optional action identifier associated with the element. Omitted from JSON when empty.
- `Bounds`: The rectangular screen region occupied by the element, in cell coordinates.
- `Children`: Optional nested child elements. Omitted from JSON when empty.

**Behavior:**

- Elements form a hierarchical tree via the `Children` slice.
- Empty strings in optional fields are omitted from JSON output.

---

### `Bounds`

```go
type Bounds struct {
    X int `json:"x"`
    Y int `json:"y"`
    W int `json:"w"`
    H int `json:"h"`
}
```

Defines a rectangular region in cell coordinates.

**Fields:**

- `X`: X-coordinate of the top-left corner.
- `Y`: Y-coordinate of the top-left corner.
- `W`: Width of the region (number of columns).
- `H`: Height of the region (number of rows).

---

### `ElementProvider`

```go
type ElementProvider interface {
    Elements(width, height int) []Element
}
```

Interface implemented by panels that can expose their UI element trees.

**Methods:**

- `Elements(width, height int) []Element`: Returns the root elements describing the panel's UI for the given terminal dimensions.

**Behavior:**

- Implementations are expected to produce a complete element tree for the current layout.
- The returned slice may be empty if the panel has no semantic elements.

---

### `ElementProviderFunc`

```go
type ElementProviderFunc func(width, height int) []Element
```

Adapts a plain function to the `ElementProvider` interface.

**Methods:**

- `Elements(width, height int) []Element`: Invokes the underlying function and returns its result.

---

## Public API Functions

### `Bounds.Center`

```go
func (b Bounds) Center() (int, int)
```

Returns the center cell coordinate `(cx, cy)` of the bounds.

**Implementation Details:**

- Uses integer division: `cx = b.X + b.W/2`, `cy = b.Y + b.H/2`.
- For odd widths/heights, the center is rounded toward zero from the upper-left corner.

---

### `FindElement`

```go
func FindElement(elems []Element, role, name, action string) (Element, bool)
```

Recursively searches a slice of elements for the first element matching the given `role`, `name`, and `action`.

**Matching Rules:**

- An empty criterion string matches any value for that field.
- All provided (non-empty) criteria must match for an element to be considered a match.
- Search is depth-first: each element is checked before its children.

**Returns:**

- The matching `Element` and `true` if found.
- A zero-value `Element` and `false` if no match is found.

**Implementation Details:**

- Recursively traverses the `Children` tree.
- Time complexity is proportional to the number of nodes in the tree.

---

## Unexported Functions

### `collectElements`

```go
func collectElements(panel Panel, width, height int) []Element
```

Safely collects elements from a `Panel` if it implements `ElementProvider`.

**Behavior:**

- Returns `nil` if `panel` is `nil`.
- If the panel implements `ElementProvider`, returns the result of `panel.Elements(width, height)`.
- Otherwise returns `nil`.

**Usage:**

- Intended for internal use by panels or layouts that need to gather element trees from child panels without requiring every panel to implement `ElementProvider`.

---

## Important Implementation Details

- **JSON serialization:** All fields in `Element` and `Bounds` are serialized with lowercase keys. Optional fields (`Action`, `Children`) are omitted when empty due to the `omitempty` tag.
- **Panel integration:** `ElementProvider` is designed to be implemented by panel types in the `warp` package, allowing layouts and containers to expose their semantic structure for testing, automation, or accessibility purposes.
- **Function adapter:** `ElementProviderFunc` allows anonymous functions or package-level functions to be used wherever an `ElementProvider` is required, improving composability.
- **No mutation:** All types in this file are value types; none of the public functions modify the input element trees.

---

## Dependencies

- This file only uses the Go standard library.
- It depends on the `Panel` type defined elsewhere in the `warp` package.
