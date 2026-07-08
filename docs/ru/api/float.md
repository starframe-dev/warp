---
title: Float
description: Плавающие панели FloatPane и вспомогательная функция StripANSI.
---

# Float

`FloatPane` отображает панель поверх текущей раскладки.

## Создание

```go
NewFloatPane(panel Panel, x, y, w, h int) *FloatPane
```

Создаёт плавающую панель в позиции `x`, `y` с размером `w` × `h`.

## Закрытие по внешнему клику

```go
CloseOnOutsideClick bool
```

Если включено, панель закрывается при клике вне её границ.

## Overlay

```go
Overlay(lines []string, totalW, totalH int) []string
```

Накладывает float-панель на готовые строки экрана.

## Mouse

```go
HandleMouse(msg tea.MouseMsg) bool
```

Обрабатывает мышь. Возвращает `true`, если событие обработано.

## StripANSI

```go
StripANSI(s string) string
```

Удаляет ANSI escape-последовательности из строки.
