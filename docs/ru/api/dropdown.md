---
title: Dropdown
description: Выпадающее меню DropdownMenu и элементы DropdownItem.
---

# Dropdown

`DropdownMenu` отображает выпадающий список вариантов.

## Создание

```go
NewDropdownMenu(label string, items []DropdownItem) *DropdownMenu
```

Создаёт меню с подписью и списком элементов.

## DropdownItem

```go
type DropdownItem struct {
    Label    string
    Selected bool
}
```

`Label` — текст элемента. `Selected` показывает выбранный элемент.

## Состояние

```go
Open bool
Hovered int
OnSelect func(idx int)
```

- `Open` — открыто ли меню.
- `Hovered` — индекс элемента под курсором.
- `OnSelect` — обработчик выбора.

## Panel

`DropdownMenu` реализует интерфейс `Panel`.
