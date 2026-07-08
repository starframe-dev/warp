---
title: Scrollable
description: Прокручиваемая панель Warp.
---

# Scrollable

`Scrollable` добавляет прокрутку к любой панели.

## Создание

```go
NewScrollable(panel Panel) *Scrollable
```

Создаёт прокручиваемую обёртку над `Panel`.

## Panel

`Scrollable` реализует интерфейс `Panel`.

## Управление

Поддерживаются:

- колесо мыши;
- `PgUp` / `PgDn`;
- стрелки вверх и вниз.
