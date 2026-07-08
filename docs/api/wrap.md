---
title: Word Wrap
description: Text wrapping utilities
---

# Word Wrap

## Functions

```go
func WordWrap(text string, width int) []string
func SpaceWrap(text string, width int) []string
```

`WordWrap` wraps at word boundaries. `SpaceWrap` wraps at any space.

Both use `lipgloss.Width` for visual column counting (ANSI-aware).
