---
title: Focus
description: Focus API
---

# Focus

Warp provides explicit focus management. The developer decides which keys trigger focus switching.

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
    HandleRawKey(msg tea.KeyMsg) tea.Cmd
}
```

For PTY/terminals that need all key events without interception.

## Tab Focus Methods

```go
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

## Example

```go
func (m *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.tab.FocusNext()
        case "shift+tab":
            m.tab.FocusPrev()
        }
    }
    return m, nil
}
```
