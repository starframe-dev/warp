---
title: Фокус
description: Явное управление фокусом и запрос необработанных клавиш.
---

# Фокус

Warp предоставляет операции фокуса, но не выбирает клавиши для их вызова. Приложение само решает, должны ли `Tab`, `Shift+Tab` или другие клавиши переключать фокус.

## Focusable

```go
type Focusable interface {
    Panel
    Focus()
    Blur()
    Focused() bool
}
```

## RawKeyReceiver

```go
type RawKeyReceiver interface {
    Panel
    WantsRawKeys() bool
}
```

`WantsRawKeys` показывает, что панели терминального типа, например PTY-панели, нужен необработанный ввод с клавиатуры.

## Методы Tab

```go
func (t *Tab) Focus() Panel
func (t *Tab) SetFocus(panel Panel) tea.Cmd
func (t *Tab) FocusNext()
func (t *Tab) FocusPrev()
func (t *Tab) FocusFirst()
func (t *Tab) FocusPanel(panel Panel)
```

Обход идёт в визуальном порядке листовых фокусируемых панелей. `FocusNext` и `FocusPrev` циклические.

## Пример

```go
case tea.KeyMsg:
    switch msg.String() {
    case "tab":
        tab.FocusNext()
    case "shift+tab":
        tab.FocusPrev()
    }
```
