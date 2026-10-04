---
title: Focus & Input
description: Handle focus explicitly in Warp and route raw-key preferences to terminal-style panels.
---

# Focus & Input

Warp keeps focus explicit. Your app decides which keys move focus.

## Focusable

Panels that can receive focus implement:

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

## Tab focus helpers

```go
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)
```

Traversal follows the visual order of focusable leaves and wraps at either end.

## Key bindings

Warp does not bind `Tab` or `Shift+Tab`; bind those keys in the application when they fit its interaction model. The root `TabGroup` handles `Ctrl+T` (create a tab), `Ctrl+W` (close the active tab unless it is the last one), `Ctrl+Tab` / `Ctrl+Shift+Tab` (switch tabs), and `Ctrl+C` (quit).

## RawKeyReceiver

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

Use `RawKeyReceiver` for PTY or terminal panels that need every key before Warp shortcuts. When the focused panel returns `true`, it receives the key first, including `Ctrl+C`, so Warp does not treat that key as quit. The interface does not replace `Panel.Update`; it adds the `WantsRawKeys` signal.

## Example

```go
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.tab.FocusNext()
            return m, nil
        case "shift+tab":
            m.tab.FocusPrev()
            return m, nil
        }
    }

    _, cmd := m.warp.Update(msg)
    return m, cmd
}
```
