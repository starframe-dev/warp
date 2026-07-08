---
title: Tab
description: API вкладки Warp для компоновки панелей, фокуса, float-панелей и контекстного меню.
---

# Tab

`Tab` хранит дерево панелей вкладки и управляет раскладкой, фокусом и всплывающими элементами.

## Корневая панель

```go
RootPanel() Panel
SetRootPanel(panel Panel)
```

`RootPanel` возвращает корневую панель вкладки. `SetRootPanel` задаёт её.

## Split

```go
SplitVertical(parent Panel, fraction float64, panel Panel)
SplitHorizontal(parent Panel, fraction float64, panel Panel)
```

Делят родительскую панель по вертикали или горизонтали. `fraction` задаёт долю первой области.

## Flex

```go
FlexRow(parent Panel, items []FlexItemSpec)
FlexColumn(parent Panel, items []FlexItemSpec)
```

Создают flex-раскладку в строку или колонку.

## Float

```go
Float(panel Panel, x, y, w, h int)
CloseFloat(fp *FloatPane)
```

`Float` открывает плавающую панель. `CloseFloat` закрывает её.

## Фокус

```go
FocusNext()
FocusPrev()
FocusFirst()
FocusPanel(panel Panel)
```

Управляют фокусом между фокусируемыми панелями.

## Collapsible

```go
ToggleCollapsible(panel Panel)
```

Переключает свёрнутое состояние панели, если она поддерживает `Collapsible`.

## Контекстное меню

```go
ShowContextMenu(items []PopoverItem, x, y int)
```

Показывает контекстное меню в позиции `x`, `y`. Использует `Popover`.
