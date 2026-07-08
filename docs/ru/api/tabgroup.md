---
title: TabGroup
description: Группа вкладок Warp и позиции отображения вкладок.
---

# TabGroup

`TabGroup` управляет набором вкладок и реализует `Panel`.

## Создание

```go
NewTabGroup(pos TabPosition) *TabGroup
```

Создаёт группу вкладок с заданной позицией.

## Вкладки

```go
NewTab(name string) *Tab
ActiveTab() *Tab
NextTab()
PrevTab()
```

`NewTab` создаёт вкладку. `ActiveTab` возвращает активную вкладку. `NextTab` и `PrevTab` переключают вкладки.

## Panel

`TabGroup` реализует интерфейс `Panel`.

## TabPosition

```go
TabTop
TabBottom
TabLeft
TabRight
TabNone
```

`TabPosition` задаёт расположение панели вкладок или отключает её отображение через `TabNone`.
