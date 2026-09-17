---
title: Float
description: Состояние и поведение плавающей панели.
---

# Float

`FloatPane` описывает панель, которая рисуется поверх обычной компоновки вкладки. Для создания используется `Tab.Float`, для удаления — `Tab.CloseFloat`.

## FloatPane

```go
type FloatPane struct {
    Panel  Panel
    X, Y   int
    Width  int
    Height int
    Title  string
    CloseRequested      bool
    CloseOnOutsideClick bool
}
```

После нажатия на кнопку закрытия `CloseRequested` становится `true`; владелец `Tab` удаляет панель. `CloseOnOutsideClick` включает закрытие при клике за пределами float.

Публичного конструктора `NewFloatPane` и публичного метода наложения нет. Отрисовкой и маршрутизацией мыши управляет `Tab`.

## Создание float

```go
tab.Float(panel, 10, 4, 30, 8)
```

Float поддерживает перетаскивание, изменение размера за границы, заголовок и кнопку закрытия. Клик внутри float поднимает его наверх и не передаётся панелям под ним.

## ANSI-утилита

```go
func StripANSI(s string) string
```

Удаляет ANSI escape-последовательности. Warp учитывает визуальную ширину в терминальных ячейках при наложении.
