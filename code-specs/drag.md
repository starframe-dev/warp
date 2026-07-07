# Specification: drag.go

## Обзор

Файл `drag.go` в пакете `warp` является плейсхолдером. Он содержит только документальный комментарий, объясняющий, что функционал drag-and-drop для границ разделения панелей реализован в других файлах.

## Текущее состояние

Файл содержит исключительно комментарии:

```go
package warp

// Drag-and-drop interactions for split borders are not implemented in this
// file. They are handled in tab.go and rendered in render.go:
//
//   - tab.go implements the input handling and state management for dragging a
//     border between split panes. The relevant functions are handleMouse,
//     updateDrag, and hitBorder.
//   - render.go computes the screen positions of borders via findBorders.
```

## Публичный API

Файл не экспортирует никаких публичных типов, функций или констант. Весь функционал перенаправлен на другие модули.

## Реализованный функционал

### Отсутствует в этом файле

- Обработка событий drag-and-drop
- Управление состоянием перетаскивания границ
- Вычисление позиций границ для отображения

### Реализация в других модулях

| Файл | Функционал |
|------|------------|
| `tab.go` | Обработка ввода, управление состоянием перетаскивания границ между раздельными панелями. Функции: `handleMouse`, `updateDrag`, `hitBorder`. |
| `render.go` | Вычисление позиций экранов границ через функцию `findBorders`. |

## Обработка ошибок

Файл не содержит обработки ошибок.

## Тестирование

Файл не содержит тестов.

## Версионирование

Файл не имеет публичного API, требующего версионирования.

## Связанные файлы

- `tab.go` — реализация обработки событий drag-and-drop
- `render.go` — вычисление позиций границ
