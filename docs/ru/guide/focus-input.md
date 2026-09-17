---
title: Фокус и ввод
description: Focusable, переключение фокуса, RawKeyReceiver и обработка ввода в Warp.
---

# Фокус и ввод

Warp не навязывает правила фокуса и биндинги клавиш. Приложение само решает, какие клавиши переключают фокус.

## Focusable

Фокусируемая панель реализует:

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

## Методы фокуса Tab

```go
tab.FocusNext()
tab.FocusPrev()
tab.FocusFirst()
tab.FocusPanel(panel)
```

Обход идёт в визуальном порядке фокусируемых листьев и циклически переходит через границы.

## Биндинги клавиш

Warp не назначает `Tab` и `Shift+Tab` автоматически. Приложение может связать эти клавиши с нужными действиями.

## RawKeyReceiver

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

`RawKeyReceiver` подходит для PTY и терминальных панелей, которым нужно объявить предпочтение необработанного ввода. Интерфейс не заменяет `Panel.Update`, а добавляет признак `WantsRawKeys`.

## Пример

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "tab":
            m.Tab.FocusNext()
            return m, nil
        case "shift+tab":
            m.Tab.FocusPrev()
            return m, nil
        }
    }

    _, cmd := m.Warp.Update(msg)
    return m, cmd
}
```
