# Архитектура проекта warp

## Обзор

Warp — Go TUI layout engine на базе Bubbletea. Управляет расположением панелей через дерево Node (split/flex/leaf). Каждый компонент — Panel с View/Update.

## Структура модулей

### Ядро
- `warp.go` — tea.Model, обёртка вокруг root Panel
- `tabgroup.go` — TabGroup: Panel с таб-баром
- `tab.go` — Tab: дерево splits/flex, float-панели, фокус, mouse
- `panel.go` — интерфейс Panel
- `split.go` — Node, SplitConfig, FlexConfig
- `render.go` — renderNode, findBorders, padContent
- `float.go` — FloatPane, overlayFloat, StripANSI
- `styles.go` — Gruvbox Dark стили
- `focus.go` — Focusable, RawKeyReceiver

### Компоненты
- `collapsible.go` — сворачиваемые панели
- `scrollable.go` — скроллинг
- `dropdown.go` — выпадающие списки
- `selectable.go` — выделение текста
- `input.go` — поля ввода
- `modal.go` — модальные окна
- `popover.go` — контекстные меню
- `element.go` — ElementProvider для тестирования
- `wrap.go` — WordWrap/SpaceWrap

## Дерево панелей

```go
Node { Panel Panel | Split *SplitConfig | Flex *FlexConfig }
SplitConfig { Direction, Fraction, First *Node, Second *Node, Dragging bool }
FlexConfig  { Direction, Items []FlexItem }
```

Leaf → `panel.View(w, h)`. Split → рекурсивный рендеринг детей с границей `│`/`─`.
Flex → распределение Grow-весов. Float → overlay поверх.

## Рендеринг

1. `renderNode(node, w, h)` — рекурсивно рендерит дерево
2. `overlayFloat(lines, fp, w, h)` — ANSI-aware наложение float поверх
3. `padContent(content, w, h)` — обрезка/дополнение до точных размеров

## Событийная модель

- `Tab.handleMouse` — проверка float'ов (z-order), потом border drag, потом клик по панели
- `Tab.handleKeys` — форвард в focused panel
- `TabGroup.handleKeyMsg` — Ctrl+Tab, Ctrl+W, Ctrl+T, потом делегат в Tab
- Warp **не перехватывает** клавиши — все сообщения идут в root Panel

## Фокус

- `Focusable` — Panel с Focus/Blur/Focused
- `Tab.FocusNext()/FocusPrev()` — явное переключение (разработчик сам биндит клавиши)
- `RawKeyReceiver` — для PTY/терминалов, получает все клавиши без перехвата

## Стили

Gruvbox Dark палитра. Все цвета в `styles.go`. Нет публичного API кастомизации.
