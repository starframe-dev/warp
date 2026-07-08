---
title: Collapsible
description: Сворачиваемая панель Warp.
---

# Collapsible

`Collapsible` оборачивает панель и позволяет сворачивать её содержимое.

## Создание

```go
NewCollapsible(title string, panel Panel) *Collapsible
```

Создаёт сворачиваемую панель с заголовком.

## Состояние

```go
Toggle()
Collapsed bool
```

`Toggle` переключает состояние. `Collapsed` показывает, свёрнута ли панель.

## Panel

`Collapsible` реализует интерфейс `Panel`.
