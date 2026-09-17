# wrap.go

Text-wrapping utilities for the `warp` package. All width measurements
are *visual* widths, computed with `lipgloss.Width` (a fork of
go-runewidth) — this means ANSI escape sequences, zero-width joiners and
other control characters are handled correctly when deciding where to
break a line. A *visual width* of 0 characters means "the characters a
human eye sees", not the byte length of the string.

## Public API

### `WordWrap`

``` typescript
func WordWrap(text string, width int) []string
```

Wraps text at *word* boundaries so no line exceeds `width` visual
columns. Words that are longer than `width` are broken in the middle
(hard break). A line in the input text that is itself shorter than
`width` is left as-is. The input is split on `"\n"` first, then each
line is wrapped independently — output lines never contain `\n`.

- If `width <= 0`, the result is `nil`.
- Input text without any characters wider than `width` is returned
  unchanged.
- Each element of the returned slice is one visual line, at most `width`
  columns wide.

``` go
lines := warp.WordWrap("the quick brown fox", 10)
// lines[0] == "the quick"
// lines[1] == "brown fox"
// lines[2] == nil (if width <= 0)
```

### `SpaceWrap`

``` typescript
func SpaceWrap(text string, width int) []string
```

Wraps text at *space* boundaries so no line exceeds `width` visual
columns. Unlike `WordWrap`, this does *not* break words in the middle —
a single word longer than `width` overflows the line and the result is
still produced. Word boundaries are taken from `strings.Fields` (any run
of whitespace collapses to a single space).

- If `width <= 0`, the result is `nil`.
- Empty input returns a single empty line.

``` go
lines := warp.SpaceWrap("hello   world foo", 12)
// lines[0] == "hello"
// lines[1] == "world foo"
```

### `WrapToString`

``` typescript
func WrapToString(text string, width int, useSpaceWrap bool) string
```

Convenience wrapper around `WordWrap` / `SpaceWrap` that joins the
resulting lines with `"\n"`. It is the only exported function that
returns a string directly; the slice-returning functions above give the
caller control over how lines are rendered (e.g. a TUI renderer that
needs to know how many lines were emitted).

``` go
s := warp.WrapToString("a b c d e f g h i j k", 8, false)
// s == "a b c d\nf g h i\nj k"
```

## Implementation Notes

- `wrapLine` (used by `WordWrap`) walks the line character-by-character,
  keeping track of the last word boundary it has seen, so a long word is
  only broken if no natural break point exists before the visual width
  is exceeded.
- `wrapAtSpaces` (used by `SpaceWrap`) first tokenizes with
  `strings.Fields`, then greedily packs words into lines, adding a
  single space between them. The space separator counts toward the
  visual width.
- The `isWordBreak` helper exists as a byte-level convenience around
  `unicode.IsSpace`; it is currently unused in the package but kept for
  potential external use.
- Both wrappers rely on `lipgloss.Width` for visual width. That means a
  string like `"\x1b[31mred\x1b[0m"` is measured as 3 characters, not as
  the length of the escape sequence.
- The two strategies differ in one way only: `WordWrap` hard-breaks
  words that are longer than the width; `SpaceWrap` does not, and will
  emit a single line that overflows the width if no spaces are
  available.

## Dependencies

- `github.com/charmbracelet/lipgloss` — provides `lipgloss.Width`, a
  fork of go-runewidth, used as the source of truth for visual character
  width.

## Limitations

- Both functions only handle `"\n"` as the line separator; carriage
  returns and other Unicode line breaks in the input are not treated as
  line breaks and will be preserved in the output as-is (wrapped by the
  width logic).
- There is no concept of hyphenation or word-joining across a wrap
  boundary; a wrap always occurs at a whitespace boundary (or mid-word
  for `WordWrap`).
