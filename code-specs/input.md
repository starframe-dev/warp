# `input.go` Specification

## Overview

`input.go` implements a single-line text input component for the `warp` package. It is designed to integrate with the [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework and uses [Lipgloss](https://github.com/charmbracelet/lipgloss) for styling. The component supports keyboard-driven editing, cursor movement, focus handling, and two rendering modes: inline and bordered box.

## Type: `Input`

```go
type Input struct {
    Value   string
    Cursor  int
    Prompt  string
    Width   int
    focused bool
}
```

A single-line text input field.

| Field | Description |
|-------|-------------|
| `Value` | Current text value. |
| `Cursor` | Cursor position measured in runes (Unicode code points). |
| `Prompt` | Static prefix shown before the value (e.g., a label or `> `). |
| `Width` | Desired width; `0` means the width is inferred from the `View` argument. |
| `focused` | Internal focus state. |

## Public API

### Construction

```go
func NewInput(prompt string) *Input
```

Creates a new empty input with the given prompt. The cursor starts at position `0` and the value is empty.

### Value & Cursor

```go
func (in *Input) SetValue(v string)
func (in *Input) SetCursor(pos int)
```

- `SetValue` replaces the entire value and moves the cursor to the end.
- `SetCursor` sets the cursor position in runes and clamps it to the valid range.

### Focus

```go
func (in *Input) Focused() bool
func (in *Input) Focus()
func (in *Input) Blur()
```

- `Focus` gives the input focus.
- `Blur` removes focus.
- `Focused` reports the current focus state.

Focus affects visual rendering (border color changes when focused) and determines whether keyboard input is processed.

### Rendering

```go
func (in *Input) View(w, h int) string
```

Renders the input as a string of dimensions `w × h`.

- If `h >= 3`, the input is rendered inside a bordered box (`viewBoxed`).
- Otherwise, it is rendered inline (`viewInline`).

The rendered output is designed to be placed in a terminal grid. The component uses the provided width to truncate or pad the content line as necessary.

### Bubble Tea Update Loop

```go
func (in *Input) Update(msg tea.Msg) tea.Cmd
```

Handles keyboard input when the component is focused. Non-key messages are ignored. When blurred, no key events are processed. Returns `nil` (no command) for all cases.

## Keyboard Handling

When focused, the input responds to the following keys:

| Key | Action |
|-----|--------|
| `backspace` | Deletes the rune before the cursor. |
| `delete` | Deletes the rune at the cursor. |
| `left` | Moves the cursor one rune left. |
| `right` | Moves the cursor one rune right. |
| `home` | Moves the cursor to the start of the value. |
| `end` | Moves the cursor to the end of the value. |
| `tab` / `shift+tab` | Intentionally ignored; reserved for parent focus traversal. |
| `enter` | No-op by default (submit hook is not implemented). |
| Printable characters | Inserts the character at the cursor position. |

The cursor is always clamped to the valid range `[0, len([]rune(Value))]` after each operation.

## Rendering Details

### Inline Mode (`h < 3`)

The prompt and value are rendered on a single line using `inputStyle`. The line is padded on the right to width `w` and duplicated across all `h` rows. This produces a filled block of `h` identical lines.

### Boxed Mode (`h >= 3`)

Draws a rounded box border using `inputBorderStyle` (unfocused) or `inputFocusBorderStyle` (focused). The interior is width `w-2` and height `h-2`. The content line is vertically centered; if the interior height is greater than one, the content line is also drawn on the top interior row to help centering.

The box uses the following Unicode drawing characters:
- `╭` / `╰` corners
- `─` horizontal line
- `│` vertical line

### Cursor Highlight

The cursor is rendered as a highlighted character using ANSI inverse video (`\x1b[7m ... \x1b[0m`). If the cursor is positioned after the last rune, a highlighted space is drawn to indicate the end-of-text cursor.

### Truncation

If `Prompt + Value` exceeds the available width, the value is truncated from the left so the cursor remains visible. The function `truncateTailToWidth` keeps a window of runes centered around the cursor near the right side of the visible area.

## Internal Functions

```go
func (in *Input) renderLine(maxW int) string
func (in *Input) insertAtCursor(s string)
func (in *Input) deleteBeforeCursor()
func (in *Input) deleteAtCursor()
func (in *Input) clampCursor()
func truncateTailToWidth(s string, maxW, cursor int) string
```

All text mutations operate on rune slices to ensure correct Unicode handling. The cursor position is always maintained in runes, not bytes.

## Styles

The component uses package-level Lipgloss styles:

```go
var (
    inputStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color(gbLight1))
    inputBorderStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color(gbDark4))
    inputFocusBorderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(gbBlue))
)
```

- `inputStyle` colors the text.
- `inputBorderStyle` colors the border when unfocused.
- `inputFocusBorderStyle` colors the border when focused.

These styles reference palette constants (`gbLight1`, `gbDark4`, `gbBlue`) defined elsewhere in the `warp` package.

## Dependencies

- `github.com/charmbracelet/bubbletea` — Bubble Tea message types and update loop.
- `github.com/charmbracelet/lipgloss` — Terminal styling and width measurement.
- `strings` — String manipulation utilities.

## Important Notes

- The cursor position is measured in **runes**, not bytes. This is essential for correct handling of multi-byte Unicode characters.
- `Update` returns `nil` for all messages, including `enter`; callers should implement submission logic externally if needed.
- Tab navigation is intentionally not handled inside the component; it is expected to be managed by a parent container.
- The `Width` field exists but is not used in `View`; the actual width is taken from the `w` argument passed to `View`.
- The boxed layout assumes a minimum height of 3; values below this fall back to inline rendering.
