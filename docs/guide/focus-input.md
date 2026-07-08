---
title: Focus & Input
description: Handle focus explicitly in Warp and route raw keys to terminal-style panels.
---

# Focus & Input

Warp keeps focus explicit. Your app decides which keys move focus.

## Focusable

Panels that can receive focus implement `Focusable`:

```go
type Focusable interface {
	Focus()
	Blur()
	Focused() bool
}
```

## Tab focus helpers

`Tab` provides helpers for focus movement:

```go
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)
```

Use these from your app update loop or command handlers.

## Key bindings

Warp does not bind `Tab` or `Shift+Tab` for you. Bind those keys in your app if they fit your interaction model.

## RawKeyReceiver

Use `RawKeyReceiver` for PTY or terminal panels that need every key without Warp intercepting it.

## Example

Switch focus on `Tab` from your app's `Update()` method:

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

	m.warp.Update(msg)
	return m, nil
}
```
