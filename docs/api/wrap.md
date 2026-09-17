---
title: Word Wrap
description: ANSI-aware text wrapping helpers.
---

# Word Wrap

```go
func WordWrap(text string, width int) []string
func SpaceWrap(text string, width int) []string
func WrapToString(text string, width int, useSpaceWrap bool) string
```

`WordWrap` wraps at word boundaries and breaks a word when it cannot fit. `SpaceWrap` only wraps at spaces. Both use terminal display width rather than byte length.

`WrapToString` selects `SpaceWrap` when `useSpaceWrap` is true and joins the resulting lines with newlines.
