# wrap.go

Text-wrapping utilities for the `warp` package. Widths are measured in terminal cells by `github.com/charmbracelet/x/ansi`; ANSI escape sequences are not counted as visible columns, and Unicode grapheme clusters are handled by that package.

## Public API

### `WordWrap`

```go
func WordWrap(text string, width int) []string
```

Wraps text using `ansi.Wrap`, with spaces as word-break opportunities. Long words are hard-wrapped when possible. Each input line is wrapped independently; input is split at `\n`, and output lines are separated into slice elements. Any wrapped line that still measures wider than `width` is truncated to `width` terminal cells.

- If `width <= 0`, the result is `nil`.
- Whitespace and ANSI sequences are handled by the underlying `ansi` package; whitespace is not normalized through `strings.Fields`.

```go
lines := warp.WordWrap("the quick brown fox", 10)
// lines contains "the quick" and "brown fox"
```

### `SpaceWrap`

```go
func SpaceWrap(text string, width int) []string
```

Wraps text using `ansi.Wordwrap`, with spaces as word-break opportunities. It does not hard-wrap words that are longer than `width`, so such words can produce lines wider than the requested width. Each input line is wrapped independently, after splitting the input at `\n`.

- If `width <= 0`, the result is `nil`.
- Whitespace handling follows the underlying `ansi` package; whitespace is not tokenized with `strings.Fields`.

```go
lines := warp.SpaceWrap("hello   world foo", 12)
// The words are wrapped at available breakpoints; whitespace handling follows ansi.Wordwrap.
```

### `WrapToString`

```go
func WrapToString(text string, width int, useSpaceWrap bool) string
```

Calls `SpaceWrap` when `useSpaceWrap` is true and `WordWrap` otherwise, then joins the returned lines with `\n`. If `width <= 0`, the wrapped slice is nil and the result is an empty string.

```go
s := warp.WrapToString("the quick brown fox", 10, false)
// s == "the quick\nbrown fox"
```

## Implementation notes

- `WordWrap` uses `ansi.Wrap(line, width, " ")`. If any returned line has an ANSI-aware visual width greater than `width`, it is passed through `ansi.Truncate` with an empty suffix.
- `SpaceWrap` uses `ansi.Wordwrap(line, width, " ")` and does not apply a post-wrap truncation step.
- `isWordBreak` is an unexported helper that is not used by these wrapping functions.
- Both functions return `nil` for non-positive widths. Otherwise, each processes the results of `strings.Split(text, "\n")` and appends the underlying wrapper's lines.

## Dependency

The wrapping and terminal-width behavior comes from `github.com/charmbracelet/x/ansi`.
