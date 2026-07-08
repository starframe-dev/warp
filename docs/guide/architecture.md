---
title: Architecture
description: Core concepts behind Warp's panel tree, layout nodes, floats, focus, styles, and test element tree.
---

# Architecture

Warp is a Bubbletea-based TUI layout engine. It wraps one root `Panel`. By default, that root panel is a `TabGroup`.

## Panel

A `Panel` renders itself and reacts to Bubbletea messages:

```go
type Panel interface {
	View(w, h int) string
	Update(msg tea.Msg)
}
```

Warp builds layouts from a node tree. The tree can contain:

- Leaf panels.
- `SplitConfig` nodes for vertical and horizontal splits with a fraction.
- `FlexConfig` nodes for rows or columns with grow weights.

## Floats

Floats render on top of the normal layout output. A tab owns its floating panels and draws them over the split or flex tree.

## Tabs

`TabGroup` is a `Panel`. It renders a tab bar and a content area.

Each `Tab` manages its own tree of splits, flex nodes, and floats.

## Focus

Panels that accept focus implement `Focusable`:

```go
type Focusable interface {
	Focus()
	Blur()
	Focused() bool
}
```

Warp exposes explicit focus operations such as `FocusNext` and `FocusPrev`. Your app decides which keys trigger them.

## Styles

Warp uses the Gruvbox Dark palette from `styles.go`.

## Element tree

Panels can expose structured test data through `ElementProvider`. End-to-end tests can inspect this element tree without parsing terminal output.
