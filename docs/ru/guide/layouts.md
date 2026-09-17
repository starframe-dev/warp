---
title: Компоновка
description: Split, Flex, Float, вложенные TabGroup, Collapsible и Scrollable в Warp.
---

# Компоновка

Начните с `Tab` и используйте его методы, чтобы заменять панели в дереве компоновки.

## Split

Split делит область на две части с перетаскиваемой границей:

```go
tab.SplitVertical(parent, 0.5, rightPanel)
tab.SplitHorizontal(parent, 0.5, bottomPanel)
```

`fraction` задаёт долю первого дочернего узла. Warp ограничивает её диапазоном `0.1..0.9`.

## Flex

Flex распределяет несколько дочерних панелей в строке или колонке:

```go
tab.FlexRow(parent, []warp.FlexItemSpec{
    {Panel: leftPanel, Grow: 1},
    {Panel: rightPanel, Grow: 2},
})
tab.FlexColumn(parent, items)
```

Значения `Grow` делят доступное пространство между элементами. Отрицательное значение превращается в ноль.

## Float

Панель можно разместить поверх основной компоновки:

```go
tab.Float(panel, 10, 4, 30, 8)
```

Float поддерживает перетаскивание, изменение размера за границы, кнопку закрытия и изменение z-order при клике.

## Вложенные компоновки

`TabGroup` уже реализует `Panel`, поэтому его можно вставить прямо в split или flex:

```go
tg := warp.NewTabGroup(warp.TabLeft)
tg.NewTab("inspector")
tab.FlexRow(tab.RootPanel(), []warp.FlexItemSpec{
    {Panel: tg, Grow: 1},
    {Panel: content, Grow: 2},
})
```

Для полного вложенного Warp используйте `inner.AsPanel()`.

## Collapsible

```go
collapsed := warp.NewCollapsible("Details", panel)
tab.SetRootPanel(collapsed)
tab.ToggleCollapsible(collapsed)
```

## Scrollable

```go
scroll := warp.NewScrollable(panel)
tab.SetRootPanel(scroll)
```

Поддерживаются колесо мыши, `PgUp`, `PgDn` и построчная навигация.
