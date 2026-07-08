---
title: Modal
description: Модальные окна Warp, кнопки, сообщения открытия и закрытия.
---

# Modal

`Modal` отображает модальное окно поверх текущего экрана.

## Создание

```go
NewModal(title, content string, buttons []ModalButton, onClose func()) *Modal
```

Создаёт модальное окно с заголовком, содержимым, кнопками и обработчиком закрытия.

## ModalButton

```go
type ModalButton struct {
    Label  string
    Action func()
}
```

`Label` — текст кнопки. `Action` вызывается при нажатии.

## Сообщения

```go
ShowModalMsg
CloseModalMsg
```

`ShowModalMsg` показывает модальное окно. `CloseModalMsg` закрывает его.

## Overlay

```go
Overlay(lines []string, totalW, totalH int) []string
```

Накладывает модальное окно на строки экрана.

## Mouse

```go
HandleMouse(msg tea.MouseMsg) bool
```

Обрабатывает мышь. Возвращает `true`, если событие обработано.

## Перетаскивание

Модальное окно можно перетаскивать мышью.
