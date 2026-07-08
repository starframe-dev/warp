# Архитектура

> Полная спецификация: [`specs/arhitektura.md`](https://github.com/starframe-dev/warp/blob/main/specs/arhitektura.md)

## Обзор

Warp построен на системе element-based rendering. Каждый UI-компонент — самодостаточная единица с собственной логикой рендеринга и обработки событий.

## Модули

### Ядро
- `render.go` — система рендеринга, обработчики событий
- `element.go` — базовый тип элемента UI
- `styles.go` — система стилей (Gruvbox Dark)
- `wrap.go` — обёртки для элементов

### Компоненты
- `modal.go` — модальные окна
- `tab.go` / `tabgroup.go` — вкладки
- `dropdown.go` — выпадающие списки
- `input.go` — поля ввода
- `scrollable.go` — скроллинг
- `selectable.go` — выделение текста
- `popover.go` — контекстные меню
- `float.go` — плавающие панели
- `collapsible.go` — сворачиваемые панели
- `focus.go` — управление фокусом

## Дерево панелей

```
Node { Panel | Split | Flex }
SplitConfig { Direction, Fraction, First, Second }
FlexConfig  { Direction, Items[] }
```
