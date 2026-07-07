# `panel.go` Specification

## Overview

`panel.go` defines the core user-facing abstraction for content in the **warp** pane system. It declares the `Panel` interface that all custom panel implementations must satisfy, and provides a no-op default implementation via `BasePanel`.

## Behavior

- A `Panel` is a self-contained unit that knows how to render itself at a given size and how to react to Bubbletea messages.
- The `Update` method is invoked only with messages that arrived while the panel had focus.
- Panels may represent any kind of content: terminals, text, graphics, forms, etc.
- A panel is not responsible for layout or focus management; it only defines how it renders and responds to messages.

## Public API

### Type `Panel` (interface)

```go
type Panel interface {
    View(width, height int) string
    Update(msg tea.Msg) tea.Cmd
}
```

Methods:

- `View(width, height int) string`
  - Renders the panel content for the provided width and height (in cells).
  - Returns a string, typically the rendered view that the caller will place in a pane.

- `Update(msg tea.Msg) tea.Cmd`
  - Processes a Bubbletea message, such as a key press or mouse event.
  - Returns a `tea.Cmd` (or `nil`) to schedule asynchronous work or produce future messages.

### Type `BasePanel` (struct)

```go
type BasePanel struct{}
```

A default embeddable implementation of `Panel` with no behavior. Useful for rapid prototyping: embed `BasePanel` and override only the methods needed.

- `func (BasePanel) View(width, height int) string`
  - Always returns an empty string (`""`).

- `func (BasePanel) Update(msg tea.Msg) tea.Cmd`
  - Always returns `nil`.

## Important Implementation Details

- `Panel` is intentionally minimal: it depends only on `github.com/charmbracelet/bubbletea` (`tea.Msg` and `tea.Cmd`).
- `BasePanel` has zero fields and produces no output, making it safe to embed without adding state.
- The contract that `Update` receives only focus-time messages is enforced by the surrounding warp framework, not by this file.
- Future extensions to `Panel` (e.g. lifecycle hooks, mouse capture flags) would expand this interface; currently it is stable and purposefully narrow.
