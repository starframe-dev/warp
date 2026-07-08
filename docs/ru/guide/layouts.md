---
title: Компоновка
description: Split, Flex, Float, вложенные TabGroup, Collapsible и Scrollable в Warp.
---

# Компоновка

Warp поддерживает несколько способов построения интерфейса: split, flex, float и вложенные панели.

## Split

Split делит родительскую область на две части.

```go
warp.SplitVertical(parent, fraction, panel)
warp.SplitHorizontal(parent, fraction, panel)
```

- `SplitVertical` делит область по вертикали;
- `SplitHorizontal` делит область по горизонтали;
- `fraction` задаёт долю пространства для одной из частей;
- границы split-компоновки можно перетаскивать мышью.

## Flex

Flex распределяет несколько элементов в строке или колонке.

```go
warp.FlexRow(parent, items)
warp.FlexColumn(parent, items)
```

Элементы используют веса `Grow`. Чем больше вес, тем больше пространства получает элемент относительно остальных.

## Float

Float создаёт плавающее окно поверх контента.

```go
warp.Float(panel, x, y, w, h)
```

Float-панели поддерживают:

- перетаскивание;
- изменение размера;
- закрытие;
- изменение Z-order при клике.

При клике окно поднимается выше остальных float-окон.

## Вложенность

`TabGroup` можно использовать как обычную `Panel` через `AsPanel()`.

Это позволяет вкладывать вкладки внутрь других компоновок и даже запускать Warp внутри Warp.

## Collapsible

Collapsible-панель можно свернуть и развернуть.

```go
panel := warp.NewCollapsible(title, panel)
warp.ToggleCollapsible(panel)
```

`NewCollapsible(title, panel)` создаёт сворачиваемую панель.

`ToggleCollapsible(panel)` переключает её состояние.

## Scrollable

Scrollable добавляет прокрутку для панели.

```go
panel := warp.NewScrollable(panel)
```

Поддерживаются:

- колесо мыши;
- `PgUp`;
- `PgDn`.
