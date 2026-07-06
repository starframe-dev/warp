# Specification: `scrollable.go`

## Overview

`scrollable.go` provides a `Scrollable` wrapper component in the `warp` package that adds vertical scrolling behavior to any `Panel` component. When the wrapped panel renders more lines than the available height, the user can scroll the viewport using the mouse wheel or keyboard.

## Behavior

- `Scrollable` renders a fixed-height viewport of a child panel by slicing a virtual full-height render into a visible window.
- If the content does not exceed the viewport height, `Offset` is clamped to `0` and no scrolling occurs.
- If the child panel is `nil`, the component renders a blank area of the requested height by returning `h` newline characters.
- Each visible line is padded with spaces or truncated to the requested width `w` so that the viewport is rectangular.

## Public Types

### `Scrollable`

```go
type Scrollable struct {
    Content Panel
    Offset  int
}
```

| Field | Type | Description |
|-------|------|-------------|
| `Content` | `Panel` | The child panel being wrapped. |
| `Offset`  | `int`   | Current scroll offset measured in lines from the top of the rendered content. |

### `Panel` dependency

`Scrollable` depends on the `Panel` interface (defined elsewhere in the package). It calls:

- `Panel.View(w, h int) string`
- `Panel.Update(msg tea.Msg) tea.Cmd`

## Public API

### `func NewScrollable(content Panel) *Scrollable`

Creates a new `Scrollable` instance wrapping the given `content` panel. The initial `Offset` is `0`.

### `func (s *Scrollable) View(w, h int) string`

Renders the visible viewport of the content.

- Renders the full content at width `w` with effectively unlimited height (`9999` lines).
- Splits the full content into lines and clamps `Offset` to the valid range `[0, max(0, len(lines) - h)]`.
- Returns `h` lines of output. If a content line exists at `Offset + i`, it is used; otherwise a blank line is emitted.
- Each line is padded or truncated to width `w` using `padLine`.

### `func (s *Scrollable) Update(msg tea.Msg) tea.Cmd`

Processes scroll input and forwards the message to the wrapped panel.

Supported input:

| Message | Action |
|---------|--------|
| `tea.MouseMsg` with `tea.MouseButtonWheelUp` | Decrease `Offset` by `3` (scroll up), clamped to `>= 0`. |
| `tea.MouseMsg` with `tea.MouseButtonWheelDown` | Increase `Offset` by `3` (scroll down). |
| `tea.KeyMsg` with `"up"` | Decrease `Offset` by `1`, clamped to `>= 0`. |
| `tea.KeyMsg` with `"down"` | Increase `Offset` by `1`. |
| `tea.KeyMsg` with `"pgup"` | Decrease `Offset` by `10`, clamped to `>= 0`. |
| `tea.KeyMsg` with `"pgdown"` | Increase `Offset` by `10`. |

After handling scroll input, the message is passed to `s.Content.Update(msg)` if `Content` is not `nil`. The child may return a command that is propagated to the caller.

## Implementation Details

### `padLine(line string, w int) string`

Private helper that ensures a rendered line fits exactly within the viewport width.

- If `lipgloss.Width(line)` is less than `w`, it appends spaces to fill the width.
- If the line is wider than `w`, it iterates over byte indices and returns the longest prefix whose display width is `<= w` (specifically, it returns the first prefix that exceeds `w` minus the last byte). If no truncation is needed, the original line is returned.

### Rendering strategy

The wrapper uses a large constant height (`9999`) to obtain the full height of the child panel without imposing an explicit height constraint. This is simple but assumes that panel content fits within that limit. In practice, content taller than `9999` lines would be truncated at rendering time.

### Offset management

`Offset` is mutated in place during `Update`. Clamping for the upper bound happens only inside `View`, which means after a wheel/key event the offset can momentarily exceed the valid range. `View` corrects it before rendering. This avoids needing to know the content height during `Update`.

### Keyboard handling

Key messages are compared using `msg.String()`, which is the human-readable representation returned by Bubble Tea. This matches strings such as `"up"`, `"down"`, `"pgup"`, and `"pgdown"`.

## Dependencies

- `strings` – line splitting and padding.
- `github.com/charmbracelet/bubbletea` – input messages and commands.
- `github.com/charmbracelet/lipgloss` – display width calculation for styling-aware truncation and padding.
