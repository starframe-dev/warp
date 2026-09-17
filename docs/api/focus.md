---
title: Focus
description: Explicit focus management and raw-key preferences.
---

# Focus

Warp exposes focus operations but does not choose focus key bindings. The application decides whether `Tab`, `Shift+Tab`, or another key should move focus.

## Focusable

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

## RawKeyReceiver

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

`WantsRawKeys` identifies terminal-style panels, such as PTY-backed panels, that need raw keyboard input.

## Tab methods

```go
func (t *Tab) Focus() Panel
func (t *Tab) SetFocus(panel Panel) tea.Cmd
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

Focus traversal follows the visual order of focusable leaves in the panel tree. `FocusNext` and `FocusPrev` wrap around.

## Example

```go
case tea.KeyMsg:
    switch msg.String() {
    case "tab":
        tab.FocusNext()
    case "shift+tab":
        tab.FocusPrev()
    }
```
