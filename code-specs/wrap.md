# Specification: `wrap.go`

## Overview

`wrap.go` provides text-wrapping utilities for terminal-rendered strings. All functions use `github.com/charmbracelet/lipgloss.Width` to measure visible width, so ANSI escape sequences and multi-cell characters are handled correctly. The package supports wrapping at word boundaries, wrapping only at spaces (allowing overflow), and joining results back into a single string.

## Public API

### Functions

#### `WordWrap(text string, width int) []string`

- **Behavior**: Wraps `text` so that no output line exceeds `width` visible cells.
- **Input lines**: Splits `text` on `\n` first, preserving paragraph boundaries.
- **Word boundaries**: Preferentially breaks at whitespace. If a word is longer than `width`, it is broken mid-word at a visual boundary.
- **Width handling**: Returns `nil` if `width <= 0`.
- **Return**: Slice of wrapped lines, without trailing newline characters.

#### `SpaceWrap(text string, width int) []string`

- **Behavior**: Wraps `text` at spaces. No word is ever split; words longer than `width` are allowed to overflow onto their own line.
- **Input lines**: Splits `text` on `\n` first, preserving paragraph boundaries.
- **Whitespace handling**: Uses `strings.Fields` to tokenize by whitespace; multiple spaces collapse to a single separator.
- **Width handling**: Returns `nil` if `width <= 0`.
- **Return**: Slice of wrapped lines.

#### `WrapToString(text string, width int, useSpaceWrap bool) string`

- **Behavior**: Convenience wrapper that calls `SpaceWrap` or `WordWrap` and joins the resulting lines with `\n`.
- **Parameters**:
  - `text` — input text to wrap.
  - `width` — maximum visible width per line.
  - `useSpaceWrap` — if `true`, uses `SpaceWrap`; otherwise uses `WordWrap`.
- **Return**: A single string with embedded newline characters. No trailing newline is appended.

## Private implementation details

### `wrapLine(line string, width int) []string`

- Core implementation of `WordWrap` for a single line.
- **Short-circuit**: If `lipgloss.Width(line) <= width`, returns `[]string{line}`.
- **Long-word overflow**: When a word boundary cannot be found within the width, it advances by one visible cell boundary at a time using `lipgloss.Width(line[start:i+1])`.
- **Whitespace trimming**: Each emitted line has trailing whitespace removed via `strings.TrimRightFunc(..., unicode.IsSpace)`.
- **Leading whitespace**: After a break, leading whitespace on the next line is skipped.

### `wrapAtSpaces(line string, width int) []string`

- Core implementation of `SpaceWrap` for a single line.
- **Short-circuit**: If `lipgloss.Width(line) <= width`, returns `[]string{line}`.
- **Empty whitespace-only lines**: If `strings.Fields` returns no words, returns `[]string{line}` unchanged.
- **Greedy packing**: Builds each line greedily: adds a word if the current width plus one space plus the word width fits within `width`.
- **Overflow behavior**: A word wider than `width` is placed on its own line and will exceed the requested width.

### `isWordBreak(b byte) bool`

- Reports whether `b` is a whitespace byte using `unicode.IsSpace`.
- Currently unused by the active wrapping logic; kept as a helper for future word-boundary detection.

## Types

This file defines no new types or structs. All functions operate on plain `string` and `[]string` values.

## Important notes

- **Visible width, not byte count**: Width is measured with `lipgloss.Width`, which accounts for ANSI styles and wide runes.
- **Width `<= 0`**: All public entry points return `nil` (or an empty string through `WrapToString`) for non-positive widths.
- **Newlines preserved**: Any existing `\n` in the input are treated as line breaks before wrapping logic is applied.
- **Space collapsing**: `SpaceWrap` collapses consecutive whitespace and ignores leading/trailing whitespace within a line via `strings.Fields`.
- **Trailing whitespace trimmed**: `WordWrap` removes trailing whitespace from each emitted line.
