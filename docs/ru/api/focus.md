---
title: Focus
description: Интерфейсы фокуса, raw key handling и методы управления фокусом во вкладке.
---

# Focus

Warp поддерживает фокусируемые панели и прямую обработку клавиш.

## Focusable

```go
type Focusable interface {
    Focus()
    Blur()
    Focused() bool
}
```

`Focusable` описывает компонент, который может получать и терять фокус.

## RawKeyReceiver

```go
type RawKeyReceiver interface {
    HandleRawKey(msg tea.KeyMsg) tea.Cmd
}
```

`RawKeyReceiver` получает необработанные клавиатурные события.

## Фокус вкладки

```go
Tab.FocusNext()
Tab.FocusPrev()
Tab.FocusFirst()
Tab.FocusPanel(panel Panel)
```

Методы `Tab` переключают фокус между панелями, выбирают первую фокусируемую панель или фокусируют конкретную панель.
