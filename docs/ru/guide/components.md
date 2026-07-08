---
title: Компоненты
description: Input, DropdownMenu, Selectable, Modal, Popover и утилиты переноса текста в Warp.
---

# Компоненты

Warp включает готовые компоненты для ввода, выбора, модальных окон, контекстных меню и работы с текстом.

## Input

```go
input := warp.NewInput(prompt)
```

`Input` — поле ввода с приглашением.

Поддерживается:

- курсор;
- `Backspace`;
- стрелки;
- `Home` / `End`;
- `Delete`.

## DropdownMenu

```go
menu := warp.NewDropdownMenu(label, items)
```

`DropdownMenu` показывает кнопку `▼`, список элементов и состояние наведения.

Поддерживается:

- открытие меню;
- ховер;
- выбор элемента.

## Selectable

```go
selectable := warp.NewSelectable(panel)
```

`Selectable` добавляет выделение текста для панели.

Поддерживается:

- выделение мышью;
- `Shift` + стрелки;
- `Ctrl+A`;
- `Esc`;
- копирование через OSC 52.

OSC 52 позволяет отправлять выделенный текст в буфер обмена терминала.

## Modal

```go
modal := warp.NewModal(title, content, buttons, onClose)
```

`Modal` отображается поверх основного интерфейса.

Поддерживается:

- наложение поверх контента;
- перетаскивание;
- закрытие;
- обработчик `onClose`.

## Popover

```go
popover := &warp.Popover{
	Items:   items,
	X:       x,
	Y:       y,
	OnClose: onClose,
}
```

`Popover` подходит для контекстных меню.

Поддерживается:

- показ списка действий;
- позиционирование по координатам;
- закрытие по клику вне меню.

## WordWrap и SpaceWrap

`WordWrap` и `SpaceWrap` — утилиты переноса текста.

Они помогают подготавливать текст для отображения в ограниченной ширине панели.
