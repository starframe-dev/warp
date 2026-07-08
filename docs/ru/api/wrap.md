---
title: Word Wrap
description: Функции переноса текста WordWrap и SpaceWrap.
---

# Word Wrap

Warp предоставляет функции переноса текста по визуальной ширине.

## WordWrap

```go
WordWrap(text string, width int) []string
```

Переносит текст по словам и возвращает строки, укладывающиеся в заданную ширину.

## SpaceWrap

```go
SpaceWrap(text string, width int) []string
```

Переносит текст с учётом пробельных символов.

## Визуальная ширина

`WordWrap` и `SpaceWrap` используют `lipgloss.Width` для подсчёта визуальных колонок, поэтому корректнее работают с ANSI-стилями и широкими символами.
