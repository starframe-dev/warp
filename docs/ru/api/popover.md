---
title: Popover
description: Всплывающее меню Popover и элементы PopoverItem.
---

# Popover

`Popover` отображает всплывающий список действий.

## Структура

```go
type Popover struct {
    Items   []PopoverItem
    X       int
    Y       int
    Width   int
    OnClose func()
}
```

- `Items` — элементы меню.
- `X`, `Y` — позиция.
- `Width` — ширина.
- `OnClose` — обработчик закрытия.

## PopoverItem

```go
type PopoverItem struct {
    Name   string
    Action func()
}
```

`Name` — подпись действия. `Action` вызывается при выборе.

## Overlay

```go
Overlay(lines []string, totalW, totalH int) []string
```

Накладывает popover на строки экрана.

## Обработка событий

```go
HandleMouse(msg tea.MouseMsg) bool
HandleKey(msg tea.KeyMsg) bool
```

Возвращают `true`, если событие мыши или клавиатуры обработано.
