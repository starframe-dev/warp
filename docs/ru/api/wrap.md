---
title: Перенос текста
description: Утилиты переноса текста с учётом ANSI и ширины терминала.
---

# Перенос текста

```go
func WordWrap(text string, width int) []string
func SpaceWrap(text string, width int) []string
func WrapToString(text string, width int, useSpaceWrap bool) string
```

`WordWrap` переносит по границам слов и разбивает слово, если оно не помещается. `SpaceWrap` переносит только по пробелам. Обе функции используют отображаемую ширину терминала, а не длину в байтах.

`WrapToString` выбирает `SpaceWrap`, если `useSpaceWrap == true`, и соединяет полученные строки переводами строк.
